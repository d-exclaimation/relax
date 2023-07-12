package ai

import (
	"context"
	"time"

	"d-exclaimation.me/relax/lib/f"
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
	if !ok || time.Since(prev.start) > 2*time.Minute {
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
		Model:            openai.GPT3Dot5Turbo,
		MaxTokens:        f.SumBy(prev.messages, func(m openai.ChatCompletionMessage) int { return len(m.Content) }) + 3000,
		FrequencyPenalty: 0.6,
		Temperature:      0.5,
		PresencePenalty:  1,
		Messages:         prev.messages,
		User:             userId,
	})
	if err != nil {
		return "", err
	}

	answer := resp.Choices[0].Message.Content

	prev.messages = append(prev.messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: answer,
	})

	history[userId] = prev

	return answer, nil
}
