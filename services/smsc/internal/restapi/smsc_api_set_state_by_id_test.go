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

type SetStateByIdTestConfiguration struct {
	name       string
	username   string
	target     string
	err        error
	request    restapi.UpdateSmscState
	AssertWith func(*testing.T, *httptest.ResponseRecorder, string, restapi.UpdateSmscState) error
}

func TestSmscApi_SetStateById(t *testing.T) {
	executeSetStateByIdTest(t, assertUpdateSmscStateResponseWhenNoContent, []SetStateByIdTestConfiguration{
		{
			name:   "value=ACTIVATED",
			target: "0",
			request: restapi.UpdateSmscState{
				Value: restapi.ActivatedSmscState,
			},
			AssertWith: nil,
		},
		{
			name:   "value=DEACTIVATED",
			target: "0",
			request: restapi.UpdateSmscState{
				Value: restapi.DeactivatedSmscState,
			},
			AssertWith: nil,
		},
	})
}

func TestSmscApi_SetStateById_when_value_not_valid(t *testing.T) {
	executeSetStateByIdTest(t, assertUpdateSmscStateResponseWhenBadInput, []SetStateByIdTestConfiguration{
		{
			name:       "value=nil",
			target:     "0",
			request:    restapi.UpdateSmscState{},
			AssertWith: nil,
		},
		{
			name:   "value=<value/>",
			target: "0",
			request: restapi.UpdateSmscState{
				Value: "<value/>",
			},
			AssertWith: nil,
		},
	})
}

func executeSetStateByIdTest(t *testing.T, assertWith func(*testing.T, *httptest.ResponseRecorder, string, restapi.UpdateSmscState) error, arr []SetStateByIdTestConfiguration) {
	if arr == nil {
		return
	}
	for _, definition := range arr {
		smscApi := &SmscApi{
			service: &TestSmscService{
				err: definition.err,
			},
		}
		username := "dmarime"
		if definition.username != "" {
			username = definition.username
		}
		r := getGinEngine(&HardCodedAuthenticator{username: username}, smscApi)
		t.Run(definition.name, func(t *testing.T) {
			requestData, _ := json.Marshal(definition.request)
			req, _ := http.NewRequest("PUT", smscEndpoint+"/"+definition.target+"/state", bytes.NewBuffer(requestData))
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

func assertUpdateSmscStateResponseWhenNoContent(t *testing.T, w *httptest.ResponseRecorder, username string, req restapi.UpdateSmscState) error {
	return assertResponseWhenNoContent[restapi.UpdateSmscState](t, w, username, req)
}

func assertResponseWhenNoContent[T any](t *testing.T, w *httptest.ResponseRecorder, _ string, _ T) error {
	fmt.Println(w.Code, w.Body.String())
	require.Equal(t, 204, w.Code)
	return nil
}

func assertUpdateSmscStateResponseWhenBadInput(t *testing.T, w *httptest.ResponseRecorder, username string, settings restapi.UpdateSmscState) error {
	return createAssertResponseWhenBadInput[restapi.UpdateSmscState](EditSmscStateOperationId)(t, w, username, settings)
}
