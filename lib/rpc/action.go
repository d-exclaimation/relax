package rpc

import (
	"log"
	"strings"

	"d-exclaimation.me/relax/lib/async"
	"d-exclaimation.me/relax/lib/f"
)

type Resolver[C any] func(args string, ctx C) error

type Handler[C any] func(string, func() C) error

type Action[C any] struct {
	resolver Resolver[C]
	trigger  func(string) bool
}

func Exact[C any](path string, resolver Resolver[C]) Action[C] {
	return Action[C]{
		resolver: resolver,
		trigger: func(event string) bool {
			return strings.ToLower(strings.TrimSpace(event)) == path
		},
	}
}

func Contains[C any](path string, resolver Resolver[C]) Action[C] {
	return Action[C]{
		resolver: resolver,
		trigger: func(event string) bool {
			return strings.Contains(strings.ToLower(event), path)
		},
	}
}

type ActionsRouter[C any] struct {
	actions  []Action[C]
	fallback Resolver[C]
}

func Actions[C any](routes ...Action[C]) ActionsRouter[C] {
	return ActionsRouter[C]{
		actions: routes,
	}
}

func (r ActionsRouter[C]) Else(resolver Resolver[C]) ActionsRouter[C] {
	r.fallback = resolver
	return r
}

func (r *ActionsRouter[C]) HandleMentionAsync(message string, ctx func() C) async.Task[async.Unit] {
	return async.New(func() (async.Unit, error) {
		args := f.Filter(strings.Split(strings.TrimSpace(message), " "), func(word string) bool {
			return !strings.HasPrefix(word, "<@") && !strings.HasSuffix(word, ">")
		})
		event := args[0]

		for _, route := range r.actions {
			if route.trigger(event) {
				err := route.resolver(strings.Join(args[1:], " "), ctx())
				if err != nil {
					log.Fatalln(err)
				}
				return async.Done, nil
			}
		}

		if r.fallback != nil {
			err := r.fallback(strings.Join(args[1:], " "), ctx())
			if err != nil {
				log.Fatalln(err)
			}
		}

		return async.Done, nil
	})
}

func (r *ActionsRouter[C]) HandleCommandAsync(command string, ctx func() C) async.Task[async.Unit] {
	return async.New(func() (async.Unit, error) {
		args := strings.Split(strings.TrimSpace(f.TailString(command)), " ")
		event := args[0]
		for _, route := range r.actions {
			if route.trigger(event) {
				err := route.resolver(strings.Join(args[1:], " "), ctx())
				if err != nil {
					log.Fatalln(err)
				}
				return async.Done, nil
			}
		}
		return async.Done, nil
	})
}
