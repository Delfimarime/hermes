package smpp

import (
	"context"
	"fmt"
	"github.com/delfimarime/hermes/services/smsc/internal/model"
	"github.com/delfimarime/hermes/services/smsc/pkg/config"
	"github.com/fiorix/go-smpp/smpp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
	"sync"
)

type SimpleConnectorManager struct {
	mutex              sync.Mutex
	repository         Repository
	connectorFactory   ConnectorFactory
	pduListenerFactory *PduListenerFactory
	configuration      config.Configuration
	connectors         []Connector
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
	//	ctx, cancel := context.WithTimeout(context.Background(),
	//		common.MillisToDuration(instance.configuration.Smsc.StartupTimeout))
	//	defer cancel()
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

func (instance *SimpleConnectorManager) newConnectorFrom(connector Client, def model.Smpp) *SimpleConnector {
	status := StartupConnectorLifecycleState
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
		connector:                   connector,
		sendMessageCountMetric:      metrics[0],
		sendMessageErrorCountMetric: metrics[1],
		refreshCountMetric:          metrics[2],
		refreshErrorCountMetric:     metrics[3],
		bindCountMetric:             metrics[4],
		bindErrorCountMetric:        metrics[5],
	}
}

func (instance *SimpleConnectorManager) doBindConnector(ctx context.Context, wg *sync.WaitGroup,
	connector Client, def model.Smpp) {
	defer wg.Done()
	status := StartupConnectorLifecycleState
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
	if status == StartupConnectorLifecycleState {
		select {
		case <-ctx.Done():
			zap.L().Warn("Cannot bind smsc[id="+def.Id+"]",
				zap.String(smscIdAttribute, def.Id),
				zap.String(smscAliasAttribute, def.Alias),
				zap.String(smscNameAttribute, def.Name),
				zap.Int64("duration", instance.configuration.Smsc.StartupTimeout),
			)
			status = ErrorConnectorLifecycleState
		case prob := <-instance.bindConnector(connector):
			if prob != nil {
				zap.L().Error("Cannot bind smsc[id="+connector.GetId()+"]",
					zap.String(smscIdAttribute, def.Id),
					zap.String(smscAliasAttribute, def.Alias),
					zap.String(smscNameAttribute, def.Name),
					zap.Error(prob),
				)
			} else {
				zap.L().Info("Bind smsc[id="+connector.GetId()+"]",
					zap.String(smscIdAttribute, def.Id),
					zap.String(smscAliasAttribute, def.Alias),
					zap.String(smscNameAttribute, def.Name),
				)
			}
		}
	}
	m := SimpleConnector{
		state:                       status,
		alias:                       def.Alias,
		connector:                   connector,
		sendMessageCountMetric:      metrics[0],
		sendMessageErrorCountMetric: metrics[1],
		refreshCountMetric:          metrics[2],
		refreshErrorCountMetric:     metrics[3],
		bindCountMetric:             metrics[4],
		bindErrorCountMetric:        metrics[5],
	}
	instance.connectorCache[def.Id] = &m
	instance.connectors = append(instance.connectors, &m)
	m.state = ReadyConnectorLifecycleState
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

func (instance *SimpleConnectorManager) newClient(definition model.Smpp) (Client, error) {
	var f smpp.HandlerFunc
	var connector Client
	if definition.Type == model.ReceiverType || definition.Type == model.TransceiverType {
		f = instance.pduListenerFactory.New(definition)
	}
	switch definition.Type {
	case model.ReceiverType:
		connector = instance.connectorFactory.NewListenerConnector(definition, f)
		break
	case model.TransceiverType:
		connector = instance.connectorFactory.NewTransmitterConnector(definition, f)
		break
	case model.TransmitterType:
		connector = instance.connectorFactory.NewTransmitterConnector(definition, f)
		break
	default:
		return nil, fmt.Errorf("type=%s isn't supported", definition.Type)
	}
	return connector, nil
}

func (instance *SimpleConnectorManager) bindConnector(connector Client) <-chan error {
	ch := make(chan error, 1)
	go func() {
		err := connector.Bind()
		ch <- err
		close(ch)
	}()
	return ch
}

func (instance *SimpleConnectorManager) refresh() {
	var wg sync.WaitGroup
	instance.mutex.Lock()
	zap.L().Info("Starting connector(s)")
	for _, connector := range instance.connectors {
		if connector.GetState() != StartupConnectorLifecycleState {
			continue
		}
		zap.L().Info(fmt.Sprintf("Starting smsc[id=%s] connector", connector.GetId()),
			zap.String(smscIdAttribute, connector.GetId()),
			zap.String(smscAliasAttribute, connector.GetAlias()),
			zap.String(smscStateAttribute, connector.GetState()),
		)
		wg.Add(1)
		each := connector
		go func() {
			target := each.(*SimpleConnector)
			if err := target.doBind(); err != nil {
				target.state = ErrorConnectorLifecycleState
			} else {
				target.state = ReadyConnectorLifecycleState
			}
			wg.Done()
		}()
	}
	wg.Wait()
	instance.mutex.Unlock()
}
