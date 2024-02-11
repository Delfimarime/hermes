package restapi

import (
	"encoding/json"
	"fmt"
	"github.com/delfimarime/hermes/services/smsc/internal/repository/sdk"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

type RemoveByIdTestConfiguration struct {
	name       string
	username   string
	target     string
	err        error
	AssertWith func(*testing.T, *httptest.ResponseRecorder, string) error
}

func TestSmscApi_RemoveById(t *testing.T) {
	executeRemoveByIdTest(t, assertRemoveByIdResponseWhenNoContent, []RemoveByIdTestConfiguration{
		{
			name:       "id=0",
			target:     "0",
			AssertWith: nil,
		},
	})
}

func TestSmscApi_RemoveById_when_entity_doesnt_exist(t *testing.T) {
	executeRemoveByIdTest(t, assertRemoveByIdResponseWhenNoSuchEntity, []RemoveByIdTestConfiguration{
		{
			target:     "0",
			AssertWith: nil,
			name:       "id=1",
			err: &sdk.EntityNotFoundError{
				Id:   "1",
				Type: "SMSC",
			},
		},
	})
}

func executeRemoveByIdTest(t *testing.T, assertWith func(*testing.T, *httptest.ResponseRecorder, string) error, arr []RemoveByIdTestConfiguration) {
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
		r := getGinEngine(&HardCodedAuthenticator{username: username}, smscApi, nil)
		t.Run(definition.name, func(t *testing.T) {
			req, _ := http.NewRequest("DELETE", smscEndpoint+"/"+definition.target, nil)
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

func assertRemoveByIdResponseWhenNoContent(t *testing.T, w *httptest.ResponseRecorder, _ string) error {
	fmt.Println(w.Code, w.Body.String())
	require.Equal(t, 204, w.Code)
	return nil
}

func assertRemoveByIdResponseWhenNoSuchEntity(t *testing.T, w *httptest.ResponseRecorder, username string) error {
	return createAssertResponseWhenNoSuchEntity[any](RemoveSmscOperationId)(t, w, username, nil)
}

func createAssertResponseWhenNoSuchEntity[T any](operationId string) func(t *testing.T, w *httptest.ResponseRecorder, _ string, _ T) error {
	return func(t *testing.T, w *httptest.ResponseRecorder, _ string, _ T) error {
		fmt.Println(w.Code, w.Body.String())
		require.Equal(t, 404, w.Code)
		zalandoProblem := make(map[string]any)
		if err := json.Unmarshal([]byte(w.Body.String()), &zalandoProblem); err != nil {
			return err
		}
		require.Equal(t, float64(404), zalandoProblem[zalandoStatusPath])
		require.Equal(t, PageNotFoundTitle, zalandoProblem[zalandoTitlePath])
		require.Empty(t, zalandoProblem[zalandoOperationIdPath])
		require.Equal(t, PageNotFoundType, zalandoProblem[zalandoTypePath])
		require.Equal(t, PageNotFoundDetail, zalandoProblem[zalandoDetailPath])
		return nil
	}
}
