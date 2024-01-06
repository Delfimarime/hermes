package connect

import (
	"errors"
	"github.com/fiorix/go-smpp/smpp"
)

type SmppReceiverClientAdapter struct {
	receiver *smpp.Receiver
}

func (instance *SmppReceiverClientAdapter) Close() error {
	return instance.receiver.Close()
}

func (instance *SmppReceiverClientAdapter) Bind() <-chan smpp.ConnStatus {
	return instance.receiver.Bind()
}

func (instance *SmppReceiverClientAdapter) Submit(sm *smpp.ShortMessage) (*smpp.ShortMessage, error) {
	return nil, errors.New("not supported")
}
