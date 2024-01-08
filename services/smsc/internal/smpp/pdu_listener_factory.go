package smpp

import (
	"context"
	"fmt"
	"github.com/delfimarime/hermes/services/smsc/internal/model"
	"github.com/fiorix/go-smpp/smpp"
	"github.com/fiorix/go-smpp/smpp/pdu"
	"github.com/fiorix/go-smpp/smpp/pdu/pdufield"
	"github.com/fiorix/go-smpp/smpp/pdu/pdutlv"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
	"time"
)

type PduListenerFactory struct {
	meter                         metric.Meter
	smsEventListener              SmsEventListener
	pduCounterMetric              metric.Float64Counter
	unsupportedEventCounterMetric metric.Float64Counter
	consumedEventCounterMetric    metric.Float64Counter
}

func NewPduListenerFactory(meter metric.Meter, smsEventListener SmsEventListener) (*PduListenerFactory, error) {
	prefix := "hermes_smsc_pdu_listener"
	pduCounterMetric, err := meter.Float64Counter(fmt.Sprintf("%s_%s", prefix, "_number_of_pdu_events_received"),
		metric.WithDescription("The number of PDU's events received"),
	)
	if err != nil {
		return nil, err
	}
	unsupportedEventCounterMetric, err := meter.Float64Counter(fmt.Sprintf("%s_%s", prefix,
		"_number_of_unsupported_type_events"),
		metric.WithDescription("The number of PDU's event that cannot be consumed due to type"),
	)
	if err != nil {
		return nil, err
	}
	consumedEventCounterMetric, err := meter.Float64Counter(fmt.Sprintf("%s_%s", prefix,
		"_number_of_consumed_pdu_events"),
		metric.WithDescription("The number of PDU's event that have been consumed"),
	)
	if err != nil {
		return nil, err
	}
	return &PduListenerFactory{
		meter:                         meter,
		smsEventListener:              smsEventListener,
		pduCounterMetric:              pduCounterMetric,
		unsupportedEventCounterMetric: unsupportedEventCounterMetric,
		consumedEventCounterMetric:    consumedEventCounterMetric,
	}, nil
}

func (instance *PduListenerFactory) New(definition model.Smpp) smpp.HandlerFunc {
	return func(event pdu.Body) {
		instance.onPduEvent(definition.Id, definition.Alias, event)
	}
}

func (instance *PduListenerFactory) onPduEvent(id, alias string, event pdu.Body) {
	zap.L().Info("Pdu Event received",
		zap.String("sms_id", id),
		zap.String("sms_alias", alias),
	)
	attrs := []attribute.KeyValue{
		{
			Key:   attribute.Key("type"),
			Value: attribute.StringValue(event.Header().ID.String()),
		},
		{
			Key:   attribute.Key("status"),
			Value: attribute.Float64Value(float64(event.Header().Status)),
		},
		{
			Key:   attribute.Key("alias"),
			Value: attribute.StringValue(alias),
		},
	}
	for _, name := range event.FieldList() {
		zap.L().Debug("Adding Pdu.Field["+string(name)+"] into Pdu Counter metric",
			zap.String("sms_id", id),
			zap.String("sms_alias", alias),
		)
		attrs = append(attrs, attribute.KeyValue{
			Key:   attribute.Key(name),
			Value: attribute.StringValue(event.Header().ID.String()),
		})
	}
	zap.L().Debug("Incrementing Pdu event metric counter",
		zap.String("sms_id", id),
		zap.String("sms_alias", alias),
	)
	instance.pduCounterMetric.Add(context.TODO(), 1, metric.WithAttributes(attrs...))
	zap.L().Debug("Consuming Pdu event",
		zap.String("sms_id", id),
		zap.String("sms_alias", alias),
	)
	err := instance.handleEvent(id, alias, event)
	if err == nil {
		zap.L().Debug("Pdu event consumed",
			zap.String("sms_id", id),
			zap.String("sms_alias", alias),
		)
		return
	}
	zap.L().Error("Failure during Pdu event consumption",
		zap.String("sms_id", id),
		zap.String("sms_alias", alias),
	)
}

func (instance *PduListenerFactory) handleEvent(id, alias string, event pdu.Body) error {
	var err error
	switch event.Header().ID {
	case pdu.DeliverSMID:
		err = instance.onDeliverySM(id, alias, event)
		break
	default:
		return instance.onUnsupportedPduEvent(id, alias, event)
	}
	if err != nil {
		return err
	}
	instance.consumedEventCounterMetric.Add(context.TODO(), 1, metric.WithAttributes(
		attribute.KeyValue{
			Key:   "sms_id",
			Value: attribute.StringValue(id),
		},
		attribute.KeyValue{
			Key:   "sms_id",
			Value: attribute.StringValue(id),
		},
		attribute.KeyValue{
			Key:   "sms_alias",
			Value: attribute.StringValue(alias),
		},
		attribute.KeyValue{
			Key:   "operation",
			Value: attribute.StringValue(event.Header().ID.String()),
		},
		attribute.KeyValue{
			Key:   "status",
			Value: attribute.Float64Value(float64(event.Header().Status)),
		},
	))
	return nil
}

func (instance *PduListenerFactory) onUnsupportedPduEvent(id, alias string, event pdu.Body) error {
	instance.unsupportedEventCounterMetric.Add(context.TODO(), 1, metric.WithAttributes(
		attribute.KeyValue{
			Key:   "sms_id",
			Value: attribute.StringValue(id),
		},
		attribute.KeyValue{
			Key:   "sms_id",
			Value: attribute.StringValue(id),
		},
		attribute.KeyValue{
			Key:   "sms_alias",
			Value: attribute.StringValue(alias),
		},
		attribute.KeyValue{
			Key:   "operation",
			Value: attribute.StringValue(event.Header().ID.String()),
		},
		attribute.KeyValue{
			Key:   "status",
			Value: attribute.Float64Value(float64(event.Header().Status)),
		},
	))
	zap.L().Warn("Cannot consume Pdu event",
		zap.String("sms_id", id),
		zap.String("sms_alias", alias),
		zap.String("operation", event.Header().ID.String()),
		zap.Float64("status", float64(event.Header().Status)),
	)
	return nil
}

func (instance *PduListenerFactory) onDeliverySM(id, alias string, event pdu.Body) error {
	receivedAt := time.Now()
	esmClass := event.Fields()[pdufield.ESMClass].Raw().(uint8)
	switch esmClass {
	case 0x00:
		zap.L().Info("Pdu event identified recognized as SmsRequest",
			zap.String("sms_id", id),
			zap.String("sms_alias", alias),
			zap.String("esm_class", "0x00"),
			zap.String("operation", event.Header().ID.String()),
		)
		zap.L().Debug("Converting PduEvent into PduObject",
			zap.String("sms_id", id),
			zap.String("sms_alias", alias),
		)
		fromRequest := getPduObject(event)
		zap.L().Debug("Converting PduObject into ReceivedSmsRequest",
			zap.String("sms_id", id),
			zap.String("sms_alias", alias),
		)
		r := ReceivedSmsRequest{
			ReceivedAt: receivedAt,
			SmscId:     id,
			Message:    fromRequest.content,
		}
		if fromRequest.messageId != "" {
			r.Id = fromRequest.messageId
		}
		if fromRequest.sourceAddr != "" {
			r.From = fromRequest.sourceAddr
		}
		zap.L().Debug("Submit ReceivedSmsRequest into listener",
			zap.String("sms_id", id),
			zap.String("sms_alias", alias),
		)
		instance.smsEventListener.OnSmsRequest(r)
		break
	case 0x04:
		zap.L().Info("Pdu event identified recognized as SmsDeliveryRequest",
			zap.String("sms_id", id),
			zap.String("sms_alias", alias),
			zap.String("esm_class", "0x04"),
			zap.String("operation", event.Header().ID.String()),
		)
		zap.L().Debug("PduEvent converted into SmsDeliveryRequest",
			zap.String("sms_id", id),
			zap.String("sms_alias", alias),
		)
		r := SmsDeliveryRequest{
			SmscId:     id,
			ReceivedAt: receivedAt,
			Status:     int(event.Header().Status),
			Id:         event.TLVFields()[pdutlv.TagReceiptedMessageID].String(),
		}
		zap.L().Debug("Submit SmsDeliveryRequest into listener",
			zap.String("sms_id", id),
			zap.String("sms_alias", alias),
		)
		instance.smsEventListener.OnSmsDelivered(r)
		break
	}
	return nil
}
