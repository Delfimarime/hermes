package restapi

import (
	"github.com/delfimarime/hermes/services/smsc/internal/sdk"
	"github.com/delfimarime/hermes/services/smsc/pkg/restapi"
	"github.com/gin-gonic/gin"
)

const (
	AddSmscOperationId = "addSmsc"
)

type SmscApi struct {
	service              sdk.SmscService
	getAuthenticatedUser getAuthenticatedUser
}

func (instance *SmscApi) New(c *gin.Context) {
	withAuthenticatedUser(instance.getAuthenticatedUser, c, AddSmscOperationId, func(username string) error {
		withRequestBody[restapi.NewSmscRequest](c, AddSmscOperationId, func(request *restapi.NewSmscRequest) error {
			response, err := instance.service.Add(username, *request)
			if err != nil {
				sendProblem(c, AddSmscOperationId, err)
				return nil
			}
			c.JSON(200, response)
			return nil
		})
		return nil
	})
}
