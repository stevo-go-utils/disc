package webhooker_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/disgoorg/disgo/discord"
	"github.com/joho/godotenv"
	"github.com/matryer/is"
	"github.com/stevo-go-utils/disc/webhooker"
	"github.com/stevohuncho/gofile"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
}

func TestSend(t *testing.T) {
	is := is.New(t)

	proxies, err := gofile.SimpleCsv("proxies.csv")
	is.NoErr(err)
	fmt.Println(proxies)

	errCh := make(chan error)
	go func() {
		for err := range errCh {
			fmt.Println(err)
		}
	}()
	c := webhooker.NewClient(
		webhooker.ErrChClientOpt(errCh),
		webhooker.MaxRetriesClientOpt(3),
		webhooker.EnableLoggingClientOpt(),
		webhooker.ProxiesClientOpt(proxies...),
	)
	for i := 0; i < 1000; i++ {
		err := c.Send(os.Getenv("WEBHOOK_URL"), discord.NewWebhookMessageCreateBuilder().SetContent(fmt.Sprint(i)).Build())
		is.NoErr(err)
	}
	<-make(chan struct{})
}
