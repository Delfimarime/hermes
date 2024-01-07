package connect

import "github.com/delfimarime/hermes/services/smsc/internal/data"

type ConnectorRegistry struct {
	ConnectorFactory
	Repository data.SmppRepository
	Cache      map[string]Connector
}

func (instance *ConnectorRegistry) Start() error {
	seq, err := instance.Repository.FindAll()
	if err != nil {
		return err
	}
	for _, config := range seq {
		connector := instance.New(config)
	}
	return nil
}

func (instance *ConnectorRegistry) Close() error {
	if instance.Cache == nil {
		return nil
	}
	for _, v := range instance.Cache {
		defer func() {
			v.Close() //TODO LOG ERROR
		}()
	}
	return nil
}
