package outbound

import (
	"github.com/delfimarime/hermes/services/smsc/internal/smpp"
	"github.com/google/uuid"
)

type TestConnectorManager struct {
	seq []smpp.Connector
}

func (t *TestConnectorManager) Close() error {
	return nil
}

func (t *TestConnectorManager) GetList() []smpp.Connector {
	return t.seq
}

func (t *TestConnectorManager) AfterPropertiesSet() error {
	return nil
}

func (t *TestConnectorManager) GetById(id string) smpp.Connector {
	if t.seq == nil {
		return nil
	}
	for _, each := range t.seq {
		if each.GetId() == id {
			return each
		}
	}
	return nil
}

type TestConnector struct {
	err            error
	trackDelivery  bool
	id             string
	connectionType string
	alias          string
	state          smpp.State
}

func (t TestConnector) GetId() string {
	return t.id
}

func (t TestConnector) GetType() string {
	return t.connectionType
}

func (t TestConnector) GetAlias() string {
	return t.alias
}

func (t TestConnector) GetState() smpp.State {
	return t.state
}

func (t TestConnector) IsTrackingDelivery() bool {
	return t.trackDelivery
}

func (t TestConnector) SendMessage(_, _ string) (smpp.SendMessageResponse, error) {
	if t.err != nil {
		return smpp.SendMessageResponse{}, t.err
	}
	return smpp.SendMessageResponse{
		Id: uuid.New().String(),
	}, nil
}
