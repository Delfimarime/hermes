package restapi

import "github.com/gin-gonic/gin"

const (
	AddSmscOperationId       = "AddSmsc"
	RemoveSmscOperationId    = "RemoveSmscById"
	EditSmscOperationId      = "EditSmscById"
	EditSmscStateOperationId = "EditSmscStateById"
	EditSmscSettingsId       = "EidSmscSettingsById"
)

const (
	smscEndpoint             = "/smscs"
	smscByIdEndpoint         = "/smscs/:id"
	smscStateByIdEndpoint    = "/smscs/:id/state"
	smscSettingsByIdEndpoint = "/smscs/:id/state"
)

func getGinEngine(authenticator Authenticator, smscApi *SmscApi) *gin.Engine {
	r := gin.Default()
	r.POST(smscEndpoint, withUser(AddSmscOperationId, authenticator, smscApi.New))
	r.PUT(smscByIdEndpoint, withUser(EditSmscOperationId, authenticator, smscApi.EditById))
	r.DELETE(smscByIdEndpoint, withUser(RemoveSmscOperationId, authenticator, smscApi.RemoveById))
	r.PUT(smscStateByIdEndpoint, withUser(EditSmscStateOperationId, authenticator, smscApi.EditStateById))
	r.POST(smscSettingsByIdEndpoint, withUser(EditSmscSettingsId, authenticator, smscApi.EditSettingsById))
	return r
}

func withClient(operationId string, authenticator Authenticator, f func(operationId, clientId string, c *gin.Context) error) func(*gin.Context) {
	return withPrincipal(operationId, authenticator, authenticator.GetClientId, f)
}

func withUser(operationId string, authenticator Authenticator, f func(operationId, username string, c *gin.Context) error) func(*gin.Context) {
	return withPrincipal(operationId, authenticator, authenticator.GetPrincipal, f)
}

func withPrincipal(operationId string, authenticator Authenticator, extract func(*gin.Context) string, exec func(operationId, username string, c *gin.Context) error) func(*gin.Context) {
	return func(c *gin.Context) {
		if authenticator == nil {
			setUnauthenticatedResponse(operationId, c)
			return
		}
		principal := extract(c)
		if principal == "" {
			setUnauthenticatedResponse(operationId, c)
			return
		}
		if err := exec(operationId, principal, c); err != nil {
			sendProblem(c, operationId, err)
			return
		}
	}
}
