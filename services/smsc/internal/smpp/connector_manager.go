package smpp

import (
	"context"
	"fmt"
	"github.com/delfimarime/hermes/services/smsc/internal/model"
	"github.com/delfimarime/hermes/services/smsc/pkg/common"
	"github.com/delfimarime/hermes/services/smsc/pkg/config"
	"github.com/fiorix/go-smpp/smpp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
	"sync"
)

type SimpleConnectorManager struct {
	mutex              sync.Mutex
	repository         SmppRepository
	connectorFactory   ConnectorFactory
	pduListenerFactory *PduListenerFactory
	configuration      config.Configuration
	cache              map[string]ManagedConnector
}

func (instance *SimpleConnectorManager) AfterPropertiesSet() error {
	instance.mutex.Lock()
	defer instance.mutex.Unlock()
	seq, err := instance.repository.FindAll()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(),
		common.MillisToDuration(instance.configuration.Smsc.StartupTimeout))
	defer cancel()
	instance.setConnectors(ctx, seq...)
	return nil
}

func (instance *SimpleConnectorManager) Close() error {
	instance.mutex.Lock()
	defer instance.mutex.Unlock()
	if instance.cache == nil {
		return nil
	}
	for _, v := range instance.cache {
		each := v
		if err := each.Close(); err != nil {
			if err != nil {
				zap.L().Error("Cannot close smsc[id="+v.GetId()+"]",
					zap.String("smsc_id", v.GetId()),
					zap.Error(err),
				)
			}
		}
	}
	instance.cache = nil
	return nil
}

func (instance *SimpleConnectorManager) StateOf(id string) string {
	if fromCache, hasValue := instance.cache[id]; hasValue {
		return fromCache.state
	}
	return ""
}

func (instance *SimpleConnectorManager) Refresh(_ string) error {
	//TODO implement me
	panic("implement me")
}

func (instance *SimpleConnectorManager) setConnectors(ctx context.Context, seq ...model.Smpp) {
	var wg sync.WaitGroup
	for _, definition := range seq {
		connector, prob := instance.newConnector(definition)
		if prob != nil {
			zap.L().Warn("Cannot initialize smsc[id="+definition.Id+"]",
				zap.String("smsc_id", definition.Id),
				zap.String("smsc_name", definition.Name),
				zap.Error(prob),
			)
		}
		wg.Add(1)
		go instance.doBindConnector(ctx, &wg, connector, definition)
	}
	wg.Wait()
}

func (instance *SimpleConnectorManager) doBindConnector(ctx context.Context, wg *sync.WaitGroup,
	connector Connector, def model.Smpp) {
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
			zap.String("smsc_id", def.Id),
			zap.String("smsc_alias", def.Alias),
			zap.String("smsc_name", def.Name),
			zap.Error(err),
		)
		status = ErrorConnectorLifecycleState
	}
	metrics = arr
	if status == StartupConnectorLifecycleState {
		select {
		case <-ctx.Done():
			zap.L().Warn("Cannot bind smsc[id="+def.Id+"]",
				zap.String("smsc_id", def.Id),
				zap.String("smsc_alias", def.Alias),
				zap.String("smsc_name", def.Name),
				zap.Int64("duration", instance.configuration.Smsc.StartupTimeout),
			)
			status = ErrorConnectorLifecycleState
		case prob := <-instance.bindConnector(connector):
			if prob != nil {
				zap.L().Error("Cannot bind smsc[id="+connector.GetId()+"]",
					zap.String("smsc_id", def.Id),
					zap.String("smsc_alias", def.Alias),
					zap.String("smsc_name", def.Name),
					zap.Error(prob),
				)
			} else {
				zap.L().Info("Bind smsc[id="+connector.GetId()+"]",
					zap.String("smsc_id", def.Id),
					zap.String("smsc_alias", def.Alias),
					zap.String("smsc_name", def.Name),
				)
			}
		}
	}
	m := ManagedConnector{
		state:                       status,
		connector:                   connector,
		SendMessageCountMetric:      metrics[0],
		SendMessageErrorCountMetric: metrics[1],
		RefreshCountMetric:          metrics[2],
		RefreshErrorCountMetric:     metrics[3],
		BindCountMetric:             metrics[4],
		BindErrorCountMetric:        metrics[5],
		canSendSms:                  def.Type == model.TransmitterType || def.Type == model.TransceiverType,
	}
	instance.cache[def.Id] = m
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

func (instance *SimpleConnectorManager) newConnector(definition model.Smpp) (Connector, error) {
	var f smpp.HandlerFunc
	var connector Connector
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
	// TODO: DECORATE CONNECTOR
	return connector, nil
}

func (instance *SimpleConnectorManager) bindConnector(connector Connector) <-chan error {
	ch := make(chan error, 1)
	go func() {
		err := connector.Bind()
		ch <- err
		close(ch)
	}()
	return ch
}
