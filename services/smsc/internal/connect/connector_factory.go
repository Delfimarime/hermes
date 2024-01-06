package connect

import (
	"github.com/delfimarime/hermes/services/smsc/internal/model"
	"github.com/fiorix/go-smpp/smpp"
	"reflect"
	"time"
)

type ConnectorFactory struct {
	SmsEventListener SmsEventListener
}

func (instance *ConnectorFactory) New(config model.Smpp) *Connector {
	c := &Connector{
		sourceAddress: config.SourceAddr,
		client:        instance.newSmppClientFrom(config),
	}
	return c
}

func (instance *ConnectorFactory) newSmppClientFrom(config model.Smpp) SmppClient {
	switch config.Type {
	case model.TransmitterType:
		return instance.newSmppClient(config, reflect.TypeOf(smpp.Transmitter{}))
	case model.TransceiverType:
		return instance.newSmppClient(config, reflect.TypeOf(smpp.Transceiver{}))
	case model.ReceiverType:
		return instance.newSmppClient(config, reflect.TypeOf(smpp.Receiver{}))
	default:
		return nil
	}
}

func (instance *ConnectorFactory) newSmppClient(config model.Smpp, dType reflect.Type) SmppClient {
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
		t := time.Duration(config.Bind.Timeout * int64(time.Nanosecond))
		bindInterval = &t
	}
	if config.Enquire != nil {
		t := time.Duration(config.Enquire.Link * int64(time.Nanosecond))
		enquireLinkTimeout = &t
		v := time.Duration(config.Enquire.LinkTimeout * int64(time.Nanosecond))
		enquireLinkTimeout = &v
	}
	if config.Response != nil {
		t := time.Duration(config.Response.Timeout * int64(time.Nanosecond))
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
				Set(reflect.ValueOf(time.Duration(config.Merge.Interval * int64(time.Nanosecond))))
			smppObject.Elem().FieldByName("MergeCleanupInterval").
				Set(reflect.ValueOf(time.Duration(config.Merge.CleanupInterval * int64(time.Nanosecond))))
		}
		if instance.SmsEventListener != nil {
			smppObject.Elem().FieldByName("Handler").
				Set(reflect.ValueOf(NewPduListener(config.Id, instance.SmsEventListener)))
		}
	}
	if isReceiver {
		return &SmppReceiverClientAdapter{receiver: smppObject.Interface().(*smpp.Receiver)}
	}
	return smppObject.Interface().(SmppClient)
}
