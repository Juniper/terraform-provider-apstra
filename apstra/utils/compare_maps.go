package utils

func MapsMatch[A, B comparable](a, b map[A]B) bool {
	if len(a) != len(b) {
		return false
	}

	for k, va := range a {
		if vb, ok := b[k]; ok {
			// element found
			if va != vb {
				// element value mismatch
				return false
			}
		} else {
			// element not found
			return false
		}
	}

	return true
}
