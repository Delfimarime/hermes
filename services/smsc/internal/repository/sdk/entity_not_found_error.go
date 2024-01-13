package sdk

import "fmt"

type EntityNotFoundError struct {
	Id   string
	Type string
}

func (instance *EntityNotFoundError) Error() string {
	return fmt.Sprintf("%s[id=%s] cannot be retrived from SmppRepository",
		instance.Type, instance.Id)
}
