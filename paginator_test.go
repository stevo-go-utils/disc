package disc_test

import (
	"fmt"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/matryer/is"
	"github.com/stevo-go-utils/disc"
	"github.com/stevo-go-utils/structures"
)

func TestPaginator(t *testing.T) {
	list := make([]any, 20)
	for i := 0; i < 20; i++ {
		list[i] = fmt.Sprintf("String %d", i+1)
	}
	is := is.New(t)
	c, err := createAndStartBotClient()
	is.NoErr(err)
	defer c.Close()
	handlerErrCh := make(chan error)
	go func() {
		for err := range handlerErrCh {
			t.Log(err)
		}
	}()
	c.StartCmds(&discordgo.ApplicationCommand{
		Name:        "paginator",
		Description: "Paginator command",
	})
	c.Handle()
	c.AddAppCmdHandler("paginator", func(data disc.AppCmdHandlerData) error {
		paginator, paginatorHandlers := c.NewPaginatorBuilder().
			SetTitle("Paginator").
			SetDesc("This is a paginator message").
			SetFieldsFunc(func(p *disc.Paginator) []*discordgo.MessageEmbedField {
				fields := []*discordgo.MessageEmbedField{}
				for i, item := range structures.ParseAnyArr[string](p.CurPageItems()) {
					fields = append(fields, &discordgo.MessageEmbedField{
						Name:   fmt.Sprintf("Field %d", i+1),
						Value:  item,
						Inline: false,
					})
				}
				return fields
			}).
			SetFooterFunc(func(p *disc.Paginator) *discordgo.MessageEmbedFooter {
				return &discordgo.MessageEmbedFooter{
					Text: fmt.Sprintf("Page %d of %d", p.Page(), p.LastPage()),
				}
			}).
			SetInitialItems(list).
			UseEphemeralResponse().
			Build(5)

		c.AddMsgComponentHandlers(paginatorHandlers, handlerErrCh)
		data.S.InteractionRespond(data.I.Interaction, paginator.Response())
		return nil
	}, handlerErrCh)
	<-make(chan struct{})
}
