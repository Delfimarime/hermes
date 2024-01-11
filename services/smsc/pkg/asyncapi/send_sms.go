package asyncapi

type SendSmsRequest struct {
	Id       string    `json:"id"`
	To       string    `json:"to"`
	Tags     []string  `json:"tags"`
	From     string    `json:"from"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Id      string `json:"id"`
	Content string `json:"content"`
}
