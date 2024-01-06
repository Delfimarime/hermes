package connect

import (
	"fmt"
	"github.com/fiorix/go-smpp/smpp"
	"github.com/fiorix/go-smpp/smpp/pdu"
	"github.com/fiorix/go-smpp/smpp/pdu/pdufield"
	"github.com/fiorix/go-smpp/smpp/pdu/pdutlv"
	"time"
)

func NewPduListener(smscId string, ev SmsEventListener) smpp.HandlerFunc {
	return func(p pdu.Body) {
		switch p.Header().ID {
		case pdu.DeliverSMID:
			whenDeliverySM(smscId, ev, p)
			break
		case pdu.DeliverSMRespID:
			whenDeliverSMResp(smscId, ev, p)
			break
		default:
			//TODO PRINT
			break
		}
	}
}

func whenDeliverSMResp(smscId string, ev SmsEventListener, p pdu.Body) {
	receivedAt := time.Now()
	messageId := p.Fields()[pdufield.MessageID]
	correlationId := p.TLVFields()[pdutlv.TagReceiptedMessageID]
	if messageId == nil {
		messageId = p.Fields()[pdufield.SMDefaultMsgID]
	}
	r := SmsDeliveryResponse{
		SmscId:     smscId,
		ReceivedAt: receivedAt,
		Id:         messageId.String(),
		Status:     int(p.Header().Status),
	}
	if correlationId != nil {
		r.CorrelationId = correlationId.String()
	}
	ev.OnSmsDelivered(r)
}

func whenDeliverySM(smscId string, ev SmsEventListener, p pdu.Body) {
	receivedAt := time.Now()
	id := p.Fields()[pdufield.MessageID]
	sourceAddr := p.Fields()[pdufield.SourceAddr]
	messageContent := p.Fields()[pdufield.ShortMessage]
	if id == nil {
		id = p.Fields()[pdufield.SMDefaultMsgID]
	}
	r := ReceivedSmsRequest{
		SmscId:     smscId,
		ReceivedAt: receivedAt,
		Message:    messageContent.String(),
	}
	if id != nil {
		r.Id = id.String()
	}
	if sourceAddr != nil {
		r.From = sourceAddr.String()
	}
	fmt.Println("TIMESTAMP", p.Fields()[pdufield.ScheduleDeliveryTime])
	ev.OnSmsRequest(r)
}
