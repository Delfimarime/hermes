package restapi

import (
	"github.com/delfimarime/hermes/services/smsc/pkg/restapi/common"
	"github.com/delfimarime/hermes/services/smsc/pkg/restapi/smsc"
	"github.com/google/uuid"
	"time"
)

type TestSmscService struct {
	err                 error
	getSmscByIdResponse smsc.GetSmscByIdResponse
	smscSearchResponse  common.Page[smsc.PaginatedSmsc]
}

func (t *TestSmscService) Add(username string, request smsc.NewSmscRequest) (smsc.NewSmscResponse, error) {
	if t.err != nil {
		return smsc.NewSmscResponse{}, t.err
	}
	response := smsc.NewSmscResponse{
		NewSmscRequest: smsc.NewSmscRequest{
			UpdateSmscRequest: smsc.UpdateSmscRequest{
				Name:        request.Name,
				Type:        request.Type,
				PoweredBy:   request.PoweredBy,
				Description: request.Description,
				Settings: smsc.Settings{
					Bind:        request.Settings.Bind,
					Merge:       request.Settings.Merge,
					Enquire:     request.Settings.Enquire,
					Response:    request.Settings.Response,
					Delivery:    request.Settings.Delivery,
					ServiceType: request.Settings.ServiceType,
					Host: smsc.Host{
						Username: request.Settings.Host.Username,
						Address:  request.Settings.Host.Address,
					},
					SourceAddr: request.Settings.SourceAddr,
				},
			},
			Alias: request.Alias,
		},
		CreatedBy: username,
		CreatedAt: time.Now(),
		Id:        uuid.New().String(),
	}
	return response, nil
}

func (t *TestSmscService) EditById(username string, id string, request smsc.UpdateSmscRequest) (smsc.UpdateSmscResponse, error) {
	if t.err != nil {
		return smsc.UpdateSmscResponse{}, t.err
	}
	lastUpdatedAt := time.Date(2024, time.February, 10, 17, 35, 0, 0, time.UTC)
	createdAt := lastUpdatedAt.Add(-time.Hour * 24 * 7)
	return smsc.UpdateSmscResponse{
		LastUpdatedBy: username,
		LastUpdatedAt: lastUpdatedAt,
		NewSmscResponse: smsc.NewSmscResponse{
			NewSmscRequest: smsc.NewSmscRequest{
				UpdateSmscRequest: smsc.UpdateSmscRequest{
					Name:        request.Name,
					Type:        request.Type,
					PoweredBy:   request.PoweredBy,
					Description: request.Description,
					Settings: smsc.Settings{
						Bind:        request.Settings.Bind,
						Merge:       request.Settings.Merge,
						Enquire:     request.Settings.Enquire,
						Response:    request.Settings.Response,
						Delivery:    request.Settings.Delivery,
						ServiceType: request.Settings.ServiceType,
						Host: smsc.Host{
							Username: request.Settings.Host.Username,
							Address:  request.Settings.Host.Address,
						},
						SourceAddr: request.Settings.SourceAddr,
					},
				},
			},
			CreatedBy: username,
			CreatedAt: createdAt,
			Id:        uuid.New().String(),
		},
	}, nil
}

func (t *TestSmscService) EditSettingsById(_ string, _ string, _ smsc.UpdateSmscSettingsRequest) error {
	return t.err
}

func (t *TestSmscService) EditStateById(_ string, _ string, _ smsc.UpdateSmscState) error {
	return t.err
}

func (t *TestSmscService) RemoveById(_ string, _ string) error {
	return t.err
}

func (t *TestSmscService) FindAll(_ smsc.SearchCriteriaRequest) (common.Page[smsc.PaginatedSmsc], error) {
	return t.smscSearchResponse, t.err
}

func (t *TestSmscService) FindById(id string) (smsc.GetSmscByIdResponse, error) {
	if t.err != nil {
		return smsc.GetSmscByIdResponse{}, t.err
	}
	resp := t.getSmscByIdResponse
	resp.Settings.Host.Password = ""
	return resp, nil
}
