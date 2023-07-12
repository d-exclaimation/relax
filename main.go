package main

import (
	"log"

	"d-exclaimation.me/relax/app"
	"d-exclaimation.me/relax/app/ai"
	"d-exclaimation.me/relax/config"
	"d-exclaimation.me/relax/lib/async"
	"github.com/slack-go/slack"
)

func main() {
	config.Env.Load()

	client := slack.New(
		config.Env.OAuth(),
		slack.OptionAppLevelToken(config.Env.OAuthApp()),
	)

	ai := ai.New(config.Env.AIToken())

	task1 := async.New(func() (async.Unit, error) {
		app.Listen(client, ai)
		return async.Done, nil
	})

	errors := async.AwaitAllUnit(
		task1,
	)

	for _, err := range errors {
		if err != nil {
			log.Fatalf("error: %s", err)
		}
	}
}
