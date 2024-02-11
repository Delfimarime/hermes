package restapi

import (
	"errors"
	"github.com/delfimarime/hermes/services/smsc/internal/service/vault"
	"github.com/delfimarime/hermes/services/smsc/pkg/restapi/common"
	pkgRestapi "github.com/delfimarime/hermes/services/smsc/pkg/restapi/integration"
	"github.com/gin-gonic/gin"
)

type VaultApi struct {
	service vault.ManagementService
}

// New registers a new Integration.
//
// Parameters:
// - operationId: the ID of the operation
// - username: the username
// - c: the gin.Context object
//
// Returns:
// - error: an error if the operation fails, nil otherwise
func (instance *VaultApi) New(operationId, username string, c *gin.Context) error {
	request, err := readBody[pkgRestapi.NewVaultRequest](operationId, c)
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

// FindById retrieves an Integration by ID
//
// Parameters:
// - operationId: the ID of the operation
// - c: the gin.Context object
//
// Returns:
// - error: an error if the operation fails, nil otherwise
func (instance *VaultApi) FindById(c *gin.Context) error {
	response, err := instance.service.FindById(c.Param("id"))
	if err != nil {
		return err
	}
	c.JSON(200, response)
	return nil
}

// FindAll retrieves Integration based on a Criteria
//
// Parameters:
// - operationId: the ID of the operation
// - c: the gin.Context object
//
// Returns:
// - error: an error if the operation fails, nil otherwise
func (instance *VaultApi) FindAll(operationId string, c *gin.Context) error {
	request, err := readQuery[pkgRestapi.SearchCriteria](operationId, c)
	if err != nil {
		return err
	}
	if request.Type != "" && !AnyOf(request.Type, pkgRestapi.HashicorpVaultIntegrationType,
		pkgRestapi.AwsSecretManagerIntegrationType) {
		return GoValidationError{
			Err:  WebpageInputError(errors.New("SearchCriteria.Type, Field validation for 'Type' failed on the 'oneof' tag")),
			From: "query",
		}
	}

	page, err := instance.service.FindAll(*request)
	if err != nil {
		return err
	}
	c.JSON(200, common.ResponsePage[pkgRestapi.ListableVault]{
		Self:  page.Self,
		Next:  page.Next,
		Items: page.Items,
		Prev:  page.Previous,
	})
	return nil
}

// EditById updates an Integration by ID.
//
// Parameters:
// - username: the username
// - c: the gin.Context object
// - operationId: the ID of the operation
//
// Returns:
// - error: an error if the operation fails, nil otherwise
func (instance *VaultApi) EditById(operationId string, username string, c *gin.Context) error {
	request, err := readBody[pkgRestapi.UpdateVaultRequest](operationId, c)
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

// RemoveById removes an SMSC by ID.
//
// Parameters:
// - username: the username
// - c: the gin.Context object
//
// Returns:
// - error: an error if the operation fails, nil otherwise
func (instance *VaultApi) RemoveById(_, username string, c *gin.Context) error {
	err := instance.service.RemoveById(username, c.Param("id"))
	if err != nil {
		return err
	}
	c.JSON(204, nil)
	return nil
}
