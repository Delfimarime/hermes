package asyncapi

type SendSmsRequest struct {
	Id       string               `json:"id"`
	To       string               `json:"to"`
	Tags     []string             `json:"tags"`
	From     string               `json:"from"`
	Messages []SendSmsRequestPart `json:"messages"`
}

type SendSmsRequestPart struct {
	Id      string `json:"id"`
	Content string `json:"content"`
}

type SendSmsResponse struct {
	Id       string                `json:"id"`
	Messages []SendSmsResponsePart `json:"messages"`
}

type SendSmsResponsePart struct {
	Id     string `json:"id"`
	SmppId string `json:"smpp_id"`
	SmscId string `json:"smsc_id"`
}
