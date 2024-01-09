package context

import "github.com/delfimarime/hermes/services/smsc/pkg/config"

type AppContext struct {
	Configuration config.Configuration
}

func (instance *AppContext) GetConfiguration() config.Configuration {
	return instance.Configuration
}
