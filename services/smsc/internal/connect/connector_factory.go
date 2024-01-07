package connect

import (
	"github.com/delfimarime/hermes/services/smsc/internal/model"
	"github.com/delfimarime/hermes/services/smsc/pkg/common"
	"github.com/fiorix/go-smpp/smpp"
	"reflect"
	"time"
)

type ConnectorFactory struct {
	SmsEventListener SmsEventListener
}

func (instance *ConnectorFactory) New(config model.Smpp) *Connector {
	c := &Connector{
		id:                  config.Id,
		sourceAddress:       config.SourceAddr,
		smsEventListener:    instance.SmsEventListener,
		state:               StartupConnectorLifecycleState,
		submitsMessage:      config.Type != model.ReceiverType,
		awaitDeliveryReport: config.Type != model.TransmitterType,
	}
	c.client = instance.newSmppClientFrom(config, c.Listen)
	return c
}

func (instance *ConnectorFactory) newSmppClientFrom(config model.Smpp, f smpp.HandlerFunc) SmppClient {
	switch config.Type {
	case model.ReceiverType:
		return instance.newSmppClient(config, reflect.TypeOf(smpp.Receiver{}), f)
	case model.TransmitterType:
		return instance.newSmppClient(config, reflect.TypeOf(smpp.Transmitter{}), f)
	case model.TransceiverType:
		return instance.newSmppClient(config, reflect.TypeOf(smpp.Transceiver{}), f)
	default:
		return nil
	}
}

func (instance *ConnectorFactory) newSmppClient(config model.Smpp, dType reflect.Type, f smpp.HandlerFunc) SmppClient {
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
		if instance.SmsEventListener != nil {
			smppObject.Elem().FieldByName("Handler").Set(reflect.ValueOf(f))
		}
	}
	if isReceiver {
		return &SmppReceiverClientAdapter{target: smppObject.Interface().(*smpp.Receiver)}
	}
	return smppObject.Interface().(SmppClient)
}
