package connect

import (
	"github.com/fiorix/go-smpp/smpp"
	"github.com/fiorix/go-smpp/smpp/pdu/pdufield"
	"github.com/fiorix/go-smpp/smpp/pdu/pdutext"
	"github.com/fiorix/go-smpp/smpp/pdu/pdutlv"
	"github.com/google/uuid"
)

type Connector struct {
	SupportsDeliveryReport bool
	sourceAddress          string
	client                 SmppClient
}

func (instance *Connector) DoBind() error {
	conn := instance.client.Bind()
	if status := <-conn; status.Error() != nil {
		return status.Error()
	}
	return nil
}

func (instance *Connector) SendMessage(id, destination, message string) error {
	msgId := id
	if msgId == "" {
		msgId = uuid.New().String()
	}
	deliverySetting := pdufield.NoDeliveryReceipt
	if instance.SupportsDeliveryReport {
		deliverySetting = pdufield.FinalDeliveryReceipt
	}
	_, err := instance.client.Submit(&smpp.ShortMessage{
		Dst:      destination,
		Text:     pdutext.Raw(message),
		Src:      instance.sourceAddress,
		Register: deliverySetting,
		TLVFields: pdutlv.Fields{
			pdutlv.TagReceiptedMessageID: pdutlv.CString(msgId),
		},
	})
	return err
}

func (instance *Connector) Close() error {
	return instance.client.Close()
}
