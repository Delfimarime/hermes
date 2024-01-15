package outbound

import (
	"github.com/delfimarime/hermes/services/smsc/pkg/asyncapi"
	"github.com/delfimarime/hermes/services/smsc/pkg/common"
	"io"
)

type SendSmsRequestListener interface {
	io.Closer
	common.PostConfigurable
	ListenTo(request asyncapi.SendSmsRequest) (asyncapi.SendSmsResponse, error)
}
