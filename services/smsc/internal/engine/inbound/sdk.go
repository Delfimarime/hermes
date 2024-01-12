package inbound

import (
	"github.com/delfimarime/hermes/services/smsc/pkg/asyncapi"
	"io"
)

type SendSmsRequestListener interface {
	io.Closer
	ListenTo(request asyncapi.SendSmsRequest) (asyncapi.SendSmsResponse, error)
}
