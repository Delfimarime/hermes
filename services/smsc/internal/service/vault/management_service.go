package vault

import (
	"github.com/delfimarime/hermes/services/smsc/pkg/restapi/common"
	"github.com/delfimarime/hermes/services/smsc/pkg/restapi/integration"
)

// ManagementService defines an interface for managing vaults
type ManagementService interface {
	// Add creates a new vault from the provided configuration.
	//
	// Parameters:
	//  - createdBy: the username of the user that performs the operation.
	//  - request: the configuration for the new vault.
	//
	// Returns:
	//  - integration.NewVaultResponse: represents the newly created integration point.
	//  - error: an error if the operation fails, nil otherwise.
	Add(createdBy string, request integration.NewVaultRequest) (integration.NewVaultResponse, error)

	// FindById retrieves a vault by its specific id.
	//
	// Parameters:
	//  - id: the identifier of the vault to retrieve.
	//
	// Returns:
	//  - integration.GetVaultResponse: represents the retrieved vault.
	//  - error: an error if the operation fails, nil otherwise.
	FindById(id string) (integration.GetVaultResponse, error)

	// EditById modifies an existing vault identified by id.
	//
	// Parameters:
	//  - modifiedBy: the username of the user that performs the operation.
	//  - id: the identifier of the vault to modify.
	//  - request: the new configuration for the integration point.
	//
	// Returns:
	//  - integration.UpdateVaultResponse: represents the vault after modification.
	//  - error: an error if the operation fails, nil otherwise.
	EditById(modifiedBy string, id string, request integration.UpdateVaultRequest) (integration.UpdateVaultResponse, error)

	// FindAll retrieves a list of vaults that match given search criteria.
	//
	// Parameters:
	//  - searchCriteria: the criteria to search against.
	//
	// Returns:
	//  - common.Page[integration.ListableVault]: a paginated list of integration points.
	//  - error: an error if the operation fails, nil otherwise.
	FindAll(searchCriteria integration.SearchCriteria) (common.Page[integration.ListableVault], error)

	// RemoveById removes a vault point by its id.
	//
	// Parameters:
	//  - id: the identifier of the integration point to remove.
	//
	// Returns:
	//  - username: the username of the user that performs the operation.
	//  - error: an error if the operation fails, nil otherwise.
	RemoveById(username, id string) error
}
