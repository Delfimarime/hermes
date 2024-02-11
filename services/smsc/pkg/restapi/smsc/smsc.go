package smsc

import "time"

const (
	ReceiverType    Type = "RECEIVER"
	TransmitterType Type = "TRANSMITTER"
	TransceiverType Type = "TRANSCEIVER"
)

const (
	ActivatedSmscState   string = "ACTIVATED"
	DeactivatedSmscState string = "DEACTIVATED"
)

type Type string

type NewSmscRequest struct {
	UpdateSmscRequest
	Alias string `json:"alias,omitempty" binding:"required,gte=3,lte=20"`
}

type NewSmscResponse struct {
	NewSmscRequest
	Id        string    `json:"id" binding:"required"`
	CreatedAt time.Time `json:"created_at" binding:"required"`
	CreatedBy string    `json:"created_by,omitempty" binding:"required"`
}

type UpdateSmscRequest struct {
	Settings    Settings `json:"settings" binding:"required"`
	Name        string   `json:"name,omitempty" binding:"required,gte=3,lte=50"`
	PoweredBy   string   `json:"powered_by,omitempty" binding:"omitempty,lte=45"`
	Description string   `json:"description,omitempty" binding:"required,gte=2,lte=255"`
	Type        Type     `json:"type" binding:"required,oneof=TRANSMITTER TRANSCEIVER RECEIVER"`
}

type UpdateSmscResponse struct {
	NewSmscResponse
	LastUpdatedAt time.Time `json:"last_modified_at" binding:"required"`
	LastUpdatedBy string    `json:"last_modified_by,omitempty" binding:"required"`
}

type Settings struct {
	Bind        *Bind     `json:"bind,omitempty"`
	Merge       *Merge    `json:"merge,omitempty"`
	Enquire     *Enquire  `json:"enquire,omitempty"`
	Response    *Response `json:"response,omitempty"`
	Delivery    *Delivery `json:"delivery,omitempty"`
	ServiceType string    `json:"service_type,omitempty"`
	Host        Host      `json:"host" binding:"required"`
	SourceAddr  string    `json:"source_address,omitempty" binding:"omitempty,ipv4"`
}

type Delivery struct {
	AwaitReport bool `json:"await_report" binding:"required"`
}

type Host struct {
	Username string `json:"username,omitempty" binding:"required"`
	Password string `json:"password,omitempty" binding:"required"`
	Address  string `json:"address,omitempty" binding:"required,hostname_port"`
}

type Bind struct {
	Timeout int64 `json:"timeout" binding:"required,gte=1000"`
}

type Enquire struct {
	Link        int64 `json:"link" binding:"required,gte=1000"`
	LinkTimeout int64 `json:"link_timeout" binding:"required,gte=1000"`
}

type Response struct {
	Timeout int64 `json:"timeout" binding:"required,gte=1000"`
}

type Merge struct {
	Interval        int64 `json:"interval" binding:"required,gte=1000"`
	CleanupInterval int64 `json:"cleanup_interval" binding:"required,gte=1000"`
}

type UpdateSmscSettingsRequest Settings

type UpdateSmscState struct {
	Value string `json:"value" binding:"required,oneof=ACTIVATED DEACTIVATED"`
}

type GetSmscByIdResponse UpdateSmscResponse

type SearchCriteriaRequest struct {
	Cursor        string `form:"cursor"`
	CreatedBy     string `form:"created_by"`
	LastUpdatedBy string `form:"last_updated_by"`
	S             string `form:"s" binding:"lte=50"`
	PoweredBy     string `form:"powered_by" binding:"lte=45"`
	State         string `form:"state"`
	Type          Type   `form:"type"`
	Sort          string `form:"sort"`
}

func GetSmscSearchRequestSortOpts() []string {
	return []string{
		"+name", "-name", "+created_by", "-created_at",
		"+last_modified_at", "-last_modified_at", "+state", "-state",
		"+powered_by", "-powered_by",
	}
}

type PaginatedSmsc struct {
	Id          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Alias       string `json:"alias,omitempty"`
	PoweredBy   string `json:"powered_by,omitempty"`
	Description string `json:"description,omitempty"`
	Type        Type   `json:"type"`
}
