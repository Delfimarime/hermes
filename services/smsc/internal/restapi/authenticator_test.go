package restapi

import "github.com/gin-gonic/gin"

type HardCodedAuthenticator struct {
	username string
	clientId string
}

func (h *HardCodedAuthenticator) GetClientId(_ *gin.Context) string {
	return h.clientId
}

func (h *HardCodedAuthenticator) GetPrincipal(_ *gin.Context) string {
	return h.username
}
