package connect

import "time"

type SmsEventListener interface {
	OnSmsRequest(request ReceivedSmsRequest)
	OnSmsDelivered(request SmsDeliveryRequest)
}

type ReceivedSmsRequest struct {
	Id         string
	From       string
	Message    string
	SmscId     string
	ReceivedAt time.Time
}

type SmsDeliveryRequest struct {
	Status     int
	SmscId     string
	Id         string
	ReceivedAt time.Time
}
