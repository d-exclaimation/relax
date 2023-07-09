package f

// IfElse is a ternary operator (not lazy, both a and b are evaluated immediately)
func IfElse[T any](cond bool, a T, b T) T {
	if cond {
		return a
	}
	return b
}

// IfElseF is a ternary operator (lazy using functions, a and b are evaluated only when needed)
func IfElseF[T any](cond bool, a func() T, b func() T) T {
	if cond {
		return a()
	}
	return b()
}
