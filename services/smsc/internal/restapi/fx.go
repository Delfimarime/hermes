package restapi

import "github.com/gin-gonic/gin"

func getGinEngine(smscApi *SmscApi) *gin.Engine {
	r := gin.Default()
	r.POST("/smscs", smscApi.New)
	return r
}
