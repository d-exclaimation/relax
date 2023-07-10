package f

type Summable interface {
	int | float32 | float64
}

// SumBy applies a function that returns a Summable to each element of an array and returns the sum
func SumBy[T any, K Summable](arr []T, fn func(T) K) K {
	res := K(0)
	for _, item := range arr {
		res += fn(item)
	}
	return res
}

// Sum produces the sum of an array of Summable
func Sum[T Summable](arr []T) T {
	return SumBy[T, T](arr, func(item T) T {
		return item
	})
}
