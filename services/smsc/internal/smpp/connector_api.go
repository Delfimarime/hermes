package smpp

import (
	"github.com/delfimarime/hermes/services/smsc/internal/model"
	"github.com/fiorix/go-smpp/smpp"
)

const (
	ReadyConnectorLifecycleState   = "READY"
	ErrorConnectorLifecycleState   = "ERROR"
	StartupConnectorLifecycleState = "STARTUP"
)

type ConnectorManager interface {
	Close() error
	GetList() []Connector
	AfterPropertiesSet() error
	GetById(id string) Connector
}

type ConnectorFactory interface {
	NewListenerConnector(smpp model.Smpp, f smpp.HandlerFunc) Client
	NewTransmitterConnector(smpp model.Smpp, f smpp.HandlerFunc) Client
}

type Connector interface {
	GetId() string
	GetType() string
	GetState() string
	SendMessage(destination, message string) (SendMessageResponse, error)
}

type SendMessageResponse struct {
	Id string
}
