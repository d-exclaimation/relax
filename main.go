package main

import (
	"log"

	"d-exclaimation.me/relax/config"
	"d-exclaimation.me/relax/lib/team"
	"github.com/slack-go/slack"
)

func main() {
	config.Env.Load()

	client := slack.New(config.Env.OAuth(), slack.OptionDebug(true))

	teamMembers, err := team.GetMembers(client)

	if err != nil {
		log.Fatalf("Cannot find team member because %s\n", err)
		log.Fatalln("Exiting...")
		return
	}

	attachments := make([]slack.Attachment, len(teamMembers))

	for i, member := range teamMembers {
		attachments[i] = slack.Attachment{
			Text:  member.Profile.RealName,
			Color: "#6366F1",
		}
	}

	_, _, err = client.PostMessage(config.Env.Channels()[0], slack.MsgOptionAttachments(attachments...))

	if err != nil {
		log.Fatalf("Cannot post message because %s\n", err)
		log.Fatalln("Exiting...")
		return
	}
}
