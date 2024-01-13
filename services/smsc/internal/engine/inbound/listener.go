package inbound

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
	conditionsRetrievedFromRepositoryF = "Creating inbound.SendSmsRequestPredicate for smpp.Connector" +
		" from %d conditions retrieved from SmppRepository"
)

type SmppSendSmsRequestListener struct {
	mutex          sync.Mutex
	smsRepository  sdk.SmsRepository
	smppRepository sdk.SmppRepository
	manager        smpp.ConnectorManager
	cache          map[string]SendSmsRequestPredicate
}

func (instance *SmppSendSmsRequestListener) ListenTo(req asyncapi.SendSmsRequest) (asyncapi.SendSmsResponse, error) {
	zap.L().Info("Listening to asyncapi.SendSmsRequest", zap.String("id", req.Id))
	fromDb, err := instance.smsRepository.FindById(req.Id)
	if err != nil {
		return asyncapi.SendSmsResponse{}, err
	}
	if fromDb == nil {
		fromDb = &model.Sms{
			Id:             "",
			To:             "",
			Type:           "",
			From:           "",
			Tags:           nil,
			ListenedAt:     time.Time{},
			TrackDelivery:  false,
			NumberOfParts:  0,
			SentParts:      nil,
			MaxSizePerPart: 0,
			Smpp:           asyncapi.ObjectId{},
		}
	}
	return instance.doListenTo(req)
}

func (instance *SmppSendSmsRequestListener) doListenTo(req asyncapi.SendSmsRequest) (asyncapi.SendSmsResponse, error) {
	sendSmsResponse := asyncapi.SendSmsResponse{
		Id: req.Id,
	}
	zap.L().Debug("Retrieving []smpp.Connector  in order to process asyncapi.SendSmsRequest",
		zap.String("id", req.Id))
	for _, each := range instance.manager.GetList() {
		zap.L().Debug(fmt.Sprintf("Fetching inbound.SendSmsRequestPredicate smpp.Connector[id=%s]",
			each.GetId()), zap.String("id", req.Id), zap.String(smpp.SmscIdAttribute, each.GetId()),
			zap.String(smpp.SmscAliasAttribute, each.GetAlias()),
		)
		predicate, hasValue := instance.cache[each.GetId()]
		if !hasValue {
			zap.L().Warn(fmt.Sprintf("Cannot fetch inbound.SendSmsRequestPredicate smpp.Connector[id=%s]",
				each.GetId()), zap.String("id", req.Id), zap.String(smpp.SmscIdAttribute, each.GetId()),
				zap.String(smpp.SmscAliasAttribute, each.GetAlias()),
			)
			continue
		}
		zap.L().Debug(fmt.Sprintf("Checking if smpp.Connector[id=%s] can sendMessage",
			each.GetId()), zap.String("id", req.Id), zap.String(smpp.SmscIdAttribute, each.GetId()),
			zap.String(smpp.SmscAliasAttribute, each.GetAlias()),
		)
		isCapableOfSendSms := predicate(req)
		if !isCapableOfSendSms {
			zap.L().Debug(fmt.Sprintf("smpp.Connector[id=%s] isn't capable of sendMessage",
				each.GetId()), zap.String("id", req.Id), zap.String(smpp.SmscIdAttribute, each.GetId()),
				zap.String(smpp.SmscAliasAttribute, each.GetAlias()),
			)
			continue
		}
		zap.L().Debug(fmt.Sprintf("Starting sendMessage on smpp.Connector[id=%s]",
			each.GetId()), zap.String("id", req.Id), zap.String(smpp.SmscIdAttribute, each.GetId()),
			zap.String(smpp.SmscAliasAttribute, each.GetAlias()),
		)
		sendMessageResponse, err := each.SendMessage(req.To, req.Content)
		if err != nil {
			var unavailableConnectorError *smpp.UnavailableConnectorError
			if errors.As(err, &unavailableConnectorError) {
				zap.L().Warn(fmt.Sprintf("Cannot sendMessage on smpp.Connector[id=%s] since it's not available",
					each.GetId()), zap.String("id", req.Id), zap.String(smpp.SmscIdAttribute, each.GetId()),
					zap.String(smpp.SmscAliasAttribute, each.GetAlias()),
					zap.Error(err),
				)
				continue
			}
			return asyncapi.SendSmsResponse{}, err
		}
		sendSmsResponse.Type = asyncapi.ShortSendSmsResponseType
		sendSmsResponse.Delivery = asyncapi.NotTrackingDeliveryStrategy
		sendSmsResponse.Smsc = asyncapi.ObjectId{Id: each.GetId()}
		if len(sendMessageResponse.Parts) > 1 {
			sendSmsResponse.Type = asyncapi.LongSendSmsResponseType
		}
		if each.IsTrackingDelivery() {
			sendSmsResponse.Delivery = asyncapi.TrackingDeliveryStrategy
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
	zap.L().Info("Setting up inbound.SendSmsRequestListener")
	for _, each := range instance.manager.GetList() {
		zap.L().Debug("Setting up inbound.SendSmsRequestPredicate for smpp.Connector",
			zap.String(smpp.SmscIdAttribute, each.GetId()),
			zap.String(smpp.SmscTypeAttribute, each.GetType()),
			zap.String(smpp.SmscAliasAttribute, each.GetAlias()),
		)
		seq, err := instance.smppRepository.GetConditionsFrom(each.GetId())
		if err != nil {
			zap.L().Warn("Cannot setup inbound.SendSmsRequestPredicate for smpp.Connector",
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
			zap.L().Warn("Cannot setup inbound.SendSmsRequestPredicate for smpp.Connector",
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
