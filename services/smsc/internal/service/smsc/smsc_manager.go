package smsc

import (
	"fmt"
	"github.com/delfimarime/hermes/services/smsc/internal/asyncapi"
	"github.com/delfimarime/hermes/services/smsc/internal/model"
	"github.com/delfimarime/hermes/services/smsc/internal/repository/sdk"
	"github.com/delfimarime/hermes/services/smsc/internal/service/resolve"
	"github.com/delfimarime/hermes/services/smsc/pkg/restapi/smsc"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"time"
)

const (
	nameProperty           = "name"
	aliasProperty          = "alias"
	validationDetailErrorF = "$.%s is not permitted since another smsc is using it"
)

type DefaultSmscService struct {
	repository  sdk.SmppRepository
	smppChannel asyncapi.SmppChannel
	resolver    resolve.ValueResolver
}

func (instance *DefaultSmscService) Add(createdBy string, request smsc.NewSmscRequest) (smsc.NewSmscResponse, error) {
	request, err := instance.resolveCredentials(request)
	if err != nil {
		return smsc.NewSmscResponse{}, err
	}
	data := instance.modelFromRequest(request)
	data.CreatedAt = time.Now()
	data.CreatedBy = createdBy
	if err = instance.repository.Save(data); err != nil {
		return instance.mapError(request, err)
	}
	response := smsc.NewSmscResponse{
		NewSmscRequest: request,
		Id:             data.Id,
		CreatedAt:      data.CreatedAt,
		CreatedBy:      data.CreatedBy,
	}
	err = instance.smppChannel.SubmitSmscAddedEvent(response)
	if err != nil {
		zap.L().Error("Cannot publish SmscAddedEvent", zap.String("name", request.Name),
			zap.String("alias", request.Alias), zap.Error(err))
	}
	return response, nil
}

func (instance *DefaultSmscService) resolveCredentials(request smsc.NewSmscRequest) (smsc.NewSmscRequest, error) {
	var err error
	_, err = instance.resolver.Get(request.Settings.Host.Username)
	if err != nil {
		return request, err
	}
	_, err = instance.resolver.Get(request.Settings.Host.Password)
	return request, err
}

func (instance *DefaultSmscService) mapError(request smsc.NewSmscRequest, causedBy error) (smsc.NewSmscResponse, error) {
	if prob, isConstraintError := causedBy.(*sdk.FieldConstraintError); isConstraintError {
		if prob.Field == nameProperty || prob.Field == aliasProperty {
			return smsc.NewSmscResponse{}, newUniquenessValidationError(prob.Field)
		}
	}
	zap.L().Error("Cannot persist model.Smpp into sdk.SmppRepository", zap.String("name", request.Name),
		zap.String("alias", request.Alias), zap.Error(causedBy))
	return smsc.NewSmscResponse{}, causedBy
}

func newUniquenessValidationError(field string) ValidationError {
	return ValidationError{
		Field:  field,
		Rule:   "not-unique",
		Detail: fmt.Sprintf(validationDetailErrorF, field),
	}
}

func (instance *DefaultSmscService) modelFromRequest(request smsc.NewSmscRequest) model.Smpp {
	return model.Smpp{
		Name:        request.Name,
		Alias:       request.Alias,
		PoweredBy:   request.PoweredBy,
		Description: request.Description,
		Id:          uuid.New().String(),
		Type:        string(request.Type),
		Settings:    instance.settingsFromRequest(request),
	}
}

func (instance *DefaultSmscService) settingsFromRequest(request smsc.NewSmscRequest) model.Settings {
	value := model.Settings{
		SourceAddr:  request.Settings.SourceAddr,
		ServiceType: request.Settings.ServiceType,
		Host: model.Host{
			Address:  request.Settings.Host.Address,
			Username: request.Settings.Host.Username,
			Password: request.Settings.Host.Password,
		},
		Bind:     nil,
		Merge:    nil,
		Enquire:  nil,
		Response: nil,
		Delivery: nil,
	}
	if request.Settings.Bind != nil {
		value.Bind = &model.Bind{
			Timeout: request.Settings.Bind.Timeout,
		}
	}
	if request.Settings.Merge != nil {
		value.Merge = &model.Merge{
			Interval:        request.Settings.Merge.Interval,
			CleanupInterval: request.Settings.Merge.CleanupInterval,
		}
	}
	if request.Settings.Enquire != nil {
		value.Enquire = &model.Enquire{
			Link:        request.Settings.Enquire.Link,
			LinkTimeout: request.Settings.Enquire.LinkTimeout,
		}
	}
	if request.Settings.Response != nil {
		value.Response = &model.Response{
			Timeout: request.Settings.Response.Timeout,
		}
	}
	if request.Settings.Delivery != nil {
		value.Delivery = &model.Delivery{
			AwaitReport: request.Settings.Delivery.AwaitReport,
		}
	}
	return value
}
