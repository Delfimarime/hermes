package publish

import (
	"errors"
	"fmt"
	"github.com/delfimarime/hermes/services/smsc/internal/model"
	"github.com/delfimarime/hermes/services/smsc/pkg/asyncapi"
	"regexp"
)

type Subject string
type SendSmsRequestPredicate func(request asyncapi.SendSmsRequest) bool
type SendSmsRequestPredicateFactory func(model.Predicate) (SendSmsRequestPredicate, error)

const (
	Destination Subject = "to"
	Tag         Subject = "tag"
	Source      Subject = "from"
	Content     Subject = "content"
)

const (
	UnsupportedSubjectErrorf = "subject=%s isn't supported"
)

var predicateFactories = map[Subject]SendSmsRequestPredicateFactory{
	Tag:         newTagPredicate,
	Source:      newSourcePredicate,
	Content:     newContentPredicate,
	Destination: newDestinationPredicate,
}

func toPredicate(conditions ...model.Condition) (SendSmsRequestPredicate, error) {
	if conditions == nil || len(conditions) == 0 {
		return alwaysTrue, nil
	}
	var predicates []SendSmsRequestPredicate
	for _, condition := range conditions {
		predicate, err := createPredicateFromDefinition(condition.Predicate)
		if err != nil {
			return nil, err
		}
		predicates = append(predicates, predicate)
	}
	return anyMatch(predicates), nil
}

func createPredicateFromDefinition(definition model.Predicate) (SendSmsRequestPredicate, error) {
	if definition.Subject == nil {
		return newLogicalPredicate(definition)
	}
	subject := Subject(*definition.Subject)
	predicate, hasValue := predicateFactories[subject]
	if !hasValue {
		return nil, fmt.Errorf(UnsupportedSubjectErrorf, subject)
	}
	return predicate(definition)
}

func newDestinationPredicate(def model.Predicate) (SendSmsRequestPredicate, error) {
	return newTextPredicate(def, func(req asyncapi.SendSmsRequest) string {
		return req.To
	})
}

func newContentPredicate(def model.Predicate) (SendSmsRequestPredicate, error) {
	return newTextPredicate(def, func(req asyncapi.SendSmsRequest) string {
		return req.Content
	})
}

func newSourcePredicate(def model.Predicate) (SendSmsRequestPredicate, error) {
	return newTextPredicate(def, func(req asyncapi.SendSmsRequest) string {
		return req.From
	})
}

func newTagPredicate(def model.Predicate) (SendSmsRequestPredicate, error) {
	return func(request asyncapi.SendSmsRequest) bool {
		for _, tag := range request.Tags {
			if match := performTextPredicateEvaluation(def, tag); match {
				return true
			}
		}
		return false
	}, nil
}

func newLogicalPredicate(def model.Predicate) (SendSmsRequestPredicate, error) {
	if def.MinimumLength != nil || def.MaximumLength != nil {
		return nil, newBadDefinitionError("Logical predicate cannot have length constraints")
	}
	if def.Pattern != nil || def.EqualTo != nil {
		return nil, newBadDefinitionError("Logical predicate cannot have pattern or equal destination constraints")
	}
	if def.AllMatch != nil && def.AnyMatch != nil {
		return nil, newBadDefinitionError("Logical predicate cannot have both all match and any match")
	}

	var predicates []SendSmsRequestPredicate
	for _, each := range def.AnyMatch {
		predicate, err := createPredicateFromDefinition(each)
		if err != nil {
			return nil, err
		}
		predicates = append(predicates, predicate)
	}

	if def.AnyMatch != nil {
		return anyMatch(predicates), nil
	}
	return allMatch(predicates), nil
}

func newTextPredicate(def model.Predicate, getValue func(asyncapi.SendSmsRequest) string) (SendSmsRequestPredicate, error) {
	if def.AllMatch != nil || def.AnyMatch != nil {
		return nil, newBadDefinitionError("Text predicate cannot have all match or any match")
	}
	return func(request asyncapi.SendSmsRequest) bool {
		return performTextPredicateEvaluation(def, getValue(request))
	}, nil
}

func performTextPredicateEvaluation(def model.Predicate, target string) bool {
	return equalTo(target, def.EqualTo) && matchesPattern(target, def.Pattern) &&
		isOfRange(def.MinimumLength, def.MaximumLength, func() int { return len(target) })
}

func isOfRange(min *int, max *int, getLength func() int) bool {
	length := getLength()
	if length < 0 {
		return false
	}
	if min != nil && length < *min {
		return false
	}
	if max != nil && length > *max {
		return false
	}
	return true
}

func equalTo(text string, equalTo *string) bool {
	if equalTo != nil {
		return text == *equalTo
	}
	return true
}

func matchesPattern(text string, pattern *string) bool {
	if pattern != nil {
		matched, err := regexp.MatchString(*pattern, text)
		if err != nil {
			return false
		}
		return matched
	}
	return true
}

func newBadDefinitionError(msg string) error {
	return errors.New("Predicate isn't well defined: " + msg)
}

func alwaysTrue(_ asyncapi.SendSmsRequest) bool {
	return true
}

func anyMatch(predicates []SendSmsRequestPredicate) SendSmsRequestPredicate {
	return func(request asyncapi.SendSmsRequest) bool {
		for _, predicate := range predicates {
			if predicate(request) {
				return true
			}
		}
		return false
	}
}

func allMatch(predicates []SendSmsRequestPredicate) SendSmsRequestPredicate {
	return func(request asyncapi.SendSmsRequest) bool {
		for _, predicate := range predicates {
			if !predicate(request) {
				return false
			}
		}
		return true
	}
}
