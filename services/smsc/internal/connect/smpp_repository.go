package connect

import "github.com/delfimarime/hermes/services/smsc/internal/model"

type SmppRepository interface {
	FindAll() ([]model.Smpp, error)
}
