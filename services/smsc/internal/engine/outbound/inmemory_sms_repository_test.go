package outbound

import (
	"github.com/delfimarime/hermes/services/smsc/internal/model"
)

type InMemorySmsRepository struct {
	Err error
	Arr []model.Sms
}

func (i *InMemorySmsRepository) Save(sms *model.Sms) error {
	if i.Err != nil {
		return i.Err
	}
	if sms == nil {
		return nil
	}
	if i.Arr == nil {
		i.Arr = make([]model.Sms, 0)
	}
	i.Arr = append(i.Arr, *sms)
	return nil
}

func (i *InMemorySmsRepository) FindById(id string) (*model.Sms, error) {
	if i.Err != nil {
		return nil, i.Err
	}
	if i.Arr == nil {
		return nil, nil
	}
	for _, each := range i.Arr {
		if each.Id == id {
			v := each
			return &v, nil
		}
	}
	return nil, nil
}
