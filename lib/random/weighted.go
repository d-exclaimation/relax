package random

import (
	"math/rand"
	"time"

	"d-exclaimation.me/relax/lib/f"
)

// WeightedValue is a value with a weight (higher weight means higher chance of being selected)
type WeightedValue[T any] struct {
	// Value is the value to be returned
	Value T

	// Weight is the weight of the value
	Weight int
}

// Weighted returns a random value using a simple weighted random algorithm
func Weighted[T any](values ...WeightedValue[T]) T {
	generator := rand.New(rand.NewSource(time.Now().UnixNano()))

	sum := f.SumBy(values, func(value WeightedValue[T]) int {
		return value.Weight
	})

	res := generator.Intn(sum)

	for _, value := range values {
		if res < value.Weight {
			return value.Value
		}
		res -= value.Weight
	}
	return values[0].Value
}
