package smpp

import (
	"fmt"
)

type SendMessageResponse struct {
	Id string
}

type UnavailableConnectorError struct {
	causedBy error
	state    string
}

func (u UnavailableConnectorError) Error() string {
	if u.causedBy != nil {
		return u.causedBy.Error()
	}
	msg := "connector isn't ready"
	if u.state != "" {
		msg += fmt.Sprintf(" due to current state=%s", u.state)
	}
	return msg
}
