package disc

import (
	"errors"

	"github.com/bwmarrin/discordgo"
)

func GetInteractorUserID(i *discordgo.InteractionCreate) string {
	if i.Member != nil {
		return i.Member.User.ID
	}
	if i.User != nil {
		return i.User.ID
	}
	return ""
}

func GetAppCmdOptByName[T any](opts []*discordgo.ApplicationCommandInteractionDataOption, name string) (res T, err error) {
	for _, v := range opts {
		if v.Name == name {
			switch v := v.Value.(type) {
			case T:
				return v, nil
			default:
				return res, errors.New("incorrect type for opt")
			}
		}
	}
	return
}

func MustGetAppCmdOptByName[T any](opts []*discordgo.ApplicationCommandInteractionDataOption, name string) (res T) {
	for _, v := range opts {
		if v.Name == name {
			switch v := v.Value.(type) {
			case T:
				return v
			}
		}
	}
	return
}

func GetAppCmdOptByIdx[T any](opts []*discordgo.ApplicationCommandInteractionDataOption, idx int) (res T, err error) {
	for i, v := range opts {
		if i == idx {
			switch v := v.Value.(type) {
			case T:
				return v, nil
			default:
				return res, errors.New("incorrect type for opt")
			}
		}
	}
	return
}

func MustGetAppCmdOptByIdx[T any](opts []*discordgo.ApplicationCommandInteractionDataOption, idx int) (res T) {
	for i, v := range opts {
		if i == idx {
			switch v := v.Value.(type) {
			case T:
				return v
			}
		}
	}
	return
}
