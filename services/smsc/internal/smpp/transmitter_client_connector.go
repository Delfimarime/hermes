package smpp

import (
	"github.com/fiorix/go-smpp/smpp"
	"github.com/fiorix/go-smpp/smpp/pdu/pdufield"
	"github.com/fiorix/go-smpp/smpp/pdu/pdutext"
)

type TransmitterClientConnector struct {
	NoOpConnector
	awaitDeliveryReport bool
}

func (instance *TransmitterClientConnector) SendMessage(destination, message string) (SendMessageResponse, error) {
	deliverySetting := pdufield.NoDeliveryReceipt
	if instance.awaitDeliveryReport {
		deliverySetting = pdufield.FinalDeliveryReceipt
	}
	sm, err := instance.NoOpConnector.smppClient.(TransmitterClient).Submit(&smpp.ShortMessage{
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
