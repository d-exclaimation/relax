package mr

import (
	"fmt"

	"d-exclaimation.me/relax/lib/f"
	"github.com/slack-go/slack"
)

const (
	FOOTER_IMG = "https://raw.githubusercontent.com/d-exclaimation/relax/main/assets/relax.png"
)

// ReviewerStatusBlock represents the block for a single user in terms of review status
func ReviewerStatusBlock(user *slack.User) slack.Block {
	// Description and image of the reviewer
	return slack.NewSectionBlock(

		// Description of the reviewer
		slack.NewTextBlockObject(
			"mrkdwn",
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

func ChosenReviewerBlock(user *slack.User) slack.Block {
	// Description and image of the reviewer
	return slack.NewSectionBlock(

		// Description of the reviewer
		slack.NewTextBlockObject(
			"mrkdwn",
			fmt.Sprintf(
				":party-deno: *Chosen reviewer* is <@%s>\n%s\n>:microservices: Has done *%d* review(s)\n>:hourglass:%s",
				user.ID,
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
