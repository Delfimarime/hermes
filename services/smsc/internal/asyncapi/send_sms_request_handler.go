package asyncapi

import (
	"github.com/delfimarime/hermes/services/smsc/pkg/asyncapi"
	"github.com/delfimarime/hermes/services/smsc/pkg/common"
	"io"
)

type SendSmsRequestHandler interface {
	io.Closer
	common.PostConfigurable
	Accept(request asyncapi.SendSmsRequest) (asyncapi.SendSmsResponse, error)
}
