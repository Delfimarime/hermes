package outbound

import (
	"github.com/delfimarime/hermes/services/smsc/internal/model"
	"github.com/delfimarime/hermes/services/smsc/internal/repository/sdk"
)

type SequenceBasedSmppRepository struct {
	Err       error
	Arr       []model.Smpp
	Condition map[string]model.Condition
}

func (s SequenceBasedSmppRepository) FindAll() ([]model.Smpp, error) {
	if s.Err != nil {
		return nil, s.Err
	}
	if s.Arr == nil {
		return make([]model.Smpp, 0), nil
	}
	return s.Arr, nil
}

func (s SequenceBasedSmppRepository) FindById(id string) (model.Smpp, error) {
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

func (s SequenceBasedSmppRepository) GetConditionsFrom(id string) ([]model.Condition, error) {
	valueFrom, hasValue := s.Condition[id]
	if !hasValue {
		return nil, nil
	}
	return []model.Condition{valueFrom}, nil
}
