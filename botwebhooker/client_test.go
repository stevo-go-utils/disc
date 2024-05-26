package botwebhooker_test

import (
	"os"
	"sync"
	"testing"

	"github.com/disgoorg/disgo/discord"
	"github.com/joho/godotenv"
	"github.com/matryer/is"
	"github.com/stevo-go-utils/disc/botwebhooker"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}
}

func TestSend(t *testing.T) {
	is := is.New(t)
	errCh := make(chan error)
	go func() {
		for err := range errCh {
			t.Log(err)
		}
	}()
	c, err := botwebhooker.NewClient(os.Getenv("TOKEN"), os.Getenv("APP_ID"), botwebhooker.ErrChClientOpt(errCh))
	is.NoErr(err)
	defer c.Close()
	c.Start()
	wg := &sync.WaitGroup{}
	wg.Add(100)
	for i := 0; i < 100; i++ {
		c.SendWithWg(os.Getenv("CHANNEL_ID"), discord.NewWebhookMessageCreateBuilder().SetContentf("%d", i).Build(), wg)
	}
	wg.Wait()
}
