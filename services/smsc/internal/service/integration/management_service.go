package integration

import (
	"github.com/delfimarime/hermes/services/smsc/pkg/restapi/common"
	"github.com/delfimarime/hermes/services/smsc/pkg/restapi/integration"
)

// ManagementService defines an interface for managing integrations in order to resolve SMSC credentials
type ManagementService interface {
	// Add creates a new integration point from the provided configuration.
	//
	// Parameters:
	//  - createdBy: the username of the user that performs the operation.
	//  - request: the configuration for the new integration point.
	//
	// Returns:
	//  - integration.NewIntegrationResponse: represents the newly created integration point.
	//  - error: an error if the operation fails, nil otherwise.
	Add(createdBy string, request integration.NewIntegrationRequest) (integration.NewIntegrationResponse, error)

	// FindById retrieves an integration point by its specific id.
	//
	// Parameters:
	//  - id: the identifier of the integration point to retrieve.
	//
	// Returns:
	//  - integration.GetIntegrationResponse: represents the retrieved integration point.
	//  - error: an error if the operation fails, nil otherwise.
	FindById(id string) (integration.GetIntegrationResponse, error)

	// EditById modifies an existing integration point identified by id.
	//
	// Parameters:
	//  - modifiedBy: the username of the user that performs the operation.
	//  - id: the identifier of the integration point to modify.
	//  - request: the new configuration for the integration point.
	//
	// Returns:
	//  - integration.UpdateIntegrationResponse: represents the integration point after modification.
	//  - error: an error if the operation fails, nil otherwise.
	EditById(modifiedBy string, id string, request integration.UpdateIntegrationRequest) (integration.UpdateIntegrationResponse, error)

	// FindAll retrieves a list of integration points that match given search criteria.
	//
	// Parameters:
	//  - searchCriteria: the criteria to search against.
	//
	// Returns:
	//  - common.Page[integration.ListableIntegration]: a paginated list of integration points.
	//  - error: an error if the operation fails, nil otherwise.
	FindAll(searchCriteria integration.SearchCriteria) (common.Page[integration.ListableIntegration], error)

	// RemoveById removes an integration point by its id.
	//
	// Parameters:
	//  - id: the identifier of the integration point to remove.
	//
	// Returns:
	//  - error: an error if the operation fails, nil otherwise.
	RemoveById(id string) error
}
