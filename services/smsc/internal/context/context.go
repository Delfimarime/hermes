package context

import "github.com/delfimarime/hermes/services/smsc/pkg/config"

type Context interface {
	GetConfiguration() config.Configuration
}
