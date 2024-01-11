package model

type Condition struct {
	Predicate
	Id          string
	Name        string
	Description string
}

type Predicate struct {
	MinimumLength *int
	MaximumLength *int
	Subject       *string
	Pattern       *string
	EqualTo       *string
	AllMatch      []Predicate
	AnyMatch      []Predicate
}
