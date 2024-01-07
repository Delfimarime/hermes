package connect

import (
	"errors"
	"github.com/fiorix/go-smpp/smpp"
	"github.com/fiorix/go-smpp/smpp/pdu"
	"github.com/fiorix/go-smpp/smpp/pdu/pdufield"
	"github.com/fiorix/go-smpp/smpp/pdu/pdutext"
	"github.com/fiorix/go-smpp/smpp/pdu/pdutlv"
	"time"
)

const (
	StartupConnectorLifecycleState = "STARTUP"
	ReadyConnectorLifecycleState   = "READY"
)

type Connector struct {
	awaitDeliveryReport bool
	submitsMessage      bool
	id                  string
	sourceAddress       string
	state               string
	client              SmppClient
	smsEventListener    SmsEventListener
}

func (instance *Connector) Close() error {
	return instance.client.Close()
}

func (instance *Connector) Refresh() error {
	return nil
}

func (instance *Connector) SendMessage(destination, message string) (string, error) {
	if !instance.submitsMessage {
		return "", errors.New("connector isn't allowed to send sms")
	}
	if instance.state != ReadyConnectorLifecycleState {
		return "", errors.New("connector cannot to send sms due to state=" + instance.state)
	}
	deliverySetting := pdufield.NoDeliveryReceipt
	if instance.awaitDeliveryReport {
		deliverySetting = pdufield.FinalDeliveryReceipt
	}
	sm, err := instance.client.Submit(&smpp.ShortMessage{
		Dst:      destination,
		Text:     pdutext.Raw(message),
		Src:      instance.sourceAddress,
		Register: deliverySetting,
	})
	return sm.Resp().Fields()[pdufield.MessageID].String(), err
}

func (instance *Connector) Listen(p pdu.Body) {
	switch p.Header().ID {
	case pdu.DeliverSMID:
		instance.onDeliverySM(p)
	}
}

func (instance *Connector) doBind() error {
	conn := instance.client.Bind()
	if status := <-conn; status.Error() != nil {
		return status.Error()
	}
	return nil
}

func (instance *Connector) onDeliverySM(p pdu.Body) {
	esmClass := p.Fields()[pdufield.ESMClass].Raw().(uint8)
	switch esmClass {
	case 0x00:
		receivedAt := time.Now()
		fromRequest := getPduObject(p)
		r := ReceivedSmsRequest{
			ReceivedAt: receivedAt,
			SmscId:     instance.id,
			Message:    fromRequest.content,
		}
		if fromRequest.messageId != "" {
			r.Id = fromRequest.messageId
		}
		if fromRequest.sourceAddr != "" {
			r.From = fromRequest.sourceAddr
		}
		instance.smsEventListener.OnSmsRequest(r)
		break
	case 0x04:
		instance.smsEventListener.OnSmsDelivered(SmsDeliveryRequest{
			ReceivedAt: time.Now(),
			SmscId:     instance.id,
			Status:     int(p.Header().Status),
			Id:         p.TLVFields()[pdutlv.TagReceiptedMessageID].String(),
		})
		break
	}
}
