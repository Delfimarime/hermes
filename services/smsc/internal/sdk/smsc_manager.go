package sdk

import "github.com/delfimarime/hermes/services/smsc/pkg/restapi"

type SmscService interface {
	Add(username string, request restapi.NewSmscRequest) (restapi.NewSmscResponse, error)
}
