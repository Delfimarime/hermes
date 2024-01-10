package smpp

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"sync"
)

type SimpleConnector struct {
	mutex                       sync.Mutex
	state                       string
	alias                       string
	connector                   Client
	sendMessageCountMetric      metric.Float64Counter // number_of_messages_sent
	sendMessageErrorCountMetric metric.Float64Counter // number_of_messages_send_that_failed
	refreshCountMetric          metric.Float64Counter // number_of_refreshes
	refreshErrorCountMetric     metric.Float64Counter // number_of_refreshes_that_failed
	bindCountMetric             metric.Float64Counter // number_of_bindings
	bindErrorCountMetric        metric.Float64Counter // number_of_bindings_that_failed
}

func (instance *SimpleConnector) GetState() string {
	return instance.state
}

func (instance *SimpleConnector) GetType() string {
	return instance.connector.GetType()
}

func (instance *SimpleConnector) GetId() string {
	return instance.connector.GetId()
}

func (instance *SimpleConnector) GetAlias() string {
	return instance.alias
}

func (instance *SimpleConnector) SendMessage(destination, message string) (SendMessageResponse, error) {
	resp, err := instance.connector.(Client).SendMessage(destination, message)
	defer func() {
		instance.increaseMetricCounter(instance.sendMessageCountMetric, instance.sendMessageErrorCountMetric, err)
	}()
	return resp, err
}

func (instance *SimpleConnector) doBind() error {
	instance.mutex.Lock()
	defer instance.mutex.Unlock()
	err := instance.connector.Bind()
	defer func() {
		instance.increaseMetricCounter(instance.bindCountMetric, instance.bindErrorCountMetric, err)
	}()
	return err
}

func (instance *SimpleConnector) doClose() error {
	instance.mutex.Lock()
	defer instance.mutex.Unlock()
	return instance.connector.Close()
}

func (instance *SimpleConnector) doRefresh() error {
	instance.mutex.Lock()
	defer instance.mutex.Unlock()
	err := instance.connector.Refresh()
	defer func() {
		instance.increaseMetricCounter(instance.refreshCountMetric, instance.refreshErrorCountMetric, err)
	}()
	return err
}

func (instance *SimpleConnector) increaseMetricCounter(onSuccess, onFailure metric.Float64Counter, err error) {
	if err != nil {
		onFailure.Add(context.TODO(), 1, metric.WithAttributes(
			attribute.String(smscIdAttribute, instance.GetId()),
			attribute.String(smscAliasAttribute, instance.GetAlias()),
			attribute.String(errorAttribute, truncate(err.Error(), 15)),
		))
	} else {
		onSuccess.Add(context.TODO(), 1, metric.WithAttributes(
			attribute.String(smscIdAttribute, instance.GetId()),
			attribute.String(smscAliasAttribute, instance.GetAlias()),
		))
	}
}
