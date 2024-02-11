package smsc

import "github.com/delfimarime/hermes/services/smsc/pkg/restapi"

// ManagementService defines an interface for managing SMS Communication Centers (SMSC).
type ManagementService interface {
	// Add creates a new SMSC from the provided configuration
	//
	// Parameters:
	//  - username: the username of the user that performs the operation
	//	-  request: the configuration of the SMSC that must be created
	//
	// Returns:
	//  - restapi.NewSmscResponse: that represents the newly created SMSC
	//  - error: an error if the operation fails, nil otherwise
	Add(username string, request restapi.NewSmscRequest) (restapi.NewSmscResponse, error)

	// FindAll retrieves a page of SMSC that match given criteria
	//
	// Parameters:
	//	- request: the criteria to fetch against
	//
	// Returns:
	//  - restapi.Page: that has restapi.PaginatedSmsc representing SMSCs
	//  - error: an error if the operation fails, nil otherwise
	FindAll(request restapi.SmscSearchRequest) (restapi.Page[restapi.PaginatedSmsc], error)

	// FindById retrieves an SMSC with a specific id
	//
	// Parameters:
	//	- id: identifies the SMSC that needs to be retrieved
	//
	// Returns:
	//  - error: an error if the operation fails, nil otherwise
	FindById(id string) (restapi.GetSmscByIdResponse, error)

	// EditById modifies an SMSC with a specific id
	//
	// Parameters:
	//  - username: the username of the user that performs the operation
	//	- id: identifies the SMSC that needs to be modified
	//	- request: the configuration that must be applied to the SMSC
	//
	// Returns:
	//  - restapi.UpdateSmscResponse: that represents the SMSC after apply the configuration
	//  - error: an error if the operation fails, nil otherwise
	EditById(username string, id string, request restapi.UpdateSmscRequest) (restapi.UpdateSmscResponse, error)

	// EditSettingsById modifies an SMSC with a specific id, updating its settings
	//
	// Parameters:
	//  - username: the username of the user that performs the operation
	//	- id: identifies the SMSC that needs to be modified
	//	- request: the configuration that must be applied to the SMSC
	//
	// Returns:
	//  - error: an error if the operation fails, nil otherwise
	EditSettingsById(username string, id string, request restapi.UpdateSmscSettingsRequest) error

	// EditStateById modifies an SMSC with a specific id, updating its state
	//
	// Parameters:
	//  - username: the username of the user that performs the operation
	//	- id: identifies the SMSC that needs to be modified
	//	- request: the configuration that must be applied to the SMSC
	//
	// Returns:
	//  - error: an error if the operation fails, nil otherwise
	EditStateById(username string, id string, request restapi.UpdateSmscState) error

	// RemoveById removes an SMSC with a specific id
	//
	// Parameters:
	//  - username: the username of the user that performs the operation
	//	- id: identifies the SMSC that needs to be removed
	//
	// Returns:
	//  - error: an error if the operation fails, nil otherwise
	RemoveById(username string, id string) error
}
