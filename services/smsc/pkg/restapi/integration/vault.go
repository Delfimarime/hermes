package integration

import "time"

const (
	HashicorpVaultIntegrationType   = "HASHICORP_VAULT"
	AwsSecretManagerIntegrationType = "AWS_SECRET_MANAGER"
)

type NewVaultRequest struct {
	HashicorpVault
	AwsSecretManager
	Name        string `json:"name" binding:"gte=2,lte=50"`
	Description string `json:"description" binding:"lte=255"`
	Type        string `json:"type" binding:"oneof=HASHICORP_VAULT AWS_SECRET_MANAGER"`
}

type NewVaultResponse struct {
	Id string `json:"id"`
	NewVaultRequest
	CreatedAt time.Time `json:"created_at" binding:"required"`
	CreatedBy string    `json:"created_by,omitempty" binding:"required"`
}

type GetVaultResponse struct {
	NewVaultResponse
	LastUpdatedAt *time.Time `json:"last_modified_at,omitempty" binding:"required"`
	LastUpdatedBy string     `json:"last_modified_by,omitempty" binding:"required"`
}

type UpdateVaultRequest struct {
	HashicorpVault
	AwsSecretManager
	Name        string `json:"name" binding:"gte=2,lte=50"`
	Description string `json:"description" binding:"lte=255"`
}

type UpdateVaultResponse GetVaultResponse

type SearchCriteria struct {
	S      string `json:"s,omitempty"`
	Cursor string `json:"cursor,omitempty"`
	Type   string `json:"type,omitempty"`
}

type ListableVault struct {
	Id          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty" binding:"gte=2,lte=50"`
	Description string `json:"description,omitempty" binding:"lte=255"`
	Type        string `json:"type,omitempty" binding:"oneof=HASHICORP_VAULT AWS_SECRET_MANAGER"`
}
