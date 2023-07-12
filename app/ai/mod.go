package ai

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
)

var history = map[string][]openai.ChatCompletionMessage{}

func Answer(ai *openai.Client, userId string, event string) (string, error) {
	background := context.Background()

	prev, ok := history[userId]
	if ok {
		prev = []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "The following is a conversation with a Slack assistant bot called relax. The bot is helpful, creative, clever, and very friendly.",
			},
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "The following bot should only use this format. *text* represents bold, _text_ represents italic, and ~text~ represents strikethrough. ```code``` represents a code block (no language support).",
			},
		}
	}

	messages := append(prev, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: event,
	})

	resp, err := ai.CreateChatCompletion(background, openai.ChatCompletionRequest{
		Model:            openai.GPT3Dot5Turbo,
		MaxTokens:        4000,
		Temperature:      0.9,
		FrequencyPenalty: 1,
		PresencePenalty:  1,
		Messages:         messages,
	})
	if err != nil {
		return "", err
	}

	history[userId] = messages

	answer := resp.Choices[0].Message.Content

	return answer, nil
}
