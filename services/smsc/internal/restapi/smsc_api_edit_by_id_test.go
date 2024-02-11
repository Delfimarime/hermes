package restapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/delfimarime/hermes/services/smsc/pkg/restapi/smsc"
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
	request    smsc.UpdateSmscRequest
	AssertWith func(*testing.T, *httptest.ResponseRecorder, string, smsc.UpdateSmscRequest) error
}

func TestSmscApi_EditById(t *testing.T) {
	executeEditByIdTest(t, assertUpdateSmscResponseWhenOK, []EditByTestConfiguration{
		{
			name:   "type=transmitter",
			target: "0",
			request: createUpdateSmscRequest(func(r *smsc.UpdateSmscRequest) {
				r.Type = smsc.TransmitterType
			}),
			AssertWith: nil,
		},
		{
			name:   "type=receiver",
			target: "0",
			request: createUpdateSmscRequest(func(r *smsc.UpdateSmscRequest) {
				r.Type = smsc.ReceiverType
			}),
			AssertWith: nil,
		},
		{
			name:   "type=transceiver",
			target: "0",
			request: createUpdateSmscRequest(func(r *smsc.UpdateSmscRequest) {
				r.Type = smsc.TransceiverType
			}),
			AssertWith: nil,
		},
		{
			name:   "powered_by=nil",
			target: "0",
			request: createUpdateSmscRequest(func(r *smsc.UpdateSmscRequest) {
				r.PoweredBy = ""
			}),
			AssertWith: nil,
		},
		{
			name:   "powered_by=<value>",
			target: "0",
			request: createUpdateSmscRequest(func(r *smsc.UpdateSmscRequest) {
				r.PoweredBy = "<value>"
			}),
			AssertWith: nil,
		},
		{
			name:   "settings.bind.timeout=1000",
			target: "0",
			request: createUpdateSmscRequest(func(r *smsc.UpdateSmscRequest) {
				r.Settings.Bind = &smsc.Bind{
					Timeout: 1000,
				}
			}),
			AssertWith: nil,
		},
	})

}

func TestSmscApi_EditById_when_type_is_not_valid(t *testing.T) {
	executeEditByIdTest(t, assertUpdateSmscResponseWhenBadInput, []EditByTestConfiguration{
		{
			name:   "type=nil",
			target: "0",
			request: createUpdateSmscRequest(func(r *smsc.UpdateSmscRequest) {
				r.Type = ""
			}),
			AssertWith: nil,
		},
		{
			name:   "type=<value>",
			target: "0",
			request: createUpdateSmscRequest(func(r *smsc.UpdateSmscRequest) {
				r.Type = "<value>"
			}),
			AssertWith: nil,
		},
	})

}

func TestSmscApi_EditById_when_name_is_not_valid(t *testing.T) {
	executeEditByIdTest(t, assertUpdateSmscResponseWhenBadInput, []EditByTestConfiguration{
		{
			name:   "name=nil",
			target: "0",
			request: createUpdateSmscRequest(func(r *smsc.UpdateSmscRequest) {
				r.Name = ""
			}),
			AssertWith: nil,
		},
		{
			name:   "len(name)==2",
			target: "0",
			request: createUpdateSmscRequest(func(r *smsc.UpdateSmscRequest) {
				r.Name = stringWithCharset(2)
			}),
			AssertWith: nil,
		},
		{
			name:   "len(name)==51",
			target: "0",
			request: createUpdateSmscRequest(func(r *smsc.UpdateSmscRequest) {
				r.Name = stringWithCharset(51)
			}),
			AssertWith: nil,
		},
	})

}

func TestSmscApi_EditById_when_description_is_not_valid(t *testing.T) {
	executeEditByIdTest(t, assertUpdateSmscResponseWhenBadInput, []EditByTestConfiguration{
		{
			name:   "description=nil",
			target: "0",
			request: createUpdateSmscRequest(func(r *smsc.UpdateSmscRequest) {
				r.Description = ""
			}),
			AssertWith: nil,
		},
		{
			name:   "len(description)==1",
			target: "0",
			request: createUpdateSmscRequest(func(r *smsc.UpdateSmscRequest) {
				r.Description = stringWithCharset(1)
			}),
			AssertWith: nil,
		},
		{
			name:   "len(description)==256",
			target: "0",
			request: createUpdateSmscRequest(func(r *smsc.UpdateSmscRequest) {
				r.Name = stringWithCharset(256)
			}),
			AssertWith: nil,
		},
	})

}

func TestSmscApi_EditById_when_source_addr_from_settings_is_not_valid(t *testing.T) {
	executeEditByIdTest(t, assertUpdateSmscResponseWhenBadInput, []EditByTestConfiguration{
		{
			name:   "settings.source_address=google.com",
			target: "0",
			request: createUpdateSmscRequest(func(r *smsc.UpdateSmscRequest) {
				r.Settings.SourceAddr = "google.com"
			}),
			AssertWith: nil,
		},
	})

}

func TestSmscApi_EditById_when_host_from_settings_is_not_valid(t *testing.T) {
	executeEditByIdTest(t, assertUpdateSmscResponseWhenBadInput, []EditByTestConfiguration{
		{
			name:   "settings.host=nil",
			target: "0",
			request: createUpdateSmscRequest(func(r *smsc.UpdateSmscRequest) {
				r.Settings.Host = smsc.Host{}
			}),
			AssertWith: nil,
		},
		{
			name:   "host.username=nil",
			target: "0",
			request: createUpdateSmscRequest(func(r *smsc.UpdateSmscRequest) {
				r.Settings.Host.Username = ""
			}),
			AssertWith: nil,
		},
		{
			name:   "host.password=nil",
			target: "0",
			request: createUpdateSmscRequest(func(r *smsc.UpdateSmscRequest) {
				r.Settings.Host.Password = ""
			}),
			AssertWith: nil,
		},
		{
			name:   "host.address=nil",
			target: "0",
			request: createUpdateSmscRequest(func(r *smsc.UpdateSmscRequest) {
				r.Settings.Host.Address = ""
			}),
			AssertWith: nil,
		},
		{
			name:   "host.address not hostname_port",
			target: "0",
			request: createUpdateSmscRequest(func(r *smsc.UpdateSmscRequest) {
				r.Settings.Host.Address = "google.com"
			}),
			AssertWith: nil,
		},
	})

}

func TestSmscApi_EditById_when_bind_from_settings_is_not_valid(t *testing.T) {
	executeEditByIdTest(t, assertUpdateSmscResponseWhenBadInput, []EditByTestConfiguration{
		{
			name:   "bind.timeout=999",
			target: "0",
			request: createUpdateSmscRequest(func(r *smsc.UpdateSmscRequest) {
				r.Settings.Bind = &smsc.Bind{
					Timeout: 999,
				}
			}),
			AssertWith: nil,
		},
		{
			name:   "bind.timeout=0",
			target: "0",
			request: createUpdateSmscRequest(func(r *smsc.UpdateSmscRequest) {
				r.Settings.Bind = &smsc.Bind{
					Timeout: 0,
				}
			}),
			AssertWith: nil,
		},
	})
}

func TestSmscApi_EditById_when_enquire_from_settings_is_not_valid(t *testing.T) {
	executeEditByIdTest(t, assertUpdateSmscResponseWhenBadInput, []EditByTestConfiguration{
		{
			name:   "enquiry.link=999",
			target: "0",
			request: createUpdateSmscRequest(func(r *smsc.UpdateSmscRequest) {
				r.Settings.Enquire = &smsc.Enquire{
					Link:        999,
					LinkTimeout: 1000,
				}
			}),
			AssertWith: nil,
		},
		{
			name:   "enquiry.link=0",
			target: "0",
			request: createUpdateSmscRequest(func(r *smsc.UpdateSmscRequest) {
				r.Settings.Enquire = &smsc.Enquire{
					Link:        0,
					LinkTimeout: 1000,
				}
			}),
			AssertWith: nil,
		},
		{
			name:   "enquiry.link_timeout=999",
			target: "0",
			request: createUpdateSmscRequest(func(r *smsc.UpdateSmscRequest) {
				r.Settings.Enquire = &smsc.Enquire{
					Link:        1000,
					LinkTimeout: 999,
				}
			}),
			AssertWith: nil,
		},
		{
			name:   "enquiry.link_timeout=0",
			target: "0",
			request: createUpdateSmscRequest(func(r *smsc.UpdateSmscRequest) {
				r.Settings.Enquire = &smsc.Enquire{
					Link:        1000,
					LinkTimeout: 999,
				}
			}),
			AssertWith: nil,
		},
	})
}

func TestSmscApi_EditById_when_response_from_settings_is_not_valid(t *testing.T) {
	executeEditByIdTest(t, assertUpdateSmscResponseWhenBadInput, []EditByTestConfiguration{
		{
			name:   "response.timeout=999",
			target: "0",
			request: createUpdateSmscRequest(func(r *smsc.UpdateSmscRequest) {
				r.Settings.Response = &smsc.Response{
					Timeout: 999,
				}
			}),
			AssertWith: nil,
		},
		{
			name:   "response.timeout=0",
			target: "0",
			request: createUpdateSmscRequest(func(r *smsc.UpdateSmscRequest) {
				r.Settings.Response = &smsc.Response{
					Timeout: 0,
				}
			}),
			AssertWith: nil,
		},
	})
}

func TestSmscApi_EditById_when_merge_from_settings_is_not_valid(t *testing.T) {
	executeEditByIdTest(t, assertUpdateSmscResponseWhenBadInput, []EditByTestConfiguration{
		{
			name:   "merge.interval=999",
			target: "0",
			request: createUpdateSmscRequest(func(r *smsc.UpdateSmscRequest) {
				r.Settings.Merge = &smsc.Merge{
					Interval:        999,
					CleanupInterval: 1000,
				}
			}),
			AssertWith: nil,
		},
		{
			name:   "merge.interval=0",
			target: "0",
			request: createUpdateSmscRequest(func(r *smsc.UpdateSmscRequest) {
				r.Settings.Merge = &smsc.Merge{
					Interval:        0,
					CleanupInterval: 1000,
				}
			}),
			AssertWith: nil,
		},
		{
			name:   "merge.cleanup_interval=999",
			target: "0",
			request: createUpdateSmscRequest(func(r *smsc.UpdateSmscRequest) {
				r.Settings.Merge = &smsc.Merge{
					Interval:        1000,
					CleanupInterval: 999,
				}
			}),
			AssertWith: nil,
		},
		{
			name:   "merge.cleanup_interval=0",
			target: "0",
			request: createUpdateSmscRequest(func(r *smsc.UpdateSmscRequest) {
				r.Settings.Merge = &smsc.Merge{
					Interval:        1000,
					CleanupInterval: 0,
				}
			}),
			AssertWith: nil,
		},
	})
}

func createUpdateSmscRequest(f func(request *smsc.UpdateSmscRequest)) smsc.UpdateSmscRequest {
	r := &smsc.UpdateSmscRequest{
		PoweredBy: "raitonbl.com",
		Settings: smsc.Settings{
			Host: smsc.Host{
				Username: "admin",
				Password: "admin",
				Address:  "localhost:4000",
			},
		},
		Name:        "raitonbl",
		Description: "<description/>",
		Type:        smsc.TransmitterType,
	}
	if f != nil {
		f(r)
	}
	return *r
}

func executeEditByIdTest(t *testing.T, assertWith func(*testing.T, *httptest.ResponseRecorder, string, smsc.UpdateSmscRequest) error, arr []EditByTestConfiguration) {
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

func assertUpdateSmscResponseWhenBadInput(t *testing.T, w *httptest.ResponseRecorder, _ string, _ smsc.UpdateSmscRequest) error {
	require.Equal(t, 422, w.Code)
	zalandoProblem := make(map[string]any)
	if err := json.Unmarshal([]byte(w.Body.String()), &zalandoProblem); err != nil {
		return err
	}
	require.Equal(t, float64(422), zalandoProblem[zalandoStatusPath])
	require.Equal(t, httpValidationTitle, zalandoProblem[zalandoTitlePath])
	require.Equal(t, EditSmscOperationId, zalandoProblem[zalandoOperationIdPath])
	require.Equal(t, fmt.Sprintf(constraintViolationF, EditSmscOperationId), zalandoProblem[zalandoTypePath])
	require.Equal(t, fmt.Sprintf(httpValidationDetailWithLocationF, "body", EditSmscOperationId), zalandoProblem[zalandoDetailPath])
	return nil
}

func assertUpdateSmscResponseWhenOK(t *testing.T, w *httptest.ResponseRecorder, username string, smscRequest smsc.UpdateSmscRequest) error {
	require.Equal(t, 200, w.Code)
	smscResponse := smsc.UpdateSmscResponse{}
	if err := json.Unmarshal([]byte(w.Body.String()), &smscResponse); err != nil {
		return err
	}
	require.NotNil(t, smscResponse.LastUpdatedAt)
	require.Equal(t, username, smscResponse.LastUpdatedBy)
	return assertNewSmscResponseWhenOK(t, w, username, smscRequest, "")
}

func assertNewSmscResponseWhenOK(t *testing.T, w *httptest.ResponseRecorder, username string, smscRequest smsc.UpdateSmscRequest, alias string) error {
	require.Equal(t, 200, w.Code)
	smscResponse := smsc.NewSmscResponse{}
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
