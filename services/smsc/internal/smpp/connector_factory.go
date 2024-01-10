package smpp

import (
	"github.com/delfimarime/hermes/services/smsc/internal/model"
	"github.com/delfimarime/hermes/services/smsc/pkg/common"
	"github.com/fiorix/go-smpp/smpp"
	"reflect"
	"time"
)

type SimpleConnectorFactory struct {
}

func (instance *SimpleConnectorFactory) NewListenerConnector(
	config model.Smpp, f smpp.HandlerFunc) Client {
	target := instance.newNoOpConnector(config, model.ReceiverType, f)
	return &target
}

func (instance *SimpleConnectorFactory) NewTransmitterConnector(
	config model.Smpp, f smpp.HandlerFunc) Client {
	awaitDeliveryReport := f != nil
	if config.Settings.Delivery != nil {
		awaitDeliveryReport = config.Settings.Delivery.AwaitReport
	}
	definitionType := model.TransmitterType
	if f != nil {
		definitionType = model.TransceiverType
	}
	client := instance.newNoOpConnector(config, definitionType, f)
	client.awaitDeliveryReport = awaitDeliveryReport
	return &client
}

func (instance *SimpleConnectorFactory) newNoOpConnector(
	config model.Smpp, definitionType string, f smpp.HandlerFunc) SimpleClient {
	return SimpleClient{
		smppType:   definitionType,
		smppClient: instance.newSmppClientFrom(config.Settings, definitionType, f),
	}
}

func (instance *SimpleConnectorFactory) newSmppClientFrom(config model.Settings, definitionType string, f smpp.HandlerFunc) smpp.ClientConn {
	switch definitionType {
	case model.ReceiverType, model.TransceiverType:
		var sample any
		if definitionType == model.ReceiverType {
			sample = smpp.Receiver{}
		} else {
			sample = smpp.Transceiver{}
		}
		return instance.withHandlerFunction(instance.newSmppClient(config, reflect.TypeOf(sample)), f)
	case model.TransmitterType:
		return instance.newSmppClient(config, reflect.TypeOf(smpp.Transmitter{}))
	default:
		return nil
	}
}

func (instance *SimpleConnectorFactory) withHandlerFunction(client smpp.ClientConn, f smpp.HandlerFunc) smpp.ClientConn {
	reflect.ValueOf(client).Elem().FieldByName("Handler").Set(reflect.ValueOf(f))
	return client
}

func (instance *SimpleConnectorFactory) newSmppClient(config model.Settings, dType reflect.Type) smpp.ClientConn {
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
