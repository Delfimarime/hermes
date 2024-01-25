package outbound

import (
	goContext "context"
	"github.com/delfimarime/hermes/services/smsc/internal/context"
	"github.com/delfimarime/hermes/services/smsc/internal/model"
	"github.com/delfimarime/hermes/services/smsc/internal/repository/sdk"
	"github.com/delfimarime/hermes/services/smsc/internal/smpp"
	"github.com/delfimarime/hermes/services/smsc/pkg/asyncapi"
	"github.com/delfimarime/hermes/services/smsc/pkg/common"
	"github.com/delfimarime/hermes/services/smsc/pkg/config"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
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
				return &SequenceBasedSmppRepository{}
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

func TestSmppSendSmsRequestListener_Accept_when_no_connectors(t *testing.T) {
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
	require.Equal(t, req.Id, resp.Id)
	require.Nil(t, resp.Smsc)
	require.Nil(t, resp.CanceledAt)
	require.Empty(t, resp.Delivery)
	require.NotNil(t, resp.Problem)
	require.Equal(t, CannotSendSmsRequestProblemType, resp.Problem.Type)
	require.Equal(t, CannotSendSmsRequestProblemTitle, resp.Problem.Title)
	require.Equal(t, CannotSendSmsRequestProblemDetail, resp.Problem.Detail)
}

// Expects a response that informs that there aren't any connector capable of sending such message
func TestSmppSendSmsRequestListener_Accept_when_no_connector_rule_match(t *testing.T) {
	req := asyncapi.SendSmsRequest{
		Id:      uuid.New().String(),
		To:      "+25884990XXXX",
		Tags:    []string{"banking", "onboard"},
		From:    "onboard-workflow",
		Content: "Welcome to our bank",
	}
	connectorId := uuid.New().String()
	var resp asyncapi.SendSmsResponse
	app := NewApp(t, TestAppConfig{
		ConnectorManager: &TestConnectorManager{
			seq: []smpp.Connector{
				TestConnector{
					trackDelivery:  false,
					alias:          "vm",
					connectionType: "TRANSMITTER",
					id:             connectorId,
					state:          smpp.ReadyConnectorLifecycleState,
				},
			},
		},
		SmppRepository: SequenceBasedSmppRepository{
			Arr: []model.Smpp{
				{
					Id:          connectorId,
					Name:        "vodacom",
					Description: "<DESCRIPTION/>",
					Type:        "TRANSMITTER",
					Alias:       "vm",
				},
			},
			Condition: map[string]model.Condition{
				connectorId: {
					Predicate: model.Predicate{
						EqualTo: common.ToStrPointer("25884XXX"),
						Subject: common.ToStrPointer(string(Destination)),
					},
				},
			},
		},
	},
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
	require.Equal(t, req.Id, resp.Id)
	require.Nil(t, resp.Smsc)
	require.Nil(t, resp.CanceledAt)
	require.Empty(t, resp.Delivery)
	require.NotNil(t, resp.Problem)
	require.Equal(t, CannotSendSmsRequestProblemType, resp.Problem.Type)
	require.Equal(t, CannotSendSmsRequestProblemTitle, resp.Problem.Title)
	require.Equal(t, CannotSendSmsRequestProblemDetail, resp.Problem.Detail)
}

// Expects a response that informs message has been sent successfully
func TestSmppSendSmsRequestListener_Accept_when_a_connector_can_send_the_message(t *testing.T) {
}

// Expects a response that informs message has been sent successfully
func TestSmppSendSmsRequestListener_Accept_when_the_first_connector_is_unavailable_and_a_secondary_is_available(t *testing.T) {
}

// Expects an error[Service-Not-Available]
func TestSmppSendSmsRequestListener_Accept_when_a_connectors_are_unavailable(t *testing.T) {
}

// Expects a response that informs that the message have been canceled
func TestSmppSendSmsRequestListener_Accept_when_the_request_is_canceled(t *testing.T) {
}

// Expects a response that informs that the message has been sent before [rebuilding the response]
func TestSmppSendSmsRequestListener_Accept_when_the_request_has_been_sent_previous(t *testing.T) {
}

// Expects a response that informs that the message has terminated with an error [rebuilding the response]
func TestSmppSendSmsRequestListener_Accept_when_the_request_has_been_previous_processed_with_error(t *testing.T) {
}
