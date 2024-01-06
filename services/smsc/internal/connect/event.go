package connect

import "time"

type ReceivedSmsRequest struct {
	Id         string
	From       string
	Message    string
	SmscId     string
	ReceivedAt time.Time
}

type SmsDeliveryResponse struct {
	Status        int
	SmscId        string
	Id            string
	CorrelationId string
	ReceivedAt    time.Time
}

type SmsEventListener interface {
	OnSmsRequest(request ReceivedSmsRequest)
	OnSmsDelivered(request SmsDeliveryResponse)
}
