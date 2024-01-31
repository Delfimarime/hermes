package restapi

func AnyOf(value any, from ...any) bool {
	if from == nil {
		return false
	}
	for _, each := range from {
		if each == value {
			return true
		}
	}
	return false
}
