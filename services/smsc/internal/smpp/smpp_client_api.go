package smpp

import (
	"github.com/fiorix/go-smpp/smpp"
	"io"
)

type TransmitterConn interface {
	smpp.ClientConn
	Submit(sm *smpp.ShortMessage) (*smpp.ShortMessage, error)
}

type Client interface {
	io.Closer
	Bind() error
	GetId() string
	Refresh() error
	GetType() string
	SendMessage(destination, message string) (SendMessageResponse, error)
}
