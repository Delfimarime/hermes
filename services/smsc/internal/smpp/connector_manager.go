package smpp

import (
	"fmt"
	"github.com/delfimarime/hermes/services/smsc/internal/model"
	"github.com/delfimarime/hermes/services/smsc/internal/repository/sdk"
	"github.com/delfimarime/hermes/services/smsc/pkg/config"
	"github.com/fiorix/go-smpp/smpp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sync"
)

type SimpleConnectorManager struct {
	mutex              sync.Mutex
	repository         sdk.SmppRepository
	connectors         []Connector
	pduListenerFactory *PduListenerFactory
	configuration      config.Configuration
	connectorsCache    map[string]*SimpleConnector
}

func (instance *SimpleConnectorManager) GetList() []Connector {
	return instance.connectors
}

func (instance *SimpleConnectorManager) GetById(id string) Connector {
	c, _ := instance.connectorsCache[id]
	return c
}

func (instance *SimpleConnectorManager) AfterPropertiesSet() error {
	instance.mutex.Lock()
	defer instance.mutex.Unlock()
	zap.L().Debug("Retrieving smsc configuration from SmppRepository")
	seq, err := instance.repository.FindAll()
	if err != nil {
		return err
	}
	zap.L().Info(fmt.Sprintf("%d smsc's retrieved from SmppRepository", len(seq)))
	instance.setConnectors(seq...)
	go func() {
		instance.refresh()
	}()
	return nil
}

func (instance *SimpleConnectorManager) Close() error {
	instance.mutex.Lock()
	defer instance.mutex.Unlock()
	if instance.connectorsCache == nil {
		return nil
	}
	for _, v := range instance.connectorsCache {
		each := v
		if err := each.doClose(); err != nil {
			if err != nil {
				zap.L().Error(fmt.Sprintf("Unable to close smpp.connector[id=%s]", v.GetId()),
					zap.String(SmscIdAttribute, v.GetId()), zap.String(SmscAliasAttribute, v.GetAlias()),
					zap.Error(err),
				)
			}
		}
	}
	instance.connectors = nil
	instance.connectorsCache = nil
	return nil
}

func (instance *SimpleConnectorManager) setConnectors(seq ...model.Smpp) {
	for _, definition := range seq {
		zap.L().Info(fmt.Sprintf("Creating smpp.Client for smsc[id=%s]", definition.Id),
			zap.String(SmscIdAttribute, definition.Id), zap.String(SmscAliasAttribute, definition.Alias),
			zap.String(SmscNameAttribute, definition.Name),
		)
		client, prob := instance.newClient(definition)
		if prob != nil {
			zap.L().Error(fmt.Sprintf("Unable to create smpp.Client for smsc[id=%s]", definition.Id),
				zap.String(SmscIdAttribute, definition.Id), zap.String(SmscNameAttribute, definition.Name),
				zap.String(SmscAliasAttribute, definition.Alias), zap.Error(prob),
			)
		}
		zap.L().Debug(fmt.Sprintf("Creating smpp.connector for smsc[id=%s]", definition.Id),
			zap.String(SmscIdAttribute, definition.Id), zap.String(SmscNameAttribute, definition.Name),
			zap.String(SmscAliasAttribute, definition.Alias), zap.String(SmscTypeAttribute, definition.Type),
			zap.Error(prob),
		)
		connector := instance.newConnectorFrom(client, definition)
		zap.L().Debug(fmt.Sprintf("Registering smpp.connector for smsc[id=%s]", definition.Id),
			zap.String(SmscIdAttribute, definition.Id), zap.String(SmscNameAttribute, definition.Name),
			zap.String(SmscAliasAttribute, definition.Alias), zap.String(SmscTypeAttribute, definition.Type),
			zap.Error(prob),
		)
		instance.connectorsCache[definition.Id] = connector
		instance.connectors = append(instance.connectors, connector)
	}
}

func (instance *SimpleConnectorManager) newClient(definition model.Smpp) (Client, error) {
	zap.L().Debug(fmt.Sprintf("Creating smpp.Client for smsc[id=%s]", definition.Id),
		zap.String(SmscIdAttribute, definition.Id), zap.String(SmscNameAttribute, definition.Name),
		zap.String(SmscAliasAttribute, definition.Alias),
	)
	var connector Client
	var f smpp.HandlerFunc
	if definition.Type == model.ReceiverType || definition.Type == model.TransceiverType {
		zap.L().Debug(fmt.Sprintf("Creating PduListener for smsc[id=%s,type=%s]", definition.Id, definition.Type),
			zap.String(SmscIdAttribute, definition.Id), zap.String(SmscNameAttribute, definition.Name),
			zap.String(SmscTypeAttribute, definition.Type), zap.String(SmscAliasAttribute, definition.Alias),
		)
		f = instance.pduListenerFactory.New(definition)
	}
	switch definition.Type {
	case model.ReceiverType:
		connector = newReceiverClient(definition, instance.onClientConnEvent(definition.Id), f)
		break
	case model.TransceiverType, model.TransmitterType:
		connector = newTransmitterClient(definition, instance.onClientConnEvent(definition.Id), f)
		break
	default:
		return nil, fmt.Errorf("type=%s isn't supported", definition.Type)
	}
	return connector, nil
}

func (instance *SimpleConnectorManager) newConnectorFrom(cl Client, def model.Smpp) *SimpleConnector {
	state := StartupConnectorLifecycleState
	namesOfMetrics := []string{
		"number_of_messages_sent_sms",
		"number_of_messages_send_sms_attempts",
		"number_of_refreshes",
		"number_of_refresh_attempts",
		"number_of_bindings",
		"number_of_bindings_attempts",
	}
	zap.L().Debug(fmt.Sprintf("Creating smpp.connector metrics for smsc[id=%s]", def.Id),
		zap.String(SmscIdAttribute, def.Id), zap.String(SmscNameAttribute, def.Name),
		zap.String(SmscAliasAttribute, def.Alias), zap.Strings("names_of_metrics", namesOfMetrics))
	metrics := make([]metric.Float64Counter, len(namesOfMetrics))
	arr, err := instance.metricsFrom(otel.Meter("hermes_smsc"), def, namesOfMetrics)
	if err != nil {
		zap.L().Error(fmt.Sprintf("Cannot create smpp.connector metrics for smsc[id=%s]", def.Id),
			zap.String(SmscIdAttribute, def.Id),
			zap.String(SmscAliasAttribute, def.Alias),
			zap.String(SmscNameAttribute, def.Name),
			zap.Error(err),
		)
		state = ErrorConnectorLifecycleState
	}
	metrics = arr
	return &SimpleConnector{
		client:                      cl,
		state:                       state,
		alias:                       def.Alias,
		sendMessageCountMetric:      metrics[0],
		sendMessageErrorCountMetric: metrics[1],
		refreshCountMetric:          metrics[2],
		refreshErrorCountMetric:     metrics[3],
		bindCountMetric:             metrics[4],
		bindErrorCountMetric:        metrics[5],
	}
}

func (instance *SimpleConnectorManager) metricsFrom(meter metric.Meter, definition model.Smpp, seq []string) ([]metric.Float64Counter, error) {
	producer := func(metricName string) (metric.Float64Counter, error) {
		zap.L().Debug(fmt.Sprintf("Creating smpp.connector metric for smsc[id=%s]", definition.Id),
			zap.String(SmscIdAttribute, definition.Id), zap.String(SmscNameAttribute, definition.Name),
			zap.String(SmscAliasAttribute, definition.Alias), zap.String("metric_name", metricName),
			zap.String("metric_type", "Float64Counter"))
		return meter.Float64Counter(metricName)
	}
	r := make([]metric.Float64Counter, 0)
	for _, name := range seq {
		metricName := fmt.Sprintf("hermes_smsc_%s_%s", definition.Alias, name)
		m, err := producer(metricName)
		if err != nil {
			zap.L().Debug(fmt.Sprintf("Couldn't create smpp.connector metric for smsc[id=%s]", definition.Id),
				zap.String(SmscIdAttribute, definition.Id), zap.String(SmscNameAttribute, definition.Name),
				zap.String(SmscAliasAttribute, definition.Alias), zap.String("metric_name", metricName),
				zap.String("metric_type", "Float64Counter"),
				zap.Error(err),
			)
			return nil, err
		}
		r = append(r, m)
	}
	return r, nil
}

func (instance *SimpleConnectorManager) refresh() {
	instance.mutex.Lock()
	defer instance.mutex.Unlock()
	zap.L().Info("Starting managed smpp.connector(s)")
	var wg sync.WaitGroup
	for _, connector := range instance.connectors {
		if connector.GetState() != StartupConnectorLifecycleState {
			zap.L().Warn(fmt.Sprintf("Cannot initialize smpp.connector[id=%s] due to it's state", connector.GetId()),
				zap.String(SmscIdAttribute, connector.GetId()), zap.String(SmscAliasAttribute, connector.GetAlias()),
				zap.String(smscStateAttribute, connector.GetState().string()),
			)
			continue
		}
		wg.Add(1)
		each := connector
		go func() {
			zap.L().Debug(fmt.Sprintf("Starting managed smpp.connector[id=%s]", each.GetId()),
				zap.String(SmscIdAttribute, each.GetId()), zap.String(SmscAliasAttribute, each.GetAlias()),
				zap.String(smscStateAttribute, each.GetState().string()),
			)
			instance.bindOrRefresh(each.(*SimpleConnector), nil)
			defer wg.Done()
			zap.L().Debug(fmt.Sprintf("Successfully started smpp.connector[id=%s]", each.GetId()),
				zap.String(SmscIdAttribute, each.GetId()), zap.String(SmscAliasAttribute, each.GetAlias()),
			)
		}()
	}
	wg.Wait()
}

func (instance *SimpleConnectorManager) bindOrRefresh(c *SimpleConnector, e *ClientConnEvent) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	desc := "Starting"
	if e != nil {
		desc = "Refreshing"
	}
	zap.L().Debug(fmt.Sprintf("%s managed smpp.connector[id=%s] binding", desc, c.GetId()),
		zap.String(SmscIdAttribute, c.GetId()), zap.String(SmscAliasAttribute, c.GetAlias()),
		zap.String(smscStateAttribute, c.GetState().string()),
	)
	c.state = WaitConnectorLifecycleState
	zap.L().Debug(fmt.Sprintf("Changing managed smpp.connector[id=%s] state to %s",
		c.GetId(), c.GetState().string()), zap.String(SmscIdAttribute, c.GetId()),
		zap.String(SmscAliasAttribute, c.GetAlias()), zap.String(smscStateAttribute, c.GetState().string()),
	)
	if e != nil {
		absenceLogFactory := func(l zapcore.Level, msg string, err error) {
			fieldOpts := []zap.Field{
				zap.String(SmscIdAttribute, c.GetId()),
			}
			if e != nil {
				fieldOpts = append(fieldOpts, zap.String("event_type", string(e.Type)))
				if e.Err != nil {
					fieldOpts = append(fieldOpts, zap.String("event_error", e.Err.Error()))
				}
			}
			if err != nil {
				fieldOpts = append(fieldOpts, zap.Error(err))
			}
			zap.L().Log(l, msg, fieldOpts...)
		}
		zap.L().Debug(fmt.Sprintf("Closing smpp.connector[id=%s] smpp.Client", c.GetId()),
			zap.String(SmscIdAttribute, c.GetId()), zap.String(SmscAliasAttribute, c.GetAlias()),
			zap.String(smscStateAttribute, c.GetState().string()),
		)
		err := c.client.Close()
		if err != nil {
			zap.L().Warn(fmt.Sprintf("Closing smpp.connector[id=%s] smpp.Client", c.GetId()),
				zap.String(SmscIdAttribute, c.GetId()), zap.String(SmscAliasAttribute, c.GetAlias()),
				zap.String(smscStateAttribute, c.GetState().string()),
				zap.Error(err),
			)
		}
		zap.L().Debug(fmt.Sprintf("Retriving smpp.connector[id=%s] configuration from SmppRepository", c.GetId()),
			zap.String(SmscIdAttribute, c.GetId()), zap.String(SmscAliasAttribute, c.GetAlias()),
			zap.String(smscStateAttribute, c.GetState().string()),
		)
		definition, err := instance.repository.FindById(c.GetId())
		if err == nil {
			client, prob := instance.newClient(definition)
			if prob == nil {
				c.client = client
			}
			err = prob
		}
		if err != nil {
			absenceLogFactory(zap.ErrorLevel, "Cannot trigger refresh connector", err)
		}
		c.increaseMetricCounter(c.refreshCountMetric, c.refreshErrorCountMetric, err)
		return
	}
	zap.L().Debug(fmt.Sprintf("Binding managed smpp.connector[id=%s]", c.GetId()),
		zap.String(SmscIdAttribute, c.GetId()), zap.String(SmscAliasAttribute, c.GetAlias()),
		zap.String(smscStateAttribute, c.GetState().string()),
	)
	c.client.Bind()
}

func (instance *SimpleConnectorManager) onClientConnEvent(id string) ClientConnEventListener {
	return func(event ClientConnEvent) {
		zap.L().Debug(fmt.Sprintf("Handling smpp.Client[id=%s] smpp.ClientConnEvent", id),
			zap.String(SmscIdAttribute, id),
			zap.String(smppClientEventTypeAttribute, string(event.Type)),
		)
		absenceLogFactory := func(l zapcore.Level, msg string, err error) {
			fieldOpts := []zap.Field{
				zap.String(SmscIdAttribute, id),
				zap.String(smppClientEventTypeAttribute, string(event.Type)),
			}
			if err != nil {
				zap.Error(err)
			}
			if event.Err != nil {
				fieldOpts = append(fieldOpts, zap.String(smppClientEventProblemAttribute, event.Err.Error()))
			}
			zap.L().Log(l, msg, fieldOpts...)
		}
		c, hasValue := instance.connectorsCache[id]
		if !hasValue {
			absenceLogFactory(zap.WarnLevel,
				fmt.Sprintf("Cannot consume smpp.ClientConnEvent since  smpp.Client[id=%s] isn't present", id),
				nil)
			return
		}
		switch event.Type {
		case ClientConnBindErrorEventType, ClientConnErrorEventType:
			c.setState(ErrorConnectorLifecycleState)
			c.increaseMetricCounter(c.bindCountMetric, c.bindErrorCountMetric, event.Err)
			break
		case ClientConnBoundEventType:
			c.setState(ReadyConnectorLifecycleState)
			c.increaseMetricCounter(c.bindCountMetric, c.bindErrorCountMetric, nil)
			break
		case ClientConnDisconnectEventType:
			c.setState(ClosedConnectorLifecycleState)
			break
		case ClientConnInterruptedEventType:
			instance.bindOrRefresh(c, &event)
			break
		}
	}
}
