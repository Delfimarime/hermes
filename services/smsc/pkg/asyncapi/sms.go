package asyncapi

type SendSmsRequest struct {
	Id       string `json:"id"`
	To       string `json:"to"`
	Messages []Sms  `json:"messages"`
}

type Sms struct {
	Id      string
	Content string
}
