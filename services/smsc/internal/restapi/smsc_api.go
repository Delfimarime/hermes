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
	username := instance.getAuthenticatedUser(c)
	if username == "" {
		sendUnauthorizedResponse(c, AddSmscOperationId, "")
		return
	}
	request := &restapi.NewSmscRequest{}
	if !bindAndValidate(c, request, AddSmscOperationId) {
		return
	}
	response, err := instance.service.Add(username, *request)
	if err != nil {
		sendProblem(c, AddSmscOperationId, err)
		return
	}
	c.JSON(200, response)
}
