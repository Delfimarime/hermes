package restapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/delfimarime/hermes/services/smsc/pkg/restapi"
	"github.com/stretchr/testify/require"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	zalandoTypePath        = "type"
	zalandoDetailPath      = "detail"
	zalandoTitlePath       = "title"
	zalandoStatusPath      = "status"
	zalandoOperationIdPath = "operationId"
)

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func TestSmscApi_New_and_expect_success(t *testing.T) {
	doTestSmscApiNewWithSuccess(t, []struct {
		name     string
		username string
		request  restapi.NewSmscRequest
	}{
		{
			name: "type=transmitter",
			request: restapi.NewSmscRequest{
				UpdateSmscRequest: restapi.UpdateSmscRequest{
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
				},
				Alias: "raitonbl",
			},
		},
		{
			name:     "type=receiver",
			username: "anonymous",
			request: restapi.NewSmscRequest{
				UpdateSmscRequest: restapi.UpdateSmscRequest{
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
					Type:        restapi.ReceiverType,
				},
				Alias: "raitonbl",
			},
		},
		{
			name: "type=transceiver",
			request: restapi.NewSmscRequest{
				UpdateSmscRequest: restapi.UpdateSmscRequest{
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
					Type:        restapi.TransceiverType,
				},
				Alias: "raitonbl",
			},
		},
		{
			name: "powered_by=nil",
			request: restapi.NewSmscRequest{
				UpdateSmscRequest: restapi.UpdateSmscRequest{
					Settings: restapi.SmscSettingsRequest{
						Host: restapi.Host{
							Username: "admin",
							Password: "admin",
							Address:  "localhost:4000",
						},
					},
					Name:        "raitonbl",
					Description: "<description/>",
					Type:        restapi.TransceiverType,
				},
				Alias: "raitonbl",
			},
		},
		{
			name: "powered_by=nil",
			request: restapi.NewSmscRequest{
				UpdateSmscRequest: restapi.UpdateSmscRequest{
					Settings: restapi.SmscSettingsRequest{
						Bind: &restapi.Bind{
							Timeout: 1000,
						},
						Host: restapi.Host{
							Username: "admin",
							Password: "admin",
							Address:  "localhost:4000",
						},
					},
					Name:        "raitonbl",
					Description: "<description/>",
					Type:        restapi.TransceiverType,
				},
				Alias: "raitonbl",
			},
		},
	})
}

func TestSmscApi_New_when_type_is_not_valid(t *testing.T) {
	doTestSmscApiNewWithBadInput(t, []struct {
		name    string
		request restapi.NewSmscRequest
	}{
		{
			name: "type=nil",
			request: restapi.NewSmscRequest{
				UpdateSmscRequest: restapi.UpdateSmscRequest{
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
				},
				Alias: "raitonbl",
			},
		},
		{
			name: "type=<type/>",
			request: restapi.NewSmscRequest{
				UpdateSmscRequest: restapi.UpdateSmscRequest{
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
					Type:        "<type/>",
				},
				Alias: "raitonbl",
			},
		},
	})
}

func TestSmscApi_New_when_name_is_not_valid(t *testing.T) {
	doTestSmscApiNewWithBadInput(t, []struct {
		name    string
		request restapi.NewSmscRequest
	}{
		{
			name: "name=nil",
			request: restapi.NewSmscRequest{
				UpdateSmscRequest: restapi.UpdateSmscRequest{
					PoweredBy: "raitonbl.com",
					Settings: restapi.SmscSettingsRequest{
						Host: restapi.Host{
							Username: "admin",
							Password: "admin",
							Address:  "localhost:4000",
						},
					},
					Description: "<description/>",
					Type:        restapi.TransmitterType,
				},
				Alias: "raitonbl",
			},
		},
		{
			name: "len(name)==2",
			request: restapi.NewSmscRequest{
				UpdateSmscRequest: restapi.UpdateSmscRequest{
					PoweredBy: "raitonbl.com",
					Settings: restapi.SmscSettingsRequest{
						Host: restapi.Host{
							Username: "admin",
							Password: "admin",
							Address:  "localhost:4000",
						},
					},
					Name:        stringWithCharset(2),
					Description: "<description/>",
					Type:        restapi.TransmitterType,
				},
				Alias: "raitonbl",
			},
		},
		{
			name: "len(name)==51",
			request: restapi.NewSmscRequest{
				UpdateSmscRequest: restapi.UpdateSmscRequest{

					PoweredBy: "raitonbl.com",
					Settings: restapi.SmscSettingsRequest{
						Host: restapi.Host{
							Username: "admin",
							Password: "admin",
							Address:  "localhost:4000",
						},
					},
					Name:        stringWithCharset(51),
					Description: "<description/>",
					Type:        restapi.TransmitterType,
				},
				Alias: "raitonbl",
			},
		},
	})
}

func TestSmscApi_New_when_alias_is_not_valid(t *testing.T) {
	doTestSmscApiNewWithBadInput(t, []struct {
		name    string
		request restapi.NewSmscRequest
	}{
		{
			name: "alias=nil",
			request: restapi.NewSmscRequest{
				UpdateSmscRequest: restapi.UpdateSmscRequest{
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
				},
			},
		},
		{
			name: "len(alias)==2",
			request: restapi.NewSmscRequest{
				UpdateSmscRequest: restapi.UpdateSmscRequest{
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
				},
				Alias: stringWithCharset(2),
			},
		},
		{
			name: "len(alias)==21",
			request: restapi.NewSmscRequest{
				UpdateSmscRequest: restapi.UpdateSmscRequest{
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
				},
				Alias: stringWithCharset(51),
			},
		},
	})
}

func TestSmscApi_New_when_description_is_not_valid(t *testing.T) {
	doTestSmscApiNewWithBadInput(t, []struct {
		name    string
		request restapi.NewSmscRequest
	}{
		{
			name: "description=nil",
			request: restapi.NewSmscRequest{
				UpdateSmscRequest: restapi.UpdateSmscRequest{
					PoweredBy: "raitonbl.com",
					Settings: restapi.SmscSettingsRequest{
						Host: restapi.Host{
							Username: "admin",
							Password: "admin",
							Address:  "localhost:4000",
						},
					},
					Name: "raitonbl",
					Type: restapi.TransmitterType,
				},
				Alias: "raitonbl",
			},
		},
		{
			name: "len(description)==1",
			request: restapi.NewSmscRequest{
				UpdateSmscRequest: restapi.UpdateSmscRequest{
					PoweredBy: "raitonbl.com",
					Settings: restapi.SmscSettingsRequest{
						Host: restapi.Host{
							Username: "admin",
							Password: "admin",
							Address:  "localhost:4000",
						},
					},
					Description: stringWithCharset(1),
					Name:        "raitonbl",
					Type:        restapi.TransmitterType,
				},
				Alias: "raitonbl",
			},
		},
		{
			name: "len(description)==21",
			request: restapi.NewSmscRequest{
				UpdateSmscRequest: restapi.UpdateSmscRequest{
					PoweredBy: "raitonbl.com",
					Settings: restapi.SmscSettingsRequest{
						Host: restapi.Host{
							Username: "admin",
							Password: "admin",
							Address:  "localhost:4000",
						},
					},
					Name:        "raitonbl",
					Description: stringWithCharset(256),
					Type:        restapi.TransmitterType,
				},
				Alias: "raitonbl",
			},
		},
	})
}

func TestSmscApi_New_when_source_addr_from_settings_is_not_valid(t *testing.T) {
	doTestSmscApiNewWithBadInput(t, []struct {
		name    string
		request restapi.NewSmscRequest
	}{
		{
			name: "settings.source_address=google.com",
			request: restapi.NewSmscRequest{
				UpdateSmscRequest: restapi.UpdateSmscRequest{
					PoweredBy: "raitonbl.com",
					Settings: restapi.SmscSettingsRequest{
						Host: restapi.Host{
							Username: "admin",
							Password: "admin",
							Address:  "localhost:4000",
						},
						SourceAddr: "google.com",
					},
					Name:        "raitonbl",
					Description: "<description/>",
					Type:        restapi.TransmitterType,
				},
				Alias: "raitonbl",
			},
		},
	})
}

func TestSmscApi_New_when_host_from_settings_is_not_valid(t *testing.T) {
	doTestSmscApiNewWithBadInput(t, []struct {
		name    string
		request restapi.NewSmscRequest
	}{
		{
			name: "host=nil",
			request: restapi.NewSmscRequest{
				UpdateSmscRequest: restapi.UpdateSmscRequest{
					PoweredBy:   "raitonbl.com",
					Settings:    restapi.SmscSettingsRequest{},
					Name:        "raitonbl",
					Description: "<description/>",
					Type:        restapi.TransmitterType,
				},
				Alias: "raitonbl",
			},
		},
		{
			name: "host.username=nil",
			request: restapi.NewSmscRequest{
				UpdateSmscRequest: restapi.UpdateSmscRequest{
					PoweredBy: "raitonbl.com",
					Settings: restapi.SmscSettingsRequest{
						Host: restapi.Host{
							Password: "admin",
							Address:  "localhost:4000",
						},
					},
					Name:        "raitonbl",
					Description: "<description/>",
					Type:        restapi.TransmitterType,
				},
				Alias: "raitonbl",
			},
		},
		{
			name: "host.password=nil",
			request: restapi.NewSmscRequest{
				UpdateSmscRequest: restapi.UpdateSmscRequest{
					PoweredBy: "raitonbl.com",
					Settings: restapi.SmscSettingsRequest{
						Host: restapi.Host{
							Username: "admin",
							Address:  "localhost:4000",
						},
					},
					Name:        "raitonbl",
					Description: "<description/>",
					Type:        restapi.TransmitterType,
				},
				Alias: "raitonbl",
			},
		},
		{
			name: "host.address=nil",
			request: restapi.NewSmscRequest{
				UpdateSmscRequest: restapi.UpdateSmscRequest{
					PoweredBy: "raitonbl.com",
					Settings: restapi.SmscSettingsRequest{
						Host: restapi.Host{
							Username: "admin",
							Password: "admin",
						},
					},
					Name:        "raitonbl",
					Description: "<description/>",
					Type:        restapi.TransmitterType,
				},
				Alias: "raitonbl",
			},
		},
		{
			name: "host.address not hostname_port",
			request: restapi.NewSmscRequest{
				UpdateSmscRequest: restapi.UpdateSmscRequest{
					PoweredBy: "raitonbl.com",
					Settings: restapi.SmscSettingsRequest{
						Host: restapi.Host{
							Username: "admin",
							Password: "admin",
							Address:  "google",
						},
					},
					Name:        "raitonbl",
					Description: "<description/>",
					Type:        restapi.TransmitterType,
				},
				Alias: "raitonbl",
			},
		},
	})
}

func TestSmscApi_New_when_bind_from_settings_is_not_valid(t *testing.T) {
	doTestSmscApiNewWithBadInput(t, []struct {
		name    string
		request restapi.NewSmscRequest
	}{
		{
			name: "bind.timeout=999",
			request: restapi.NewSmscRequest{
				UpdateSmscRequest: restapi.UpdateSmscRequest{
					PoweredBy: "raitonbl.com",
					Settings: restapi.SmscSettingsRequest{
						Bind: &restapi.Bind{
							Timeout: 999,
						},
						Host: restapi.Host{
							Username: "admin",
							Password: "admin",
							Address:  "localhost:4000",
						},
					},
					Name:        "raitonbl",
					Description: "<description/>",
					Type:        restapi.TransmitterType,
				},
				Alias: "raitonbl",
			},
		},
		{
			name: "bind.timeout=0",
			request: restapi.NewSmscRequest{
				UpdateSmscRequest: restapi.UpdateSmscRequest{
					PoweredBy: "raitonbl.com",
					Settings: restapi.SmscSettingsRequest{
						Bind: &restapi.Bind{},
						Host: restapi.Host{
							Username: "admin",
							Password: "admin",
							Address:  "localhost:4000",
						},
					},
					Name:        "raitonbl",
					Description: "<description/>",
					Type:        restapi.TransmitterType,
				},
				Alias: "raitonbl",
			},
		},
	})
}

func TestSmscApi_New_when_enquire_from_settings_is_not_valid(t *testing.T) {
	doTestSmscApiNewWithBadInput(t, []struct {
		name    string
		request restapi.NewSmscRequest
	}{
		{
			name: "enquiry.link=999",
			request: restapi.NewSmscRequest{
				UpdateSmscRequest: restapi.UpdateSmscRequest{
					PoweredBy: "raitonbl.com",
					Settings: restapi.SmscSettingsRequest{
						Enquire: &restapi.Enquire{
							Link:        999,
							LinkTimeout: 1000,
						},
						Host: restapi.Host{
							Username: "admin",
							Password: "admin",
							Address:  "localhost:4000",
						},
					},
					Name:        "raitonbl",
					Description: "<description/>",
					Type:        restapi.TransmitterType,
				},
				Alias: "raitonbl",
			},
		},
		{
			name: "enquiry.link=0",
			request: restapi.NewSmscRequest{
				UpdateSmscRequest: restapi.UpdateSmscRequest{
					PoweredBy: "raitonbl.com",
					Settings: restapi.SmscSettingsRequest{
						Enquire: &restapi.Enquire{
							LinkTimeout: 1000,
						},
						Host: restapi.Host{
							Username: "admin",
							Password: "admin",
							Address:  "localhost:4000",
						},
					},
					Name:        "raitonbl",
					Description: "<description/>",
					Type:        restapi.TransmitterType,
				},
				Alias: "raitonbl",
			},
		},
		{
			name: "enquiry.link_timeout=999",
			request: restapi.NewSmscRequest{
				UpdateSmscRequest: restapi.UpdateSmscRequest{
					PoweredBy: "raitonbl.com",
					Settings: restapi.SmscSettingsRequest{
						Enquire: &restapi.Enquire{
							Link:        1000,
							LinkTimeout: 999,
						},
						Host: restapi.Host{
							Username: "admin",
							Password: "admin",
							Address:  "localhost:4000",
						},
					},
					Name:        "raitonbl",
					Description: "<description/>",
					Type:        restapi.TransmitterType,
				},
				Alias: "raitonbl",
			},
		},
		{
			name: "enquiry.link_timeout=0",
			request: restapi.NewSmscRequest{
				UpdateSmscRequest: restapi.UpdateSmscRequest{
					PoweredBy: "raitonbl.com",
					Settings: restapi.SmscSettingsRequest{
						Enquire: &restapi.Enquire{
							Link: 1000,
						},
						Host: restapi.Host{
							Username: "admin",
							Password: "admin",
							Address:  "localhost:4000",
						},
					},
					Name:        "raitonbl",
					Description: "<description/>",
					Type:        restapi.TransmitterType,
				},
				Alias: "raitonbl",
			},
		},
	})
}

func TestSmscApi_New_when_response_from_settings_is_not_valid(t *testing.T) {
	doTestSmscApiNewWithBadInput(t, []struct {
		name    string
		request restapi.NewSmscRequest
	}{
		{
			name: "response.timeout=999",
			request: restapi.NewSmscRequest{
				UpdateSmscRequest: restapi.UpdateSmscRequest{
					PoweredBy: "raitonbl.com",
					Settings: restapi.SmscSettingsRequest{
						Response: &restapi.Response{
							Timeout: 999,
						},
						Host: restapi.Host{
							Username: "admin",
							Password: "admin",
							Address:  "localhost:4000",
						},
					},
					Name:        "raitonbl",
					Description: "<description/>",
					Type:        restapi.TransmitterType,
				},
				Alias: "raitonbl",
			},
		},
		{
			name: "response.timeout=0",
			request: restapi.NewSmscRequest{
				UpdateSmscRequest: restapi.UpdateSmscRequest{
					PoweredBy: "raitonbl.com",
					Settings: restapi.SmscSettingsRequest{
						Response: &restapi.Response{},
						Host: restapi.Host{
							Username: "admin",
							Password: "admin",
							Address:  "localhost:4000",
						},
					},
					Name:        "raitonbl",
					Description: "<description/>",
					Type:        restapi.TransmitterType,
				},
				Alias: "raitonbl",
			},
		},
	})
}

func TestSmscApi_New_when_merge_from_settings_is_not_valid(t *testing.T) {
	doTestSmscApiNewWithBadInput(t, []struct {
		name    string
		request restapi.NewSmscRequest
	}{
		{
			name: "merge.interval=999",
			request: restapi.NewSmscRequest{
				UpdateSmscRequest: restapi.UpdateSmscRequest{
					PoweredBy: "raitonbl.com",
					Settings: restapi.SmscSettingsRequest{
						Merge: &restapi.Merge{
							Interval:        999,
							CleanupInterval: 1000,
						},
						Host: restapi.Host{
							Username: "admin",
							Password: "admin",
							Address:  "localhost:4000",
						},
					},
					Name:        "raitonbl",
					Description: "<description/>",
					Type:        restapi.TransmitterType,
				},
				Alias: "raitonbl",
			},
		},
		{
			name: "merge.interval=0",
			request: restapi.NewSmscRequest{
				UpdateSmscRequest: restapi.UpdateSmscRequest{
					PoweredBy: "raitonbl.com",
					Settings: restapi.SmscSettingsRequest{
						Merge: &restapi.Merge{
							CleanupInterval: 1000,
						},
						Host: restapi.Host{
							Username: "admin",
							Password: "admin",
							Address:  "localhost:4000",
						},
					},
					Name:        "raitonbl",
					Description: "<description/>",
					Type:        restapi.TransmitterType,
				},
				Alias: "raitonbl",
			},
		},
		{
			name: "merge.cleanup_interval=999",
			request: restapi.NewSmscRequest{
				UpdateSmscRequest: restapi.UpdateSmscRequest{
					PoweredBy: "raitonbl.com",
					Settings: restapi.SmscSettingsRequest{
						Merge: &restapi.Merge{
							Interval:        1000,
							CleanupInterval: 999,
						},
						Host: restapi.Host{
							Username: "admin",
							Password: "admin",
							Address:  "localhost:4000",
						},
					},
					Name:        "raitonbl",
					Description: "<description/>",
					Type:        restapi.TransmitterType,
				},
				Alias: "raitonbl",
			},
		},
		{
			name: "merge.cleanup_interval=0",
			request: restapi.NewSmscRequest{
				UpdateSmscRequest: restapi.UpdateSmscRequest{
					PoweredBy: "raitonbl.com",
					Settings: restapi.SmscSettingsRequest{
						Merge: &restapi.Merge{
							Interval: 1000,
						},
						Host: restapi.Host{
							Username: "admin",
							Password: "admin",
							Address:  "localhost:4000",
						},
					},
					Name:        "raitonbl",
					Description: "<description/>",
					Type:        restapi.TransmitterType,
				},
				Alias: "raitonbl",
			},
		},
	})
}

func TestSmscApi_New_when_smsc_service_has_error(t *testing.T) {
	doTestSmscApiNewAndCatchError(t, []struct {
		err  error
		name string
	}{
		{
			name: "errors.New",
			err:  errors.New("<error/>"),
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
		service: &TestSmscService{},
	}
	for _, definition := range arr {
		username := "dmarime"
		if definition.username != "" {
			username = definition.username
		}
		r := getGinEngine(&HardCodedAuthenticator{username: username}, smscApi)
		t.Run(definition.name, func(t *testing.T) {
			smscRequest := definition.request
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

func doTestSmscApiNewWithBadInput(t *testing.T, arr []struct {
	name    string
	request restapi.NewSmscRequest
}) {
	if arr == nil {
		return
	}
	smscApi := &SmscApi{
		service: &TestSmscService{},
	}
	r := getGinEngine(&HardCodedAuthenticator{username: "dmarime"}, smscApi)
	for _, definition := range arr {
		t.Run(definition.name, func(t *testing.T) {
			smscRequest := definition.request
			requestData, _ := json.Marshal(smscRequest)
			req, _ := http.NewRequest("POST", "/smscs", bytes.NewBuffer(requestData))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			require.Equal(t, 422, w.Code)
			zalandoProblem := make(map[string]any)
			if err := json.Unmarshal([]byte(w.Body.String()), &zalandoProblem); err != nil {
				t.Fatal(err)
			}
			require.Equal(t, float64(422), zalandoProblem[zalandoStatusPath])
			require.Equal(t, httpValidationTitle, zalandoProblem[zalandoTitlePath])
			require.Equal(t, AddSmscOperationId, zalandoProblem[zalandoOperationIdPath])
			require.Equal(t, fmt.Sprintf(constraintViolationF, AddSmscOperationId), zalandoProblem[zalandoTypePath])
			require.Equal(t, fmt.Sprintf(httpValidationDetailF, AddSmscOperationId), zalandoProblem[zalandoDetailPath])
		})
	}
}

func doTestSmscApiNewAndCatchError(t *testing.T, arr []struct {
	err  error
	name string
}) {
	if arr == nil {
		return
	}
	smscApi := &SmscApi{
		service: &TestSmscService{},
	}
	r := getGinEngine(&HardCodedAuthenticator{username: "dmarime"}, smscApi)
	for _, definition := range arr {
		smscApi.service.(*TestSmscService).err = definition.err
		t.Run(definition.name, func(t *testing.T) {
			smscRequest := restapi.NewSmscRequest{
				UpdateSmscRequest: restapi.UpdateSmscRequest{
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
				},
				Alias: "raitonbl",
			}
			requestData, _ := json.Marshal(smscRequest)
			req, _ := http.NewRequest("POST", "/smscs", bytes.NewBuffer(requestData))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			require.Equal(t, 500, w.Code)
			zalandoProblem := make(map[string]any)
			if err := json.Unmarshal([]byte(w.Body.String()), &zalandoProblem); err != nil {
				t.Fatal(err)
			}
			require.Equal(t, float64(500), zalandoProblem[zalandoStatusPath])
			require.Equal(t, somethingWentWrongTitle, zalandoProblem[zalandoTitlePath])
			require.Equal(t, AddSmscOperationId, zalandoProblem[zalandoOperationIdPath])
			require.Equal(t, fmt.Sprintf(somethingWentWrongF, AddSmscOperationId), zalandoProblem[zalandoTypePath])
			require.Equal(t, fmt.Sprintf(somethingWentWrongDetailF, AddSmscOperationId), zalandoProblem[zalandoDetailPath])
		})
	}
}

func stringWithCharset(length int) string {
	charset := "aAbBcCdDeEfFgGhHiIjJkKlLmMnNoOpPqQrRsStTuUvVwWxXyYzZ"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset)-1)]
	}
	return string(b)
}
