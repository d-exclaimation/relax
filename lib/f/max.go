package f

type Comparable interface {
	int | float32 | float64
}

// MaxBy applies a function that returns a Comparable to each element of an array and returns the max
func MaxBy[T any, K Comparable](arr []T, fn func(T) K) K {
	res := K(0)
	for _, item := range arr {
		if fn(item) > res {
			res = fn(item)
		}
	}
	return res
}
