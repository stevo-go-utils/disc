package webhooker

import (
	"log/slog"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	disgo_webhook "github.com/disgoorg/disgo/webhook"
	"github.com/stevo-go-utils/structures"
)

type Client struct {
	recChMap *structures.SafeMap[string, *webhooker]
	asyncCh  chan asyncSendMsg
	*ClientOpts
}

type ClientOpts struct {
	errCh               chan error
	errDelay            time.Duration
	maxRetries          int
	maxRateLimitRetries int
	enableLogging       bool
	logger              *slog.Logger
	async               bool
}

type ClientOptFunc func(opts *ClientOpts)

func DefaultClientOpts() *ClientOpts {
	return &ClientOpts{
		maxRetries:          0,
		maxRateLimitRetries: -1,
		errDelay:            time.Second * 2,
		enableLogging:       false,
		logger:              slog.New(slog.Default().Handler()),
	}
}

func ErrChClientOpt(errCh chan error) ClientOptFunc {
	return func(co *ClientOpts) {
		co.errCh = errCh
	}
}

func MaxRetriesClientOpt(maxRetries int) ClientOptFunc {
	return func(co *ClientOpts) {
		co.maxRetries = maxRetries
	}
}

func MaxRateLimitRetriesClientOpt(maxRateLimitRetries uint64) ClientOptFunc {
	return func(co *ClientOpts) {
		co.maxRateLimitRetries = int(maxRateLimitRetries)
	}
}

func ErrDelayClientOpt(errDelay time.Duration) ClientOptFunc {
	return func(co *ClientOpts) {
		co.errDelay = errDelay
	}
}

func EnableLoggingClientOpt() ClientOptFunc {
	return func(co *ClientOpts) {
		co.enableLogging = true
	}
}

func LoggerClientOpt(logger *slog.Logger) ClientOptFunc {
	return func(co *ClientOpts) {
		co.logger = logger
	}
}

func AsyncClientOpt() ClientOptFunc {
	return func(co *ClientOpts) {
		co.async = true
	}
}

func NewClient(opts ...ClientOptFunc) (c *Client) {
	o := DefaultClientOpts()
	for _, opt := range opts {
		opt(o)
	}
	c = &Client{
		recChMap:   structures.NewSafeMap[string, *webhooker](),
		asyncCh:    make(chan asyncSendMsg),
		ClientOpts: o,
	}
	go func() {
		for msg := range c.asyncCh {
			msg.w.send(sendMsg{msg: msg.msg, opts: msg.opts})
		}
	}()
	return
}

type SendOpts struct {
	RequestOpts []rest.RequestOpt
}

type SendOptsFunc func(opts *SendOpts)

func DefaultSendOpts() *SendOpts {
	return &SendOpts{}
}

func SendRequestOpts(opts ...rest.RequestOpt) SendOptsFunc {
	return func(so *SendOpts) {
		so.RequestOpts = opts
	}
}

type sendMsg struct {
	msg  discord.WebhookMessageCreate
	opts *SendOpts
}

type asyncSendMsg struct {
	msg  discord.WebhookMessageCreate
	opts *SendOpts
	w    *webhooker
}

func (c *Client) Send(webhookURL string, msg discord.WebhookMessageCreate, opts ...SendOptsFunc) (err error) {
	w, ok := c.recChMap.Get(webhookURL)
	if !ok {
		c.recChMap.Set(webhookURL, w)
		wc, err := disgo_webhook.NewWithURL(webhookURL)
		if err != nil {
			return err
		}
		w = c.newWebhooker(wc.URL())
		c.recChMap.Set(webhookURL, w)
		go w.listen()
	}
	o := DefaultSendOpts()
	for _, opt := range opts {
		opt(o)
	}
	if c.async {
		c.asyncCh <- asyncSendMsg{msg: msg, opts: o, w: w}
	} else {
		w.ch <- sendMsg{msg: msg, opts: o}
	}
	return
}
