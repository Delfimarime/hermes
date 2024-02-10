package restapi

import "github.com/gin-gonic/gin"

type Authenticator interface {
	GetClientId(c *gin.Context) string
	GetPrincipal(c *gin.Context) string
}
