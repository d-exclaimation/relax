package ai

import (
	"context"
	"errors"
	"io"
	"time"

	"d-exclaimation.me/relax/config"
	openai "github.com/sashabaranov/go-openai"
)

// Conversation is a struct that holds the conversation history for the LLM to use as context
type Conversation struct {
	start    time.Time
	messages []openai.ChatCompletionMessage
}

// LLM is a struct that holds the AI LLM model and act as a concurrent-safe actor to handle the conversation history
type LLM struct {
	model         *openai.Client
	conversations map[string]Conversation
	setter        chan struct {
		userId       string
		conversation Conversation
	}
	getter chan struct {
		userId string
		out    chan Conversation
	}
}

// New is a constructor for the LLM struct and run the actor
func New(token string) *LLM {
	l := LLM{
		model:         openai.NewClient(token),
		conversations: make(map[string]Conversation),
		setter: make(chan struct {
			userId       string
			conversation Conversation
		}),
		getter: make(chan struct {
			userId string
			out    chan Conversation
		}),
	}

	go func() {
		for {
			select {
			case s := <-l.setter:
				l.conversations[s.userId] = s.conversation
			case g := <-l.getter:
				conversation, ok := l.conversations[g.userId]
				if !ok || time.Since(conversation.start) > 5*time.Minute {
					conversation = Conversation{
						start: time.Now(),
						messages: []openai.ChatCompletionMessage{
							{
								Role:    openai.ChatMessageRoleSystem,
								Content: config.Env.AIContext(),
							},
						},
					}
				}
				g.out <- conversation
			}
		}
	}()

	return &l
}

// Set is a setter for the conversation history through the actor
func (l *LLM) Set(userId string, conversation Conversation) {
	l.setter <- struct {
		userId       string
		conversation Conversation
	}{userId, conversation}
}

// Get is a getter for the conversation history through the actor
func (l *LLM) Get(userId string) Conversation {
	out := make(chan Conversation)
	l.getter <- struct {
		userId string
		out    chan Conversation
	}{userId, out}
	return <-out
}

// ClearHistory is a function to clear the conversation history for a user through the actor
func (l *LLM) ClearHistory(userId string) {
	l.Set(userId, Conversation{
		start: time.Now(),
		messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: config.Env.AIContext(),
			},
		},
	})
}

// StreamChat is a function to stream the chat response from the AI LLM model
// It returns a channel of string that will be batched and throlled for every 1.5 seconds (40 emits/minute)
func (l *LLM) StreamChat(userId string, event string) (<-chan string, error) {
	background := context.Background()

	prev := l.Get(userId)
	prev.messages = append(prev.messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: event,
	})

	deltas, err := l.model.CreateChatCompletionStream(background, openai.ChatCompletionRequest{
		Model:           openai.GPT3Dot5Turbo,
		Temperature:     1.5,
		PresencePenalty: 2,
		Messages:        prev.messages,
		Stream:          true,
	})
	if err != nil {
		return nil, err
	}

	stream := make(chan string)

	go func() {
		answer := ""
		last := time.Now().Add(-250 * time.Millisecond)

		for {
			response, err := deltas.Recv()
			if errors.Is(err, io.EOF) {
				break
			}

			if err != nil {
				break
			}

			answer += response.Choices[0].Delta.Content

			if time.Since(last) > 1500*time.Millisecond {
				stream <- answer
				last = time.Now()
			}
		}

		time.Sleep(1500*time.Millisecond - time.Since(last))
		stream <- answer

		prev.messages = append(prev.messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: answer,
		})

		l.Set(userId, prev)

		close(stream)
	}()

	return stream, nil
}
