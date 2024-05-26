package botwebhooker

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/disgoorg/disgo/discord"
)

func ConvertWebhookToMessage(webhook discord.WebhookMessageCreate) (msg *discordgo.MessageSend) {
	msg = &discordgo.MessageSend{
		Content: webhook.Content,
		TTS:     webhook.TTS,
		Flags:   discordgo.MessageFlags(int(webhook.Flags)),
	}
	for _, embed := range webhook.Embeds {
		msgEmbed := &discordgo.MessageEmbed{}
		msgEmbed.URL = embed.URL
		msgEmbed.Title = embed.Title
		msgEmbed.Type = discordgo.EmbedType(embed.Type)
		msgEmbed.Description = embed.Description
		msgEmbed.Timestamp = fmt.Sprint(embed.Timestamp)
		msgEmbed.Color = embed.Color
		if embed.Footer != nil {
			msgEmbed.Footer = &discordgo.MessageEmbedFooter{
				Text:         embed.Footer.Text,
				IconURL:      embed.Footer.IconURL,
				ProxyIconURL: embed.Footer.ProxyIconURL,
			}
		}
		if embed.Image != nil {
			msgEmbed.Image = &discordgo.MessageEmbedImage{
				URL:      embed.Image.URL,
				ProxyURL: embed.Image.ProxyURL,
				Height:   embed.Image.Height,
				Width:    embed.Image.Width,
			}
		}
		if embed.Thumbnail != nil {
			msgEmbed.Thumbnail = &discordgo.MessageEmbedThumbnail{
				URL:      embed.Thumbnail.URL,
				ProxyURL: embed.Thumbnail.ProxyURL,
				Height:   embed.Thumbnail.Height,
				Width:    embed.Thumbnail.Width,
			}
		}
		if embed.Video != nil {
			msgEmbed.Video = &discordgo.MessageEmbedVideo{
				URL:    embed.Video.URL,
				Height: embed.Video.Height,
				Width:  embed.Video.Width,
			}
		}
		if embed.Provider != nil {
			msgEmbed.Provider = &discordgo.MessageEmbedProvider{
				Name: embed.Provider.Name,
				URL:  embed.Provider.URL,
			}
		}
		if embed.Author != nil {
			msgEmbed.Author = &discordgo.MessageEmbedAuthor{
				Name:         embed.Author.Name,
				URL:          embed.Author.URL,
				IconURL:      embed.Author.IconURL,
				ProxyIconURL: embed.Author.ProxyIconURL,
			}
		}
		for _, field := range embed.Fields {
			msgEmbed.Fields = append(msgEmbed.Fields, &discordgo.MessageEmbedField{
				Name:   field.Name,
				Value:  field.Value,
				Inline: *field.Inline,
			})
		}
		msg.Embeds = append(msg.Embeds, msgEmbed)
	}

	return msg
}
