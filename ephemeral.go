package disc

import "github.com/bwmarrin/discordgo"

func EphemeralResponse(data *discordgo.InteractionResponseData) *discordgo.InteractionResponse {
	data.Flags = 1 << 6
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: data,
	}
}

func EphemeralContentResponse(content string) *discordgo.InteractionResponse {
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   1 << 6,
		},
	}
}

func EphemeralEmbedResponse(embeds ...*discordgo.MessageEmbed) *discordgo.InteractionResponse {
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: embeds,
			Flags:  1 << 6,
		},
	}
}
