package restapi

import (
	"github.com/delfimarime/hermes/services/smsc/internal/sdk"
	"github.com/delfimarime/hermes/services/smsc/pkg/restapi"
	"github.com/gin-gonic/gin"
)

type SmscApi struct {
	service sdk.SmscService
}

func (instance *SmscApi) New(operationId string, username string, c *gin.Context) error {
	request, err := readBody[restapi.NewSmscRequest](operationId, c)
	if err != nil {
		return err
	}
	response, err := instance.service.Add(username, *request)
	if err != nil {
		return err
	}
	c.JSON(200, response)
	return nil
}
