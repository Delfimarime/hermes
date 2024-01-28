package publish

import (
	"github.com/delfimarime/hermes/services/smsc/internal/smpp"
)

type Smsc struct {
	connector               smpp.Connector
	sendSmsRequestPredicate SendSmsRequestPredicate
}
