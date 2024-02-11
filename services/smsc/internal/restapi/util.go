package restapi

func AnyOf(value string, compareTo ...string) bool {
	if compareTo == nil {
		return false
	}
	for _, e := range compareTo {
		if e == value {
			return true
		}
	}
	return false
}
