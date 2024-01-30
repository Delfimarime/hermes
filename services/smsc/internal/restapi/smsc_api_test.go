package restapi

import "testing"

func TestSmscApi_New(t *testing.T) {
	instance := SmscApi{
		service:         nil,
		securityContext: nil,
	}
	instance.New(nil)
}
