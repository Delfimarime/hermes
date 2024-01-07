package connect

import (
	"context"
	"github.com/delfimarime/hermes/services/smsc/internal/model"
	"github.com/delfimarime/hermes/services/smsc/pkg/config"
	"go.uber.org/zap"
	"sync"
)

type ConnectorRegistry struct {
	mutex            sync.Mutex
	Logger           *zap.Logger
	Repository       SmppRepository
	ConnectorFactory ConnectorFactory
	Configuration    config.Configuration
	cache            map[string]*Connector
}

func (instance *ConnectorRegistry) Start() error {
	instance.mutex.Lock()
	defer instance.mutex.Unlock()
	if instance.Logger == nil {
		instance.Logger = zap.L()
	}
	if instance.cache == nil {
		instance.cache = map[string]*Connector{}
	}
	seq, err := instance.Repository.FindAll()
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(),
		MillisToDuration(instance.Configuration.Smsc.StartupTimeout))
	defer cancel()
	for _, definition := range seq {
		connector := instance.ConnectorFactory.New(definition)
		wg.Add(1)
		go func(def model.Smpp, c *Connector) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				instance.Logger.Warn("Cannot bind smsc[id="+c.Id+"]",
					zap.String("smsc_id", c.Id),
					zap.String("smsc_name", def.Name),
					zap.Int64("duration", instance.Configuration.Smsc.StartupTimeout),
				)
			case problem := <-instance.bindConnector(c):
				if problem != nil {
					instance.Logger.Error("Cannot bind smsc[id="+c.Id+"]",
						zap.String("smsc_id", c.Id),
						zap.String("smsc_name", def.Name),
						zap.Error(problem),
					)
				} else {
					instance.Logger.Info("Bind smsc[id="+c.Id+"]",
						zap.String("smsc_id", c.Id),
						zap.String("smsc_name", def.Name),
					)
				}
			}
			instance.cache[c.Id] = c
		}(definition, connector)
	}
	wg.Wait()
	return nil
}

func (instance *ConnectorRegistry) bindConnector(connector *Connector) <-chan error {
	ch := make(chan error, 1)
	go func() {
		err := connector.DoBind()
		ch <- err
		close(ch)
	}()
	return ch
}
func (instance *ConnectorRegistry) Close() error {
	instance.mutex.Lock()
	defer instance.mutex.Unlock()
	if instance.cache == nil {
		return nil
	}
	for _, v := range instance.cache {
		each := v
		if err := each.Close(); err != nil {
			if err != nil {
				instance.Logger.Error("Cannot close smsc[id="+v.Id+"]",
					zap.String("smsc_id", v.Id),
					zap.Error(err),
				)
			}
		}
	}
	instance.cache = nil
	return nil
}
