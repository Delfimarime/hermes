package restapi

import (
	"github.com/delfimarime/hermes/services/smsc/pkg/restapi"
	"github.com/google/uuid"
	"time"
)

type TestSmscService struct {
	err error
}

func (t *TestSmscService) Add(username string, request restapi.NewSmscRequest) (restapi.NewSmscResponse, error) {
	if t.err != nil {
		return restapi.NewSmscResponse{}, t.err
	}
	response := restapi.NewSmscResponse{
		NewSmscRequest: restapi.NewSmscRequest{
			UpdateSmscRequest: restapi.UpdateSmscRequest{
				Name:        request.Name,
				Type:        request.Type,
				PoweredBy:   request.PoweredBy,
				Description: request.Description,
				Settings: restapi.SmscSettingsRequest{
					Bind:        request.Settings.Bind,
					Merge:       request.Settings.Merge,
					Enquire:     request.Settings.Enquire,
					Response:    request.Settings.Response,
					Delivery:    request.Settings.Delivery,
					ServiceType: request.Settings.ServiceType,
					Host: restapi.Host{
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

func (t *TestSmscService) EditById(username string, id string, request restapi.UpdateSmscRequest) (restapi.UpdateSmscResponse, error) {
	if t.err != nil {
		return restapi.UpdateSmscResponse{}, t.err
	}
	lastUpdatedAt := time.Date(2024, time.February, 10, 17, 35, 0, 0, time.UTC)
	createdAt := lastUpdatedAt.Add(-time.Hour * 24 * 7)
	return restapi.UpdateSmscResponse{
		LastUpdatedBy: username,
		LastUpdatedAt: lastUpdatedAt,
		NewSmscResponse: restapi.NewSmscResponse{
			NewSmscRequest: restapi.NewSmscRequest{
				UpdateSmscRequest: restapi.UpdateSmscRequest{
					Name:        request.Name,
					Type:        request.Type,
					PoweredBy:   request.PoweredBy,
					Description: request.Description,
					Settings: restapi.SmscSettingsRequest{
						Bind:        request.Settings.Bind,
						Merge:       request.Settings.Merge,
						Enquire:     request.Settings.Enquire,
						Response:    request.Settings.Response,
						Delivery:    request.Settings.Delivery,
						ServiceType: request.Settings.ServiceType,
						Host: restapi.Host{
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

func (t *TestSmscService) EditSettingsById(username string, id string, request restapi.UpdateSmscSettingsRequest) error {
	panic("implement me")
}

func (t *TestSmscService) EditStateById(username string, id string, request restapi.UpdateSmscState) error {
	panic("implement me")
}

func (t *TestSmscService) RemoveById(username string, id string) error {
	panic("implement me")
}
