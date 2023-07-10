package router

import "strings"

type Resolver[C any] func(ctx C) error

type Handler[C any] func(string, func() C) error

type Route[C any] struct {
	resolver Resolver[C]
	matcher  func(string) bool
}

func Exact[C any](path string, resolver Resolver[C]) Route[C] {
	return Route[C]{
		resolver: resolver,
		matcher: func(event string) bool {
			return strings.ToLower(strings.TrimSpace(event)) == path
		},
	}
}

func Contains[C any](path string, resolver Resolver[C]) Route[C] {
	return Route[C]{
		resolver: resolver,
		matcher: func(event string) bool {
			return strings.Contains(strings.ToLower(event), path)
		},
	}
}

func Router[C any](routes ...Route[C]) Handler[C] {
	return func(event string, ctx func() C) error {
		for _, route := range routes {
			if route.matcher(event) {
				return route.resolver(ctx())
			}
		}
		return nil
	}
}
