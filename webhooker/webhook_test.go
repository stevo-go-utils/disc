package webhooker_test

import (
	"os"
	"testing"

	"github.com/disgoorg/disgo/discord"
	"github.com/joho/godotenv"
	"github.com/matryer/is"
	"github.com/stevo-go-utils/disc/webhooker"
	"github.com/stevo-go-utils/structures"
)

func TestSendAndEdit(t *testing.T) {
	is := is.New(t)
	is.NoErr(godotenv.Load())
	webhookURL := os.Getenv("WEBHOOK_URL")
	is.NoErr(webhooker.Send(webhookURL, discord.WebhookMessageCreate{
		Content: "test",
	}))
	m, err := webhooker.SendWithWait(webhookURL, discord.WebhookMessageCreate{
		Content: "test",
	})
	is.NoErr(err)
	is.Equal(m.Content, "test")

	is.NoErr(webhooker.Edit(webhookURL, m.ID.String(), discord.WebhookMessageUpdate{
		Content: structures.Ptr("test2"),
	}))
}
