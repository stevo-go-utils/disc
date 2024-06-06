package webhooker

import (
	"log/slog"
	"sync"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	disgo_webhook "github.com/disgoorg/disgo/webhook"
	"github.com/stevo-go-utils/structures"
)

type Client struct {
	recChMap *structures.SafeMap[string, *webhooker]
	wg       *sync.WaitGroup
	*ClientOpts
}

type ClientOpts struct {
	errCh               chan error
	errDelay            time.Duration
	maxRetries          int
	maxRateLimitRetries int
	enableLogging       bool
	logger              *slog.Logger
	proxies             []string
	useWg               bool
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

func ProxiesClientOpt(proxies ...string) ClientOptFunc {
	return func(co *ClientOpts) {
		co.proxies = proxies
	}
}

func UseWgClientOpt() ClientOptFunc {
	return func(co *ClientOpts) {
		co.useWg = true
	}
}

func NewClient(opts ...ClientOptFunc) *Client {
	o := DefaultClientOpts()
	for _, opt := range opts {
		opt(o)
	}
	return &Client{
		recChMap:   structures.NewSafeMap[string, *webhooker](),
		wg:         &sync.WaitGroup{},
		ClientOpts: o,
	}
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
	done func()
}

func (c *Client) Send(webhookURL string, msg discord.WebhookMessageCreate, opts ...SendOptsFunc) (err error) {
	w, ok := c.recChMap.Get(webhookURL)
	if !ok {
		w = c.newWebhooker()
		c.recChMap.Set(webhookURL, w)
		wc, err := disgo_webhook.NewWithURL(webhookURL)
		if err != nil {
			return err
		}
		go w.listen(wc)
	}
	o := DefaultSendOpts()
	for _, opt := range opts {
		opt(o)
	}
	if c.useWg {
		c.wg.Wait()
		c.wg.Add(1)
		w.ch <- sendMsg{msg: msg, opts: o, done: func() { c.wg.Done() }}
	} else {
		w.ch <- sendMsg{msg: msg, opts: o}
	}
	return
}
