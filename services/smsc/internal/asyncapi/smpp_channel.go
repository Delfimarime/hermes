package asyncapi

import (
	"github.com/delfimarime/hermes/services/smsc/pkg/restapi/smsc"
)

type SmppChannel interface {
	SubmitSmscAddedEvent(event smsc.NewSmscResponse) error
}
