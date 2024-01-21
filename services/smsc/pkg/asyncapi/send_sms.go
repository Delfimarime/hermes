package asyncapi

import "time"

type DeliveryStrategy string
type SendSmsResponseType string

const (
	TrackingDeliveryStrategy    DeliveryStrategy = "TRACKING"
	NotTrackingDeliveryStrategy DeliveryStrategy = "NOT_TRACKING"
)

type SendSmsRequest struct {
	Id      string   `json:"id"`
	To      string   `json:"to"`
	Tags    []string `json:"tags"`
	From    string   `json:"from"`
	Content string   `json:"content"`
}

type SendSmsResponse struct {
	Id         string           `json:"id"`
	Problem    *Problem         `json:"problem"`
	Delivery   DeliveryStrategy `json:"delivery"`
	CanceledAt *time.Time       `json:"canceled_at"`
	Smsc       *ObjectId        `json:"send_through"`
}

type Problem struct {
	Title  string `json:"title"`
	Type   string `json:"type"`
	Detail string `json:"detail"`
}

type ObjectId struct {
	Id string `json:"id"`
}
