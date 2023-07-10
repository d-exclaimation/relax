package mr

import (
	"d-exclaimation.me/relax/app/team"
	"d-exclaimation.me/relax/lib/f"
	"d-exclaimation.me/relax/lib/random"
	"github.com/slack-go/slack"
)

func randomReviewer(reviewers []Reviewer) Reviewer {
	reviews := f.SumBy(reviewers, func(reviewer Reviewer) int {
		return reviewer.ReviewCount
	})

	if reviews == 0 {
		reviews = 1
	}

	values := f.Map(reviewers, func(reviewer Reviewer) random.WeightedValue[Reviewer] {
		return random.WeightedValue[Reviewer]{
			Value: reviewer,
			Weight: f.IfElseF(
				reviewer.ReviewCount == 0,
				func() int { return reviews * len(reviewers) },
				func() int { return reviews / reviewer.ReviewCount },
			),
		}
	})

	return random.Weighted[Reviewer](values...)
}

// RandomReviewerResolver is a resolver that picks a random reviewer from the team and returns an appropriate message
func RandomReviewerResolver(client *slack.Client) (slack.MsgOption, error) {
	teamMembers, err := team.GetMembers(client, "team")

	if err != nil {
		return nil, err
	}

	reviewer := randomReviewer(
		f.Map(teamMembers, func(user slack.User) Reviewer {
			return Reviewer{
				User:        &user,
				ReviewCount: f.IfElse(user.Name == "vno16", 0, 10),
			}
		}),
	)

	msg := slack.MsgOptionBlocks(
		reviewer.ChosenReviewerBlock(),
		slack.NewDividerBlock(),
		slack.NewSectionBlock(
			slack.NewTextBlockObject(
				"mrkdwn",
				"To select a new random reviewer, please run `/reviewer` again",
				false,
				false,
			),
			nil,
			nil,
		),
	)

	return msg, nil
}
