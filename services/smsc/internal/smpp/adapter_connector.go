package smpp

import (
	"go.opentelemetry.io/otel/metric"
)

type AdapterConnector struct {
	state                       string
	connector                   Client
	SendMessageCountMetric      metric.Float64Counter // number_of_messages_sent
	SendMessageErrorCountMetric metric.Float64Counter // number_of_messages_send_that_failed
	RefreshCountMetric          metric.Float64Counter // number_of_refreshes
	RefreshErrorCountMetric     metric.Float64Counter // number_of_refreshes_that_failed
	BindCountMetric             metric.Float64Counter // number_of_bindings
	BindErrorCountMetric        metric.Float64Counter // number_of_bindings_that_failed
}

func (instance *AdapterConnector) GetState() string {
	return instance.state
}

func (instance *AdapterConnector) GetType() string {
	return instance.connector.GetType()
}

func (instance *AdapterConnector) GetId() string {
	return instance.connector.GetId()
}

func (instance *AdapterConnector) SendMessage(destination, message string) (SendMessageResponse, error) {
	return instance.connector.(Client).SendMessage(destination, message)
}

func (instance *AdapterConnector) doBind() error {
	return instance.connector.Bind()
}

func (instance *AdapterConnector) doClose() error {
	return instance.connector.Close()
}

func (instance *AdapterConnector) doRefresh() error {
	return instance.connector.Refresh()
}
