package connect

import (
	"github.com/fiorix/go-smpp/smpp"
)

type SmppClient interface {
	Close() error
	Bind() <-chan smpp.ConnStatus
	Submit(sm *smpp.ShortMessage) (*smpp.ShortMessage, error)
}
