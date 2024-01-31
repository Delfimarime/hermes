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
	LastUpdatedAt *time.Time `json:"last_updated_at,omitempty"`
	LastUpdatedBy string     `json:"last_updated_by,omitempty"`
}

type NewSmscRequest struct {
	PoweredBy   string              `json:"powered_by,omitempty" binding:"max=45"`
	Type        SmscType            `json:"type" binding:"required" validator:"min=8,max=11"`
	Settings    SmscSettingsRequest `json:"settings" binding:"required"`
	Name        string              `json:"name,omitempty" binding:"required" validator:"min=3,max=50"`
	Alias       string              `json:"alias,omitempty" binding:"required" validator:"min=3,max=20"`
	Description string              `json:"description,omitempty" binding:"required" validator:"min=1,max=255"`
}

type SmscSettingsRequest struct {
	Bind        *Bind     `json:"bind,omitempty"`
	Merge       *Merge    `json:"merge,omitempty"`
	Enquire     *Enquire  `json:"enquire,omitempty"`
	Response    *Response `json:"response,omitempty"`
	Delivery    *Delivery `json:"delivery,omitempty"`
	ServiceType string    `json:"service_type,omitempty"`
	Host        Host      `json:"host" binding:"required"`
	SourceAddr  string    `json:"source_address,omitempty" validator:"hostname_port"`
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
	Timeout int64 `json:"timeout" binding:"required" validator:"min:1000"`
}

type Enquire struct {
	Link        int64 `json:"link" binding:"required" validator:"min:1000"`
	LinkTimeout int64 `json:"link_timeout" binding:"required" validator:"min:1000"`
}

type Response struct {
	Timeout int64 `json:"timeout" binding:"required" validator:"min:1000"`
}

type Merge struct {
	Interval        int64 `json:"interval" binding:"required" validator:"min:1000"`
	CleanupInterval int64 `json:"cleanup_interval" binding:"required" validator:"min:1000"`
}
