package smpp

import (
	"fmt"
	"github.com/delfimarime/hermes/services/smsc/internal/model"
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
	repository         Repository
	connectors         []Connector
	pduListenerFactory *PduListenerFactory
	configuration      config.Configuration
	connectorCache     map[string]*SimpleConnector
}

func (instance *SimpleConnectorManager) GetList() []Connector {
	return instance.connectors
}

func (instance *SimpleConnectorManager) GetById(id string) Connector {
	c, _ := instance.connectorCache[id]
	return c
}

func (instance *SimpleConnectorManager) AfterPropertiesSet() error {
	instance.mutex.Lock()
	defer instance.mutex.Unlock()
	seq, err := instance.repository.FindAll()
	if err != nil {
		return err
	}
	instance.setConnectors(seq...)
	go func() {
		instance.refresh()
	}()
	return nil
}

func (instance *SimpleConnectorManager) Close() error {
	instance.mutex.Lock()
	defer instance.mutex.Unlock()
	if instance.connectorCache == nil {
		return nil
	}
	for _, v := range instance.connectorCache {
		each := v
		if err := each.doClose(); err != nil {
			if err != nil {
				zap.L().Error("Cannot close smsc[id="+v.GetId()+"]",
					zap.String(smscIdAttribute, v.GetId()),
					zap.Error(err),
				)
			}
		}
	}
	instance.connectorCache = nil
	instance.connectors = nil
	return nil
}

func (instance *SimpleConnectorManager) setConnectors(seq ...model.Smpp) {
	for _, definition := range seq {
		client, prob := instance.newClient(definition)
		if prob != nil {
			zap.L().Warn("Cannot initialize smsc[id="+definition.Id+"]",
				zap.String(smscIdAttribute, definition.Id),
				zap.String(smscNameAttribute, definition.Name),
				zap.Error(prob),
			)
		}
		connector := instance.newConnectorFrom(client, definition)
		instance.connectorCache[definition.Id] = connector
		instance.connectors = append(instance.connectors, connector)
	}
}

func (instance *SimpleConnectorManager) newClient(definition model.Smpp) (Client, error) {
	var f smpp.HandlerFunc
	var connector Client
	if definition.Type == model.ReceiverType || definition.Type == model.TransceiverType {
		f = instance.pduListenerFactory.New(definition)
	}
	switch definition.Type {
	case model.ReceiverType:
		connector = newListenerConnector(definition, instance.onClientConnEvent(definition.Id), f)
		break
	case model.TransceiverType, model.TransmitterType:
		connector = newTransmitterConnector(definition, instance.onClientConnEvent(definition.Id), f)
		break
	default:
		return nil, fmt.Errorf("type=%s isn't supported", definition.Type)
	}
	return connector, nil
}

func (instance *SimpleConnectorManager) newConnectorFrom(connector Client, def model.Smpp) *SimpleConnector {
	status := WaitConnectorLifecycleState
	namesOfMetrics := []string{
		"number_of_messages_sent_sms",
		"number_of_messages_send_sms_attempts",
		"number_of_refreshes",
		"number_of_refresh_attempts",
		"number_of_bindings",
		"number_of_bindings_attempts",
	}
	metrics := make([]metric.Float64Counter, len(namesOfMetrics))
	arr, err := instance.metricsFrom(otel.Meter("hermes_smsc"), def, namesOfMetrics)
	if err != nil {
		zap.L().Error("Cannot create metrics for smsc[id="+connector.GetId()+"]",
			zap.String(smscIdAttribute, def.Id),
			zap.String(smscAliasAttribute, def.Alias),
			zap.String(smscNameAttribute, def.Name),
			zap.Error(err),
		)
		status = ErrorConnectorLifecycleState
	}
	metrics = arr
	return &SimpleConnector{
		state:                       status,
		alias:                       def.Alias,
		client:                      connector,
		sendMessageCountMetric:      metrics[0],
		sendMessageErrorCountMetric: metrics[1],
		refreshCountMetric:          metrics[2],
		refreshErrorCountMetric:     metrics[3],
		bindCountMetric:             metrics[4],
		bindErrorCountMetric:        metrics[5],
	}
}

func (instance *SimpleConnectorManager) metricsFrom(meter metric.Meter, definition model.Smpp, seq []string) ([]metric.Float64Counter, error) {
	producer := func(name string) (metric.Float64Counter, error) {
		return meter.Float64Counter(fmt.Sprintf("hermes_smsc_%s_%s", definition.Alias, name))
	}
	r := make([]metric.Float64Counter, 0)
	for _, name := range seq {
		m, err := producer(name)
		if err != nil {
			return nil, err
		}
		r = append(r, m)
	}
	return r, nil
}

func (instance *SimpleConnectorManager) refresh() {
	instance.mutex.Lock()
	defer instance.mutex.Unlock()
	zap.L().Info("Starting client(s)")
	var wg sync.WaitGroup
	for _, connector := range instance.connectors {
		if connector.GetState() != WaitConnectorLifecycleState {
			continue
		}
		zap.L().Info(fmt.Sprintf("Starting smsc[id=%s] client", connector.GetId()),
			zap.String(smscIdAttribute, connector.GetId()),
			zap.String(smscAliasAttribute, connector.GetAlias()),
			zap.String(smscStateAttribute, connector.GetState()),
		)
		wg.Add(1)
		each := connector
		go func() {
			instance.bindOrRefresh(each.(*SimpleConnector), nil)
			wg.Done()
		}()
	}
	wg.Wait()
}

func (instance *SimpleConnectorManager) bindOrRefresh(c *SimpleConnector, e *ClientConnEvent) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.state = WaitConnectorLifecycleState
	if e != nil {
		absenceLogFactory := func(l zapcore.Level, msg string, err error) {
			fieldOpts := []zap.Field{
				zap.String(smscIdAttribute, c.GetId()),
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
		_ = c.client.Close()
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
	c.client.Bind()
}

func (instance *SimpleConnectorManager) onClientConnEvent(id string) ClientConnEventListener {
	return func(event ClientConnEvent) {
		c, hasValue := instance.connectorCache[id]
		absenceLogFactory := func(l zapcore.Level, msg string, err error) {
			fieldOpts := []zap.Field{
				zap.String(smscIdAttribute, id),
				zap.String("event_type", string(event.Type)),
			}
			if err != nil {
				zap.Error(err)
			}
			if event.Err != nil {
				fieldOpts = append(fieldOpts, zap.String("event_error", event.Err.Error()))
			}
			zap.L().Log(l, msg, fieldOpts...)
		}
		if !hasValue {
			absenceLogFactory(zap.WarnLevel, "Cannot consume client event since client isn't present", nil)
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
