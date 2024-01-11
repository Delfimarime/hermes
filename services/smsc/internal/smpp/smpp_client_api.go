package smpp

import (
	"github.com/fiorix/go-smpp/smpp"
	"io"
)

type ClientEventType string

const (
	ClientConnBoundEventType       ClientEventType = "BIND"
	ClientConnErrorEventType       ClientEventType = "ERROR"
	ClientConnInterruptedEventType ClientEventType = "RECONNECT"
	ClientConnBindErrorEventType   ClientEventType = "BIND_ERROR"
	ClientConnDisconnectEventType  ClientEventType = "DISCONNECT"
)

type TransmitterConn interface {
	smpp.ClientConn
	Submit(sm *smpp.ShortMessage) (*smpp.ShortMessage, error)
}

type Client interface {
	io.Closer
	Bind()
	GetId() string
	GetType() string
	SendMessage(destination, message string) (SendMessageResponse, error)
}

type ClientConnEvent struct {
	Err  error
	Type ClientEventType
}
