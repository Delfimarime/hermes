package smpp

import (
	goContext "context"
	"github.com/delfimarime/hermes/services/smsc/internal/context"
	"go.opentelemetry.io/otel"
	"go.uber.org/fx"
)

func GetUberFxModule() fx.Option {
	return fx.Module("smpp",
		fx.Provide(
			func(smsEventListener SmsEventListener) (*PduListenerFactory, error) {
				return NewPduListenerFactory(otel.Meter("hermes_smsc_listener"), smsEventListener)
			},
			func(
				ctx context.Context,
				repository Repository,
				pduListenerFactory *PduListenerFactory,
			) ConnectorManager {
				return &SimpleConnectorManager{
					repository:         repository,
					pduListenerFactory: pduListenerFactory,
					connectors:         make([]Connector, 0),
					configuration:      ctx.GetConfiguration(),
					connectorCache:     make(map[string]*SimpleConnector),
				}
			},
		),
		fx.Invoke(func(lifecycle fx.Lifecycle, cm ConnectorManager) {
			lifecycle.Append(
				fx.Hook{
					OnStart: func(ctx goContext.Context) error {
						return cm.AfterPropertiesSet()
					},
					OnStop: func(ctx goContext.Context) error {
						return cm.Close()
					},
				},
			)
		}),
	)
}
