package publish

type ErrorType string

const (
	GenericProblemTitle  = "Cannot send async.SendSmsRequest"
	GenericProblemType   = "/smsc/sendSmsRequest/something-went-wrong"
	GenericProblemDetail = "Cannot send async.SendSmsRequest due to an error"
)

const (
	ServiceNotAvailable   ErrorType = "SERVICE_NOT_AVAILABLE"
	ServiceNotFoundDetail string    = "Cannot send async.SendSmsRequest due to connector(s) unavailable"
)

const (
	CannotSendSmsRequestProblemTitle  string = "Cannot send async.SendSmsRequest"
	CannotSendSmsRequestProblemType   string = "/smsc/sendSmsRequest/no-connector-found"
	CannotSendSmsRequestProblemDetail string = "Couldn't determine smpp.Connector capable of sending asyncapi.SendSmsRequest"
)

type CannotHandleSendSmsRequestError struct {
	Detail string
	Type   ErrorType
}

func (instance *CannotHandleSendSmsRequestError) Error() string {
	return instance.Detail
}

func NewServiceNotAvailable(msg string) error {
	detail := msg
	if detail == "" {
		detail = ServiceNotFoundDetail
	}
	return &CannotHandleSendSmsRequestError{
		Detail: detail,
		Type:   ServiceNotAvailable,
	}
}
