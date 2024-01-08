package smpp

import (
	"errors"
	"go.opentelemetry.io/otel/metric"
)

type ManagedConnector struct {
	canSendSms                  bool
	state                       string
	connector                   Connector
	SendMessageCountMetric      metric.Float64Counter // number_of_messages_sent
	SendMessageErrorCountMetric metric.Float64Counter // number_of_messages_send_that_failed
	RefreshCountMetric          metric.Float64Counter // number_of_refreshes
	RefreshErrorCountMetric     metric.Float64Counter // number_of_refreshes_that_failed
	BindCountMetric             metric.Float64Counter // number_of_bindings
	BindErrorCountMetric        metric.Float64Counter // number_of_bindings_that_failed
}

func (instance *ManagedConnector) GetId() string {
	return instance.connector.GetId()
}

func (instance *ManagedConnector) Bind() error {
	return instance.connector.Bind()
}

func (instance *ManagedConnector) Close() error {
	return instance.connector.Close()
}

func (instance *ManagedConnector) Refresh() error {
	return instance.connector.Refresh()
}

func (instance *ManagedConnector) SendMessage(destination, message string) (SendMessageResponse, error) {
	if !instance.canSendSms {
		return SendMessageResponse{}, errors.New("operation not allowed")
	}
	return instance.connector.(TransmitterConnector).SendMessage(destination, message)
}

func (instance *ManagedConnector) GetState() string {
	return instance.state
}
