package smpp

import (
	"github.com/delfimarime/hermes/services/smsc/internal/model"
	"github.com/fiorix/go-smpp/smpp"
)

const (
	StartupConnectorLifecycleState = "STARTUP"
	ReadyConnectorLifecycleState   = "READY"
	ErrorConnectorLifecycleState   = "ERROR"
)

type Client interface {
	Close() error
	Bind() <-chan smpp.ConnStatus
}

type TransmitterClient interface {
	Client
	Submit(sm *smpp.ShortMessage) (*smpp.ShortMessage, error)
}

type Connector interface {
	GetId() string
	Bind() error
	Close() error
	Refresh() error
}

type TransmitterConnector interface {
	Connector
	SendMessage(destination, message string) (SendMessageResponse, error)
}

type ConnectorManager interface {
	Close() error
	Refresh(id string) error
	AfterPropertiesSet() error
	StateOf(id string) string
}

type ConnectorFactory interface {
	NewListenerConnector(smpp model.Smpp, f smpp.HandlerFunc) Connector
	NewTransmitterConnector(smpp model.Smpp, f smpp.HandlerFunc) Connector
}

type SendMessageResponse struct {
	Id string
}
