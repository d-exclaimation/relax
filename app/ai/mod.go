package ai

import (
	"context"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

type Conversation struct {
	start    time.Time
	messages []openai.ChatCompletionMessage
}

var history = map[string]Conversation{}

func Answer(ai *openai.Client, userId string, event string) (string, error) {
	background := context.Background()

	prev, ok := history[userId]
	if !ok || time.Since(prev.start) > 5*time.Minute {
		prev = Conversation{
			start: time.Now(),
			messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "The following is a conversation with a Slack assistant bot called relax. The bot is helpful, creative, clever, and very friendly.",
				},
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "The following bot should only use this format. *text* represents bold, _text_ represents italic, and ~text~ represents strikethrough. ```code``` represents a code block (no language support).",
				},
			},
		}
	}

	prev.messages = append(prev.messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: event,
	})

	resp, err := ai.CreateChatCompletion(background, openai.ChatCompletionRequest{
		Model:     openai.GPT3Dot5Turbo,
		MaxTokens: 4000,
		Messages:  prev.messages,
	})
	if err != nil {
		return "", err
	}

	history[userId] = prev

	answer := resp.Choices[0].Message.Content

	return answer, nil
}
