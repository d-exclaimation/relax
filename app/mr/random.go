package mr

import (
	"errors"
	"log"

	"d-exclaimation.me/relax/app/emoji"
	"d-exclaimation.me/relax/lib/async"
	"d-exclaimation.me/relax/lib/f"
	"d-exclaimation.me/relax/lib/kv"
	"d-exclaimation.me/relax/lib/random"
	"github.com/slack-go/slack"
)

func randomlyPickReviewer(reviewers []Reviewer) Reviewer {
	reviews := f.MaxBy(reviewers, func(reviewer Reviewer) int {
		return reviewer.ReviewCount
	}) + 1

	values := f.Map(reviewers, func(reviewer Reviewer) random.WeightedValue[Reviewer] {
		partial := reviews - reviewer.ReviewCount
		return random.WeightedValue[Reviewer]{
			Value:  reviewer,
			Weight: partial * partial,
		}
	})

	return random.Weighted[Reviewer](values...)
}

// ReadonlyRandomReviewer picks a random reviewer from the team, excluding the given user
func ReadonlyRandomReviewer(client *slack.Client, excluding func(slack.User) bool) (Reviewer, error) {
	teamMembers, err := GetMembers(client, "team").Await()

	if err != nil {
		return Reviewer{}, err
	}

	filteredMembers := f.Filter(teamMembers, func(user slack.User) bool {
		return !excluding(user) && !user.IsBot && user.Profile.StatusEmoji != emoji.BRB
	})
	keys := f.Map(filteredMembers, func(member slack.User) string {
		return "reviews:" + member.ID
	})
	reviews, err := async.New(func() ([]int, error) {
		data, err := kv.GetAll(keys...).Await()
		if err != nil {
			return nil, err
		}

		return f.Map(data, func(res kv.KVPacket[string]) int { return f.ParseInt(res.Result) }), nil
	}).Await()

	if err != nil {
		return Reviewer{}, nil
	}

	reviewers := make([]Reviewer, len(filteredMembers))
	for i, member := range filteredMembers {
		reviewers[i] = Reviewer{
			User:        member,
			ReviewCount: reviews[i],
		}
	}
	reviewer := randomlyPickReviewer(reviewers)
	return reviewer, nil
}

// RandomReviewer picks a random reviewer from the team, excluding the given user
func RandomReviewer(client *slack.Client, excluding func(slack.User) bool) (Reviewer, error) {
	teamMembers, err := GetMembers(client, "team").Await()

	if err != nil {
		return Reviewer{}, err
	}

	filteredMembers := f.Filter(teamMembers, func(user slack.User) bool {
		return !excluding(user) && !user.IsBot && user.Profile.StatusEmoji != emoji.BRB
	})
	keys := f.Map(filteredMembers, func(member slack.User) string {
		return "reviews:" + member.ID
	})
	reviews, err := async.New(func() ([]int, error) {
		data, err := kv.GetAll(keys...).Await()
		if err != nil {
			return nil, err
		}

		return f.Map(data, func(res kv.KVPacket[string]) int { return f.ParseInt(res.Result) }), nil
	}).Await()

	if err != nil {
		return Reviewer{}, nil
	}

	log.Print("Selecting reviewers: ")
	reviewers := make([]Reviewer, len(filteredMembers))
	for i, member := range filteredMembers {
		reviewers[i] = Reviewer{
			User:        member,
			ReviewCount: reviews[i],
		}
		log.Printf(" %s (%d)", reviewers[i].User.Name, reviewers[i].ReviewCount)
	}
	log.Println()

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
func SelfReviewerStatus(client *slack.Client, userID string) (slack.MsgOption, error) {
	members, err := GetMembers(client, "team").Await()
	if err != nil {
		return nil, err
	}

	keys := f.Map(members, func(member slack.User) string {
		return "reviews:" + member.ID
	})
	reviews, err := async.New(func() ([]int, error) {
		data, err := kv.GetAll(keys...).Await()
		if err != nil {
			return nil, err
		}

		return f.Map(data, func(res kv.KVPacket[string]) int { return f.ParseInt(res.Result) }), nil
	}).Await()

	if err != nil {
		return nil, err
	}

	_, userIndex, ok := f.FindIndexOf(members, func(member slack.User) bool { return member.ID == userID })
	if !ok {
		return nil, errors.New("user not found")
	}

	max := f.MaxBy(reviews, func(review int) int { return review }) + 1

	reviewees := make([]Reviewee, len(members))
	for i, member := range members {
		if member.ID == userID {
			reviewees[i] = Reviewee{
				User:           member,
				ReviewerChance: 0,
			}
			continue
		}

		weight := (max - reviews[userIndex]) * (max - reviews[userIndex])

		totalWeight := f.SumBy(reviews, func(review int) int {
			partial := max - review
			return partial * partial
		})

		totalWeight -= (max - reviews[i]) * (max - reviews[i])

		if totalWeight == 0 {
			reviewees[i] = Reviewee{
				User:           member,
				ReviewerChance: 100 / (len(members) - 1),
			}
			continue
		}

		reviewees[i] = Reviewee{
			User:           member,
			ReviewerChance: weight * 100 / totalWeight,
		}
	}

	reviewer := FullReviewerProfile{
		User:        members[userIndex],
		IsAvailable: members[userIndex].Profile.StatusEmoji != emoji.BRB,
		ReviewCount: reviews[userIndex],
		Odds:        reviewees,
	}

	msg := slack.MsgOptionBlocks(
		reviewer.ReviewerFullProfileBlocks()...,
	)

	return msg, nil
}
