package outbound

import (
	"errors"
	"fmt"
	"github.com/delfimarime/hermes/services/smsc/internal/model"
	"github.com/delfimarime/hermes/services/smsc/internal/repository/sdk"
	"github.com/delfimarime/hermes/services/smsc/internal/smpp"
	"github.com/delfimarime/hermes/services/smsc/pkg/asyncapi"
	"go.uber.org/zap"
	"sync"
	"time"
)

const (
	requestIdAttribute = "send_sms_request_id"
)

const (
	conditionsRetrievedFromRepositoryF = "Creating outbound.SendSmsRequestPredicate for smpp.Connector" +
		" from %d conditions retrieved from SmppRepository"
)

type SendSmsResponse struct {
	Id string
	asyncapi.SendSmsResponse
}

type SmppSendSmsRequestListener struct {
	mutex          sync.Mutex
	smsRepository  sdk.SmsRepository
	smppRepository sdk.SmppRepository
	manager        smpp.ConnectorManager
	cache          map[string]SendSmsRequestPredicate
}

func (instance *SmppSendSmsRequestListener) ListenTo(req asyncapi.SendSmsRequest) (asyncapi.SendSmsResponse, error) {
	instant := time.Now()
	zap.L().Info("Listening to asyncapi.SendSmsRequest", zap.String(requestIdAttribute, req.Id))
	zap.L().Debug("Fetching model.Sms from Repository", zap.String(requestIdAttribute, req.Id))
	fromDb, err := instance.smsRepository.FindById(req.Id)
	if err != nil {
		zap.L().Error("Cannot fetch model.Sms from Repository", zap.String(requestIdAttribute, req.Id), zap.Error(err))
		return asyncapi.SendSmsResponse{}, err
	}
	if fromDb != nil {
		zap.L().Debug("Retrieved model.Sms from Repository. No action will be taken",
			zap.String(requestIdAttribute, req.Id))
		deliveryStrategy := asyncapi.NotTrackingDeliveryStrategy
		if fromDb.TrackDelivery {
			deliveryStrategy = asyncapi.TrackingDeliveryStrategy
		}
		return asyncapi.SendSmsResponse{
			Id:       fromDb.Id,
			Smsc:     fromDb.Smpp,
			Delivery: deliveryStrategy,
		}, nil
	}
	resp, err := instance.doListenTo(req)
	if err != nil {
		zap.L().Error("Cannot consume asyncapi.SendSmsRequest due to an error",
			zap.String(requestIdAttribute, req.Id),
			zap.Error(err))
		return asyncapi.SendSmsResponse{}, err
	}
	zap.L().Debug("Persisting model.Sms for asyncapi.SendSmsRequest",
		zap.String(requestIdAttribute, req.Id),
		zap.String(smpp.SmsIdAttribute, resp.Id),
		zap.String(smpp.SmscIdAttribute, resp.Smsc.Id),
	)
	err = instance.smsRepository.Save(&model.Sms{
		Id:         req.Id,
		TrackId:    resp.Id,
		ListenedAt: instant,
		Smpp:       resp.Smsc,
	})
	if err != nil {
		return asyncapi.SendSmsResponse{}, err
	}
	return resp.SendSmsResponse, nil
}

func (instance *SmppSendSmsRequestListener) doListenTo(req asyncapi.SendSmsRequest) (SendSmsResponse, error) {
	sendSmsResponse := SendSmsResponse{
		SendSmsResponse: asyncapi.SendSmsResponse{
			Id: req.Id,
		},
	}
	zap.L().Debug("Retrieving []smpp.Connector  in order to process asyncapi.SendSmsRequest",
		zap.String(requestIdAttribute, req.Id))
	for _, each := range instance.manager.GetList() {
		zap.L().Debug(fmt.Sprintf("Fetching outbound.SendSmsRequestPredicate smpp.Connector[id=%s]",
			each.GetId()), zap.String(requestIdAttribute, req.Id), zap.String(smpp.SmscIdAttribute, each.GetId()),
			zap.String(smpp.SmscAliasAttribute, each.GetAlias()),
		)
		predicate, hasValue := instance.cache[each.GetId()]
		if !hasValue {
			zap.L().Warn(fmt.Sprintf("Cannot fetch outbound.SendSmsRequestPredicate smpp.Connector[id=%s]",
				each.GetId()), zap.String(requestIdAttribute, req.Id), zap.String(smpp.SmscIdAttribute, each.GetId()),
				zap.String(smpp.SmscAliasAttribute, each.GetAlias()),
			)
			continue
		}
		zap.L().Debug(fmt.Sprintf("Checking if smpp.Connector[id=%s] can sendMessage",
			each.GetId()), zap.String(requestIdAttribute, req.Id), zap.String(smpp.SmscIdAttribute, each.GetId()),
			zap.String(smpp.SmscAliasAttribute, each.GetAlias()),
		)
		isCapableOfSendSms := predicate(req)
		if !isCapableOfSendSms {
			zap.L().Debug(fmt.Sprintf("smpp.Connector[id=%s] isn't capable of sendMessage",
				each.GetId()), zap.String(requestIdAttribute, req.Id), zap.String(smpp.SmscIdAttribute, each.GetId()),
				zap.String(smpp.SmscAliasAttribute, each.GetAlias()),
			)
			continue
		}
		zap.L().Debug(fmt.Sprintf("Starting sendMessage on smpp.Connector[id=%s]",
			each.GetId()), zap.String(requestIdAttribute, req.Id), zap.String(smpp.SmscIdAttribute, each.GetId()),
			zap.String(smpp.SmscAliasAttribute, each.GetAlias()),
		)
		sendMessageResponse, err := each.SendMessage(req.To, req.Content)
		if err != nil {
			var unavailableConnectorError *smpp.UnavailableConnectorError
			if errors.As(err, &unavailableConnectorError) {
				zap.L().Warn(fmt.Sprintf("Cannot sendMessage on smpp.Connector[id=%s] since it's not available",
					each.GetId()), zap.String(requestIdAttribute, req.Id), zap.String(smpp.SmscIdAttribute, each.GetId()),
					zap.String(smpp.SmscAliasAttribute, each.GetAlias()),
					zap.Error(err),
				)
				continue
			}
			return SendSmsResponse{}, err
		}
		sendSmsResponse.Id = sendMessageResponse.Id
		sendSmsResponse.Delivery = asyncapi.NotTrackingDeliveryStrategy
		sendSmsResponse.Smsc = &asyncapi.ObjectId{Id: each.GetId()}
		if each.IsTrackingDelivery() {
			sendSmsResponse.Delivery = asyncapi.TrackingDeliveryStrategy
		}
	}

	if sendSmsResponse.SendSmsResponse.Smsc == nil {
		sendSmsResponse.SendSmsResponse.Problem = &asyncapi.Problem{
			Title:  "Cannot send async.SendSmsRequest",
			Type:   "/smsc/sendSmsRequest/unprocessable-entity",
			Detail: "Couldn't determine smpp.Connector capable of sending asyncapi.SendSmsRequest",
		}
	}

	return sendSmsResponse, nil
}

func (instance *SmppSendSmsRequestListener) Close() error {
	instance.mutex.Lock()
	defer instance.mutex.Unlock()
	if instance.cache != nil {
		instance.cache = nil
	}
	return nil
}

func (instance *SmppSendSmsRequestListener) AfterPropertiesSet() error {
	instance.mutex.Lock()
	defer instance.mutex.Unlock()
	if instance.cache == nil {
		instance.cache = make(map[string]SendSmsRequestPredicate)
	}
	zap.L().Info("Setting up outbound.SendSmsRequestListener")
	for _, each := range instance.manager.GetList() {
		zap.L().Debug("Setting up outbound.SendSmsRequestPredicate for smpp.Connector",
			zap.String(smpp.SmscIdAttribute, each.GetId()),
			zap.String(smpp.SmscTypeAttribute, each.GetType()),
			zap.String(smpp.SmscAliasAttribute, each.GetAlias()),
		)
		seq, err := instance.smppRepository.GetConditionsFrom(each.GetId())
		if err != nil {
			zap.L().Warn("Cannot setup outbound.SendSmsRequestPredicate for smpp.Connector",
				zap.String(smpp.SmscIdAttribute, each.GetId()),
				zap.String(smpp.SmscTypeAttribute, each.GetType()),
				zap.String(smpp.SmscAliasAttribute, each.GetAlias()),
				zap.Error(err),
			)
			continue
		}
		zap.L().Debug(fmt.
			Sprintf(conditionsRetrievedFromRepositoryF, len(seq)),
			zap.String(smpp.SmscIdAttribute, each.GetId()),
			zap.String(smpp.SmscTypeAttribute, each.GetType()),
			zap.String(smpp.SmscAliasAttribute, each.GetAlias()),
		)
		predicate, err := toPredicate(seq...)
		if err != nil {
			zap.L().Warn("Cannot setup outbound.SendSmsRequestPredicate for smpp.Connector",
				zap.String(smpp.SmscIdAttribute, each.GetId()),
				zap.String(smpp.SmscTypeAttribute, each.GetType()),
				zap.String(smpp.SmscAliasAttribute, each.GetAlias()),
				zap.Error(err),
			)
			continue
		}
		instance.cache[each.GetId()] = predicate
	}
	return nil
}
