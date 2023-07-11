package mr

import (
	"d-exclaimation.me/relax/lib/async"
	"d-exclaimation.me/relax/lib/f"
	"d-exclaimation.me/relax/lib/kv"
	"d-exclaimation.me/relax/lib/random"
	"github.com/slack-go/slack"
)

func randomlyPickReviewer(reviewers []Reviewer) Reviewer {
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

// RandomReviewer picks a random reviewer from the team, excluding the given user
func RandomReviewer(client *slack.Client, excluding func(slack.User) bool) (Reviewer, error) {
	teamMembers, err := GetMembers(client, "team").Await()

	if err != nil {
		return Reviewer{}, err
	}

	filteredMembers := f.Filter(teamMembers, func(user slack.User) bool {
		return !excluding(user)
	})
	keys := f.Map(filteredMembers, func(member slack.User) string {
		return "reviews:" + member.ID
	})
	reviews, err := async.New(func() ([]int, error) {
		data, err := kv.MGet(keys...).Await()
		if err != nil {
			return nil, err
		}

		return f.Map(data.Result, f.ParseInt), nil
	}).Await()

	if err != nil {
		return Reviewer{}, nil
	}

	reviewers := make([]Reviewer, len(filteredMembers))
	for i, member := range filteredMembers {
		reviewers[i] = Reviewer{
			User:        &member,
			ReviewCount: reviews[i],
		}
	}
	reviewer := randomlyPickReviewer(reviewers)
	kv.Incr("reviews:" + reviewer.User.ID)

	return reviewer, nil
}

// RandomReviewerWithMessage is a resolver that picks a random reviewer from the team and returns an appropriate message
func RandomReviewerWithMessage(client *slack.Client, excluding func(slack.User) bool) (slack.MsgOption, error) {
	reviewer, err := RandomReviewer(client, excluding)

	if err != nil {
		return nil, err
	}

	msg := slack.MsgOptionBlocks(
		reviewer.ChosenReviewerBlock(),
	)

	return msg, nil
}

// SelfReviewerStatus is a resolver that returns the number of reviews a user has done
func SelfReviewerStatus(userID string) (slack.MsgOption, error) {
	data, err := kv.Get("reviews:" + userID).Await()
	if err != nil {
		return nil, err
	}

	reviewer := Reviewer{
		User:        &slack.User{ID: userID},
		ReviewCount: f.ParseInt(data.Result),
	}

	msg := slack.MsgOptionBlocks(
		reviewer.SelfReviewerStatusBlock(),
	)

	return msg, nil
}
