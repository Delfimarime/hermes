package restapi

import "github.com/gin-gonic/gin"

type SecurityContext interface {
	GetUsernameFrom(c *gin.Context) string
}
