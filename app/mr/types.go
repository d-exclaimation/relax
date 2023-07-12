package mr

import "github.com/slack-go/slack"

type Reviewer struct {
	User        slack.User
	ReviewCount int
}
