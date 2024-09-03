package common

// Contains returns true if the specified value is contained in a slice.
func Contains[T comparable](s []T, e ...T) bool {
	for _, v := range s {
		for _, v2 := range e {
			if v == v2 {
				return true
			}
		}
	}
	return false
}
