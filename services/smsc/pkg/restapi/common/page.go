package common

type ResponsePage[T any] struct {
	Self  string `json:"self,omitempty"`
	Prev  string `json:"prev,omitempty"`
	Next  string `json:"next,omitempty"`
	Last  string `json:"last,omitempty"`
	Items []T    `json:"items,omitempty"`
}

type Page[T any] struct {
	Items    []T
	Self     string
	Next     string
	Previous string
}
