package restapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/delfimarime/hermes/services/smsc/pkg/restapi"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSmscApi_New(t *testing.T) {
	r := getGinEngine(&SmscApi{
		service: &TestSmsService{},
		getAuthenticatedUser: func(c *gin.Context) string {
			return "dmarime"
		},
	})
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
				Timeout: -1,
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
	fmt.Println(w.Code, w.Body.String())
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
			Settings:    request.Settings,
		},
		CreatedBy: username,
		CreatedAt: time.Now(),
		Id:        uuid.New().String(),
	}

	return response, nil
}
