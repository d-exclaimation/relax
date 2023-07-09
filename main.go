package main

import (
	"fmt"
	"log"

	"d-exclaimation.me/relax/config"
	"d-exclaimation.me/relax/lib/f"
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

	content := make([]slack.Block, len(teamMembers)*3+2+1)

	content[0] = slack.NewSectionBlock(
		slack.NewTextBlockObject("mrkdwn", "Here’s the list of members of *Saturday* :saturday:", false, false),
		nil,
		nil,
	)
	content[1] = slack.NewDividerBlock()

	for i, member := range teamMembers {
		offset := i*3 + 2
		content[offset] = slack.NewSectionBlock(
			slack.NewTextBlockObject(
				"mrkdwn",
				fmt.Sprintf(
					"*<github.com/d-exclaimation/relax|%s>*\n%s\n>:microservices: Has done *%d* review(s)\n>:hourglass:%s",
					member.Profile.RealName,
					f.IfElse(member.Profile.StatusText != "", member.Profile.StatusText, "_No status_"),
					0,
					member.TZLabel,
				),
				false,
				false,
			),
			nil,
			slack.NewAccessory(
				slack.NewImageBlockElement(
					member.Profile.Image72,
					member.Profile.RealName,
				),
			),
		)
		content[offset+1] = slack.NewContextBlock(
			"",
			slack.NewImageBlockElement("https://w7.pngwing.com/pngs/195/984/png-transparent-slack-logo-icon-thumbnail.png", member.Profile.RealName),
			slack.NewTextBlockObject("mrkdwn", member.Name, false, false),
		)
		content[offset+2] = slack.NewDividerBlock()
	}

	content[len(content)-1] = slack.NewActionBlock(
		"",
		slack.NewButtonBlockElement(
			"random",
			"random",
			slack.NewTextBlockObject("plain_text", "Random", false, false),
		),
	)

	_, _, err = client.PostMessage(
		config.Env.Channels()[0],
		slack.MsgOptionBlocks(content...),
	)

	if err != nil {
		log.Fatalf("Cannot post message because %s\n", err)
		log.Fatalln("Exiting...")
		return
	}
}
