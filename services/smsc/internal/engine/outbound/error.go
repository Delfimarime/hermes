package outbound

type ErrorType string

const (
	ServiceNotAvailable ErrorType = "SERVICE_NOT_AVAILABLE"
)

type CannotHandleSendSmsRequestError struct {
	Detail string
	Type   ErrorType
}

func (instance *CannotHandleSendSmsRequestError) Error() string {
	return instance.Detail
}

func NewServiceNotAvailable() error {
	return &CannotHandleSendSmsRequestError{
		Detail: "Cannot",
		Type:   ServiceNotAvailable,
	}
}
