package restapi

import (
	"encoding/json"
	"fmt"
	"github.com/delfimarime/hermes/services/smsc/pkg/restapi/common"
	"github.com/delfimarime/hermes/services/smsc/pkg/restapi/smsc"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

type FindAllTestConfiguration struct {
	name           string
	username       string
	err            error
	searchCriteria *smsc.SearchCriteriaRequest
	response       common.Page[smsc.PaginatedSmsc]
	assertWith     func(*testing.T, *httptest.ResponseRecorder, common.Page[smsc.PaginatedSmsc]) error
}

func TestSmscApi_FindAll(t *testing.T) {
	executeFindAllTest(t, assertFindAllWhenOK, []FindAllTestConfiguration{
		{
			name:           "searchCriteria=nil",
			username:       "",
			err:            nil,
			searchCriteria: nil,
			response: common.Page[smsc.PaginatedSmsc]{
				Items: []smsc.PaginatedSmsc{
					{
						Id:          "1",
						Name:        "one",
						Alias:       "one",
						PoweredBy:   "one",
						Description: "one",
						Type:        smsc.TransmitterType,
					},
					{
						Id:          "2",
						Name:        "two",
						Alias:       "two",
						PoweredBy:   "two",
						Description: "two",
						Type:        smsc.ReceiverType,
					},
					{
						Id:          "3",
						Name:        "three",
						Alias:       "three",
						PoweredBy:   "three",
						Description: "three",
						Type:        smsc.TransceiverType,
					},
				},
				Self: "1",
				Next: "2",
			},
			assertWith: nil,
		},
	})
}

func TestSmscApi_FindAll_when_bad_search(t *testing.T) {
	executeFindAllTest(t, assertFindAllWhenBadInput, []FindAllTestConfiguration{
		{
			name:           "len(searchCriteria.s)>50",
			searchCriteria: &smsc.SearchCriteriaRequest{S: stringWithCharset(51)},
		},
		{
			name:           "len(searchCriteria.powered_by)>45",
			searchCriteria: &smsc.SearchCriteriaRequest{PoweredBy: stringWithCharset(46)},
		},
		{
			name:           "searchCriteria.state=<value/>",
			searchCriteria: &smsc.SearchCriteriaRequest{State: stringWithCharset(46)},
		},
		{
			name:           "searchCriteria.type=<value/>",
			searchCriteria: &smsc.SearchCriteriaRequest{Type: smsc.Type(stringWithCharset(46))},
		},
		{
			name:           "searchCriteria.sort=<value/>",
			searchCriteria: &smsc.SearchCriteriaRequest{Sort: stringWithCharset(46)},
		},
	})
}

func executeFindAllTest(t *testing.T, assertWith func(*testing.T, *httptest.ResponseRecorder, common.Page[smsc.PaginatedSmsc]) error, arr []FindAllTestConfiguration) {
	if arr == nil {
		return
	}
	for _, definition := range arr {
		smscApi := &SmscApi{
			service: &TestSmscService{
				err:                definition.err,
				smscSearchResponse: definition.response,
			},
		}
		username := "dmarime"
		if definition.username != "" {
			username = definition.username
		}
		r := getGinEngine(&HardCodedAuthenticator{username: username}, smscApi)
		t.Run(definition.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", getEndpointURL(smscEndpoint, definition.searchCriteria), nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			var err error
			if definition.assertWith == nil {
				err = assertWith(t, w, definition.response)
			} else {
				err = assertWith(t, w, definition.response)
			}
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func getEndpointURL(endpoint string, criteria *smsc.SearchCriteriaRequest) string {
	if criteria == nil {
		return endpoint
	}
	sb := strings.Builder{}
	sb.WriteString(endpoint)
	v := reflect.ValueOf(*criteria)
	typeOfS := v.Type()
	parametersAdded := false

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		tag := typeOfS.Field(i).Tag.Get("form")
		if !field.IsZero() {
			if parametersAdded {
				sb.WriteString("&")
			} else {
				sb.WriteString("?")
				parametersAdded = true
			}
			value := url.QueryEscape(fmt.Sprintf("%v", field.Interface()))
			sb.WriteString(fmt.Sprintf("%s=%s", tag, value))
		}
	}
	return sb.String()
}

func assertFindAllWhenOK(t *testing.T, w *httptest.ResponseRecorder, originalPage common.Page[smsc.PaginatedSmsc]) error {
	fmt.Println(w.Code, w.Body.String())
	require.Equal(t, 200, w.Code)
	page := common.ResponsePage[smsc.PaginatedSmsc]{}
	if err := json.Unmarshal([]byte(w.Body.String()), &page); err != nil {
		return err
	}
	require.NotNil(t, page.Items)
	require.Equal(t, originalPage.Self, page.Self)
	require.Equal(t, originalPage.Next, page.Next)
	require.Equal(t, len(originalPage.Items), len(page.Items))
	for i, item := range originalPage.Items {
		require.Equal(t, item.Id, page.Items[i].Id)
		require.Equal(t, item.Name, page.Items[i].Name)
		require.Equal(t, item.Type, page.Items[i].Type)
		require.Equal(t, item.Alias, page.Items[i].Alias)
		require.Equal(t, item.PoweredBy, page.Items[i].PoweredBy)
		require.Equal(t, item.Description, page.Items[i].Description)
	}
	return nil
}

func assertFindAllWhenBadInput(t *testing.T, w *httptest.ResponseRecorder, _ common.Page[smsc.PaginatedSmsc]) error {
	return createAssertResponseBindingWhenBadInput[any](GetSmscPageOperationId, "query")(t, w, "", nil)
}
