package connect

import (
	"fmt"
	"github.com/delfimarime/hermes/services/smsc/internal/model"
	"github.com/google/uuid"
	"testing"
	"time"
)

func TestConnector_SendMessage(t *testing.T) {
	f := ConnectorFactory{
		SmsEventListener: &TestReceivedSmsRequestListener{},
	}
	c := f.New(model.Smpp{
		//SourceAddr:  "vm.co.mz",
		Host: model.Host{
			Address:  "127.0.0.1:2775",
			Username: "transmitter",
			Password: "admin",
		},
		Type: model.TransmitterType,
	})
	err := c.DoBind()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if c != nil {
			_ = c.Close()
		}
	}()
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.SendMessage("+258842102217", "Hi")
	if err != nil {
		t.Fatal(err)
	}
}

func TestConnector_Listen(t *testing.T) {
	f := ConnectorFactory{
		SmsEventListener: &TestReceivedSmsRequestListener{},
	}
	c := f.New(model.Smpp{
		//SourceAddr:  "vm.co.mz",
		Host: model.Host{
			Address:  "127.0.0.1:2775",
			Username: "receiver",
			Password: "admin",
		},
		Type: model.ReceiverType,
	})
	err := c.DoBind()
	if err != nil {
		t.Error(err)
		return
	}
	/*
		go func() {
			TestConnector_SendMessage(t)
		}()
	*/
	time.Sleep(30 * time.Second)
}

func TestConnector_SendMessageAndListen(t *testing.T) {
	f := ConnectorFactory{
		SmsEventListener: &TestReceivedSmsRequestListener{},
	}
	c := f.New(model.Smpp{
		Id: uuid.New().String(),
		//SourceAddr:  "vm.co.mz",
		Host: model.Host{
			Address:  "127.0.0.1:2775",
			Username: "transmitter",
			Password: "admin",
		},
		Type: model.TransceiverType,
	})
	err := c.DoBind()
	if err != nil {
		t.Error(err)
		return
	}
	go func() {
		_, err = c.SendMessage("+258842102217", "Hi")
		if err != nil {
			t.Error(err)
			return
		}
	}()
	/*
		go func() {
			TestConnector_SendMessage(t)
		}()
	*/
	time.Sleep(30 * time.Second)
}

type TestReceivedSmsRequestListener struct {
}

func (instance *TestReceivedSmsRequestListener) OnSmsRequest(request ReceivedSmsRequest) {
	fmt.Println("---------ReceivedSmsRequest----------------")
	fmt.Println("Id", request.Id)
	fmt.Println("From", request.From)
	fmt.Println("SmscId", request.SmscId)
	fmt.Println("Message", request.Message)
}

func (instance *TestReceivedSmsRequestListener) OnSmsDelivered(request SmsDeliveryRequest) {
	fmt.Println("----------SmsDeliveryRequest--------------")
	fmt.Println("Id", request.Id)
	fmt.Println("SmscId", request.SmscId)
	fmt.Println("Status", request.Status)
}
