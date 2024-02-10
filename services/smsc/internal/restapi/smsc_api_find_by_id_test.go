package restapi

import (
	"fmt"
	"github.com/delfimarime/hermes/services/smsc/internal/repository/sdk"
	"github.com/delfimarime/hermes/services/smsc/pkg/restapi"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type FindByIdTestConfiguration struct {
	name       string
	username   string
	target     string
	err        error
	response   restapi.GetSmscByIdResponse
	AssertWith func(*testing.T, *httptest.ResponseRecorder, string) error
}

func TestSmscApi_FindById(t *testing.T) {
	id := uuid.New().String()
	response := restapi.GetSmscByIdResponse{
		NewSmscResponse: restapi.NewSmscResponse{
			NewSmscRequest: createNewSmscSettingsRequest(nil),
			Id:             id,
			CreatedBy:      "dmarime",
			CreatedAt:      time.Now(),
		},
	}
	executeFindByIdTest(t, func(t *testing.T, recorder *httptest.ResponseRecorder, s string) error {
		return assertNewSmscResponseWhenOK(t, recorder, s, response.NewSmscRequest.UpdateSmscRequest, response.NewSmscRequest.Alias)
	}, []FindByIdTestConfiguration{
		{
			target:     id,
			AssertWith: nil,
			response:   response,
			name:       fmt.Sprintf("id=%s", id),
		},
	})
}

func TestSmscApi_FindById_when_entity_doesnt_exist(t *testing.T) {
	executeFindByIdTest(t, assertFindByIdResponseWhenNoSuchEntity, []FindByIdTestConfiguration{
		{
			target:     "1",
			AssertWith: nil,
			name:       "id=1",
			err: &sdk.EntityNotFoundError{
				Id:   "1",
				Type: "SMSC",
			},
		},
	})
}

func executeFindByIdTest(t *testing.T, assertWith func(*testing.T, *httptest.ResponseRecorder, string) error, arr []FindByIdTestConfiguration) {
	if arr == nil {
		return
	}
	for _, definition := range arr {
		smscApi := &SmscApi{
			service: &TestSmscService{
				err:                 definition.err,
				getSmscByIdResponse: definition.response,
			},
		}
		username := "dmarime"
		if definition.username != "" {
			username = definition.username
		}
		r := getGinEngine(&HardCodedAuthenticator{username: username}, smscApi)
		t.Run(definition.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", smscEndpoint+"/"+definition.target, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			var err error
			if definition.AssertWith == nil {
				err = assertWith(t, w, username)
			} else {
				err = assertWith(t, w, username)
			}
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func assertFindByIdResponseWhenNoSuchEntity(t *testing.T, w *httptest.ResponseRecorder, username string) error {
	return createAssertResponseWhenNoSuchEntity[any](GetSmscOperationId)(t, w, username, nil)
}
