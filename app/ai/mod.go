package ai

import (
	"context"
	"errors"
	"io"
	"time"

	"d-exclaimation.me/relax/config"
	"d-exclaimation.me/relax/lib/f"
	openai "github.com/sashabaranov/go-openai"
)

type Conversation struct {
	start    time.Time
	messages []openai.ChatCompletionMessage
}

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

func (l *LLM) Set(userId string, conversation Conversation) {
	l.setter <- struct {
		userId       string
		conversation Conversation
	}{userId, conversation}
}

func (l *LLM) Get(userId string) Conversation {
	out := make(chan Conversation)
	l.getter <- struct {
		userId string
		out    chan Conversation
	}{userId, out}
	return <-out
}

func (l *LLM) StreamChat(userId string, event string) (<-chan string, error) {
	background := context.Background()

	prev := l.Get(userId)
	prev.messages = append(prev.messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: event,
	})

	deltas, err := l.model.CreateChatCompletionStream(background, openai.ChatCompletionRequest{
		Model:           openai.GPT3Dot5Turbo,
		MaxTokens:       f.SumBy(prev.messages, func(m openai.ChatCompletionMessage) int { return len(m.Content) }) + 3000,
		Temperature:     1.5,
		PresencePenalty: 2,
		Messages:        prev.messages,
		User:            userId,
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
