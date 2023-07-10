package main

import (
	"log"

	"d-exclaimation.me/relax/app/mr"
	"d-exclaimation.me/relax/config"
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

	msg, err := resolver(client)

	if err != nil {
		log.Fatalf("Cannot resolve message because %s\n", err)
		log.Fatalln("Exiting...")
		return
	}

	_, _, err = client.PostMessage(config.Env.Channels()[0], msg)

	if err != nil {
		log.Fatalf("Cannot post message because %s\n", err)
		log.Fatalln("Exiting...")
		return
	}
}
