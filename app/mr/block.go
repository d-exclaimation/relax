package mr

import (
	"fmt"

	"d-exclaimation.me/relax/app/emoji"
	"d-exclaimation.me/relax/lib/f"
	"github.com/slack-go/slack"
)

const (
	FOOTER_IMG = "https://raw.githubusercontent.com/d-exclaimation/relax/main/assets/relax.png"

	REVIEWEE_ACTION = "mr-reviewee"
	REVIEWEE_INPUT  = "mr-reviewee-input"

	CHANNEL_ACTION = "mr-channel"
	CHANNEL_INPUT  = "mr-channel-input"
)

func ReviewerWorkflowStepBlocks(reviewee string, channel string) []slack.Block {
	blocks := make([]slack.Block, 3)

	blocks[0] = slack.NewSectionBlock(
		slack.NewTextBlockObject(
			"mrkdwn",
			"*This is the configuration for choosing a random reviewer",
			false,
			false,
		),
		nil,
		nil,
	)

	blocks[1] = slack.NewInputBlock(
		REVIEWEE_INPUT,
		&slack.TextBlockObject{
			Type: slack.PlainTextType,
			Text: "Merge Request By",
		},
		nil,
		slack.SelectBlockElement{
			Type:        slack.OptTypeUser,
			ActionID:    REVIEWEE_ACTION,
			InitialUser: reviewee,
		},
	)

	blocks[2] = slack.NewInputBlock(
		CHANNEL_INPUT,
		&slack.TextBlockObject{
			Type: slack.PlainTextType,
			Text: "Channel",
		},
		nil,
		slack.SelectBlockElement{
			Type:                slack.OptTypeConversations,
			ActionID:            CHANNEL_ACTION,
			InitialConversation: channel,
		},
	)

	return blocks
}

// ReviewerStatusBlock represents the block for a single user in terms of review status
func ReviewerStatusBlock(user *slack.User) slack.Block {
	// Description and image of the reviewer
	return slack.NewSectionBlock(

		// Description of the reviewer
		slack.NewTextBlockObject(
			slack.MarkdownType,
			fmt.Sprintf(
				"*<github.com/d-exclaimation/relax|%s>*\n%s\n>:microservices: Has done *%d* review(s)\n>:hourglass:%s",
				user.Profile.RealName,
				f.IfElse(user.Profile.StatusText != "", user.Profile.StatusText, "_No status_"),
				0,
				user.TZLabel,
			),
			false,
			false,
		),
		nil,

		// Image of the reviewer
		slack.NewAccessory(
			slack.NewImageBlockElement(
				user.Profile.Image72,
				user.Profile.RealName,
			),
		),
	)
}

func (r *Reviewer) ChosenReviewerBlock() slack.Block {
	return slack.NewSectionBlock(

		// Description of the reviewer
		slack.NewTextBlockObject(
			slack.MarkdownType,
			f.Text(
				fmt.Sprintf("%s *Chosen reviewer* is <@%s>", emoji.PARTY_DENO, r.User.ID),
				fmt.Sprintf("• Has done *%d* review(s) %s", r.ReviewCount, emoji.OVERWORK),
				f.IfElse(
					r.User.Profile.StatusText != "",
					fmt.Sprintf("• Is _%s %s_", r.User.Profile.StatusText, r.User.Profile.StatusEmoji),
					fmt.Sprintf("• Is _available %s_", emoji.DONE),
				),
				fmt.Sprintf("• Using _%s_ %s", r.User.TZLabel, emoji.EARTH),
			),
			false,
			false,
		),
		nil,

		// Image of the reviewer
		slack.NewAccessory(
			slack.NewImageBlockElement(
				r.User.Profile.Image72,
				r.User.Profile.RealName,
			),
		),
	)
}
