package smpp

const (
	ReadyConnectorLifecycleState  = "READY"
	ErrorConnectorLifecycleState  = "ERROR"
	ClosedConnectorLifecycleState = "CLOSED"
	WaitConnectorLifecycleState   = "WAIT"
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
	GetState() string
	GetAlias() string
	SendMessage(destination, message string) (SendMessageResponse, error)
}

type SendMessageResponse struct {
	Id string
}
