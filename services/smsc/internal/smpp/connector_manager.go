package smpp

import (
	"context"
	"github.com/delfimarime/hermes/services/smsc/internal/model"
	"github.com/delfimarime/hermes/services/smsc/internal/smpp/pkg"
	"github.com/delfimarime/hermes/services/smsc/pkg/common"
	"github.com/delfimarime/hermes/services/smsc/pkg/config"
	"github.com/fiorix/go-smpp/smpp/pdu"
	"go.uber.org/zap"
	"sync"
)

type SimpleConnectorManager struct {
	mutex            sync.Mutex
	Logger           *zap.Logger
	Repository       pkg.SmppRepository
	Configuration    config.Configuration
	cache            map[string]Connector
	ConnectorFactory SimpleConnectorFactory
}

func (instance *SimpleConnectorManager) AfterPropertiesSet() error {
	instance.mutex.Lock()
	defer instance.mutex.Unlock()
	if instance.Logger == nil {
		instance.Logger = zap.L()
	}
	if instance.cache == nil {
		instance.cache = map[string]Connector{}
	}
	seq, err := instance.Repository.FindAll()
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(),
		common.MillisToDuration(instance.Configuration.Smsc.StartupTimeout))
	defer cancel()
	for _, definition := range seq {
		connector := instance.newConnector(definition)
		wg.Add(1)
		go func(def model.Smpp, c Connector) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				instance.Logger.Warn("Cannot bind smsc[id="+c.GetId()+"]",
					zap.String("smsc_id", def.Id),
					zap.String("smsc_name", def.Name),
					zap.Int64("duration", instance.Configuration.Smsc.StartupTimeout),
				)
			case problem := <-instance.bindConnector(c):
				if problem != nil {
					instance.Logger.Error("Cannot bind smsc[id="+c.GetId()+"]",
						zap.String("smsc_id", def.Id),
						zap.String("smsc_name", def.Name),
						zap.Error(problem),
					)
				} else {
					instance.Logger.Info("Bind smsc[id="+c.GetId()+"]",
						zap.String("smsc_id", def.Id),
						zap.String("smsc_name", def.Name),
					)
				}
			}
			// c.state = ReadyConnectorLifecycleState
			instance.cache[def.Id] = c
		}(definition, connector)
	}
	wg.Wait()
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
				instance.Logger.Error("Cannot close smsc[id="+v.GetId()+"]",
					zap.String("smsc_id", v.GetId()),
					zap.Error(err),
				)
			}
		}
	}
	instance.cache = nil
	return nil
}

func (instance *SimpleConnectorManager) Refresh(id string) error {
	//TODO implement me
	panic("implement me")
}

func (instance *SimpleConnectorManager) newConnector(definition model.Smpp) Connector {
	var connector Connector
	switch definition.Type {
	case model.ReceiverType:
		connector = instance.ConnectorFactory.NewListenerConnector(definition, instance.onPduEvent)
		break
	case model.TransceiverType:
		connector = instance.ConnectorFactory.NewTransmitterConnector(definition, instance.onPduEvent)
		break
	case model.TransmitterType:
		connector = instance.ConnectorFactory.NewTransmitterConnector(definition, nil)
		break
	default:
		return nil
	}
	// TODO: DECORATE CONNECTOR
	return connector
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

func (instance *SimpleConnectorManager) onPduEvent(p pdu.Body) {

}
