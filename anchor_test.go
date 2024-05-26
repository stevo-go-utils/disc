package disc_test

import (
	"os"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/matryer/is"
	"github.com/stevo-go-utils/disc"
)

func TestAnchorChannelMsg(t *testing.T) {
	is := is.New(t)
	c, err := createAndStartBotClient()
	is.NoErr(err)
	defer c.Close()
	msg := &discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Title:       "Test",
			Description: "Test",
		},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Test",
						Style:    discordgo.PrimaryButton,
						CustomID: "test",
					},
				},
			},
		},
	}
	err = c.Anchor(os.Getenv("CHANNEL_ID"), msg, disc.ForceClearAnchorOpt())
	is.NoErr(err)
}

func TestAnchorsChannelMsg(t *testing.T) {
	is := is.New(t)
	c, err := createAndStartBotClient()
	is.NoErr(err)
	defer c.Close()
	msgs := []*discordgo.MessageSend{
		{
			Embed: &discordgo.MessageEmbed{
				Title:       "Test1",
				Description: "Test1",
			},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Test1",
							Style:    discordgo.PrimaryButton,
							CustomID: "test1",
						},
					},
				},
			},
		},
		{
			Embed: &discordgo.MessageEmbed{
				Title:       "Test2",
				Description: "Test2",
			},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Test2",
							Style:    discordgo.PrimaryButton,
							CustomID: "test2",
						},
					},
				},
			},
		},
	}
	err = c.Anchors(os.Getenv("CHANNEL_ID"), msgs)
	is.NoErr(err)
}
