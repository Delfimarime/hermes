package restapi

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"net/http"
	"schneider.vip/problem"
)

const (
	httpValidationTitle   = "Request not compliant"
	httpValidationDetailF = `Request doesn't comply with Operation{"id"="%s"} schema`
)

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
		problem.Type(fmt.Sprintf("/problems/%s/constraint-violation", operationId)),
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
