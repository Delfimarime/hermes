package pkg

import (
	"github.com/fiorix/go-smpp/smpp/pdu"
	"github.com/fiorix/go-smpp/smpp/pdu/pdufield"
)

type PduObject struct {
	messageId  string
	sourceAddr string
	content    string
}

func getPduObject(p pdu.Body) PduObject {
	resp := PduObject{}
	messageIDField := p.Fields()[pdufield.MessageID]
	if messageIDField != nil {
		resp.messageId = messageIDField.String()
	}
	sourceAddrField := p.Fields()[pdufield.SourceAddr]
	if sourceAddrField != nil {
		resp.sourceAddr = sourceAddrField.String()
	}

	messageContentField := p.Fields()[pdufield.ShortMessage]
	if messageContentField != nil {
		resp.content = messageContentField.String()
	}
	return resp
}
