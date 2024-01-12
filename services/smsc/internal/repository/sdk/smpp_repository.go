package sdk

import (
	"fmt"
	"github.com/delfimarime/hermes/services/smsc/internal/model"
)

type Repository interface {
	FindAll() ([]model.Smpp, error)
	FindById(id string) (model.Smpp, error)
	GetConditionsFrom(id string) ([]model.Condition, error)
}

type EntityNotFoundError struct {
	Id   string
	Type string
}

func (instance *EntityNotFoundError) Error() string {
	return fmt.Sprintf("%s[id=%s] cannot be retrived from Repository",
		instance.Type, instance.Id)
}
