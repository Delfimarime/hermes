package smpp

import (
	"github.com/delfimarime/hermes/services/smsc/internal/model"
	"github.com/delfimarime/hermes/services/smsc/internal/repository/sdk"
)

type SequenceBasedSmsRepository struct {
	Err error
	Arr []model.Smpp
}

func (s SequenceBasedSmsRepository) FindAll() ([]model.Smpp, error) {
	if s.Err != nil {
		return nil, s.Err
	}
	if s.Arr == nil {
		return make([]model.Smpp, 0), nil
	}
	return s.Arr, nil
}

func (s SequenceBasedSmsRepository) FindById(id string) (model.Smpp, error) {
	if s.Arr != nil {
		for _, each := range s.Arr {
			if each.Id == id {
				return each, nil
			}
		}
	}
	return model.Smpp{}, &sdk.EntityNotFoundError{
		Id:   id,
		Type: "Smpp",
	}
}

func (s SequenceBasedSmsRepository) GetConditionsFrom(id string) ([]model.Condition, error) {
	return nil, nil
}
