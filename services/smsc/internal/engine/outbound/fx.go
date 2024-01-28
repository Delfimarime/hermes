package outbound

import (
	goContext "context"
	"github.com/delfimarime/hermes/services/smsc/internal/repository/sdk"
	sdk2 "github.com/delfimarime/hermes/services/smsc/internal/sdk"
	"github.com/delfimarime/hermes/services/smsc/internal/smpp"
	"go.uber.org/fx"
)

func GetUberFxModule() fx.Option {
	return fx.Module("outbound",
		fx.Provide(
			func(
				cm smpp.ConnectorManager,
				smsRepository sdk.SmsRepository,
				smppRepository sdk.SmppRepository,
			) sdk2.SendSmsRequestHandler {
				return &SmppSendSmsRequestHandler{
					manager:        cm,
					smsRepository:  smsRepository,
					smppRepository: smppRepository,
					predicate:      make(map[string]SendSmsRequestPredicate),
				}
			},
		),
		fx.Invoke(
			func(lifecycle fx.Lifecycle, s sdk2.SendSmsRequestHandler) {
				lifecycle.Append(
					fx.Hook{
						OnStop: func(ctx goContext.Context) error {
							return s.Close()
						},
						OnStart: func(ctx goContext.Context) error {
							return s.AfterPropertiesSet()
						},
					},
				)
			},
		),
	)
}
