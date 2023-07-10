package team

import (
	"d-exclaimation.me/relax/lib/async"
	"d-exclaimation.me/relax/lib/f"
	"github.com/slack-go/slack"
)

// GetMembers gets the members of a user group
func GetMembers(client *slack.Client, handle string) async.Task[[]slack.User] {
	return async.New(func() ([]slack.User, error) {
		return GetMembersSync(client, handle)
	})
}

// GetMembersSync gets the members of a user group (blocking)
func GetMembersSync(client *slack.Client, handle string) ([]slack.User, error) {
	userGroups, err := client.GetUserGroups()
	if err != nil {
		return nil, err
	}

	for _, group := range userGroups {
		if group.Handle != handle {
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

		members := f.FilterSized(users, uint(len(ids)), func(user slack.User) bool {
			return f.IsMember(ids, user.ID)
		})
		return members, nil
	}

	return nil, nil
}
