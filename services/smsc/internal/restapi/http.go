package restapi

import (
	"fmt"
	"github.com/delfimarime/hermes/services/smsc/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"net/http"
	"schneider.vip/problem"
)

const (
	httpValidationTitle               = "Request not compliant"
	constraintViolationF              = "/problems/%s/constraint-violation"
	httpValidationDetailF             = `Request doesn't comply with Operation{"id"="%s"} schema`
	httpValidationDetailWithLocationF = `$.%s doesn't comply with Operation{"id"="%s"} schema`

	somethingWentWrongF       = "/problems/%s"
	somethingWentWrongTitle   = "Something went wrong"
	somethingWentWrongDetailF = `Cannot proceed with Operation{"id"="%s"}, the user isn't authorized to perform it`

	UnauthorizedAccessTitle       = "Unauthorized Access"
	UnauthenticatedResponseDetail = "Cannot proceed with operation, the user and/or client cannot be determined"
	UnauthorizedResponseDetail    = "Cannot proceed with operation, the user isn't authorized to perform it"
)

func readBody[T any](operationId string, c *gin.Context) (*T, error) {
	return read[T](operationId, c, bindBody[T])

}

func readQuery[T any](operationId string, c *gin.Context) (*T, error) {
	return read[T](operationId, c, bindQuery[T])
}

func read[T any](operationId string, c *gin.Context, doBind func(operationId string, c *gin.Context, request *T) error) (*T, error) {
	request := new(T)
	err := doBind(operationId, c, request)
	return request, err
}

func bindBody[T any](operationId string, c *gin.Context, request *T) error {
	return doBind[T](operationId, c, binding.JSON, request)
}

func bindQuery[T any](operationId string, c *gin.Context, request *T) error {
	return doBind[T](operationId, c, binding.Query, request)
}

func doBind[T any](operationId string, c *gin.Context, b binding.Binding, request *T) error {
	err := c.ShouldBindWith(request, b)
	if err != nil {
		zap.L().Error("Cannot bind gin.Context into target Object",
			zap.String("operationId", operationId),
			zap.String("uri", c.Request.RequestURI),
			zap.String("binding_tag", b.Name()),
			zap.Error(err),
		)
		switch b.Name() {
		case binding.JSON.Name(), binding.Query.Name(), binding.Header.Name():
			name := b.Name()
			if name == binding.JSON.Name() {
				name = "body"
			}
			return RequestValidationError{
				error: err,
				From:  name,
			}
		default:
			return err
		}
	}
	return nil
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
	handler := getErrorHandler(causedBy)
	if handler != nil {
		handler(c, operationId, causedBy)
		return
	}
	sendErrorResponse(c, operationId, somethingWentWrongTitle, fmt.Sprintf(somethingWentWrongDetailF, operationId), "", http.StatusInternalServerError)
}

func getErrorHandler(err error) func(c *gin.Context, operationId string, causedBy error) {
	switch err.(type) {
	case service.TransactionProblem:
		return handleTransactionProblem
	case validator.ValidationErrors, *validator.InvalidValidationError:
		return handleValidationErrors
	case *RequestValidationError:
		return handleRequestValidationError
	default:
		return nil
	}
}

func handleTransactionProblem(c *gin.Context, operationId string, causedBy error) {
	t := causedBy.(service.TransactionProblem)
	sendErrorResponse(c, operationId, t.GetTitle(), t.GetDetail(), t.GetErrorType(), t.GetStatusCode())
}

func handleValidationErrors(c *gin.Context, operationId string, causedBy error) {
	statusCode := http.StatusUnprocessableEntity
	if _, ok := causedBy.(*validator.InvalidValidationError); ok {
		statusCode = http.StatusBadRequest
	}
	sendRequestValidationResponse(c, statusCode, operationId, fmt.Sprintf(httpValidationDetailF, operationId))
}

func handleRequestValidationError(c *gin.Context, operationId string, causedBy error) {
	t := causedBy.(*RequestValidationError)
	statusCode := http.StatusBadRequest
	detail := fmt.Sprintf(httpValidationDetailWithLocationF, operationId, t.From)

	switch t.error.(type) {
	case validator.ValidationErrors:
		statusCode = http.StatusUnprocessableEntity
	case *validator.InvalidValidationError:
		// statusCode remains http.StatusBadRequest
	}
	sendRequestValidationResponse(c, statusCode, operationId, detail)
}

func sendErrorResponse(c *gin.Context, operationId, title, detail, errorType string, statusCode int) {
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

type RequestValidationError struct {
	error error
	From  string
}

func (e RequestValidationError) Error() string {
	return e.error.Error()
}
