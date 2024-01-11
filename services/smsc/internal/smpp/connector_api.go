package smpp

type State string

const (
	ReadyConnectorLifecycleState   State = "READY"
	ErrorConnectorLifecycleState   State = "ERROR"
	ClosedConnectorLifecycleState  State = "CLOSED"
	WaitConnectorLifecycleState    State = "WAIT"
	StartupConnectorLifecycleState State = "STARTUP"
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
	SendMessage(destination, message string) (SendMessageResponse, error)
}

type SendMessageResponse struct {
	Id string
}

type UnavailableConnectorError struct {
	CausedBy error
}

func (u UnavailableConnectorError) Error() string {
	if u.CausedBy != nil {
		return u.CausedBy.Error()
	}
	return "Connector isn't ready"
}
