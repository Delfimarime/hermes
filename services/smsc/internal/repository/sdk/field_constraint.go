package sdk

import "fmt"

type FieldConstraintError struct {
	Value  any
	Reason string
	Field  string
}

func (u FieldConstraintError) Error() string {
	return fmt.Sprintf(`%v cannot be set on Field["name":"%s"]`, u.Value, u.Field)
}
