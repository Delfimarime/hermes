package metric

import (
	goContext "context"
	"github.com/delfimarime/hermes/services/smsc/internal/context"
	"go.uber.org/fx"
)

func GetUberFxModule() fx.Option {
	return fx.Module("openTelemetry",
		fx.Invoke(func(ctx context.Context) {
			Configure(goContext.TODO(), ctx.GetConfiguration())
		}),
	)
}
