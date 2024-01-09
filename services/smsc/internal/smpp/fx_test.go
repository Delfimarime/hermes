package smpp

import (
	goContext "context"
	"fmt"
	"github.com/delfimarime/hermes/services/smsc/internal/context"
	"github.com/delfimarime/hermes/services/smsc/internal/metric"
	"github.com/delfimarime/hermes/services/smsc/internal/model"
	"github.com/delfimarime/hermes/services/smsc/pkg/config"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"strconv"
	"sync"
	"testing"
	"time"
)

// TestAppConfig contains configuration for setting up the test application.
type TestAppConfig struct {
	SenderId         string
	ReceiverId       string
	SenderType       string
	SmsEventListener SmsEventListener
	Configuration    *config.Configuration
}

// NewApp creates a common setup for tests.
func NewApp(t *testing.T, cfg TestAppConfig, fxOptions ...fx.Option) *fxtest.App {
	senderType := model.TransmitterType
	if cfg.SenderType != "" {
		senderType = cfg.SenderType
	}
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
	var smsListener SmsEventListener = &TestReceivedSmsRequestListener{}
	if cfg.SmsEventListener != nil {
		smsListener = cfg.SmsEventListener
	}
	options := []fx.Option{
		fx.StartTimeout(15 * time.Second),
		fx.Module("context", fx.Provide(func() context.Context {
			return &context.AppContext{
				Configuration: configuration,
			}
		})),
		metric.GetUberFxModule(),
		fx.Provide(func() SmsEventListener {
			return smsListener
		}),
		fx.Provide(func() SmppRepository {
			smppList := []model.Smpp{
				{
					Id:          cfg.SenderId,
					Name:        "Vodacom Principal",
					Description: "SMSC that sends messages through Vodacom Network",
					PoweredBy:   "Vodacom",
					Contact:     []model.Person{{Name: "Delfim Marime", Email: "delfimarime@raitonbl.com", Phone: []string{"+25884990XXXX"}}},
					Type:        senderType,
					Settings:    model.Settings{Host: model.Host{Address: "127.0.0.1:2775", Username: "transmitter", Password: "admin"}},
					Alias:       "vodacom-mozambique-principal",
				},
			}
			if cfg.ReceiverId != "" {
				smppList = append(smppList, model.Smpp{
					Id:          cfg.ReceiverId,
					Name:        "Vodacom Listener",
					Description: "SMSC that listens messages sent through Vodacom Network",
					PoweredBy:   "Vodacom",
					Contact:     []model.Person{{Name: "Delfim Marime", Email: "delfimarime@raitonbl.com", Phone: []string{"+25884990XXXX"}}},
					Type:        model.ReceiverType,
					Settings:    model.Settings{Host: model.Host{Address: "127.0.0.1:2775", Username: "receiver", Password: "admin"}},
					Alias:       "vodacom-mozambique-listener",
				})
			}
			return SequenceBasedSmsRepository{Arr: smppList}
		}),
		GetUberFxModule(), // smpp
	}
	options = append(options, fxOptions...)
	return fxtest.New(t, options...)
}

func TestSendSms(t *testing.T) {
	var resp *SendMessageResponse
	senderId := uuid.New().String()
	app := NewApp(t, TestAppConfig{
		SenderId: senderId,
	}, fx.Invoke(func(lifecycle fx.Lifecycle, cm ConnectorManager) {
		lifecycle.Append(
			fx.Hook{
				OnStart: func(ctx goContext.Context) error {
					time.Sleep(2 * time.Second)
					c := cm.(*SimpleConnectorManager).cache[senderId]
					r, err := c.SendMessage("+258849900000", "Hi")
					if err != nil {
						return err
					}
					resp = &r
					return nil
				},
			},
		)
	}))
	defer app.RequireStart().RequireStop()
	idFromResponse, err := strconv.Atoi(resp.Id)
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, resp, "Response mustn't be nil")
	require.GreaterOrEqual(t, idFromResponse, 1, "SendMessageResponse.Id must be greater or equal to 1")
}

func TestListenToMessage(t *testing.T) {
	senderId := uuid.New().String()
	receiverId := uuid.New().String()
	destination := "+258849900000"
	msg := fmt.Sprintf("Hi %s", uuid.New().String())
	var wg sync.WaitGroup
	listener := &TestReceivedSmsRequestListener{
		AfterReceivedSmsRequestTrigger: func(request ReceivedSmsRequest) {
			wg.Done()
		},
	}
	app := NewApp(t, TestAppConfig{
		SenderId:         senderId,
		SmsEventListener: listener,
		ReceiverId:       receiverId,
	},
		fx.Invoke(func(lifecycle fx.Lifecycle, cm ConnectorManager) {
			lifecycle.Append(
				fx.Hook{
					OnStart: func(ctx goContext.Context) error {
						time.Sleep(1 * time.Second)
						wg.Add(1)
						go func() {
							c := cm.(*SimpleConnectorManager).cache[senderId]
							if _, err := c.SendMessage(destination, msg); err != nil {
								t.Error(err)
							}
						}()
						return nil
					},
				},
			)
		}),
		fx.Invoke(func(lifecycle fx.Lifecycle, cm ConnectorManager) {
			lifecycle.Append(
				fx.Hook{
					OnStart: func(ctx goContext.Context) error {
						wg.Wait()
						return nil
					},
				},
			)
		}))
	defer app.RequireStart().RequireStop()

	require.NotNil(t, listener.ReceivedSmsRequests, "Listener didn't catch any message")
	require.Len(t, listener.ReceivedSmsRequests, 1, "Listener received messages size must be 1")
	require.NotEmptyf(t, listener.ReceivedSmsRequests[0].Id, "Id mustn't be empty")
	require.Equal(t, listener.ReceivedSmsRequests[0].SmscId, receiverId, "Receiving SMSC.id must be the same")
	require.NotEmptyf(t, listener.ReceivedSmsRequests[0].From, "From mustn't be empty")
}

func TestSendSmsAndCatchDeliveryReport(t *testing.T) {
	// Create a logger with the desired log level
	logger, err := zap.NewDevelopment(zap.IncreaseLevel(zap.DebugLevel))
	if err != nil {
		t.Fatal(err)
	}
	zap.ReplaceGlobals(logger)

	senderId := uuid.New().String()
	receiverId := uuid.New().String()
	destination := "+258849900000"
	msg := fmt.Sprintf("Hi %s", uuid.New().String())
	var wg sync.WaitGroup
	listener := &TestReceivedSmsRequestListener{
		AfterSmsDeliveryRequestTrigger: func(request SmsDeliveryRequest) {
			go func() {
				wg.Done()
			}()
		},
	}
	app := NewApp(t, TestAppConfig{
		SmsEventListener: listener,
		SenderId:         senderId,
		ReceiverId:       receiverId,
		SenderType:       model.TransceiverType,
	}, fx.Invoke(func(lifecycle fx.Lifecycle, cm ConnectorManager) {
		lifecycle.Append(fx.Hook{
			OnStart: func(ctx goContext.Context) error {
				wg.Add(1)
				go func() {
					time.Sleep(1 * time.Second)
					c := cm.(*SimpleConnectorManager).cache[senderId]
					if _, err := c.SendMessage(destination, msg); err != nil {
						t.Error(err)
					}
				}()
				wg.Wait()
				return nil
			},
		})
	}))
	defer app.RequireStart().RequireStop()

	require.NotNil(t, listener.SmsDeliveryRequests, "Listener didn't catch any delivery report")
	require.Len(t, listener.SmsDeliveryRequests, 1, "Listener delivery reports size must be 1")
	require.NotEmptyf(t, listener.SmsDeliveryRequests[0].Id, "Id mustn't be empty")
	require.Equal(t, listener.SmsDeliveryRequests[0].SmscId, senderId, "Receiving SMSC.id must be the same")
	require.Equal(t, listener.SmsDeliveryRequests[0].Status, 0, "From mustn't be empty")
}

type TestReceivedSmsRequestListener struct {
	mutex                          sync.Mutex
	ReceivedSmsRequests            []ReceivedSmsRequest
	SmsDeliveryRequests            []SmsDeliveryRequest
	AfterReceivedSmsRequestTrigger func(request ReceivedSmsRequest)
	AfterSmsDeliveryRequestTrigger func(request SmsDeliveryRequest)
}

func (instance *TestReceivedSmsRequestListener) OnSmsRequest(request ReceivedSmsRequest) {
	instance.mutex.Lock()
	if instance.ReceivedSmsRequests == nil {
		instance.ReceivedSmsRequests = make([]ReceivedSmsRequest, 0)
	}
	instance.ReceivedSmsRequests = append(instance.ReceivedSmsRequests, request)
	instance.mutex.Unlock()
	if instance.AfterReceivedSmsRequestTrigger != nil {
		instance.AfterReceivedSmsRequestTrigger(request)
	}
}

func (instance *TestReceivedSmsRequestListener) OnSmsDelivered(request SmsDeliveryRequest) {
	instance.mutex.Lock()
	if instance.SmsDeliveryRequests == nil {
		instance.SmsDeliveryRequests = make([]SmsDeliveryRequest, 0)
	}
	instance.SmsDeliveryRequests = append(instance.SmsDeliveryRequests, request)
	instance.mutex.Unlock()
	if instance.SmsDeliveryRequests != nil {
		instance.AfterSmsDeliveryRequestTrigger(request)
	}
}
