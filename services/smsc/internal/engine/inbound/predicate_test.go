package inbound

import (
	"github.com/delfimarime/hermes/services/smsc/internal/model"
	"github.com/delfimarime/hermes/services/smsc/pkg/asyncapi"
	"github.com/delfimarime/hermes/services/smsc/pkg/common"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestToPredicate_with_subject_destination_equal_to(t *testing.T) {
	predicate, err := toPredicate(model.Condition{
		Predicate: model.Predicate{
			Subject: common.ToStrPointer(string(Destination)),
			EqualTo: common.ToStrPointer("+25884990XXXX"),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	var definitions = []struct {
		result      bool
		destination string
		name        string
	}{
		{name: "destination without prefix so it doesn't match", result: false, destination: "84990XXXX"},
		{name: "destination with prefix[Mozambique]", result: true, destination: "+25884990XXXX"},
	}
	for _, tt := range definitions {
		t.Run(tt.name, func(t *testing.T) {
			r := predicate(asyncapi.SendSmsRequest{
				To: tt.destination,
			})
			require.Equal(t, tt.result, r)
		})
	}
}

func TestToPredicate_with_subject_destination_pattern(t *testing.T) {
	predicate, err := toPredicate(model.Condition{
		Predicate: model.Predicate{
			Subject: common.ToStrPointer(string(Destination)),
			Pattern: common.ToStrPointer("^84"),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	var definitions = []struct {
		result      bool
		destination string
		name        string
	}{
		{name: "destination without prefix so it matches", result: true, destination: "84990XXXX"},
		{name: "destination with prefix[Mozambique] so it doesn't match", result: false, destination: "+25884990XXXX"},
		{name: "destination without prefix and different network so it doesn't matches", result: false, destination: "82990XXXX"},
	}
	for _, tt := range definitions {
		t.Run(tt.name, func(t *testing.T) {
			r := predicate(asyncapi.SendSmsRequest{
				To: tt.destination,
			})
			require.Equal(t, tt.result, r)
		})
	}
}

func TestToPredicate_with_subject_destination_min_and_max_length(t *testing.T) {
	predicate, err := toPredicate(model.Condition{
		Predicate: model.Predicate{
			MinimumLength: common.ToIntPointer(9),
			MaximumLength: common.ToIntPointer(13),
			Subject:       common.ToStrPointer(string(Destination)),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	var definitions = []struct {
		result      bool
		destination string
		name        string
	}{
		{name: "destination without prefix", result: true, destination: "84990XXXX"},
		{name: "destination with prefix[Mozambique]", result: true, destination: "+25884990XXXX"},
		{name: "destination that violates minimum length", result: false, destination: "84990XXX"},
		{name: "destination that violates maximum length", result: false, destination: "+25884990XXXXX"},
	}
	for _, tt := range definitions {
		t.Run(tt.name, func(t *testing.T) {
			r := predicate(asyncapi.SendSmsRequest{
				To: tt.destination,
			})
			require.Equal(t, tt.result, r)
		})
	}
}

func TestToPredicate_with_subject_from_equal_to(t *testing.T) {
	predicate, err := toPredicate(model.Condition{
		Predicate: model.Predicate{
			Subject: common.ToStrPointer(string(Source)),
			EqualTo: common.ToStrPointer("ebankit"),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	var definitions = []struct {
		result bool
		from   string
		name   string
	}{
		{name: "from different system", result: false, from: "ebankit001"},
		{name: "from correct system", result: true, from: "ebankit"},
	}
	for _, tt := range definitions {
		t.Run(tt.name, func(t *testing.T) {
			r := predicate(asyncapi.SendSmsRequest{
				From: tt.from,
			})
			require.Equal(t, tt.result, r)
		})
	}
}

func TestToPredicate_with_subject_from_pattern(t *testing.T) {
	predicate, err := toPredicate(model.Condition{
		Predicate: model.Predicate{
			Subject: common.ToStrPointer(string(Source)),
			Pattern: common.ToStrPointer("^ebankit"),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	var definitions = []struct {
		result bool
		from   string
		name   string
	}{
		{name: "from system that starts with prefix", result: true, from: "ebankit002"},
		{name: "from different system that matches prefix", result: true, from: "ebankit"},
		{name: "from different system that doesn't match prefix", result: false, from: "hermes"},
	}
	for _, tt := range definitions {
		t.Run(tt.name, func(t *testing.T) {
			r := predicate(asyncapi.SendSmsRequest{
				From: tt.from,
			})
			require.Equal(t, tt.result, r)
		})
	}
}

func TestToPredicate_with_subject_from_min_and_max_length(t *testing.T) {
	predicate, err := toPredicate(model.Condition{
		Predicate: model.Predicate{
			MinimumLength: common.ToIntPointer(7),
			MaximumLength: common.ToIntPointer(10),
			Subject:       common.ToStrPointer(string(Source)),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	var definitions = []struct {
		result bool
		from   string
		name   string
	}{
		{name: "from system with minimum", result: true, from: "ebankit"},
		{name: "from system with maximum", result: true, from: "ebankit002"},
		{name: "from system with minimum but different name", result: true, from: "microgw"},
		{name: "from system that violates maximum length", result: false, from: "+25884990XXXXX"},
		{name: "from system that violates minimum length", result: false, from: "msdos"},
	}
	for _, tt := range definitions {
		t.Run(tt.name, func(t *testing.T) {
			r := predicate(asyncapi.SendSmsRequest{
				From: tt.from,
			})
			require.Equal(t, tt.result, r)
		})
	}
}

func TestToPredicate_with_subject_part_of_sms_min_and_max_length(t *testing.T) {
	predicate, err := toPredicate(model.Condition{
		Predicate: model.Predicate{
			MinimumLength: common.ToIntPointer(1),
			MaximumLength: common.ToIntPointer(2),
			Subject:       common.ToStrPointer(string(PartsOfSms)),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	var definitions = []struct {
		result bool
		name   string
		parts  []string
	}{
		{name: "request with parts=nil[violates minimum]", result: false, parts: nil},
		{name: "request with length(parts)=0[violates minimum]", result: false, parts: []string{}},
		{name: "request with length(parts)=1", result: true, parts: []string{"hello"}},
		{name: "request with length(parts)=2", result: true, parts: []string{"hello", "world"}},
		{name: "request with length(parts)=3[violates maximum]", result: false, parts: []string{"hello", "wor", "ld"}},
	}
	for _, tt := range definitions {
		t.Run(tt.name, func(t *testing.T) {
			var messages []asyncapi.Message = nil
			if tt.parts != nil {
				messages = make([]asyncapi.Message, 0)
				for _, content := range tt.parts {
					messages = append(messages, asyncapi.Message{
						Content: content,
					})
				}
			}
			r := predicate(asyncapi.SendSmsRequest{
				Messages: messages,
			})
			require.Equal(t, tt.result, r)
		})
	}
}

func TestToPredicate_with_subject_tag_equal_to(t *testing.T) {
	predicate, err := toPredicate(model.Condition{
		Predicate: model.Predicate{
			Subject: common.ToStrPointer(string(Tag)),
			EqualTo: common.ToStrPointer("transaction"),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	var definitions = []struct {
		result bool
		name   string
		tags   []string
	}{
		{name: "request when tags=nil[violation]", result: false, tags: nil},
		{name: "request when len(tags)=0[violation]", result: false, tags: make([]string, 0)},
		{name: "request when len(tags)=1", result: true, tags: []string{"transaction"}},
		{name: "request when len(tags)=2", result: true, tags: []string{"banking", "transaction"}},
		{name: "request when len(tags)=1 but doesn't match[violation]", result: false, tags: []string{"banking"}},
	}
	for _, tt := range definitions {
		t.Run(tt.name, func(t *testing.T) {
			r := predicate(asyncapi.SendSmsRequest{
				Tags: tt.tags,
			})
			require.Equal(t, tt.result, r)
		})
	}
}

func TestToPredicate_with_subject_tag_pattern(t *testing.T) {
	predicate, err := toPredicate(model.Condition{
		Predicate: model.Predicate{
			Subject: common.ToStrPointer(string(Tag)),
			Pattern: common.ToStrPointer("^transaction"),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	var definitions = []struct {
		result bool
		tags   []string
		name   string
	}{
		{name: "request when tags=nil[violation]", result: false, tags: nil},
		{name: "request when len(tags)=0[violation]", result: false, tags: make([]string, 0)},
		{name: "request when len(tags)=1", result: true, tags: []string{"transaction"}},
		{name: "request when len(tags)=2", result: true, tags: []string{"banking", "transaction"}},
		{name: "request when len(tags)=2 but different", result: true, tags: []string{"banking", "transaction001"}},
		{name: "request when len(tags)=1 but doesn't match pattern[violation]", result: false, tags: []string{"banking"}},
		{name: "request when len(tags)=1 but different", result: true, tags: []string{"transaction1000"}},
	}
	for _, tt := range definitions {
		t.Run(tt.name, func(t *testing.T) {
			r := predicate(asyncapi.SendSmsRequest{
				Tags: tt.tags,
			})
			require.Equal(t, tt.result, r)
		})
	}
}

func TestToPredicate_with_subject_tag_min_and_max_length(t *testing.T) {
	predicate, err := toPredicate(model.Condition{
		Predicate: model.Predicate{
			MinimumLength: common.ToIntPointer(11),
			MaximumLength: common.ToIntPointer(14),
			Subject:       common.ToStrPointer(string(Tag)),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	var definitions = []struct {
		result bool
		tags   []string
		name   string
	}{
		{name: "request when tags=nil[violation]", result: false, tags: nil},
		{name: "request when len(tags)=0[violation]", result: false, tags: make([]string, 0)},
		{name: "request when len(tags)=1", result: true, tags: []string{"transaction"}},
		{name: "request when len(tags)=2", result: true, tags: []string{"banking", "transaction"}},
		{name: "request when len(tags)=2[maximum length]", result: true, tags: []string{"banking", "transaction001"}},
		{name: "request when len(tags)=1 but doesn't match minimum[violation]", result: false, tags: []string{"banking"}},
		{name: "request when len(tags)=1 but doesn't match maximum[violation]", result: false, tags: []string{"transaction1000"}},
	}
	for _, tt := range definitions {
		t.Run(tt.name, func(t *testing.T) {
			r := predicate(asyncapi.SendSmsRequest{
				Tags: tt.tags,
			})
			require.Equal(t, tt.result, r)
		})
	}
}
