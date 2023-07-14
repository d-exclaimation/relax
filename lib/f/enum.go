package f

// Each applies a function to each element of an array
func Each[T any](arr []T, fn func(T)) {
	for _, item := range arr {
		fn(item)
	}
}

// Tap applies a function to each element of an array and returns the array
func Tap[T any](arr []T, fn func(T)) []T {
	Each[T](arr, fn)
	return arr
}

// Map applies a function to each element of an array and returns a new array
func Map[T any, K any](arr []T, fn func(T) K) []K {
	res := make([]K, len(arr))
	for i, item := range arr {
		res[i] = fn(item)
	}
	return res
}

// Filter returns a new array with elements that pass a test
func Filter[T any](arr []T, fn func(T) bool) []T {
	res := make([]T, 0)
	for _, item := range arr {
		if fn(item) {
			res = append(res, item)
		}
	}
	return res
}

// FilterSized returns a new array with elements that pass a test (with an expected size)
func FilterSized[T any](arr []T, size uint, fn func(T) bool) []T {
	res := make([]T, size)
	i := 0
	for _, item := range arr {
		if fn(item) {
			res[i] = item
			i++
		}
		if i == len(res) {
			return res
		}
	}
	return res
}

// Reduce applies a function to each element of an array and returns a new array
func Reduce[T any, K any](arr []T, init K, fn func(K, T) K) K {
	res := init
	for _, item := range arr {
		res = fn(res, item)
	}
	return res
}

// Some returns true if at least one element in the array passes a test
func Some[T any](arr []T, fn func(T) bool) bool {
	for _, item := range arr {
		if fn(item) {
			return true
		}
	}
	return false
}

// Every returns true if all elements in the array pass a test
func All[T any](arr []T, fn func(T) bool) bool {
	for _, item := range arr {
		if !fn(item) {
			return false
		}
	}
	return true
}

// Count returns the number of elements in the array that pass a test
func CountBy[T any](arr []T, fn func(T) bool) int {
	res := 0
	for _, item := range arr {
		if fn(item) {
			res++
		}
	}
	return res
}

// IsMember returns true if the item is in the array
func IsMember[T comparable](arr []T, item T) bool {
	for _, i := range arr {
		if i == item {
			return true
		}
	}
	return false
}

// Reversed returns a new array with the elements in reverse order
func Reversed[T any](arr []T) []T {
	res := make([]T, len(arr))
	for i, item := range arr {
		res[len(arr)-i-1] = item
	}
	return res
}

// Sliced returns a new array with the elements from start to end
func Sliced[T any](arr []T, start uint, end uint) []T {
	res := make([]T, end-start)
	for i := start; i < end; i++ {
		res[i-start] = arr[i]
	}
	return res
}

// Take returns a new array with the first n elements
func Take[T any](arr []T, n uint) []T {
	return Sliced[T](arr, 0, n)
}

// First returns a value with the first element that matches a condition
func First[T any](arr []T, fn func(T) bool) (T, bool) {
	for _, item := range arr {
		if fn(item) {
			return item, true
		}
	}
	return arr[0], false
}

// FindIndexOf returns a value with the first element that matches a condition
func FindIndexOf[T any](arr []T, fn func(T) bool) (T, int, bool) {
	for i, item := range arr {
		if fn(item) {
			return item, i, true
		}
	}
	return arr[0], -1, false
}
