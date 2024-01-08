package smpp

type NoOpConnector struct {
	id         string
	smppClient Client
}

func (instance *NoOpConnector) GetId() string {
	return instance.id
}

func (instance *NoOpConnector) Bind() error {
	conn := instance.smppClient.Bind()
	if status := <-conn; status.Error() != nil {
		return status.Error()
	}
	return nil
}

func (instance *NoOpConnector) Close() error {
	if instance.smppClient == nil {
		return nil
	}
	return instance.smppClient.Close()
}

func (instance *NoOpConnector) Refresh() error {
	//TODO implement me
	panic("implement me")
}
