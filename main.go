package main

import (
	"log"

	"d-exclaimation.me/relax/app"
	"d-exclaimation.me/relax/config"
	"d-exclaimation.me/relax/lib/async"
	openai "github.com/sashabaranov/go-openai"
	"github.com/slack-go/slack"
)

func main() {
	config.Env.Load()

	client := slack.New(
		config.Env.OAuth(),
		slack.OptionAppLevelToken(config.Env.OAuthApp()),
	)

	ai := openai.NewClient(config.Env.AIToken())

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
