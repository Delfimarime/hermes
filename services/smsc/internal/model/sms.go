package model

import (
	"github.com/delfimarime/hermes/services/smsc/pkg/asyncapi"
	"time"
)

type Sms struct {
	TrackDelivery bool
	Id            string
	TrackId       string
	ListenedAt    time.Time
	Smpp          *asyncapi.ObjectId
}
