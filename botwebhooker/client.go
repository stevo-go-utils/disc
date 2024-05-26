package botwebhooker

import (
	"sync"

	"github.com/disgoorg/disgo/discord"
	"github.com/stevo-go-utils/disc"
)

type Client struct {
	dc            *disc.Client
	sendWebhookCh chan sendWebhookMsg
	*ClientOpts
}

type ClientOpts struct {
	errCh chan error
}

type ClientOptFunc func(*ClientOpts)

func DefaultClientOpts() *ClientOpts {
	return &ClientOpts{}
}

func ErrChClientOpt(errCh chan error) ClientOptFunc {
	return func(o *ClientOpts) {
		o.errCh = errCh
	}
}

type sendWebhookMsg struct {
	cID     string
	webhook discord.WebhookMessageCreate
	wg      *sync.WaitGroup
}

func NewClient(token string, appID string, opts ...ClientOptFunc) (c *Client, err error) {
	o := DefaultClientOpts()
	for _, opt := range opts {
		opt(o)
	}
	dc, err := disc.NewClient(token, appID)
	if err != nil {
		return nil, err
	}
	err = dc.Open()
	if err != nil {
		return nil, err
	}
	return &Client{
		dc:            dc,
		sendWebhookCh: make(chan sendWebhookMsg),
		ClientOpts:    o,
	}, nil
}

func (c *Client) Close() {
	c.dc.Close()
}

func (c *Client) Start() {
	go c.sendWebhookHandler()
}

func (c *Client) Send(cID string, webhook discord.WebhookMessageCreate) {
	c.sendWebhookCh <- sendWebhookMsg{
		cID:     cID,
		webhook: webhook,
	}
}

func (c *Client) SendWithWg(cID string, webhook discord.WebhookMessageCreate, wg *sync.WaitGroup) {
	c.sendWebhookCh <- sendWebhookMsg{
		cID:     cID,
		webhook: webhook,
		wg:      wg,
	}
}
