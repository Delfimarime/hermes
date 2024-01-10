package smpp

import (
	goContext "context"
	"fmt"
	"github.com/delfimarime/hermes/services/smsc/internal/context"
	"github.com/delfimarime/hermes/services/smsc/internal/metric"
	"github.com/delfimarime/hermes/services/smsc/internal/model"
	"github.com/delfimarime/hermes/services/smsc/pkg/config"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestSendSms(t *testing.T) {
	withSmppSim(t, func(ip string, port string) {
		var resp *SendMessageResponse
		senderId := uuid.New().String()
		app := NewApp(t, TestAppConfig{
			ServerIP:   ip,
			ServerPort: port,
			SenderId:   senderId,
		}, fx.Invoke(func(lifecycle fx.Lifecycle, cm ConnectorManager) {
			lifecycle.Append(
				fx.Hook{
					OnStart: func(ctx goContext.Context) error {
						time.Sleep(2 * time.Second)
						c := cm.(*SimpleConnectorManager).connectorMap[senderId]
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
		require.GreaterOrEqual(t, idFromResponse, 0, "SendMessageResponse.Id must be greater or equal to 1")
	})
}

func TestListenToMessage(t *testing.T) {
	withSmppSim(t, func(ip string, port string) {
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
			ServerIP:         ip,
			ServerPort:       port,
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
								c := cm.(*SimpleConnectorManager).connectorMap[senderId]
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
	})
}

func TestSendSmsAndCatchDeliveryReport(t *testing.T) {
	withSmppSim(t, func(ip string, port string) {
		var resp SendMessageResponse
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
			ServerIP:         ip,
			ServerPort:       port,
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
						c := cm.(*SimpleConnectorManager).connectorMap[senderId]
						if r, err := c.SendMessage(destination, msg); err != nil {
							t.Error(err)
						} else {
							resp = r
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
		require.Equal(t, resp.Id, listener.SmsDeliveryRequests[0].Id, "Resp.Id and Delivery.Id must match")
		//		require.Equal(t, listener.SmsDeliveryRequests[0].SmscId, receiverId, "Receiving SMSC.id must match")
		require.Equal(t, listener.SmsDeliveryRequests[0].Status, 0, "From mustn't be empty")
	})

}

func withSmppSim(t *testing.T, exec func(ip string, port string)) {
	smppPort := "2775"
	managementConsolePort := "8884"
	ctx := goContext.Background()
	dockerContainer, err := testcontainers.
		GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image: "smppsim:latest",
				ExposedPorts: []string{
					fmt.Sprintf("%s/tcp", smppPort),
					fmt.Sprintf("%s/tcp", managementConsolePort),
				},
				WaitingFor: wait.ForHTTP("/").
					WithPort(nat.Port(managementConsolePort)).
					WithMethod("GET"),
				Files: []testcontainers.ContainerFile{
					{
						HostFilePath:      "./testdata/smppsim.props",
						ContainerFilePath: "/app/conf/smppsim.props",
						FileMode:          0o700,
					},
					{
						HostFilePath:      "./testdata/logging.properties",
						ContainerFilePath: "/app/conf/logging.properties",
						FileMode:          0o700,
					},
				},
			},
			Started: true,
		})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if prob := dockerContainer.Terminate(ctx); prob != nil {
			t.Error(prob)
		}
	}()
	ip, err := dockerContainer.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}
	port, err := dockerContainer.MappedPort(ctx, nat.Port(smppPort))
	if err != nil {
		return
	}
	if exec != nil {
		exec(ip, string(port.Port()))
	}
}

// TestAppConfig contains configuration for setting up the test application.
type TestAppConfig struct {
	ServerIP         string
	ServerPort       string
	SenderId         string
	ReceiverId       string
	SenderType       string
	SmsEventListener SmsEventListener
	Configuration    *config.Configuration
}

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
	serverIP := "127.0.0.1"
	serverPort := "2775"
	if cfg.ServerIP != "" {
		serverIP = cfg.ServerIP
	}
	if cfg.ServerPort != "" {
		serverPort = cfg.ServerPort
	}
	host := fmt.Sprintf("%s:%s", serverIP, serverPort)
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
					Settings:    model.Settings{Host: model.Host{Address: host, Username: "transmitter", Password: "admin"}},
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
					Settings:    model.Settings{Host: model.Host{Address: host, Username: "receiver", Password: "admin"}},
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
