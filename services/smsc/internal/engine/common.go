package engine

import (
	"fmt"
	"github.com/delfimarime/hermes/services/smsc/internal/model"
	"github.com/delfimarime/hermes/services/smsc/pkg/asyncapi"
	"github.com/pkg/errors"
)

type Subject string
type SendSmsRequestPredicate func(request asyncapi.SendSmsRequest) bool

const (
	None        Subject = ""
	Tag         Subject = "tag"
	Source      Subject = "source"
	PartsOfSms  Subject = "messages"
	Destination Subject = "destination"
)

func toPredicate(arr ...model.Condition) (SendSmsRequestPredicate, error) {
	if arr == nil || len(arr) == 0 {
		return alwaysTrueSendSmsRequestPredicate, nil
	}
	seq := make([]SendSmsRequestPredicate, 0)
	for _, definition := range arr {
		predicate, err := predicateFrom(definition.Predicate)
		if err != nil {
			return nil, err
		}
		seq = append(seq, predicate)
	}
	return anyMatchSendSmsRequestPredicate(seq...), nil
}

func predicateFrom(definition model.Predicate) (SendSmsRequestPredicate, error) {
	subject := None
	if definition.Subject != nil {
		subject = Subject(*definition.Subject)
	}
	switch subject {
	case None:
		return newLogicalPredicate(definition)
	case Tag:
		return newTagPredicate(definition)
	case Source:
		return newSourcePredicate(definition)
	case PartsOfSms:
		return newPartsOfSmsPredicate(definition)
	case Destination:

		return newDestinationPredicate(definition)
	default:
		return nil, fmt.Errorf("subject=%s isn't supported", subject)
	}
}

func newDestinationPredicate(_ model.Predicate) (SendSmsRequestPredicate, error) {
	return nil, nil
}

func newPartsOfSmsPredicate(_ model.Predicate) (SendSmsRequestPredicate, error) {
	return nil, nil
}

func newSourcePredicate(_ model.Predicate) (SendSmsRequestPredicate, error) {
	return nil, nil
}

func newTagPredicate(_ model.Predicate) (SendSmsRequestPredicate, error) {
	return nil, nil
}

func newLogicalPredicate(def model.Predicate) (SendSmsRequestPredicate, error) {
	if def.MinimumLength != nil || def.MaximumLength != nil {
		return nil, newBadDefinitionError()
	}
	if def.Pattern != nil || def.EqualTo != nil {
		return nil, newBadDefinitionError()
	}
	if def.AllMatch != nil && def.AnyMatch != nil {
		return nil, newBadDefinitionError()
	}
	isAnyMatch := def.AnyMatch != nil
	var seq []model.Predicate
	var opt func(...SendSmsRequestPredicate) SendSmsRequestPredicate
	if isAnyMatch {
		seq = def.AnyMatch
		opt = anyMatchSendSmsRequestPredicate
	} else {
		seq = def.AllMatch
		opt = allMatchSendSmsRequestPredicate
	}
	arr := make([]SendSmsRequestPredicate, 0)
	for _, each := range seq {
		t, err := predicateFrom(each)
		if err != nil {
			return nil, err
		}
		arr = append(arr, t)
	}
	return opt(arr...), nil
}

func newBadDefinitionError() error {
	return errors.New("Predicate isn;t well defined, review definition")
}

func alwaysTrueSendSmsRequestPredicate(_ asyncapi.SendSmsRequest) bool {
	return true
}

func anyMatchSendSmsRequestPredicate(v ...SendSmsRequestPredicate) SendSmsRequestPredicate {
	return func(request asyncapi.SendSmsRequest) bool {
		for _, f := range v {
			if isTrue := f(request); isTrue {
				return true
			}
		}
		return false
	}
}

func allMatchSendSmsRequestPredicate(v ...SendSmsRequestPredicate) SendSmsRequestPredicate {
	return func(request asyncapi.SendSmsRequest) bool {
		for _, f := range v {
			if isTrue := f(request); !isTrue {
				return false
			}
		}
		return true
	}
}
