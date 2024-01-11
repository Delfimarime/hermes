package smpp

import (
	"github.com/delfimarime/hermes/services/smsc/internal/model"
	"github.com/delfimarime/hermes/services/smsc/pkg/common"
	"github.com/fiorix/go-smpp/smpp"
	"reflect"
	"time"
)

func newListenerConnector(config model.Smpp, cl ClientConnEventListener, f smpp.HandlerFunc) Client {
	target := newNoOpConnector(config, model.ReceiverType, cl, f)
	return &target
}

func newTransmitterConnector(config model.Smpp, cl ClientConnEventListener, f smpp.HandlerFunc) Client {
	awaitDeliveryReport := f != nil
	if config.Settings.Delivery != nil {
		awaitDeliveryReport = config.Settings.Delivery.AwaitReport
	}
	definitionType := model.TransmitterType
	if f != nil {
		definitionType = model.TransceiverType
	}
	client := newNoOpConnector(config, definitionType, cl, f)
	client.awaitDeliveryReport = awaitDeliveryReport
	return &client
}

func newNoOpConnector(config model.Smpp, definitionType string, cl ClientConnEventListener, f smpp.HandlerFunc) SimpleClient {
	return SimpleClient{
		clientEventListener: cl,
		smppType:            definitionType,
		smppConn:            newSmppClientFrom(config.Settings, definitionType, f),
	}
}

func newSmppClientFrom(config model.Settings, definitionType string, f smpp.HandlerFunc) smpp.ClientConn {
	switch definitionType {
	case model.ReceiverType, model.TransceiverType:
		var sample any
		if definitionType == model.ReceiverType {
			sample = smpp.Receiver{}
		} else {
			sample = smpp.Transceiver{}
		}
		return withHandlerFunction(newSmppClient(config, reflect.TypeOf(sample)), f)
	case model.TransmitterType:
		return newSmppClient(config, reflect.TypeOf(smpp.Transmitter{}))
	default:
		return nil
	}
}

func withHandlerFunction(client smpp.ClientConn, f smpp.HandlerFunc) smpp.ClientConn {
	reflect.ValueOf(client).Elem().FieldByName("Handler").Set(reflect.ValueOf(f))
	return client
}

func newSmppClient(config model.Settings, dType reflect.Type) smpp.ClientConn {
	var bindInterval *time.Duration
	var enquireLink *time.Duration
	var responseTimeout *time.Duration
	var enquireLinkTimeout *time.Duration
	smppObject := reflect.New(dType)
	smppObject.Elem().FieldByName("Addr").Set(reflect.ValueOf(config.Host.Address))
	smppObject.Elem().FieldByName("User").Set(reflect.ValueOf(config.Host.Username))
	smppObject.Elem().FieldByName("Passwd").Set(reflect.ValueOf(config.Host.Password))
	smppObject.Elem().FieldByName("SystemType").Set(reflect.ValueOf(config.ServiceType))
	if config.Bind != nil {
		t := common.MillisToDuration(config.Bind.Timeout)
		bindInterval = &t
	}
	if config.Enquire != nil {
		t := common.MillisToDuration(config.Enquire.Link)
		enquireLinkTimeout = &t
		v := common.MillisToDuration(config.Enquire.LinkTimeout)
		enquireLinkTimeout = &v
	}
	if config.Response != nil {
		t := common.MillisToDuration(config.Response.Timeout)
		responseTimeout = &t
	}

	if bindInterval != nil {
		smppObject.Elem().FieldByName("BindInterval").Set(reflect.ValueOf(*bindInterval))
	}
	if enquireLink != nil {
		smppObject.Elem().FieldByName("EnquireLink").Set(reflect.ValueOf(*enquireLink))
	}
	if enquireLinkTimeout != nil {
		smppObject.Elem().FieldByName("EnquireLinkTimeout").Set(reflect.ValueOf(*enquireLinkTimeout))
	}
	if responseTimeout != nil {
		smppObject.Elem().FieldByName("RespTimeout").Set(reflect.ValueOf(*responseTimeout))
	}
	isReceiver := dType == reflect.TypeOf(smpp.Receiver{})
	isTransceiver := dType == reflect.TypeOf(smpp.Transceiver{})
	if isReceiver || isTransceiver {
		if isReceiver && config.Merge != nil {
			smppObject.Elem().FieldByName("MergeInterval").
				Set(reflect.ValueOf(common.MillisToDuration(config.Merge.Interval)))
			smppObject.Elem().FieldByName("MergeCleanupInterval").
				Set(reflect.ValueOf(common.MillisToDuration(config.Merge.CleanupInterval)))
		}
	}
	return smppObject.Interface().(smpp.ClientConn)
}
