package asyncapi

import "github.com/delfimarime/hermes/services/smsc/pkg/restapi"

type SmppChannel interface {
	SubmitSmscAddedEvent(event restapi.NewSmscResponse) error
}
