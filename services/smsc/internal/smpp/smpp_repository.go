package smpp

import "github.com/delfimarime/hermes/services/smsc/internal/model"

type Repository interface {
	FindAll() ([]model.Smpp, error)
}
