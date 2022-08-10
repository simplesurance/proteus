package slices

// Contains returns true if the slice contains the given element.
func Contains[T any](s []T, e T, compareFn func(v1, v2 T) bool) bool {
	for _, v := range s {
		if compareFn(v, e) {
			return true
		}
	}
	return false
}
