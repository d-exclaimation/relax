package ai

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
)

func Answer(ai *openai.Client, event string) (string, error) {
	background := context.Background()

	resp, err := ai.CreateChatCompletion(background, openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "The following is a conversation with a Slack assistant bot called relax. The bot is helpful, creative, clever, and very friendly.",
			},
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "The following bot should only use this format. *text* represents bold, _text_ represents italic, and ~text~ represents strikethrough. ```code``` represents a code block (no language support).",
			},
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "Make sure that you only output text that relevant, and try to optimise on the length of the response making sure it's not too long regardless of the input. Aim for at most 1000 word output and stop at around 2000 words",
			},
			{Role: openai.ChatMessageRoleUser, Content: event},
		},
	})
	if err != nil {
		return "", err
	}

	answer := resp.Choices[0].Message.Content

	return answer, nil
}
