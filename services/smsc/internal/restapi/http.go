package restapi

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
	"schneider.vip/problem"
)

func bindAndValidate[T any](c *gin.Context, request *T, operationId string) bool {
	validate, _ := c.Get("validator")
	if err := c.ShouldBindJSON(&request); err != nil {
		sendRequestValidationResponse(c, http.StatusBadRequest, operationId, err.Error())
		return false
	}
	if err := validate.(*validator.Validate).Struct(&request); err != nil {
		sendRequestValidationResponse(c, http.StatusUnprocessableEntity, operationId, err.Error())
		return false
	}
	return true
}

func sendRequestValidationResponse(c *gin.Context, statusCode int, operationId string, err string) {
	_, _ = problem.New(
		problem.Detail(err),
		problem.Status(statusCode),
		problem.Title("Request doesn't comply with defined schema"),
		problem.Type(fmt.Sprintf("/problems/%s/constraint-violation", operationId)),
	).WriteTo(c.Writer)
}

func sendUnauthorizedResponse(c *gin.Context, operationId string, err string) {
	detail := err
	if detail == "" {
		detail = "Cannot proceed with operation, the user isn't authorized to perform it"
	}
	_, _ = problem.New(
		problem.Detail(detail),
		problem.Status(http.StatusForbidden),
		problem.Title("Unauthorized Access"),
		problem.Type(fmt.Sprintf("/problems/%s/unauthorized-access", operationId)),
	).WriteTo(c.Writer)
}

func sendProblem(c *gin.Context, operationId string, causedBy error) {
	title := ""
	detail := ""
	errorType := ""
	statusCode := 0
	if t, isTransactionProblem := causedBy.(TransactionProblem); isTransactionProblem {
		title = t.GetTitle()
		detail = t.GetDetail()
		errorType = t.GetErrorType()
		statusCode = t.GetStatusCode()
	} else if detail == "" {
		detail = causedBy.Error()
	}
	if detail == "" {
		detail = "Cannot proceed with operation, the user isn't authorized to perform it"
	}
	if statusCode < 400 || statusCode > 599 {
		statusCode = http.StatusInternalServerError
	}
	if title == "" {
		title = "Something went wrong"
	}
	err := fmt.Sprintf("/problems/%s", operationId)
	if errorType != "" {
		err += "/" + errorType
	}
	_, _ = problem.New(
		problem.Title(title),
		problem.Detail(detail),
		problem.Type(errorType),
		problem.Status(statusCode),
	).WriteTo(c.Writer)
}
