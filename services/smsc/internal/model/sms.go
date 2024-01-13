package model

import (
	"github.com/delfimarime/hermes/services/smsc/pkg/asyncapi"
	"time"
)

type Sms struct {
	Id             string
	To             string
	Type           string
	From           string
	Tags           []string
	ListenedAt     time.Time
	TrackDelivery  bool
	NumberOfParts  int
	SentParts      []SmsPart
	MaxSizePerPart int
	Smpp           asyncapi.ObjectId
}

type SmsPart struct {
	Id     string
	Status string
}
