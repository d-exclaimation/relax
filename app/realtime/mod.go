package realtime

import (
	"context"
	"log"

	"d-exclaimation.me/relax/app/mr"
	"d-exclaimation.me/relax/app/router"
	"d-exclaimation.me/relax/lib/async"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

type RealtimeContext struct {
	Client *slack.Client
	Event  *slackevents.AppMentionEvent
}

// Define the handler for the realtime events using the common router interface
func handler(client *slack.Client) router.Handler[RealtimeContext] {
	return router.Router[RealtimeContext](

		// @relax hello
		router.Contains("hello", func(ctx RealtimeContext) error {
			_, _, err := ctx.Client.PostMessage(
				ctx.Event.Channel,
				slack.MsgOptionText("Hello!", false),
			)
			return err
		}),

		// @relax reviewer
		router.Contains("reviewer", func(ctx RealtimeContext) error {
			msg, err := mr.RandomReviewerResolver(ctx.Client)
			if err != nil {
				return err
			}
			_, _, err = ctx.Client.PostMessage(
				ctx.Event.Channel,
				msg,
			)
			return err
		}),
	)
}

func Listen(client *slack.Client) {
	// Create a new socket mode connection
	conn := socketmode.New(client)

	// Create an event handler
	handle := handler(client)

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())

	// Graceful shutdown
	defer cancel()

	// Start listening on a separate goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case e1 := <-conn.Events:
				switch e1.Type {

				// Connecting to Slack
				case socketmode.EventTypeConnecting:
					log.Println("Connecting to Slack with Socket Mode...")

				// Connection error
				case socketmode.EventTypeConnectionError:
					log.Println("Connection failed. Retrying later...")

				// Connected to Slack
				case socketmode.EventTypeConnected:
					log.Println("Connected to Slack with Socket Mode.")

				// Events from Slack
				case socketmode.EventTypeEventsAPI:

					// Casting event
					e2, ok := e1.Data.(slackevents.EventsAPIEvent)
					if !ok {
						continue
					}

					// Make sure Slack knows we acknowledge the event
					conn.Ack(*e1.Request)

					// Handle the event itself
					switch e2.Type {

					// Handling channel messagess
					case slackevents.CallbackEvent:
						e3 := e2.InnerEvent

						switch event := e3.Data.(type) {

						// Handling app mentions (1st way of interacting with the bot)
						case *slackevents.AppMentionEvent:
							log.Printf("Receiving mentions \"%s\" from %s\n", event.Text, event.User)

							async.New(func() (async.Unit, error) {
								err := handle(event.Text, func() RealtimeContext {
									return RealtimeContext{
										Client: client,
										Event:  event,
									}
								})
								if err != nil {
									log.Fatalln(err)
								}
								return async.Done, nil
							})
						}
					}
				}
			}
		}
	}()

	log.Println("Listening to Slack Events...")

	conn.Run()
}
