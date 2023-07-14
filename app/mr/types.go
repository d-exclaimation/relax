package mr

import "github.com/slack-go/slack"

type Reviewer struct {
	User        slack.User
	ReviewCount int
}

type Reviewee struct {
	User           slack.User
	ReviewerChance int
}

type FullReviewerProfile struct {
	User        slack.User
	IsAvailable bool
	ReviewCount int
	Odds        []Reviewee
}
