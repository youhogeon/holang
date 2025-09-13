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

func IsEqual(a any, b any) bool {
	if a == nil || b == nil {
		return a == b

	}

	switch ax := a.(type) {
	case int64:
		switch by := b.(type) {
		case int64:
			return ax == by
		case float64:
			return float64(ax) == by
		}
	case float64:
		switch by := b.(type) {
		case float64:
			return ax == by
		case int64:
			return ax == float64(by)
		}
	}

	return a == b
}

func IsNotEqual(a any, b any) bool {
	if a == nil || b == nil {
		return a == b
	}

	switch ax := a.(type) {
	case int64:
		switch by := b.(type) {
		case int64:
			return ax != by
		case float64:
			return float64(ax) != by
		}
	case float64:
		switch by := b.(type) {
		case float64:
			return ax != by
		case int64:
			return ax != float64(by)
		}
	}

	return a != b
}
