package smpp

import (
	"errors"
	"github.com/delfimarime/hermes/services/smsc/internal/model"
	"github.com/fiorix/go-smpp/smpp"
	"github.com/fiorix/go-smpp/smpp/pdu/pdufield"
	"github.com/fiorix/go-smpp/smpp/pdu/pdutext"
)

const (
	smppSingleSmsSize = 133
)

type ClientConnEventListener func(ClientConnEvent)

type SimpleClient struct {
	awaitDeliveryReport bool
	id                  string
	smppType            string
	smppConn            smpp.ClientConn
	clientEventListener ClientConnEventListener
	bindingChannel      <-chan smpp.ConnStatus
}

func (instance *SimpleClient) Close() error {
	return instance.smppConn.Close()
}

func (instance *SimpleClient) GetType() string {
	return instance.smppType
}

func (instance *SimpleClient) GetId() string {
	return instance.id
}

func (instance *SimpleClient) Bind() {
	if instance.bindingChannel != nil {
		return
	}
	instance.bindingChannel = instance.smppConn.Bind()
	go func(ch <-chan smpp.ConnStatus) {
		instance.observeClientConn(ch)
	}(instance.bindingChannel)
}

func (instance *SimpleClient) IsTrackingDelivery() bool {
	return instance.awaitDeliveryReport
}

func (instance *SimpleClient) SendMessage(destination, message string) (SendMessageResponse, error) {
	if instance.smppType != model.TransceiverType && instance.smppType != model.TransmitterType {
		return SendMessageResponse{}, errors.New("operation not supported")
	}
	deliverySetting := pdufield.NoDeliveryReceipt
	if instance.IsTrackingDelivery() {
		deliverySetting = pdufield.FinalDeliveryReceipt
	}
	var err error
	resp := SendMessageResponse{
		Parts: make([]SendMessagePart, 0),
	}
	numberOfParts := int((len(message)-1)/smppSingleSmsSize) + 1
	for i := 0; i < numberOfParts; i++ {
		content := ""
		if i == numberOfParts-1 {
			content = message[i*smppSingleSmsSize:]
		} else {
			content = message[i*smppSingleSmsSize : (i+1)*smppSingleSmsSize]
		}
		sm, prob := instance.smppConn.(TransmitterConn).Submit(&smpp.ShortMessage{
			Dst:      destination,
			Register: deliverySetting,
			Text:     pdutext.Raw(content),
		})
		if prob != nil {
			err = prob
			break
		}
		resp.Parts = append(resp.Parts, SendMessagePart{Id: sm.Resp().Fields()[pdufield.MessageID].String()})
	}
	return resp, err
}

func (instance *SimpleClient) observeClientConn(ch <-chan smpp.ConnStatus) {
	for status := range ch {
		var err error
		var eventType ClientEventType = ""
		switch status.Status() {
		case smpp.Connected:
			eventType = ClientConnBoundEventType
			break
		case smpp.Disconnected:
			eventType = ClientConnDisconnectEventType
			err = status.Error()
			break
		case smpp.BindFailed:
			err = status.Error()
			eventType = ClientConnBindErrorEventType
			break
		case smpp.ConnectionFailed:
			err = status.Error()
			eventType = ClientConnInterruptedEventType
			break
		default:
			if status.Error() != nil {
				err = status.Error()
				eventType = ClientConnErrorEventType
			}
			break
		}
		if instance.clientEventListener != nil && eventType != "" {
			instance.clientEventListener(ClientConnEvent{
				Err:  err,
				Type: eventType,
			})
		}
	}
}
