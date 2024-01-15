package asyncapi

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
	Id       string           `json:"id"`
	Delivery DeliveryStrategy `json:"delivery"`
	Smsc     ObjectId         `json:"send_through"`
}

type ObjectId struct {
	Id string `json:"id"`
}
