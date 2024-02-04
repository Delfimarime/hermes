package smsc

import (
	"fmt"
	"net/http"
)

const (
	ValidationZalandoProblemTitle = "Request not compliant"
)

type ValidationError struct {
	Field  string
	Rule   string
	Title  string
	Detail string
}

func (instance ValidationError) Error() string {
	return instance.Detail
}

func (instance ValidationError) GetTitle() string {
	if instance.Title == "" {
		return ValidationZalandoProblemTitle
	}
	return instance.Title
}

func (instance ValidationError) GetDetail() string {
	return instance.Detail
}

func (instance ValidationError) GetStatusCode() int {
	return http.StatusUnprocessableEntity
}

func (instance ValidationError) GetErrorType() string {
	if instance.Rule == "" {
		return fmt.Sprintf("/constraint-violation/%s", instance.Field)
	} else {
		return fmt.Sprintf("/constraint-violation/%s/%s", instance.Field, instance.Rule)
	}
}
