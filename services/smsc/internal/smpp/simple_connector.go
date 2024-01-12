package smpp

import (
	"context"
	"errors"
	"fmt"
	"github.com/fiorix/go-smpp/smpp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
	"sync"
)

type SimpleConnector struct {
	state                       State
	alias                       string
	client                      Client
	mutex                       sync.Mutex
	sendMessageCountMetric      metric.Float64Counter // number_of_messages_sent
	sendMessageErrorCountMetric metric.Float64Counter // number_of_messages_send_that_failed
	refreshCountMetric          metric.Float64Counter // number_of_refreshes
	refreshErrorCountMetric     metric.Float64Counter // number_of_refreshes_that_failed
	bindCountMetric             metric.Float64Counter // number_of_bindings
	bindErrorCountMetric        metric.Float64Counter // number_of_bindings_that_failed
}

func (instance *SimpleConnector) GetState() State {
	return instance.state
}

func (instance *SimpleConnector) GetType() string {
	return instance.client.GetType()
}

func (instance *SimpleConnector) GetId() string {
	return instance.client.GetId()
}

func (instance *SimpleConnector) GetAlias() string {
	return instance.alias
}

func (instance *SimpleConnector) SendMessage(destination, message string) (SendMessageResponse, error) {
	if instance.state != ReadyConnectorLifecycleState {
		zap.L().Warn(fmt.Sprintf("Cannot Send message from smpp.connector[id=%s] because of state",
			instance.GetId()), zap.String(SmscIdAttribute, instance.GetId()),
			zap.String(SmscAliasAttribute, instance.GetAlias()),
			zap.String(smscStateAttribute, instance.GetState().string()),
		)
		return SendMessageResponse{}, UnavailableConnectorError{state: instance.GetState().string()}
	}
	zap.L().Debug(fmt.Sprintf("Sending message from smpp.connector[id=%s] to specified msisdn",
		instance.GetId()), zap.String(SmscIdAttribute, instance.GetId()),
		zap.String(SmscAliasAttribute, instance.GetAlias()),
		zap.String(smscStateAttribute, instance.GetState().string()),
		zap.String(smsDestinationAttribute, destination),
	)
	resp, err := instance.client.(Client).SendMessage(destination, message)
	if err == nil {
		zap.L().Debug(fmt.Sprintf("Message sent from smpp.connector[id=%s] to specified msisdn",
			instance.GetId()), zap.String(SmscIdAttribute, instance.GetId()),
			zap.String(SmscAliasAttribute, instance.GetAlias()),
			zap.String(smscStateAttribute, instance.GetState().string()),
			zap.String(smsIdAttribute, resp.Id),
			zap.String(smsDestinationAttribute, destination),
		)
	}
	isClientConnectionError := errors.Is(err, smpp.ErrNotConnected) || errors.Is(err, smpp.ErrNotBound)
	if isClientConnectionError {
		return SendMessageResponse{}, UnavailableConnectorError{state: instance.GetState().string(), causedBy: err}
	}
	instance.increaseMetricCounter(instance.sendMessageCountMetric, instance.sendMessageErrorCountMetric, err)
	return resp, err
}

func (instance *SimpleConnector) setState(state State) {
	zap.L().Debug(fmt.Sprintf("Changing smpp.connector[id=%s] state to %s",
		instance.GetId(), instance.GetState().string()), zap.String(SmscIdAttribute, instance.GetId()),
		zap.String(SmscAliasAttribute, instance.GetAlias()), zap.String(smscStateAttribute, state.string()),
	)
	instance.mutex.Lock()
	defer instance.mutex.Unlock()
	if instance.state == ClosedConnectorLifecycleState {
		zap.L().Warn(fmt.Sprintf("Cannot change  smpp.connector[id=%s] state to %s",
			instance.GetId(), instance.GetState().string()), zap.String(SmscIdAttribute, instance.GetId()),
			zap.String(SmscAliasAttribute, instance.GetAlias()), zap.String(smscStateAttribute, state.string()),
		)
		return
	}
	instance.state = state
}

func (instance *SimpleConnector) doClose() error {
	instance.mutex.Lock()
	defer instance.mutex.Unlock()
	instance.state = ClosedConnectorLifecycleState
	return instance.client.Close()
}

func (instance *SimpleConnector) increaseMetricCounter(onSuccess, onFailure metric.Float64Counter, err error) {
	if err != nil {
		onFailure.Add(context.TODO(), 1, metric.WithAttributes(
			attribute.String(SmscIdAttribute, instance.GetId()),
			attribute.String(SmscAliasAttribute, instance.GetAlias()),
			attribute.String(errorAttribute, truncate(err.Error(), 15)),
		))
	} else {
		onSuccess.Add(context.TODO(), 1, metric.WithAttributes(
			attribute.String(SmscIdAttribute, instance.GetId()),
			attribute.String(SmscAliasAttribute, instance.GetAlias()),
		))
	}
}
