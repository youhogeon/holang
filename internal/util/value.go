package util

func IsTruthy(value any) bool {
	if value == nil {
		return false
	}

	if b, ok := value.(bool); ok {
		return b
	}

	return true
}
