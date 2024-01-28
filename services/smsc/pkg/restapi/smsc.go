package restapi

import "time"

const (
	ReceiverType    SmscType = "RECEIVER"
	TransmitterType SmscType = "TRANSMITTER"
	TransceiverType SmscType = "TRANSCEIVER"
)

type SmscType string

type NewSmscResponse struct {
	NewSmscRequest
	Id            string     `json:"id" binding:"required"`
	CreatedAt     time.Time  `json:"created_at" binding:"required"`
	CreatedBy     string     `json:"created_by,omitempty" binding:"required"`
	LastUpdatedAt *time.Time `json:"last_updated_at"`
	LastUpdatedBy string     `json:"last_updated_by,omitempty"`
}

type NewSmscRequest struct {
	PoweredBy   string              `json:"powered_by,omitempty" binding:"max=45"`
	Type        SmscType            `json:"type" binding:"min=8,max=11"`
	Settings    SmscSettingsRequest `json:"settings" binding:"required"`
	Name        string              `json:"name,omitempty" binding:"required,min=3,max=50"`
	Alias       string              `json:"alias,omitempty" binding:"required,min=3,max=20"`
	Description string              `json:"description,omitempty" binding:"required,min=1,max=255"`
}

type SmscSettingsRequest struct {
	SourceAddr  string    `json:"source_address,omitempty" binding:"hostname_port"`
	ServiceType string    `json:"service_type,omitempty"`
	Host        Host      `json:"host" binding:"required"`
	Bind        *Bind     `json:"bind"`
	Merge       *Merge    `json:"merge"`
	Enquire     *Enquire  `json:"enquire"`
	Response    *Response `json:"response"`
	Delivery    *Delivery `json:"delivery"`
}

type Delivery struct {
	AwaitReport bool `json:"await_report" binding:"required"`
}

type Host struct {
	Username string `json:"username,omitempty" binding:"required"`
	Password string `json:"password,omitempty" binding:"required"`
	Address  string `json:"address,omitempty" binding:"hostname_port"`
}

type Bind struct {
	Timeout int64 `json:"timeout" binding:"min:1000"`
}

type Enquire struct {
	Link        int64 `json:"link" binding:"min:1000"`
	LinkTimeout int64 `json:"link_timeout" binding:"min:1000"`
}

type Response struct {
	Timeout int64 `json:"timeout" binding:"min:1000"`
}

type Merge struct {
	Interval        int64 `json:"interval" binding:"min:1000"`
	CleanupInterval int64 `json:"cleanup_interval" binding:"min:1000"`
}
