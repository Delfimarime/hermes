package restapi

import "github.com/gin-gonic/gin"

type getAuthenticatedUser func(c *gin.Context) string
