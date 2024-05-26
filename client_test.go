package disc_test

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/matryer/is"
	"github.com/stevo-go-utils/disc"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
}

func createAndStartBotClient() (c *disc.Client, err error) {
	c, err = disc.NewClient(os.Getenv("BOT_TOKEN"), os.Getenv("BOT_APP_ID"))
	if err != nil {
		return nil, err
	}
	err = c.Open()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func TestCreateAndStartBotClient(t *testing.T) {
	is := is.New(t)
	c, err := createAndStartBotClient()
	is.NoErr(err)
	defer c.Close()
}
