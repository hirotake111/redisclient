package util

func Filter[T any](items []T, predicate func(T) bool) []T {
	var filtered []T
	for _, v := range items {
		if predicate(v) {
			filtered = append(filtered, v)
		}
	}
	return filtered
}
