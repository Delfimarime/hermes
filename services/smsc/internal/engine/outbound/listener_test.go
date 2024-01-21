package outbound

import (
	goContext "context"
	"encoding/json"
	"fmt"
	"github.com/delfimarime/hermes/services/smsc/internal/context"
	"github.com/delfimarime/hermes/services/smsc/internal/repository/sdk"
	"github.com/delfimarime/hermes/services/smsc/internal/smpp"
	"github.com/delfimarime/hermes/services/smsc/pkg/asyncapi"
	"github.com/delfimarime/hermes/services/smsc/pkg/config"
	"github.com/google/uuid"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"testing"
)

// TestAppConfig contains configuration for setting up the test application.
type TestAppConfig struct {
	SmsRepository    sdk.SmsRepository
	SmppRepository   sdk.SmppRepository
	Configuration    *config.Configuration
	ConnectorManager smpp.ConnectorManager
}

func NewApp(t *testing.T, cfg TestAppConfig, fxOptions ...fx.Option) *fxtest.App {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal(err)
	}
	zap.ReplaceGlobals(logger)
	configuration := config.Configuration{
		Smsc: config.Smsc{},
		Logger: config.Logger{
			Level: "INFO",
		},
		OpenTelemetry: nil,
	}
	if cfg.Configuration != nil {
		configuration = *cfg.Configuration
	}
	options := []fx.Option{
		fx.WithLogger(func() fxevent.Logger {
			return &fxevent.ZapLogger{Logger: logger}
		}),
		fx.Module("context", fx.Provide(func() context.Context {
			return &context.AppContext{
				Configuration: configuration,
			}
		})),
		fx.Provide(
			func() sdk.SmppRepository {
				if cfg.SmppRepository != nil {
					return cfg.SmppRepository
				}
				return &SequenceBasedSmsRepository{}
			},
			func() sdk.SmsRepository {
				if cfg.SmsRepository != nil {
					return cfg.SmsRepository
				}
				return &InMemorySmsRepository{}
			},
			func() smpp.ConnectorManager {
				if cfg.ConnectorManager != nil {
					return cfg.ConnectorManager
				}
				return &TestConnectorManager{}
			},
		),
		GetUberFxModule(), // outbound
	}
	options = append(options, fxOptions...)
	return fxtest.New(t, options...)
}

func TestSmppSendSmsRequestListener_ListenTo_when_no_connectors(t *testing.T) {
	req := asyncapi.SendSmsRequest{
		Id:      uuid.New().String(),
		To:      "+25884990XXXX",
		Tags:    []string{"banking", "onboard"},
		From:    "onboard-workflow",
		Content: "Welcome to our bank",
	}
	var resp asyncapi.SendSmsResponse
	app := NewApp(t, TestAppConfig{},
		fx.Invoke(func(lifecycle fx.Lifecycle, l SendSmsRequestHandler) {
			lifecycle.Append(fx.Hook{
				OnStart: func(ctx goContext.Context) error {
					response, err := l.Accept(req)
					if err != nil {
						return err
					}
					resp = response
					return nil
				},
			})
		}))
	defer app.RequireStart().RequireStop()
	b, _ := json.Marshal(resp)
	fmt.Println(string(b))
}

type TestConnectorManager struct {
	seq []smpp.Connector
}

func (t *TestConnectorManager) Close() error {
	return nil
}

func (t *TestConnectorManager) GetList() []smpp.Connector {
	return t.seq
}

func (t *TestConnectorManager) AfterPropertiesSet() error {
	return nil
}

func (t *TestConnectorManager) GetById(id string) smpp.Connector {
	if t.seq == nil {
		return nil
	}
	for _, each := range t.seq {
		if each.GetId() == id {
			return each
		}
	}
	return nil
}

type TestConnector struct {
	err            error
	trackDelivery  bool
	id             string
	connectionType string
	alias          string
	state          smpp.State
}

func (t TestConnector) GetId() string {
	return t.id
}

func (t TestConnector) GetType() string {
	return t.connectionType
}

func (t TestConnector) GetAlias() string {
	return t.alias
}

func (t TestConnector) GetState() smpp.State {
	return t.state
}

func (t TestConnector) IsTrackingDelivery() bool {
	return t.trackDelivery
}

func (t TestConnector) SendMessage(_, _ string) (smpp.SendMessageResponse, error) {
	if t.err != nil {
		return smpp.SendMessageResponse{}, t.err
	}
	return smpp.SendMessageResponse{
		Id: uuid.New().String(),
	}, nil
}
