package team

import (
	"github.com/slack-go/slack"
)

func GetMembers(client *slack.Client) ([]slack.User, error) {
	userGroups, err := client.GetUserGroups()
	if err != nil {
		return nil, err
	}

	for _, group := range userGroups {
		if group.Handle != "team" {
			continue
		}
		ids, err := client.GetUserGroupMembers(group.ID)
		if err != nil {
			return nil, err
		}
		users, err := client.GetUsers()
		if err != nil {
			return nil, err
		}

		members := make([]slack.User, len(ids))
		for i, id := range ids {
			for _, user := range users {
				if user.ID == id {
					members[i] = user
				}
			}
		}
		return members, nil
	}

	return nil, nil
}
