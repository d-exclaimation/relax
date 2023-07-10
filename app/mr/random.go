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

// FilteredRandomReviewer picks a random reviewer from the team, excluding the given user
func FilteredRandomReviewer(client *slack.Client, excluding func(slack.User) bool) (*slack.User, error) {
	teamMembers, err := team.GetMembers(client, "team").Await()

	if err != nil {
		return nil, err
	}

	filteredMembers := f.Filter(teamMembers, func(user slack.User) bool {
		return !excluding(user)
	})
	reviewer := randomReviewer(
		f.Map(filteredMembers, func(user slack.User) Reviewer {
			return Reviewer{
				User: &user,

				// TODO: Get the review count from KV
				ReviewCount: 0,
			}
		}),
	)

	return reviewer.User, nil
}

// RandomReviewerResolver is a resolver that picks a random reviewer from the team and returns an appropriate message
func RandomReviewerResolver(client *slack.Client, excluding func(slack.User) bool) (slack.MsgOption, error) {
	teamMembers, err := team.GetMembers(client, "team").Await()

	if err != nil {
		return nil, err
	}

	filteredMembers := f.Filter(teamMembers, func(user slack.User) bool {
		return !excluding(user)
	})
	reviewer := randomReviewer(
		f.Map(filteredMembers, func(user slack.User) Reviewer {
			return Reviewer{
				User: &user,

				// TODO: Get the review count from KV
				ReviewCount: 0,
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
