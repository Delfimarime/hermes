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

type EditByTestSettingsConfiguration struct {
	name       string
	username   string
	target     string
	err        error
	request    restapi.UpdateSmscSettingsRequest
	AssertWith func(*testing.T, *httptest.ResponseRecorder, string, restapi.UpdateSmscSettingsRequest) error
}

func TestSmscApi_EditSettingsById(t *testing.T) {
	executeEditSettingsByIdTest(t, assertUpdateSmscResponseWhenNoContent, []EditByTestSettingsConfiguration{
		{
			name:       "unmodified",
			target:     "0",
			request:    createUpdateSmscSettingsRequest(nil),
			AssertWith: nil,
		},
		{
			name:   "bind.timeout=1000",
			target: "0",
			request: createUpdateSmscSettingsRequest(func(r *restapi.UpdateSmscSettingsRequest) {
				r.Bind = &restapi.Bind{
					Timeout: 1000,
				}
			}),
			AssertWith: nil,
		},
	})
}

func TestSmscApi_EditSettingsById_when_source_addr_from_settings_is_not_valid(t *testing.T) {
	executeEditSettingsByIdTest(t, assertUpdateSmscSettingsResponseWhenBadInput, []EditByTestSettingsConfiguration{
		{
			name:   "source_address=google.com",
			target: "0",
			request: createUpdateSmscSettingsRequest(func(r *restapi.UpdateSmscSettingsRequest) {
				r.SourceAddr = "google.com"
			}),
			AssertWith: nil,
		},
	})

}

func TestSmscApi_EditSettingsById_when_host_from_settings_is_not_valid(t *testing.T) {
	executeEditSettingsByIdTest(t, assertUpdateSmscSettingsResponseWhenBadInput, []EditByTestSettingsConfiguration{
		{
			name:   "settings.host=nil",
			target: "0",
			request: createUpdateSmscSettingsRequest(func(r *restapi.UpdateSmscSettingsRequest) {
				r.Host = restapi.Host{}
			}),
			AssertWith: nil,
		},
		{
			name:   "host.username=nil",
			target: "0",
			request: createUpdateSmscSettingsRequest(func(r *restapi.UpdateSmscSettingsRequest) {
				r.Host.Username = ""
			}),
			AssertWith: nil,
		},
		{
			name:   "host.password=nil",
			target: "0",
			request: createUpdateSmscSettingsRequest(func(r *restapi.UpdateSmscSettingsRequest) {
				r.Host.Password = ""
			}),
			AssertWith: nil,
		},
		{
			name:   "host.address=nil",
			target: "0",
			request: createUpdateSmscSettingsRequest(func(r *restapi.UpdateSmscSettingsRequest) {
				r.Host.Address = ""
			}),
			AssertWith: nil,
		},
		{
			name:   "host.address not hostname_port",
			target: "0",
			request: createUpdateSmscSettingsRequest(func(r *restapi.UpdateSmscSettingsRequest) {
				r.Host.Address = "google.com"
			}),
			AssertWith: nil,
		},
	})

}

func TestSmscApi_EditSettingsById_when_bind_from_settings_is_not_valid(t *testing.T) {
	executeEditSettingsByIdTest(t, assertUpdateSmscSettingsResponseWhenBadInput, []EditByTestSettingsConfiguration{
		{
			name:   "bind.timeout=999",
			target: "0",
			request: createUpdateSmscSettingsRequest(func(r *restapi.UpdateSmscSettingsRequest) {
				r.Bind = &restapi.Bind{
					Timeout: 999,
				}
			}),
			AssertWith: nil,
		},
		{
			name:   "bind.timeout=0",
			target: "0",
			request: createUpdateSmscSettingsRequest(func(r *restapi.UpdateSmscSettingsRequest) {
				r.Bind = &restapi.Bind{
					Timeout: 0,
				}
			}),
			AssertWith: nil,
		},
	})
}

func TestSmscApi_EditSettingsById_when_enquire_from_settings_is_not_valid(t *testing.T) {
	executeEditSettingsByIdTest(t, assertUpdateSmscSettingsResponseWhenBadInput, []EditByTestSettingsConfiguration{
		{
			name:   "enquiry.link=999",
			target: "0",
			request: createUpdateSmscSettingsRequest(func(r *restapi.UpdateSmscSettingsRequest) {
				r.Enquire = &restapi.Enquire{
					Link:        999,
					LinkTimeout: 1000,
				}
			}),
			AssertWith: nil,
		},
		{
			name:   "enquiry.link=0",
			target: "0",
			request: createUpdateSmscSettingsRequest(func(r *restapi.UpdateSmscSettingsRequest) {
				r.Enquire = &restapi.Enquire{
					Link:        0,
					LinkTimeout: 1000,
				}
			}),
			AssertWith: nil,
		},
		{
			name:   "enquiry.link_timeout=999",
			target: "0",
			request: createUpdateSmscSettingsRequest(func(r *restapi.UpdateSmscSettingsRequest) {
				r.Enquire = &restapi.Enquire{
					Link:        1000,
					LinkTimeout: 999,
				}
			}),
			AssertWith: nil,
		},
		{
			name:   "enquiry.link_timeout=0",
			target: "0",
			request: createUpdateSmscSettingsRequest(func(r *restapi.UpdateSmscSettingsRequest) {
				r.Enquire = &restapi.Enquire{
					Link:        1000,
					LinkTimeout: 999,
				}
			}),
			AssertWith: nil,
		},
	})
}

func TestSmscApi_EditSettingsById_when_response_from_settings_is_not_valid(t *testing.T) {
	executeEditSettingsByIdTest(t, assertUpdateSmscSettingsResponseWhenBadInput, []EditByTestSettingsConfiguration{
		{
			name:   "response.timeout=999",
			target: "0",
			request: createUpdateSmscSettingsRequest(func(r *restapi.UpdateSmscSettingsRequest) {
				r.Response = &restapi.Response{
					Timeout: 999,
				}
			}),
			AssertWith: nil,
		},
		{
			name:   "response.timeout=0",
			target: "0",
			request: createUpdateSmscSettingsRequest(func(r *restapi.UpdateSmscSettingsRequest) {
				r.Response = &restapi.Response{
					Timeout: 0,
				}
			}),
			AssertWith: nil,
		},
	})
}

func TestSmscApi_EditSettingsById_when_merge_from_settings_is_not_valid(t *testing.T) {
	executeEditSettingsByIdTest(t, assertUpdateSmscSettingsResponseWhenBadInput, []EditByTestSettingsConfiguration{
		{
			name:   "merge.interval=999",
			target: "0",
			request: createUpdateSmscSettingsRequest(func(r *restapi.UpdateSmscSettingsRequest) {
				r.Merge = &restapi.Merge{
					Interval:        999,
					CleanupInterval: 1000,
				}
			}),
			AssertWith: nil,
		},
		{
			name:   "merge.interval=0",
			target: "0",
			request: createUpdateSmscSettingsRequest(func(r *restapi.UpdateSmscSettingsRequest) {
				r.Merge = &restapi.Merge{
					Interval:        0,
					CleanupInterval: 1000,
				}
			}),
			AssertWith: nil,
		},
		{
			name:   "merge.cleanup_interval=999",
			target: "0",
			request: createUpdateSmscSettingsRequest(func(r *restapi.UpdateSmscSettingsRequest) {
				r.Merge = &restapi.Merge{
					Interval:        1000,
					CleanupInterval: 999,
				}
			}),
			AssertWith: nil,
		},
		{
			name:   "merge.cleanup_interval=0",
			target: "0",
			request: createUpdateSmscSettingsRequest(func(r *restapi.UpdateSmscSettingsRequest) {
				r.Merge = &restapi.Merge{
					Interval:        1000,
					CleanupInterval: 0,
				}
			}),
			AssertWith: nil,
		},
	})
}

func createUpdateSmscSettingsRequest(f func(request *restapi.UpdateSmscSettingsRequest)) restapi.UpdateSmscSettingsRequest {
	r := &restapi.UpdateSmscSettingsRequest{
		Host: restapi.Host{
			Username: "admin",
			Password: "admin",
			Address:  "localhost:4000",
		},
	}
	if f != nil {
		f(r)
	}
	return *r
}

func executeEditSettingsByIdTest(t *testing.T, assertWith func(*testing.T, *httptest.ResponseRecorder, string, restapi.UpdateSmscSettingsRequest) error, arr []EditByTestSettingsConfiguration) {
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
			req, _ := http.NewRequest("PUT", smscEndpoint+"/"+definition.target+"/settings", bytes.NewBuffer(requestData))
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

func assertUpdateSmscSettingsResponseWhenBadInput(t *testing.T, w *httptest.ResponseRecorder, username string, settings restapi.UpdateSmscSettingsRequest) error {
	return createAssertResponseWhenBadInput[restapi.UpdateSmscSettingsRequest](EditSmscSettingsId)(t, w, username, settings)
}

func assertUpdateSmscResponseWhenNoContent(t *testing.T, w *httptest.ResponseRecorder, _ string, _ restapi.UpdateSmscSettingsRequest) error {
	fmt.Println(w.Code, w.Body.String())
	require.Equal(t, 204, w.Code)
	return nil
}

func createAssertResponseWhenBadInput[T any](operationId string) func(t *testing.T, w *httptest.ResponseRecorder, _ string, _ T) error {
	return createAssertResponseBindingWhenBadInput[T](operationId, "body")
}

func createAssertResponseBindingWhenBadInput[T any](operationId string, binding string) func(t *testing.T, w *httptest.ResponseRecorder, _ string, _ T) error {
	return func(t *testing.T, w *httptest.ResponseRecorder, _ string, _ T) error {
		fmt.Println(w.Code, w.Body.String())
		require.Equal(t, 422, w.Code)
		zalandoProblem := make(map[string]any)
		if err := json.Unmarshal([]byte(w.Body.String()), &zalandoProblem); err != nil {
			return err
		}
		require.Equal(t, float64(422), zalandoProblem[zalandoStatusPath])
		require.Equal(t, httpValidationTitle, zalandoProblem[zalandoTitlePath])
		require.Equal(t, operationId, zalandoProblem[zalandoOperationIdPath])
		require.Equal(t, fmt.Sprintf(constraintViolationF, operationId), zalandoProblem[zalandoTypePath])
		require.Equal(t, fmt.Sprintf(httpValidationDetailWithLocationF, binding, operationId), zalandoProblem[zalandoDetailPath])
		return nil
	}
}
