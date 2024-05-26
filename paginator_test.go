package disc_test

import (
	"fmt"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/matryer/is"
	"github.com/stevo-go-utils/disc"
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
		paginator := c.NewPaginatorBuilder().
			SetTitle("Paginator").
			SetDesc("This is a paginator message").
			SetFieldsFunc(func(p *disc.Paginator) []*discordgo.MessageEmbedField {
				fields := []*discordgo.MessageEmbedField{}
				for i := (p.Page() - 1) * p.PerPage(); i < (p.Page())*p.PerPage(); i++ {
					if i >= len(p.Items()) {
						break
					}
					fields = append(fields, &discordgo.MessageEmbedField{
						Name:   fmt.Sprintf("Field %d", i+1),
						Value:  p.Items()[i].(string),
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
		data.S.InteractionRespond(data.I.Interaction, paginator.Response())
		return nil
	}, handlerErrCh)
	<-make(chan struct{})
}
