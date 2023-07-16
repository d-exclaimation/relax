package app

import (
	"context"
	"fmt"
	"log"

	"d-exclaimation.me/relax/app/ai"
	"d-exclaimation.me/relax/app/emoji"
	"d-exclaimation.me/relax/app/memes"
	"d-exclaimation.me/relax/app/mr"
	"d-exclaimation.me/relax/app/quote"
	"d-exclaimation.me/relax/lib/f"
	"d-exclaimation.me/relax/lib/rpc"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

type AppContext struct {
	Client   *slack.Client
	AI       *ai.LLM
	ReplyTo  string
	UserID   string
	Channel  string
	ThreadTS string
}

// Define available workflow steps using the common rpc interface
func workflows(client *slack.Client) rpc.WorkflowsRouter[AppContext] {
	return rpc.Workflows[AppContext](
		client,

		// @relax random_reviewer | Pick a random reviewer
		rpc.Step[AppContext]("random_reviewer").
			OnEdit(func(e slack.InteractionCallback, ctx AppContext) []slack.Block {
				return mr.ReviewerWorkflowStepBlocks(e.User.ID, e.Channel.ID)
			}).
			OnSave(func(e slack.InteractionCallback, ctx AppContext) rpc.WorkflowInOut {
				values := e.View.State.Values
				user := values[mr.REVIEWEE_INPUT][mr.REVIEWEE_ACTION].SelectedOption.Value
				channel := values[mr.CHANNEL_INPUT][mr.CHANNEL_ACTION].SelectedConversation
				return rpc.WorkflowInOut{
					In: &slack.WorkflowStepInputs{
						mr.REVIEWEE_ACTION: {
							Value: user,
						},
						mr.CHANNEL_ACTION: {
							Value: channel,
						},
					},
					Out: &[]slack.WorkflowStepOutput{
						{
							Name:  mr.RANDOM_REVIEWER,
							Type:  "user",
							Label: "Random Reviewer",
						},
					},
				}
			}).
			OnExecute(func(e *slackevents.WorkflowStepExecuteEvent, ctx AppContext) rpc.WorkflowExecutionResult {
				user := (*e.WorkflowStep.Inputs)[mr.REVIEWEE_ACTION].Value
				// channel := (*e.WorkflowStep.Inputs)[mr.CHANNEL_ACTION].Value
				reviewer, err := mr.RandomReviewer(ctx.Client, func(u slack.User) bool {
					return u.IsBot || u.IsRestricted || u.ID == user
				})

				if err != nil {
					return rpc.WorkflowFailureResult{Message: err.Error()}
				}

				return rpc.WorkflowSuccessResult{
					Outputs: map[string]string{
						mr.RANDOM_REVIEWER: reviewer.User.ID,
					},
				}
			}),
	)
}

// Define the actions for the mention / commands using the common rpc interface
func actions(client *slack.Client) rpc.ActionsRouter[AppContext] {
	return rpc.Actions[AppContext](
		// @relax hello | Say hello
		rpc.Contains("hello", func(event string, ctx AppContext) error {
			_, _, err := ctx.Client.PostMessage(
				ctx.ReplyTo,
				slack.MsgOptionText("Hello!", false),
			)
			return err
		}),

		// @relax vibecheck | Vibe check
		rpc.Exact("vibecheck", func(event string, ctx AppContext) error {
			_, _, err := ctx.Client.PostMessage(
				ctx.ReplyTo,
				slack.MsgOptionBlocks(
					slack.NewSectionBlock(
						slack.NewTextBlockObject(
							slack.MarkdownType,
							":done: Vibe checked!",
							false, false,
						),
						nil,
						nil,
					),
				),
			)
			return err
		}),

		// @relax stats | Get the status and statistics of your reviews
		rpc.Exact("stats", func(event string, ctx AppContext) error {
			msg, err := mr.SelfReviewerStatus(ctx.Client, ctx.UserID)
			if err != nil {
				return err
			}
			_, _, err = ctx.Client.PostMessage(
				ctx.ReplyTo,
				msg,
			)
			return err
		}),

		// @relax reviewer | Pick a random reviewer and send a dedicated message
		rpc.Exact("reviewer", func(event string, ctx AppContext) error {
			msg, err := mr.RandomReviewerWithMessage(
				ctx.Client,
				func(u slack.User) bool {
					return u.IsBot || u.IsRestricted || u.ID == ctx.UserID
				},
			)
			if err != nil {
				return err
			}
			_, _, err = ctx.Client.PostMessage(
				ctx.ReplyTo,
				msg,
			)
			return err
		}),

		// @relax quote | Get a random quote and send a dedicated message
		rpc.Exact("quote", func(event string, ctx AppContext) error {
			quote, err := quote.Random().Await()
			if err != nil {
				return err
			}

			_, _, err = ctx.Client.PostMessage(
				ctx.ReplyTo,
				slack.MsgOptionBlocks(
					slack.NewSectionBlock(
						slack.NewTextBlockObject(
							slack.MarkdownType,
							f.Text(
								fmt.Sprintf("> _%s_", quote.Content),
								fmt.Sprintf("> • _%s_", quote.Author),
							),
							false,
							false,
						),
						nil,
						nil,
					),
				),
			)
			return err
		}),

		// @relax meme | Get a random meme and send a dedicated message
		rpc.Exact("meme", func(event string, ctx AppContext) error {
			meme, err := memes.Random().Await()
			if err != nil {
				return err
			}

			_, _, err = ctx.Client.PostMessage(
				ctx.ReplyTo,
				slack.MsgOptionBlocks(
					slack.NewImageBlock(
						meme.URL,
						meme.Title,
						"",
						nil,
					),

					slack.NewSectionBlock(
						slack.NewTextBlockObject(
							slack.MarkdownType,
							fmt.Sprintf(
								"%s <%s|%s>",
								emoji.AWS,
								meme.PostLink,
								meme.Title,
							),
							false,
							false,
						),
						nil,
						nil,
					),

					slack.NewContextBlock(
						"",
						slack.NewTextBlockObject(
							slack.PlainTextType,
							fmt.Sprintf("Author: %s", meme.Author),
							false,
							false,
						),
					),
				),
			)

			return err
		}),
	).
		// @relax | Default action (AI conversation)
		Else(func(event string, ctx AppContext) error {
			_, timestamp, err := ctx.Client.PostMessage(
				ctx.ReplyTo,
				f.IfElse(
					ctx.ThreadTS != "",
					[]slack.MsgOption{
						slack.MsgOptionText(emoji.THINK_THONK+emoji.THINK_THONK+emoji.THINK_THONK, false),
						slack.MsgOptionTS(ctx.ThreadTS),
					},
					[]slack.MsgOption{
						slack.MsgOptionText(emoji.THINK_THONK+emoji.THINK_THONK+emoji.THINK_THONK, false),
					},
				)...,
			)

			if err != nil {
				return err
			}

			stream, err := ctx.AI.StreamChat(ctx.UserID, event)

			if err != nil {
				return err
			}

			for answer := range stream {
				_, timestamp, _, err = ctx.Client.UpdateMessage(
					ctx.ReplyTo,
					timestamp,
					f.IfElse(
						ctx.ThreadTS != "",
						[]slack.MsgOption{
							slack.MsgOptionText(
								fmt.Sprintf("<@%s> %s", ctx.UserID, answer),
								false,
							),
							slack.MsgOptionTS(ctx.ThreadTS),
						},
						[]slack.MsgOption{
							slack.MsgOptionText(
								fmt.Sprintf("<@%s> %s", ctx.UserID, answer),
								false,
							),
						},
					)...,
				)
			}
			return err
		})
}

// Listen for events using Slack's Socket Mode (WebSocket / Realtime connecion)
// https://api.slack.com/apis/connections/socket
// SocketMode usually provides faster response times than the Web Events API,
// and it doesn't require a public endpoint.
func Listen(client *slack.Client, ai *ai.LLM) {
	// Create a new socket mode connection
	conn := socketmode.New(client)

	// Create action handler
	action := actions(client)

	// Create workflow handler
	workflow := workflows(client)

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
							action.HandleMentionAsync(event.Text, func() AppContext {
								return AppContext{
									Client:   client,
									AI:       ai,
									ReplyTo:  event.Channel,
									ThreadTS: event.ThreadTimeStamp,
									Channel:  event.Channel,
									UserID:   event.User,
								}
							})

						case *slackevents.WorkflowStepExecuteEvent:
							log.Printf("Receiving workflow step execute event from %s\n", event.CallbackID)

							workflow.HandleAsync(event, func() AppContext {
								return AppContext{
									Client:  client,
									AI:      ai,
									ReplyTo: event.WorkflowStep.WorkflowStepExecuteID,
								}
							})
						}

					}

				// Interactive components from Slack
				case socketmode.EventTypeInteractive:
					e2, ok := e1.Data.(slack.InteractionCallback)
					if !ok {
						continue
					}

					// Make sure Slack knows we acknowledge the event
					conn.Ack(*e1.Request)

					log.Printf("Receiving interaction callback from %s\n", e2.CallbackID)

					// Handle the workflow related event (3rd way of interacting with the bot)
					workflow.HandleInteractionAsync(e2, func() AppContext {
						return AppContext{
							Client:  client,
							AI:      ai,
							ReplyTo: e2.WorkflowStep.WorkflowID,
							UserID:  e2.User.ID,
							Channel: e2.Channel.ID,
						}
					})

				// Slash commands from Slack
				case socketmode.EventTypeSlashCommand:
					command, ok := e1.Data.(slack.SlashCommand)
					if !ok {
						continue
					}

					// Make sure Slack knows we acknowledge the event
					conn.Ack(*e1.Request)

					// Handle the event itself (2nd way of interacting with the bot)
					log.Printf("Receiving slash commands %s from %s <@%s>\n", command.Command, command.UserName, command.UserID)

					action.HandleCommandAsync(command.Command, func() AppContext {
						return AppContext{
							Client:  client,
							AI:      ai,
							ReplyTo: command.ChannelID,
							UserID:  command.UserID,
							Channel: command.ChannelID,
						}
					})
				}
			}
		}
	}()

	log.Println("Listening to Slack Events...")

	conn.Run()
}
