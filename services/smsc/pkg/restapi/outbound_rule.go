package restapi

type CreateRuleRequest struct {
	Name        string    `json:"name"`
	Condition   Condition `json:"condition"`
	Description string    `json:"description"`
}

type CreateRuleResponse struct {
	CreateRuleRequest
	Id string `json:"id"`
}

type Condition struct {
	AllMatch      []Condition `json:"all_match"`
	AnyMatch      []Condition `json:"any_match"`
	Subject       string      `json:"subject"`
	Pattern       string      `json:"pattern"`
	EqualTo       string      `json:"equal_to"`
	MinimumLength *int        `json:"minimum_length"`
	MaximumLength *int        `json:"maximum_length"`
}
