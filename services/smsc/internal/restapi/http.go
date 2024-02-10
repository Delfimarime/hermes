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

	UnauthorizedAccessTitle       = "Unauthorized Access"
	UnauthenticatedResponseDetail = "Cannot proceed with operation, the user and/or client cannot be determined"
	UnauthorizedResponseDetail    = "Cannot proceed with operation, the user isn't authorized to perform it"
)

func readBody[T any](operationId string, c *gin.Context) (*T, error) {
	request := new(T)
	err := bind[T](operationId, c, request)
	return request, err
}

func bind[T any](operationId string, c *gin.Context, request *T) error {
	err := c.ShouldBindJSON(request)
	if err != nil {
		zap.L().Error("Cannot bind JSON from gin.Context",
			zap.String("operationId", operationId),
			zap.String("uri", c.Request.RequestURI),
			zap.Error(err),
		)
	}
	return err
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
		detail = UnauthorizedResponseDetail
	}
	_, _ = problem.New(
		problem.Detail(detail),
		problem.Status(http.StatusForbidden),
		problem.Title(UnauthorizedAccessTitle),
		problem.Type(fmt.Sprintf("/problems/%s/unauthorized-access", operationId)),
		problem.Custom("operationId", operationId),
	).WriteTo(c.Writer)
}

func setUnauthenticatedResponse(operationId string, c *gin.Context) {
	_, _ = problem.New(
		problem.Status(http.StatusUnauthorized),
		problem.Title(UnauthorizedAccessTitle),
		problem.Detail(UnauthenticatedResponseDetail),
		problem.Type(fmt.Sprintf("/problems/%s/not-authenticated", operationId)),
		problem.Custom("operationId", operationId),
	).WriteTo(c.Writer)
}
func sendProblem(c *gin.Context, operationId string, causedBy error) {
	var (
		title      = somethingWentWrongTitle
		detail     = fmt.Sprintf(somethingWentWrongDetailF, operationId)
		errorType  = ""
		statusCode = http.StatusInternalServerError
	)

	switch t := causedBy.(type) {
	case service.TransactionProblem:
		title, detail, errorType, statusCode = t.GetTitle(), t.GetDetail(), t.GetErrorType(), t.GetStatusCode()
	case validator.ValidationErrors:
		sendRequestValidationResponse(c, http.StatusUnprocessableEntity, operationId, fmt.Sprintf(httpValidationDetailF, operationId))
		return
	case *validator.InvalidValidationError:
		sendRequestValidationResponse(c, http.StatusBadRequest, operationId, fmt.Sprintf(httpValidationDetailF, operationId))
		return
	}

	if statusCode < 400 || statusCode > 599 {
		statusCode = http.StatusInternalServerError
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
