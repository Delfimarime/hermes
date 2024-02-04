package sdk

import (
	"github.com/delfimarime/hermes/services/smsc/internal/model"
)

type SmppRepository interface {
	Save(smpp model.Smpp) error
	FindAll() ([]model.Smpp, error)
	FindById(id string) (model.Smpp, error)
	GetConditionsFrom(id string) ([]model.Condition, error)
}
