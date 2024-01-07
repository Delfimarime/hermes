package connect

import (
	"errors"
	"github.com/fiorix/go-smpp/smpp"
)

type SmppReceiverClientAdapter struct {
	target *smpp.Receiver
}

func (instance *SmppReceiverClientAdapter) Close() error {
	return instance.target.Close()
}

func (instance *SmppReceiverClientAdapter) Bind() <-chan smpp.ConnStatus {
	return instance.target.Bind()
}

func (instance *SmppReceiverClientAdapter) Submit(sm *smpp.ShortMessage) (*smpp.ShortMessage, error) {
	return nil, errors.New("not supported")
}
