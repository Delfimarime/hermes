package restapi

import (
	"bytes"
	"encoding/json"
	"github.com/delfimarime/hermes/services/smsc/pkg/restapi"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSmscApi_New_and_expect_success(t *testing.T) {
	doTestSmscApiNewWithSuccess(t, []struct {
		name     string
		username string
		request  restapi.NewSmscRequest
	}{
		{
			name: "Transmitter",
			request: restapi.NewSmscRequest{
				PoweredBy: "raitonbl.com",
				Settings: restapi.SmscSettingsRequest{
					Host: restapi.Host{
						Username: "admin",
						Password: "admin",
						Address:  "localhost:4000",
					},
				},
				Name:        "raitonbl",
				Alias:       "raitonbl",
				Description: "<description/>",
				Type:        restapi.TransmitterType,
			},
		},
		{
			name:     "Receiver",
			username: "anonymous",
			request: restapi.NewSmscRequest{
				PoweredBy: "raitonbl.com",
				Settings: restapi.SmscSettingsRequest{
					Host: restapi.Host{
						Username: "admin",
						Password: "admin",
						Address:  "localhost:4000",
					},
				},
				Name:        "raitonbl",
				Alias:       "raitonbl",
				Description: "<description/>",
				Type:        restapi.ReceiverType,
			},
		},
		{
			name: "Transceiver",
			request: restapi.NewSmscRequest{
				PoweredBy: "raitonbl.com",
				Settings: restapi.SmscSettingsRequest{
					Host: restapi.Host{
						Username: "admin",
						Password: "admin",
						Address:  "localhost:4000",
					},
				},
				Name:        "raitonbl",
				Alias:       "raitonbl",
				Description: "<description/>",
				Type:        restapi.TransceiverType,
			},
		},
	})

}

func doTestSmscApiNewWithSuccess(t *testing.T, arr []struct {
	name     string
	username string
	request  restapi.NewSmscRequest
}) {
	if arr == nil {
		return
	}
	smscApi := &SmscApi{
		service:              &TestSmsService{},
		getAuthenticatedUser: nil,
	}
	r := getGinEngine(smscApi)
	for _, definition := range arr {
		username := "dmarime"
		if definition.username != "" {
			username = definition.username
		}
		smscApi.getAuthenticatedUser = func(c *gin.Context) string {
			return username
		}
		t.Run(definition.name, func(t *testing.T) {
			smscRequest := restapi.NewSmscRequest{
				PoweredBy:   "raitonBL",
				Type:        restapi.TransmitterType,
				Name:        "raitonBL",
				Alias:       "raitonbl",
				Description: "<description/>",
				Settings: restapi.SmscSettingsRequest{
					Host: restapi.Host{
						Username: "admin",
						Password: "admin",
						Address:  "192.168.1.0:4000",
					},
					Bind: &restapi.Bind{
						Timeout: 1000,
					},
					Delivery: &restapi.Delivery{
						AwaitReport: true,
					},
				},
			}
			requestData, _ := json.Marshal(smscRequest)
			req, _ := http.NewRequest("POST", "/smscs", bytes.NewBuffer(requestData))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			require.Equal(t, 200, w.Code)
			smscResponse := restapi.NewSmscResponse{}
			if err := json.Unmarshal([]byte(w.Body.String()), &smscResponse); err != nil {
				t.Fatal(err)
			}
			require.NotEmpty(t, smscResponse.Id)
			require.NotNil(t, smscResponse.CreatedAt)
			require.Empty(t, smscResponse.Settings.Host.Password)
			require.Equal(t, username, smscResponse.CreatedBy)
			require.Equal(t, smscRequest.Type, smscResponse.Type)
			require.Equal(t, smscRequest.Name, smscResponse.Name)
			require.Equal(t, smscRequest.Alias, smscResponse.Alias)
			require.Equal(t, smscRequest.PoweredBy, smscResponse.PoweredBy)
			require.Equal(t, smscRequest.Description, smscResponse.Description)
			require.Equal(t, smscRequest.Settings.SourceAddr, smscResponse.Settings.SourceAddr)
			require.Equal(t, smscRequest.Settings.ServiceType, smscResponse.Settings.ServiceType)
			require.Equal(t, smscRequest.Settings.Host.Address, smscResponse.Settings.Host.Address)
			require.Equal(t, smscRequest.Settings.Host.Username, smscResponse.Settings.Host.Username)
			if smscRequest.Settings.Bind == nil {
				require.Nil(t, smscResponse.Settings.Bind)
			} else {
				require.NotNil(t, smscResponse.Settings.Bind)
				require.Equal(t, smscRequest.Settings.Bind.Timeout, smscResponse.Settings.Bind.Timeout)
			}
			if smscRequest.Settings.Merge == nil {
				require.Nil(t, smscResponse.Settings.Merge)
			} else {
				require.NotNil(t, smscResponse.Settings.Merge)
				require.Equal(t, smscRequest.Settings.Merge.Interval, smscResponse.Settings.Merge.Interval)
				require.Equal(t, smscRequest.Settings.Merge.CleanupInterval, smscResponse.Settings.Merge.CleanupInterval)
			}
			if smscRequest.Settings.Enquire == nil {
				require.Nil(t, smscResponse.Settings.Enquire)
			} else {
				require.NotNil(t, smscResponse.Settings.Enquire)
				require.Equal(t, smscRequest.Settings.Enquire.Link, smscResponse.Settings.Enquire.Link)
				require.Equal(t, smscRequest.Settings.Enquire.LinkTimeout, smscResponse.Settings.Enquire.LinkTimeout)
			}
			if smscRequest.Settings.Response == nil {
				require.Nil(t, smscResponse.Settings.Response)
			} else {
				require.NotNil(t, smscResponse.Settings.Response)
				require.Equal(t, smscRequest.Settings.Response.Timeout, smscResponse.Settings.Response.Timeout)
			}
			if smscRequest.Settings.Delivery == nil {
				require.Nil(t, smscResponse.Settings.Delivery)
			} else {
				require.NotNil(t, smscResponse.Settings.Delivery)
				require.Equal(t, smscRequest.Settings.Delivery.AwaitReport, smscResponse.Settings.Delivery.AwaitReport)
			}
		})
	}
}

type TestSmsService struct {
	err error
}

func (t *TestSmsService) Add(username string, request restapi.NewSmscRequest) (restapi.NewSmscResponse, error) {
	if t.err != nil {
		return restapi.NewSmscResponse{}, t.err
	}
	response := restapi.NewSmscResponse{
		NewSmscRequest: restapi.NewSmscRequest{
			Name:        request.Name,
			Type:        request.Type,
			Alias:       request.Alias,
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
		CreatedBy: username,
		CreatedAt: time.Now(),
		Id:        uuid.New().String(),
	}

	return response, nil
}
