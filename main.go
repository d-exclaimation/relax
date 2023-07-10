package main

import (
	"log"

	"d-exclaimation.me/relax/app/mr"
	"d-exclaimation.me/relax/config"
	"d-exclaimation.me/relax/lib/async"
	"github.com/slack-go/slack"
)

type Resolver func(client *slack.Client) (slack.MsgOption, error)

func main() {
	config.Env.Load()

	client := slack.New(config.Env.OAuth(), slack.OptionDebug(true))

	resolvers := map[string]Resolver{
		"reviewer": mr.RandomReviewerResolver,
	}

	resolver := resolvers["reviewer"]

	task := async.New(func() (async.Unit, error) {
		msg, err := resolver(client)

		if err != nil {
			return async.Unit{}, err
		}

		_, _, err = client.PostMessage(config.Env.Channels()[0], msg)

		return async.Unit{}, err
	})

	_, err := task.Await()

	if err != nil {
		log.Fatalf("Cannot post message because %s\n", err)
		log.Fatalln("Exiting...")
		return
	}
}
