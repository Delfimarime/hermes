package restapi

import (
	"github.com/delfimarime/hermes/services/smsc/internal/service/security"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"time"
)

const (
	AddSmscOperationId       = "AddSmsc"
	RemoveSmscOperationId    = "RemoveSmscById"
	EditSmscOperationId      = "EditSmscById"
	EditSmscStateOperationId = "EditSmscStateById"
	EditSmscSettingsId       = "EditSmscSettingsById"
	GetSmscOperationId       = "GetSmscById"
	GetSmscPageOperationId   = "GetSmscPage"
)

const (
	smscEndpoint             = "/smscs"
	smscByIdEndpoint         = "/smscs/:id"
	smscStateByIdEndpoint    = "/smscs/:id/state"
	smscSettingsByIdEndpoint = "/smscs/:id/settings"
)

func getGinEngine(authenticator security.Authenticator, smscApi *SmscApi) *gin.Engine {
	r := gin.Default()
	// SETUP
	r.Use(gin.Recovery())
	r.Use(ginzap.RecoveryWithZap(zap.L(), true))
	r.Use(ginzap.Ginzap(zap.L(), time.RFC3339, true))
	// SMSC
	r.GET(smscByIdEndpoint, withCatchError(GetSmscOperationId, smscApi.FindById))
	r.POST(smscEndpoint, withUser(AddSmscOperationId, authenticator, smscApi.New))
	r.GET(smscEndpoint, withCatchOperationError(GetSmscPageOperationId, smscApi.FindAll))
	r.PUT(smscByIdEndpoint, withUser(EditSmscOperationId, authenticator, smscApi.EditById))
	r.DELETE(smscByIdEndpoint, withUser(RemoveSmscOperationId, authenticator, smscApi.RemoveById))
	r.PUT(smscStateByIdEndpoint, withUser(EditSmscStateOperationId, authenticator, smscApi.EditStateById))
	r.PUT(smscSettingsByIdEndpoint, withUser(EditSmscSettingsId, authenticator, smscApi.EditSettingsById))
	return r
}

func withClient(operationId string, authenticator security.Authenticator, f func(operationId, clientId string, c *gin.Context) error) func(*gin.Context) {
	return withPrincipal(operationId, authenticator, authenticator.GetClientId, f)
}

func withUser(operationId string, authenticator security.Authenticator, f func(operationId, username string, c *gin.Context) error) func(*gin.Context) {
	return withPrincipal(operationId, authenticator, authenticator.GetPrincipal, f)
}

func withPrincipal(operationId string, authenticator security.Authenticator, extract func(*gin.Context) string, exec func(operationId, username string, c *gin.Context) error) func(*gin.Context) {
	return withCatchError(operationId, func(c *gin.Context) error {
		if authenticator == nil {
			setUnauthenticatedResponse(operationId, c)
			return nil
		}
		principal := extract(c)
		if principal == "" {
			setUnauthenticatedResponse(operationId, c)
			return nil
		}
		return exec(operationId, principal, c)
	})
}

func withCatchOperationError(operationId string, exec func(operationId string, c *gin.Context) error) func(c *gin.Context) {
	return withCatchError(operationId, func(c *gin.Context) error {
		return exec(operationId, c)
	})
}

func withCatchError(operationId string, exec func(c *gin.Context) error) func(c *gin.Context) {
	return func(c *gin.Context) {
		if err := exec(c); err != nil {
			sendProblem(c, operationId, err)
			return
		}
	}
}
