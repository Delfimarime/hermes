package restapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/delfimarime/hermes/services/smsc/pkg/restapi"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

type EditByTestConfiguration struct {
	name       string
	username   string
	target     string
	err        error
	request    restapi.UpdateSmscRequest
	AssertWith func(*testing.T, *httptest.ResponseRecorder, string, restapi.UpdateSmscRequest) error
}

func TestSmscApi_EditById_and_expect_success(t *testing.T) {
	factory := func(f func(request *restapi.UpdateSmscRequest)) restapi.UpdateSmscRequest {
		r := &restapi.UpdateSmscRequest{
			PoweredBy: "raitonbl.com",
			Settings: restapi.SmscSettingsRequest{
				Host: restapi.Host{
					Username: "admin",
					Password: "admin",
					Address:  "localhost:4000",
				},
			},
			Name:        "raitonbl",
			Description: "<description/>",
			Type:        restapi.TransmitterType,
		}
		if f != nil {
			f(r)
		}
		return *r
	}
	params := []EditByTestConfiguration{
		{
			name:   "type=transmitter",
			target: "0",
			request: factory(func(r *restapi.UpdateSmscRequest) {
				r.Type = restapi.TransmitterType
			}),
			AssertWith: nil,
		},
		{
			name:   "type=receiver",
			target: "0",
			request: factory(func(r *restapi.UpdateSmscRequest) {
				r.Type = restapi.ReceiverType
			}),
			AssertWith: nil,
		},
		{
			name:   "type=transceiver",
			target: "0",
			request: factory(func(r *restapi.UpdateSmscRequest) {
				r.Type = restapi.TransceiverType
			}),
			AssertWith: nil,
		},
		{
			name:   "powered_by=nil",
			target: "0",
			request: factory(func(r *restapi.UpdateSmscRequest) {
				r.PoweredBy = ""
			}),
			AssertWith: nil,
		},
		{
			name:   "powered_by=<value>",
			target: "0",
			request: factory(func(r *restapi.UpdateSmscRequest) {
				r.PoweredBy = "<value>"
			}),
			AssertWith: nil,
		},
		{
			name:   "settings.bind.timeout=1000",
			target: "0",
			request: factory(func(r *restapi.UpdateSmscRequest) {
				r.Settings.Bind = &restapi.Bind{
					Timeout: 1000,
				}
			}),
			AssertWith: nil,
		},
	}
	executeEditByIdTest(t, assertUpdateSmscResponseWhenOK, params)

}

func executeEditByIdTest(t *testing.T, assertWith func(*testing.T, *httptest.ResponseRecorder, string, restapi.UpdateSmscRequest) error, arr []EditByTestConfiguration) {
	if arr == nil {
		return
	}
	smscApi := &SmscApi{
		service: &TestSmscService{},
	}
	for _, definition := range arr {
		username := "dmarime"
		if definition.username != "" {
			username = definition.username
		}
		r := getGinEngine(&HardCodedAuthenticator{username: username}, smscApi)
		t.Run(definition.name, func(t *testing.T) {
			requestData, _ := json.Marshal(definition.request)
			req, _ := http.NewRequest("PUT", smscEndpoint+"/"+definition.target, bytes.NewBuffer(requestData))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			var err error
			if definition.AssertWith == nil {
				err = assertWith(t, w, username, definition.request)
			} else {
				err = assertWith(t, w, username, definition.request)
			}
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func assertUpdateSmscResponseWhenBadInput(t *testing.T, w *httptest.ResponseRecorder, _ string, _ restapi.UpdateSmscRequest) error {
	require.Equal(t, 422, w.Code)
	zalandoProblem := make(map[string]any)
	if err := json.Unmarshal([]byte(w.Body.String()), &zalandoProblem); err != nil {
		return err
	}
	require.Equal(t, float64(422), zalandoProblem[zalandoStatusPath])
	require.Equal(t, httpValidationTitle, zalandoProblem[zalandoTitlePath])
	require.Equal(t, AddSmscOperationId, zalandoProblem[zalandoOperationIdPath])
	require.Equal(t, fmt.Sprintf(constraintViolationF, EditSmscOperationId), zalandoProblem[zalandoTypePath])
	require.Equal(t, fmt.Sprintf(httpValidationDetailF, EditSmscOperationId), zalandoProblem[zalandoDetailPath])
	return nil
}

func assertUpdateSmscResponseWhenOK(t *testing.T, w *httptest.ResponseRecorder, username string, smscRequest restapi.UpdateSmscRequest) error {
	require.Equal(t, 200, w.Code)
	smscResponse := restapi.UpdateSmscResponse{}
	if err := json.Unmarshal([]byte(w.Body.String()), &smscResponse); err != nil {
		return err
	}
	require.NotNil(t, smscResponse.LastUpdatedAt)
	require.Equal(t, username, smscResponse.LastUpdatedBy)
	return assertNewSmscResponseWhenOK(t, w, username, smscRequest, "")
}

func assertNewSmscResponseWhenOK(t *testing.T, w *httptest.ResponseRecorder, username string, smscRequest restapi.UpdateSmscRequest, alias string) error {
	require.Equal(t, 200, w.Code)
	smscResponse := restapi.NewSmscResponse{}
	if err := json.Unmarshal([]byte(w.Body.String()), &smscResponse); err != nil {
		return err
	}
	require.NotEmpty(t, smscResponse.Id)
	require.NotNil(t, smscResponse.CreatedAt)
	require.Empty(t, smscResponse.Settings.Host.Password)
	require.Equal(t, alias, smscResponse.Alias)
	require.Equal(t, username, smscResponse.CreatedBy)
	require.Equal(t, smscRequest.Type, smscResponse.Type)
	require.Equal(t, smscRequest.Name, smscResponse.Name)
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
	return nil
}
