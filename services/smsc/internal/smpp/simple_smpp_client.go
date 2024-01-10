package smpp

import (
	"errors"
	"github.com/delfimarime/hermes/services/smsc/internal/model"
	"github.com/fiorix/go-smpp/smpp"
	"github.com/fiorix/go-smpp/smpp/pdu/pdufield"
	"github.com/fiorix/go-smpp/smpp/pdu/pdutext"
)

type SimpleClient struct {
	awaitDeliveryReport bool
	id                  string
	smppType            string
	smppClient          smpp.ClientConn
}

func (instance *SimpleClient) GetType() string {
	return instance.smppType
}

func (instance *SimpleClient) GetId() string {
	return instance.id
}

func (instance *SimpleClient) Bind() error {
	conn := instance.smppClient.Bind()
	if status := <-conn; status.Error() != nil {
		return status.Error()
	}
	return nil
}

func (instance *SimpleClient) Close() error {
	if instance.smppClient == nil {
		return nil
	}
	return instance.smppClient.Close()
}

func (instance *SimpleClient) Refresh() error {
	//TODO implement me
	panic("implement me")
}

func (instance *SimpleClient) SendMessage(destination, message string) (SendMessageResponse, error) {
	if instance.smppType != model.TransceiverType && instance.smppType != model.TransmitterType {
		return SendMessageResponse{}, errors.New("operation not supported")
	}
	deliverySetting := pdufield.NoDeliveryReceipt
	if instance.awaitDeliveryReport {
		deliverySetting = pdufield.FinalDeliveryReceipt
	}
	sm, err := instance.smppClient.(TransmitterConn).Submit(&smpp.ShortMessage{
		Dst:      destination,
		Register: deliverySetting,
		Text:     pdutext.Raw(message),
	})
	if err != nil {
		return SendMessageResponse{}, err
	}
	return SendMessageResponse{
		Id: sm.Resp().Fields()[pdufield.MessageID].String(),
	}, err
}
