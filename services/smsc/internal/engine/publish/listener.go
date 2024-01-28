package publish

import (
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

type SmppSendSmsRequestHandler struct {
	mutex          sync.Mutex
	smsRepository  sdk.SmsRepository
	smppRepository sdk.SmppRepository
	manager        smpp.ConnectorManager
	predicate      map[string]SendSmsRequestPredicate
}

func (instance *SmppSendSmsRequestHandler) Accept(req asyncapi.SendSmsRequest) (asyncapi.SendSmsResponse, error) {
	instant := time.Now()
	zap.L().Info("Listening to asyncapi.SendSmsRequest", zap.String(requestIdAttribute, req.Id))
	zap.L().Debug("Fetching model.Sms from Repository", zap.String(requestIdAttribute, req.Id))
	fromDb, err := instance.smsRepository.FindById(req.Id)
	if err != nil {
		zap.L().Error("Cannot fetch model.Sms from Repository",
			zap.String(requestIdAttribute, req.Id), zap.Error(err))
		return asyncapi.SendSmsResponse{}, NewServiceNotAvailable(GenericProblemDetail)
	}
	if fromDb != nil {
		return instance.getAsyncResponseFromDb(req, fromDb)
	}
	return instance.doAccept(req, instant)
}

func (instance *SmppSendSmsRequestHandler) getAsyncResponseFromDb(req asyncapi.SendSmsRequest, fromDb *model.Sms) (asyncapi.SendSmsResponse, error) {
	zap.L().Debug("Retrieved model.Sms from Repository. No action will be taken",
		zap.String(requestIdAttribute, req.Id))
	if fromDb.CanceledAt != nil {
		return asyncapi.SendSmsResponse{
			Id:         fromDb.Id,
			CanceledAt: fromDb.CanceledAt,
		}, nil
	} else if fromDb.Error != "" {
		return asyncapi.SendSmsResponse{
			Id:   fromDb.Id,
			Smsc: fromDb.Smpp,
			Problem: &asyncapi.Problem{
				Detail: fromDb.Error,
				Type:   GenericProblemType,
				Title:  GenericProblemTitle,
			},
		}, nil
	} else {
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
}

func (instance *SmppSendSmsRequestHandler) doAccept(req asyncapi.SendSmsRequest, receivedAt time.Time) (asyncapi.SendSmsResponse, error) {
	zap.L().Debug("Submitting asyncapi.SendSmsRequest into smpp.Connector(s)",
		zap.String(requestIdAttribute, req.Id))
	resp, err := instance.sendRequest(req)
	if err != nil {
		zap.L().Error("Cannot process asyncapi.SendSmsRequest due to an error",
			zap.String(requestIdAttribute, req.Id),
			zap.Error(err))
		return asyncapi.SendSmsResponse{}, err
	}
	sms := &model.Sms{
		Id:            req.Id,
		TrackId:       resp.Id,
		Smpp:          resp.Smsc,
		ListenedAt:    receivedAt,
		TrackDelivery: false,
	}
	if resp.Problem != nil {
		sms.Error = resp.Problem.Detail
	}
	opts := []zap.Field{
		zap.String(requestIdAttribute, req.Id),
		zap.String(smpp.SmsIdAttribute, resp.Id),
	}
	if resp.Smsc != nil {
		opts = append(opts, zap.String(smpp.SmscIdAttribute, resp.Smsc.Id))
	}
	if resp.Delivery == asyncapi.TrackingDeliveryStrategy {
		sms.TrackDelivery = true
	}
	zap.L().Debug("Persisting model.Sms for asyncapi.SendSmsRequest", opts...)
	err = instance.smsRepository.Save(sms)
	if err != nil {
		return asyncapi.SendSmsResponse{}, err
	}
	return resp.SendSmsResponse, nil
}

func (instance *SmppSendSmsRequestHandler) sendRequest(req asyncapi.SendSmsRequest) (SendSmsResponse, error) {
	response := SendSmsResponse{
		SendSmsResponse: asyncapi.SendSmsResponse{
			Id: req.Id,
		},
	}
	zap.L().Debug("Retrieving []smpp.Connector  in order to process asyncapi.SendSmsRequest",
		zap.String(requestIdAttribute, req.Id))
	canSendRequest := false
	for _, each := range instance.manager.GetList() {
		zap.L().Debug(fmt.Sprintf("Fetching outbound.SendSmsRequestPredicate smpp.Connector[id=%s]",
			each.GetId()), zap.String(requestIdAttribute, req.Id), zap.String(smpp.SmscIdAttribute, each.GetId()),
			zap.String(smpp.SmscAliasAttribute, each.GetAlias()),
		)
		predicate, hasValue := instance.predicate[each.GetId()]
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
		canSendRequest = true
		zap.L().Debug(fmt.Sprintf("Starting sendMessage on smpp.Connector[id=%s]",
			each.GetId()), zap.String(requestIdAttribute, req.Id), zap.String(smpp.SmscIdAttribute, each.GetId()),
			zap.String(smpp.SmscAliasAttribute, each.GetAlias()),
		)
		sendMessageResponse, err := each.SendMessage(req.To, req.Content)
		if err != nil {
			if _, isConnectorUnavailable := err.(smpp.UnavailableConnectorError); isConnectorUnavailable {
				zap.L().Warn(fmt.Sprintf("Cannot sendMessage on smpp.Connector[id=%s] since it's not available",
					each.GetId()), zap.String(requestIdAttribute, req.Id), zap.String(smpp.SmscIdAttribute, each.GetId()),
					zap.String(smpp.SmscAliasAttribute, each.GetAlias()),
					zap.Error(err),
				)
				continue
			}
			return SendSmsResponse{}, err
		}
		response.Id = sendMessageResponse.Id
		response.Delivery = asyncapi.NotTrackingDeliveryStrategy
		response.Smsc = &asyncapi.ObjectId{Id: each.GetId()}
		if each.IsTrackingDelivery() {
			response.Delivery = asyncapi.TrackingDeliveryStrategy
		}
	}
	if response.SendSmsResponse.Smsc == nil {
		if canSendRequest {
			return response, NewServiceNotAvailable("")
		}
		response.SendSmsResponse.Problem = &asyncapi.Problem{
			Type:   CannotSendSmsRequestProblemType,
			Title:  CannotSendSmsRequestProblemTitle,
			Detail: CannotSendSmsRequestProblemDetail,
		}
	}
	return response, nil
}

func (instance *SmppSendSmsRequestHandler) Close() error {
	instance.mutex.Lock()
	defer instance.mutex.Unlock()
	if instance.predicate != nil {
		instance.predicate = nil
	}
	return nil
}

func (instance *SmppSendSmsRequestHandler) AfterPropertiesSet() error {
	instance.mutex.Lock()
	defer instance.mutex.Unlock()
	if instance.predicate == nil {
		instance.predicate = make(map[string]SendSmsRequestPredicate)
	}
	zap.L().Info("Setting up outbound.SendSmsRequestHandler")
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
		instance.predicate[each.GetId()] = predicate
	}
	return nil
}
