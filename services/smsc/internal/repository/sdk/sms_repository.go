package sdk

import "github.com/delfimarime/hermes/services/smsc/internal/model"

type SmsRepository interface {
	Save(sms *model.Sms) error
	FindById(id string) (*model.Sms, error)
}
