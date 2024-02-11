package restapi

import (
	"errors"
	"github.com/delfimarime/hermes/services/smsc/internal/sdk"
	"github.com/delfimarime/hermes/services/smsc/pkg/restapi"
	"github.com/gin-gonic/gin"
)

type SmscApi struct {
	service sdk.SmscService
}

// New registers a new SMSC.
//
// Parameters:
// - operationId: the ID of the operation
// - username: the username
// - c: the gin.Context object
//
// Returns:
// - error: an error if the operation fails, nil otherwise
func (instance *SmscApi) New(operationId, username string, c *gin.Context) error {
	request, err := readBody[restapi.NewSmscRequest](operationId, c)
	if err != nil {
		return err
	}
	response, err := instance.service.Add(username, *request)
	if err != nil {
		return err
	}
	c.JSON(200, response)
	return nil
}

// FindAll retrieves SMSCs based on a Criteria
//
// Parameters:
// - operationId: the ID of the operation
// - c: the gin.Context object
//
// Returns:
// - error: an error if the operation fails, nil otherwise
func (instance *SmscApi) FindAll(operationId string, c *gin.Context) error {
	request, err := readQuery[restapi.SmscSearchRequest](operationId, c)
	if err != nil {
		return err
	}
	if request.Type != "" && !AnyOf(string(request.Type), string(restapi.TransmitterType),
		string(restapi.ReceiverType), string(restapi.TransceiverType)) {
		return GoValidationError{
			Err:  WebpageInputError(errors.New("SmscSearchRequest.Type, Field validation for 'Type' failed on the 'oneof' tag")),
			From: "query",
		}
	}
	if request.State != "" && !AnyOf(request.State, restapi.ActivatedSmscState, restapi.DeactivatedSmscState) {
		return GoValidationError{
			Err:  WebpageInputError(errors.New("SmscSearchRequest.State, Field validation for 'State' failed on the 'oneof' tag")),
			From: "query",
		}
	}
	if request.Sort != "" && !AnyOf(request.Sort, restapi.GetSmscSearchRequestSortOpts()...) {
		return GoValidationError{
			Err:  WebpageInputError(errors.New("SmscSearchRequest.Sort, Field validation for 'Sort' failed on the 'oneof' tag")),
			From: "query",
		}
	}
	page, err := instance.service.FindAll(*request)
	if err != nil {
		return err
	}
	c.JSON(200, restapi.ResponsePage[restapi.PaginatedSmsc]{
		Self:  page.Self,
		Next:  page.Next,
		Items: page.Items,
		Prev:  page.Previous,
	})
	return nil
}

// FindById retrieves an SMSC by ID
//
// Parameters:
// - operationId: the ID of the operation
// - c: the gin.Context object
//
// Returns:
// - error: an error if the operation fails, nil otherwise
func (instance *SmscApi) FindById(c *gin.Context) error {
	response, err := instance.service.FindById(c.Param("id"))
	if err != nil {
		return err
	}
	c.JSON(200, response)
	return nil
}

// EditById updates an SMSC by ID.
//
// Parameters:
// - username: the username
// - c: the gin.Context object
// - operationId: the ID of the operation
//
// Returns:
// - error: an error if the operation fails, nil otherwise
func (instance *SmscApi) EditById(operationId string, username string, c *gin.Context) error {
	request, err := readBody[restapi.UpdateSmscRequest](operationId, c)
	if err != nil {
		return err
	}
	response, err := instance.service.EditById(username, c.Param("id"), *request)
	if err != nil {
		return err
	}
	c.JSON(200, response)
	return nil
}

// EditStateById updates an SMSC state by ID.
//
// Parameters:
// - username: the username
// - c: the gin.Context object
// - operationId: the ID of the operation
//
// Returns:
// - error: an error if the operation fails, nil otherwise
func (instance *SmscApi) EditStateById(operationId string, username string, c *gin.Context) error {
	request, err := readBody[restapi.UpdateSmscState](operationId, c)
	if err != nil {
		return err
	}
	err = instance.service.EditStateById(username, c.Param("id"), *request)
	if err != nil {
		return err
	}
	c.JSON(204, nil)
	return nil
}

// EditSettingsById updates the settings of an SMSC by ID.
//
// Parameters:
// - operationId: the ID of the operation
// - username: the username
// - c: the gin.Context object
//
// Returns:
// - error: an error if the operation fails, nil otherwise
func (instance *SmscApi) EditSettingsById(operationId, username string, c *gin.Context) error {
	request, err := readBody[restapi.UpdateSmscSettingsRequest](operationId, c)
	if err != nil {
		return err
	}
	err = instance.service.EditSettingsById(username, c.Param("id"), *request)
	if err != nil {
		return err
	}
	c.JSON(204, nil)
	return nil
}

// RemoveById removes an SMSC by ID.
//
// Parameters:
// - username: the username
// - c: the gin.Context object
//
// Returns:
// - error: an error if the operation fails, nil otherwise
func (instance *SmscApi) RemoveById(_, username string, c *gin.Context) error {
	err := instance.service.RemoveById(username, c.Param("id"))
	if err != nil {
		return err
	}
	c.JSON(204, nil)
	return nil
}
