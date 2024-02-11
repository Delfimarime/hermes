package publish

import (
	goContext "context"
	"errors"
	asyncapi2 "github.com/delfimarime/hermes/services/smsc/internal/asyncapi"
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
	"time"
)

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

func TestSmppSendSmsRequestListener_Accept_when_repo_has_error(t *testing.T) {
	req := asyncapi.SendSmsRequest{
		Id:      uuid.New().String(),
		To:      "+25884990XXXX",
		Tags:    []string{"banking", "onboard"},
		From:    "onboard-workflow",
		Content: "Welcome to our bank",
	}
	var caught error
	app := NewApp(t, TestAppConfig{
		SmsRepository: &InMemorySmsRepository{Err: errors.New("<exception/>")},
	},
		fx.Invoke(func(lifecycle fx.Lifecycle, l asyncapi2.SendSmsRequestHandler) {
			lifecycle.Append(fx.Hook{
				OnStart: func(ctx goContext.Context) error {
					_, caught = l.Accept(req)
					return nil
				},
			})
		}))
	defer app.RequireStart().RequireStop()
	_, isCannotHandleSendSmsRequestError := caught.(*CannotHandleSendSmsRequestError)
	require.NotNil(t, caught)
	require.True(t, isCannotHandleSendSmsRequestError)
	require.Equal(t, GenericProblemDetail, caught.Error())
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
		fx.Invoke(func(lifecycle fx.Lifecycle, l asyncapi2.SendSmsRequestHandler) {
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
		fx.Invoke(func(lifecycle fx.Lifecycle, l asyncapi2.SendSmsRequestHandler) {
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

func TestSmppSendSmsRequestListener_Accept_when_a_connector_can_send_the_message(t *testing.T) {
	testCases := []struct {
		name             string
		trackDelivery    bool
		expectedDelivery asyncapi.DeliveryStrategy
	}{
		{"TrackDelivery", true, asyncapi.TrackingDeliveryStrategy},
		{"NoTrackDelivery", false, asyncapi.NotTrackingDeliveryStrategy},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			destination := "+25884990XXXX"
			req := newSendSmsRequest(destination)
			smppId := "e5104d81-7acb-4808-9655-091564a9bb35"
			smppRepo, smppManager := getSmppConfig([]model.Smpp{
				{
					Id:          smppId,
					Name:        "vodacom",
					Description: "<Description/>",
					PoweredBy:   "",
					Alias:       "vmz",
					Type:        model.TransmitterType,
					Settings: model.Settings{
						Delivery: &model.Delivery{AwaitReport: tc.trackDelivery},
					},
				},
			}, map[string]model.Condition{
				smppId: {
					Predicate: model.Predicate{
						EqualTo: common.ToStrPointer(destination),
						Subject: common.ToStrPointer(string(Destination)),
					},
				},
			}, nil)

			var resp asyncapi.SendSmsResponse
			app := NewApp(t, TestAppConfig{
				SmppRepository:   smppRepo,
				ConnectorManager: smppManager,
			},
				fx.Invoke(func(lifecycle fx.Lifecycle, l asyncapi2.SendSmsRequestHandler) {
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
			require.NotNil(t, resp.Smsc)
			require.Equal(t, smppId, resp.Smsc.Id)
			require.Nil(t, resp.CanceledAt)
			require.Nil(t, resp.Problem)
			require.Equal(t, tc.expectedDelivery, resp.Delivery)
		})
	}
}

func TestSmppSendSmsRequestListener_Accept_when_the_first_connector_is_unavailable_and_a_secondary_is_available(t *testing.T) {
	testCases := []struct {
		name             string
		trackDelivery    bool
		expectedSmppId   string
		expectedDelivery asyncapi.DeliveryStrategy
	}{
		{"TrackDelivery", true, "e5104d81-7acb-4808-9655-091564a9bb35", asyncapi.TrackingDeliveryStrategy},
		{"NoTrackDelivery", false, "e5104d81-7acb-4808-9655-091564a9bb35", asyncapi.NotTrackingDeliveryStrategy},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			destination := "+25884990XXXX"
			req := newSendSmsRequest(destination)
			smppIds := []string{
				"e5104d81-7acb-4808-9655-091564a9bbp6",
				tc.expectedSmppId,
			}
			smppRepo, smppManager := getSmppConfig([]model.Smpp{
				{
					Id:          smppIds[0],
					Name:        "international",
					Description: "<Description/>",
					Alias:       "int",
					Type:        model.TransmitterType,
				},
				{
					Id:          smppIds[1],
					Name:        "vodacom",
					Description: "<Description/>",
					Alias:       "vmz",
					Type:        model.TransmitterType,
					Settings: model.Settings{
						Delivery: &model.Delivery{AwaitReport: tc.trackDelivery},
					},
				},
			}, map[string]model.Condition{
				smppIds[0]: {
					Predicate: model.Predicate{
						EqualTo: common.ToStrPointer(destination),
						Subject: common.ToStrPointer(string(Destination)),
					},
				},
				smppIds[1]: {
					Predicate: model.Predicate{
						EqualTo: common.ToStrPointer(destination),
						Subject: common.ToStrPointer(string(Destination)),
					},
				},
			}, func(c *TestConnector) {
				if c.id == smppIds[0] {
					c.err = smpp.UnavailableConnectorError{}
				}
			})

			var resp asyncapi.SendSmsResponse
			app := NewApp(t, TestAppConfig{
				SmppRepository:   smppRepo,
				ConnectorManager: smppManager,
			},
				fx.Invoke(func(lifecycle fx.Lifecycle, l asyncapi2.SendSmsRequestHandler) {
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
			require.NotNil(t, resp.Smsc)
			require.Equal(t, tc.expectedSmppId, resp.Smsc.Id)
			require.Nil(t, resp.CanceledAt)
			require.Nil(t, resp.Problem)
			require.Equal(t, tc.expectedDelivery, resp.Delivery)
		})
	}
}

func TestSmppSendSmsRequestListener_Accept_when_connectors_are_unavailable(t *testing.T) {
	destination := "+25884990XXXX"
	req := newSendSmsRequest(destination)
	smppIds := []string{
		"e5104d81-7acb-4808-9655-091564a9bbp6",
		"e5104d81-7acb-4808-9655-091564a9bb35",
	}
	smppRepo, smppManager := getSmppConfig([]model.Smpp{
		{
			Id:          smppIds[0],
			Name:        "international",
			Description: "<Description/>",
			Alias:       "int",
			Type:        model.TransmitterType,
		},
		{
			Id:          smppIds[1],
			Name:        "vodacom",
			Description: "<Description/>",
			Alias:       "vmz",
			Type:        model.TransmitterType,
		},
	}, map[string]model.Condition{
		smppIds[0]: {
			Predicate: model.Predicate{
				EqualTo: common.ToStrPointer(destination),
				Subject: common.ToStrPointer(string(Destination)),
			},
		},
		smppIds[1]: {
			Predicate: model.Predicate{
				EqualTo: common.ToStrPointer(destination),
				Subject: common.ToStrPointer(string(Destination)),
			},
		},
	}, func(c *TestConnector) {
		c.err = smpp.UnavailableConnectorError{}
	})

	app := NewApp(t, TestAppConfig{
		SmppRepository:   smppRepo,
		ConnectorManager: smppManager,
	},
		fx.Invoke(func(lifecycle fx.Lifecycle, l asyncapi2.SendSmsRequestHandler) {
			lifecycle.Append(fx.Hook{
				OnStart: func(ctx goContext.Context) error {
					_, err := l.Accept(req)
					if err != nil {
						return err
					}
					return nil
				},
			})
		}))
	defer func() {
		_ = app.Stop(goContext.Background())
	}()
	err := app.Start(goContext.Background())
	_, isCannotHandleRequestError := err.(*CannotHandleSendSmsRequestError)
	require.True(t, isCannotHandleRequestError)
	require.Equal(t, ServiceNotFoundDetail, err.Error())
}

func TestSmppSendSmsRequestListener_Accept_when_the_request_is_canceled(t *testing.T) {
	testCases := []struct {
		name               string
		trackDelivery      bool
		expectedCanceledAt time.Time
	}{
		{"TrackDelivery", true, time.Now()},
		{"NoTrackDelivery", false, time.Now()},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			destination := "+25884990XXXX"
			req := newSendSmsRequest(destination)
			smppId := "e5104d81-7acb-4808-9655-091564a9bb35"
			smppRepo, smppManager := getSmppConfig([]model.Smpp{
				{
					Id:          smppId,
					Name:        "vodacom",
					Description: "<Description/>",
					PoweredBy:   "",
					Alias:       "vmz",
					Type:        model.TransmitterType,
					Settings: model.Settings{
						Delivery: &model.Delivery{AwaitReport: tc.trackDelivery},
					},
				},
			}, map[string]model.Condition{
				smppId: {
					Predicate: model.Predicate{
						EqualTo: common.ToStrPointer(destination),
						Subject: common.ToStrPointer(string(Destination)),
					},
				},
			}, nil)

			var resp asyncapi.SendSmsResponse
			app := NewApp(t, TestAppConfig{
				SmppRepository:   smppRepo,
				ConnectorManager: smppManager,
				SmsRepository: &InMemorySmsRepository{
					Arr: []model.Sms{
						{
							TrackDelivery: tc.trackDelivery,
							Id:            req.Id,
							ListenedAt:    tc.expectedCanceledAt,
							CanceledAt:    &tc.expectedCanceledAt,
						},
					},
				},
			},
				fx.Invoke(func(lifecycle fx.Lifecycle, l asyncapi2.SendSmsRequestHandler) {
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

			require.Nil(t, resp.Smsc)
			require.Nil(t, resp.Problem)
			require.Empty(t, resp.Delivery)
			require.Equal(t, req.Id, resp.Id)
			require.NotNil(t, resp.CanceledAt)
			require.Equal(t, tc.expectedCanceledAt, *resp.CanceledAt)
		})
	}
}

func TestSmppSendSmsRequestListener_Accept_when_the_request_has_been_sent_previous(t *testing.T) {
	testCases := []struct {
		name                     string
		trackDelivery            bool
		expectedDeliveryStrategy asyncapi.DeliveryStrategy
	}{
		{"WithTracking", true, asyncapi.TrackingDeliveryStrategy},
		{"WithoutTracking", false, asyncapi.NotTrackingDeliveryStrategy},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			destination := "+25884990XXXX"
			req := newSendSmsRequest(destination)
			smppId := "e5104d81-7acb-4808-9655-091564a9bb35"
			smppRepo, smppManager := getSmppConfig([]model.Smpp{
				{
					Id:          smppId,
					Name:        "vodacom",
					Description: "<Description/>",
					PoweredBy:   "",
					Alias:       "vmz",
					Type:        model.TransmitterType,
					Settings: model.Settings{
						Delivery: &model.Delivery{AwaitReport: tc.trackDelivery},
					},
				},
			}, map[string]model.Condition{
				smppId: {
					Predicate: model.Predicate{
						EqualTo: common.ToStrPointer(destination),
						Subject: common.ToStrPointer(string(Destination)),
					},
				},
			}, nil)

			instant := time.Now()
			var resp asyncapi.SendSmsResponse
			app := NewApp(t, TestAppConfig{
				SmppRepository:   smppRepo,
				ConnectorManager: smppManager,
				SmsRepository: &InMemorySmsRepository{
					Arr: []model.Sms{
						{
							TrackDelivery: tc.trackDelivery,
							Id:            req.Id,
							ListenedAt:    instant,
							Smpp: &asyncapi.ObjectId{
								Id: smppId,
							},
						},
					},
				},
			},
				fx.Invoke(func(lifecycle fx.Lifecycle, l asyncapi2.SendSmsRequestHandler) {
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
			require.NotNil(t, resp.Smsc)
			require.Equal(t, smppId, resp.Smsc.Id)
			require.Nil(t, resp.CanceledAt)
			require.Nil(t, resp.Problem)
			require.Equal(t, tc.expectedDeliveryStrategy, resp.Delivery)
		})
	}
}

func TestSmppSendSmsRequestListener_Accept_when_the_request_has_been_previous_processed_with_error(t *testing.T) {
	testCases := []struct {
		name          string
		trackDelivery bool
	}{
		{"WithTracking", true},
		{"WithoutTracking", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			destination := "+25884990XXXX"
			req := newSendSmsRequest(destination)
			smppId := "e5104d81-7acb-4808-9655-091564a9bb35"
			smppRepo, smppManager := getSmppConfig([]model.Smpp{
				{
					Id:          smppId,
					Name:        "vodacom",
					Description: "<Description/>",
					PoweredBy:   "",
					Alias:       "vmz",
					Type:        model.TransmitterType,
					Settings: model.Settings{
						Delivery: &model.Delivery{AwaitReport: tc.trackDelivery},
					},
				},
			}, map[string]model.Condition{
				smppId: {
					Predicate: model.Predicate{
						EqualTo: common.ToStrPointer(destination),
						Subject: common.ToStrPointer(string(Destination)),
					},
				},
			}, nil)

			instant := time.Now()
			var resp asyncapi.SendSmsResponse
			app := NewApp(t, TestAppConfig{
				SmppRepository:   smppRepo,
				ConnectorManager: smppManager,
				SmsRepository: &InMemorySmsRepository{
					Arr: []model.Sms{
						{
							TrackDelivery: tc.trackDelivery,
							Id:            req.Id,
							ListenedAt:    instant,
							Error:         "<EXCEPTION/>",
						},
					},
				},
			},
				fx.Invoke(func(lifecycle fx.Lifecycle, l asyncapi2.SendSmsRequestHandler) {
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
			require.Equal(t, GenericProblemType, resp.Problem.Type)
			require.Equal(t, GenericProblemTitle, resp.Problem.Title)
			require.Equal(t, "<EXCEPTION/>", resp.Problem.Detail)
		})
	}
}

func newSendSmsRequest(destination string) asyncapi.SendSmsRequest {
	return asyncapi.SendSmsRequest{
		To:      destination,
		From:    "onboard-workflow",
		Id:      uuid.New().String(),
		Content: "Welcome to our bank",
		Tags:    []string{"banking", "onboard"},
	}
}

func getSmppConfig(arr []model.Smpp, condition map[string]model.Condition, f func(*TestConnector)) (*SequenceBasedSmppRepository, *TestConnectorManager) {
	repo := &SequenceBasedSmppRepository{Arr: arr,
		Condition: condition}
	manager := &TestConnectorManager{seq: make([]smpp.Connector, 0)}

	for _, each := range arr {
		t := &TestConnector{
			err:            nil,
			trackDelivery:  false,
			id:             each.Id,
			connectionType: each.Type,
			alias:          each.Alias,
			state:          smpp.ReadyConnectorLifecycleState,
		}
		if each.Settings.Delivery != nil {
			t.trackDelivery = each.Settings.Delivery.AwaitReport
		}
		if f != nil {
			f(t)
		}
		manager.seq = append(manager.seq, t)
	}

	return repo, manager
}
