package restapi

import (
	"fmt"
	"github.com/delfimarime/hermes/services/smsc/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"net/http"
	"schneider.vip/problem"
)

const (
	httpValidationTitle   = "Request not compliant"
	constraintViolationF  = "/problems/%s/constraint-violation"
	httpValidationDetailF = `Request doesn't comply with Operation{"id"="%s"} schema`

	somethingWentWrongF       = "/problems/%s"
	somethingWentWrongTitle   = "Something went wrong"
	somethingWentWrongDetailF = `Cannot proceed with Operation{"id"="%s"}, the user isn't authorized to perform it`
)

func withAuthenticatedUser(f getAuthenticatedUser, c *gin.Context, operationId string, exec func(username string) error) {
	if f == nil {
		return
	}
	username := f(c)
	if username == "" {
		sendUnauthorizedResponse(c, operationId, "")
		return
	}
	if err := exec(username); err != nil {
		sendProblem(c, operationId, err)
	}
}

func withRequestBody[T any](c *gin.Context, operationId string, exec func(*T) error) {
	request := new(T)
	if !bindAndValidate(c, request, operationId) {
		return
	}
	if err := exec(request); err != nil {
		sendProblem(c, operationId, err)
	}
}

func bindAndValidate[T any](c *gin.Context, request *T, operationId string) bool {
	if err := c.ShouldBindJSON(request); err != nil {
		zap.L().Error("Cannot bind JSON from gin.Context",
			zap.String("operationId", operationId),
			zap.String("uri", c.Request.RequestURI),
			zap.Error(err),
		)
		if _, isValidationError := err.(validator.ValidationErrors); isValidationError {
			sendRequestValidationResponse(c, http.StatusUnprocessableEntity, operationId,
				fmt.Sprintf(httpValidationDetailF, operationId))
		} else {
			sendRequestValidationResponse(c, http.StatusBadRequest, operationId,
				fmt.Sprintf(httpValidationDetailF, operationId))
		}
		return false
	}
	return true
}

func sendRequestValidationResponse(c *gin.Context, statusCode int, operationId string, err string, opts ...problem.Option) {
	arr := []problem.Option{
		problem.Detail(err),
		problem.Status(statusCode),
		problem.Title(httpValidationTitle),
		problem.Type(fmt.Sprintf(constraintViolationF, operationId)),
		problem.Custom("operationId", operationId),
	}
	if opts != nil {
		arr = append(arr, opts...)
	}
	_, _ = problem.New(arr...).WriteTo(c.Writer)
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
		problem.Custom("operationId", operationId),
	).WriteTo(c.Writer)
}

func sendProblem(c *gin.Context, operationId string, causedBy error) {
	title := ""
	detail := ""
	errorType := ""
	statusCode := 0
	if t, isTransactionProblem := causedBy.(service.TransactionProblem); isTransactionProblem {
		title = t.GetTitle()
		detail = t.GetDetail()
		errorType = t.GetErrorType()
		statusCode = t.GetStatusCode()
	}
	if detail == "" {
		detail = fmt.Sprintf(somethingWentWrongDetailF, operationId)
	}
	if statusCode < 400 || statusCode > 599 {
		statusCode = http.StatusInternalServerError
	}
	if title == "" {
		title = somethingWentWrongTitle
	}
	determinedType := fmt.Sprintf(somethingWentWrongF, operationId)
	if errorType != "" {
		determinedType += "/" + errorType
	}
	_, _ = problem.New(
		problem.Title(title),
		problem.Detail(detail),
		problem.Type(determinedType),
		problem.Status(statusCode),
		problem.Custom("operationId", operationId),
	).WriteTo(c.Writer)
}
