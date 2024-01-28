package smpp

import (
	"fmt"
)

type State int

const (
	ClosedConnectorLifecycleState  State = -2
	StartupConnectorLifecycleState State = -1
	WaitConnectorLifecycleState    State = 0
	ReadyConnectorLifecycleState   State = 1
	ErrorConnectorLifecycleState   State = 2
)

type ConnectorManager interface {
	Close() error
	GetList() []Connector
	AfterPropertiesSet() error
	GetById(id string) Connector
}

type Connector interface {
	GetId() string
	GetType() string
	GetAlias() string
	GetState() State
	IsTrackingDelivery() bool
	SendMessage(destination, message string) (SendMessageResponse, error)
}

func (value State) string() string {
	switch value {
	case StartupConnectorLifecycleState:
		return "STARTUP"
	case WaitConnectorLifecycleState:
		return "WAITING"
	case ReadyConnectorLifecycleState:
		return "READY"
	case ErrorConnectorLifecycleState:
		return "ERROR"
	case ClosedConnectorLifecycleState:
		return "CLOSED"
	default:
		return fmt.Sprintf("%d", value)
	}
}
