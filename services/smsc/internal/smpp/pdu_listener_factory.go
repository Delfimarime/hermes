package smpp

import (
	"context"
	"fmt"
	"github.com/delfimarime/hermes/services/smsc/internal/model"
	"github.com/fiorix/go-smpp/smpp"
	"github.com/fiorix/go-smpp/smpp/pdu"
	"github.com/fiorix/go-smpp/smpp/pdu/pdufield"
	"github.com/fiorix/go-smpp/smpp/pdu/pdutlv"
	"github.com/google/uuid"
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
		instance.onPduEvent(definition, event)
	}
}

func (instance *PduListenerFactory) onPduEvent(definition model.Smpp, event pdu.Body) {
	zap.L().Info("Pdu Event received",
		zap.String(smscIdAttribute, definition.Id),
		zap.String(smscAliasAttribute, definition.Alias),
	)
	attrs := []attribute.KeyValue{
		attribute.String(pduHeaderIdAttribute, event.Header().ID.String()),
		attribute.Float64(pduHeaderStatusAttribute, float64(event.Header().Status)),
		attribute.String(smscAliasAttribute, definition.Alias),
	}
	for _, name := range event.FieldList() {
		zap.L().Debug("Adding Pdu.Field["+string(name)+"] into Pdu Counter metric",
			zap.String(smscIdAttribute, definition.Id),
			zap.String(smscAliasAttribute, definition.Alias),
		)
		attrs = append(attrs, attribute.String(fmt.
			Sprintf("%s_%s", pduFieldAttribute, name), event.Fields()[name].String()))
	}
	zap.L().Debug("Incrementing Pdu event metric counter",
		zap.String(smscIdAttribute, definition.Id),
		zap.String(smscAliasAttribute, definition.Alias),
	)
	instance.pduCounterMetric.Add(context.TODO(), 1, metric.WithAttributes(attrs...))
	zap.L().Debug("Consuming Pdu event",
		zap.String(smscIdAttribute, definition.Id),
		zap.String(smscAliasAttribute, definition.Alias),
	)
	err := instance.handleEvent(definition, event)
	if err == nil {
		zap.L().Debug("Pdu event consumed",
			zap.String(smscIdAttribute, definition.Id),
			zap.String(smscAliasAttribute, definition.Alias),
		)
		return
	}
	zap.L().Error("Failure during Pdu event consumption",
		zap.String(smscIdAttribute, definition.Id),
		zap.String(smscAliasAttribute, definition.Alias),
	)
}

func (instance *PduListenerFactory) handleEvent(definition model.Smpp, event pdu.Body) error {
	var err error
	switch event.Header().ID {
	case pdu.DeliverSMID:
		err = instance.onDeliverySM(definition, event)
		break
	default:
		return instance.onUnsupportedPduEvent(definition, event)
	}
	if err != nil {
		return err
	}
	instance.consumedEventCounterMetric.Add(context.TODO(), 1, metric.WithAttributes(
		attribute.String(smscIdAttribute, definition.Id),
		attribute.String(smscAliasAttribute, definition.Alias),
		attribute.String(pduHeaderIdAttribute, event.Header().ID.String()),
		attribute.Float64(pduHeaderStatusAttribute, float64(event.Header().Status)),
	))
	return nil
}

func (instance *PduListenerFactory) onUnsupportedPduEvent(definition model.Smpp, event pdu.Body) error {
	instance.unsupportedEventCounterMetric.Add(context.TODO(), 1, metric.WithAttributes(
		attribute.String(smscIdAttribute, definition.Id),
		attribute.String(smscAliasAttribute, definition.Alias),
		attribute.String(pduHeaderIdAttribute, event.Header().ID.String()),
		attribute.Float64(pduHeaderStatusAttribute, float64(event.Header().Status)),
	))
	zap.L().Warn("Cannot consume Pdu event",
		zap.String(smscIdAttribute, definition.Id),
		zap.String(smscAliasAttribute, definition.Alias),
		zap.String(pduHeaderIdAttribute, event.Header().ID.String()),
		zap.Float64(pduHeaderStatusAttribute, float64(event.Header().Status)),
	)
	return nil
}

func (instance *PduListenerFactory) onDeliverySM(definition model.Smpp, event pdu.Body) error {
	receivedAt := time.Now()
	esmClass := event.Fields()[pdufield.ESMClass].Raw().(uint8)
	switch esmClass {
	case 0x00:
		zap.L().Info("Pdu event identified recognized as SmsRequest",
			zap.String(smscIdAttribute, definition.Id),
			zap.String(smscAliasAttribute, definition.Alias),
			zap.String(pduFieldEsmClassAttribute, "0x00"),
			zap.String(pduHeaderIdAttribute, event.Header().ID.String()),
		)
		zap.L().Debug("Converting PduEvent into PduObject",
			zap.String(smscIdAttribute, definition.Id),
			zap.String(smscAliasAttribute, definition.Alias),
		)
		fromRequest := getPduObject(event)
		zap.L().Debug("Converting PduObject into ReceivedSmsRequest",
			zap.String(smscIdAttribute, definition.Id),
			zap.String(smscAliasAttribute, definition.Alias),
		)
		r := ReceivedSmsRequest{
			SmscId:     definition.Id,
			ReceivedAt: receivedAt,
			Id:         uuid.New().String(),
			Message:    fromRequest.content,
		}
		if fromRequest.messageId != "" {
			r.Id = fromRequest.messageId
		}
		if fromRequest.sourceAddr != "" {
			r.From = fromRequest.sourceAddr
		}
		zap.L().Debug("Submit ReceivedSmsRequest into listener",
			zap.String(smscIdAttribute, definition.Id),
			zap.String(smscAliasAttribute, definition.Alias),
		)
		instance.smsEventListener.OnSmsRequest(r)
		break
	case 0x04:
		zap.L().Info("Pdu event identified recognized as SmsDeliveryRequest",
			zap.String(smscIdAttribute, definition.Id),
			zap.String(smscAliasAttribute, definition.Alias),
			zap.String(pduFieldEsmClassAttribute, "0x04"),
			zap.String(pduHeaderIdAttribute, event.Header().ID.String()),
		)
		zap.L().Debug("PduEvent converted into SmsDeliveryRequest",
			zap.String(smscIdAttribute, definition.Id),
			zap.String(smscAliasAttribute, definition.Alias),
		)
		r := SmsDeliveryRequest{
			ReceivedAt: receivedAt,
			SmscId:     definition.Id,
			Status:     int(event.Header().Status),
			Id:         event.TLVFields()[pdutlv.TagReceiptedMessageID].String(),
		}
		zap.L().Debug("Submit SmsDeliveryRequest into listener",
			zap.String(smscIdAttribute, definition.Id),
			zap.String(smscAliasAttribute, definition.Alias),
		)
		instance.smsEventListener.OnSmsDelivered(r)
		break
	}
	return nil
}
