package publish

import (
	"github.com/delfimarime/hermes/services/smsc/internal/model"
	"github.com/delfimarime/hermes/services/smsc/internal/repository/sdk"
	"github.com/delfimarime/hermes/services/smsc/internal/smpp"
	"github.com/google/uuid"
)

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

type InMemorySmsRepository struct {
	Err error
	Arr []model.Sms
}

func (i *InMemorySmsRepository) Save(sms *model.Sms) error {
	if i.Err != nil {
		return i.Err
	}
	if sms == nil {
		return nil
	}
	if i.Arr == nil {
		i.Arr = make([]model.Sms, 0)
	}
	i.Arr = append(i.Arr, *sms)
	return nil
}

func (i *InMemorySmsRepository) FindById(id string) (*model.Sms, error) {
	if i.Err != nil {
		return nil, i.Err
	}
	if i.Arr == nil {
		return nil, nil
	}
	for _, each := range i.Arr {
		if each.Id == id {
			v := each
			return &v, nil
		}
	}
	return nil, nil
}

type SequenceBasedSmppRepository struct {
	Err       error
	Arr       []model.Smpp
	Condition map[string]model.Condition
}

func (s SequenceBasedSmppRepository) FindAll() ([]model.Smpp, error) {
	if s.Err != nil {
		return nil, s.Err
	}
	if s.Arr == nil {
		return make([]model.Smpp, 0), nil
	}
	return s.Arr, nil
}

func (s SequenceBasedSmppRepository) FindById(id string) (model.Smpp, error) {
	if s.Arr != nil {
		for _, each := range s.Arr {
			if each.Id == id {
				return each, nil
			}
		}
	}
	return model.Smpp{}, &sdk.EntityNotFoundError{
		Id:   id,
		Type: "Smpp",
	}
}

func (s SequenceBasedSmppRepository) GetConditionsFrom(id string) ([]model.Condition, error) {
	valueFrom, hasValue := s.Condition[id]
	if !hasValue {
		return nil, nil
	}
	return []model.Condition{valueFrom}, nil
}

func (s SequenceBasedSmppRepository) Save(smpp model.Smpp) error {
	panic("implement me")
}
