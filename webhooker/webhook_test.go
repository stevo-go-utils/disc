package webhooker_test

import (
	"fmt"
	"os"
	"testing"
	"time"

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

func TestWebhookRateLimit(t *testing.T) {
	is := is.New(t)
	is.NoErr(godotenv.Load())
	webhookURL := os.Getenv("WEBHOOK_URL")

	for i := 0; i < 10; i++ {
		is.NoErr(webhooker.Send(webhookURL, discord.WebhookMessageCreate{
			Content: fmt.Sprintf("test %d", i),
		}, webhooker.RateLimitWebhookOpt(1, time.Second*1)))
	}
	for i := 0; i < 10; i++ {
		m, err := webhooker.SendWithWait(webhookURL, discord.WebhookMessageCreate{
			Content: fmt.Sprintf("test %d", i),
		}, webhooker.RateLimitWebhookOpt(1, time.Second*1))
		is.NoErr(err)
		is.Equal(m.Content, fmt.Sprintf("test %d", i))
		is.NoErr(webhooker.Edit(webhookURL, m.ID.String(), discord.WebhookMessageUpdate{
			Content: structures.Ptr(fmt.Sprintf("test %d", i+1)),
		}, webhooker.RateLimitWebhookOpt(1, time.Second*1)))
	}
}
