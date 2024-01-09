package smpp

import (
	"github.com/delfimarime/hermes/services/smsc/internal/model"
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
