package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"d-exclaimation.me/relax/app/mr"
	"d-exclaimation.me/relax/config"
	"d-exclaimation.me/relax/lib/async"
	"d-exclaimation.me/relax/lib/f"
	"d-exclaimation.me/relax/lib/rpc"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

type Quote struct {
	ID           string   `json:"_id"`
	Content      string   `json:"content"`
	Author       string   `json:"author"`
	Tags         []string `json:"tags"`
	AuthorSlug   string   `json:"authorSlug"`
	Length       int      `json:"length"`
	DateAdded    string   `json:"dateAdded"`
	DateModified string   `json:"dateModified"`
}

type RealtimeContext struct {
	Client  *slack.Client
	ReplyTo string
	UserID  string
	Channel string
}

// Define available workflow steps using the common rpc interface
func workflows(client *slack.Client) rpc.WorkflowsRouter[RealtimeContext] {
	return rpc.Workflows[RealtimeContext](
		client,

		// @relax random_reviewer | Pick a random reviewer
		rpc.Step[RealtimeContext]("random_reviewer").
			OnEdit(func(e slack.InteractionCallback, ctx RealtimeContext) []slack.Block {
				return mr.ReviewerWorkflowStepBlocks(e.User.ID, e.Channel.ID)
			}).
			OnSave(func(e slack.InteractionCallback, ctx RealtimeContext) rpc.WorkflowInOut {
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
			OnExecute(func(e *slackevents.WorkflowStepExecuteEvent, ctx RealtimeContext) rpc.WorkflowExecutionResult {
				user := (*e.WorkflowStep.Inputs)[mr.REVIEWEE_ACTION].Value
				// channel := (*e.WorkflowStep.Inputs)[mr.CHANNEL_ACTION].Value
				reviewer, err := mr.FilteredRandomReviewer(ctx.Client, func(u slack.User) bool {
					return u.IsBot || u.IsRestricted || u.ID == user
				})

				if err != nil {
					return rpc.WorkflowFailureResult{Message: err.Error()}
				}

				return rpc.WorkflowSuccessResult{
					Outputs: map[string]string{
						mr.RANDOM_REVIEWER: reviewer.ID,
					},
				}
			}),
	)
}

// Define the actions for the mention / commands using the common rpc interface
func actions(client *slack.Client) rpc.ActionsRouter[RealtimeContext] {
	return rpc.Actions[RealtimeContext](
		// @relax hello | Say hello
		rpc.Contains("hello", func(event string, ctx RealtimeContext) error {
			_, _, err := ctx.Client.PostMessage(
				ctx.ReplyTo,
				slack.MsgOptionText("Hello!", false),
			)
			return err
		}),

		// @relax vibecheck | Vibe check
		rpc.Exact("vibecheck", func(event string, ctx RealtimeContext) error {
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

		// @relax reviewer | Pick a random reviewer and send a dedicated message
		rpc.Contains("reviewer", func(event string, ctx RealtimeContext) error {
			msg, err := mr.RandomReviewerResolver(
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
		rpc.Contains("quote", func(event string, ctx RealtimeContext) error {
			task := async.New(func() ([]Quote, error) {
				data := make([]Quote, 1)
				resp, err := http.Get(config.Env.QuoteAPI() + "/quotes/random?limit=1")
				if err != nil {
					return data, err
				}
				defer resp.Body.Close()

				json.NewDecoder(resp.Body).Decode(&data)
				return data, nil
			})
			quotes, err := task.Await()
			if err != nil {
				return err
			}

			quote := quotes[0]

			_, _, err = ctx.Client.PostMessage(
				ctx.ReplyTo,
				slack.MsgOptionBlocks(
					slack.NewSectionBlock(
						slack.NewTextBlockObject(
							"mrkdwn",
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
			if err != nil {
				return err
			}
			return nil
		}),
	)
}

// Listen for events using Slack's Socket Mode (WebSocket / Realtime connecion)
// https://api.slack.com/apis/connections/socket
// SocketMode usually provides faster response times than the Web Events API,
// and it doesn't require a public endpoint.
func Listen(client *slack.Client) {
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

							action.HandleMentionAsync(event.Text, func() RealtimeContext {
								return RealtimeContext{
									Client:  client,
									ReplyTo: event.Channel,
									Channel: event.Channel,
									UserID:  event.User,
								}
							})

						case *slackevents.WorkflowStepExecuteEvent:
							log.Printf("Receiving workflow step execute event from %s\n", event.CallbackID)

							workflow.HandleAsync(event, func() RealtimeContext {
								return RealtimeContext{
									Client:  client,
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
					workflow.HandleInteractionAsync(e2, func() RealtimeContext {
						return RealtimeContext{
							Client:  client,
							ReplyTo: e2.ResponseURL,
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
					log.Printf("Receiving slash commands %s from %s<%s>\n", command.Command, command.UserName, command.UserID)

					action.HandleCommandAsync(command.Command, func() RealtimeContext {
						return RealtimeContext{
							Client:  client,
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
