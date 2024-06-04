package webhooker

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	disgo_webhook "github.com/disgoorg/disgo/webhook"
	"github.com/go-resty/resty/v2"
	"github.com/stevo-go-utils/structures"
	"github.com/tidwall/gjson"
)

type webhooker struct {
	ch            chan sendMsg
	proxyBalancer *structures.Balancer[string]
	*ClientOpts
}

func (c *Client) newWebhooker() (w *webhooker) {
	w = &webhooker{
		ch:            make(chan sendMsg),
		proxyBalancer: structures.NewBalancer[string](),
		ClientOpts:    c.ClientOpts,
	}
	w.proxyBalancer.Add(c.proxies...)
	return
}

func (w *webhooker) listen(wc disgo_webhook.Client) {
	for msg := range w.ch {
		func() {
			defer func() {
				if msg.done != nil {
					msg.done()
				}
			}()

			body, contentType, err := ParseWebhook(msg.msg)
			if err != nil {
				if w.errCh != nil {
					w.errCh <- err
				}
				if w.enableLogging {
					w.logger.Error("failed to parse webhook", slog.Any("err", err))
				}
				return
			}

			rc := resty.New()
			proxyResp, ok := w.proxyBalancer.Use()
			if ok {
				proxySplit := strings.Split(proxyResp.Data(), ":")
				if len(proxySplit) == 4 {
					rc.SetProxy(fmt.Sprintf("http://%s:%s@%s:%s", proxySplit[2], proxySplit[3], proxySplit[0], proxySplit[1]))
				}
			}

			retries := 0
			rateLimitRetries := 0
			for retries <= w.maxRetries && (w.maxRateLimitRetries == -1 || rateLimitRetries <= w.maxRateLimitRetries) {
				if retries > 0 || rateLimitRetries > 0 {
					if w.enableLogging {
						w.logger.Warn("retrying webhook", slog.Any("retries", retries), slog.Any("rateLimitRetries", rateLimitRetries), slog.Any("delay", w.errDelay))
					}
					time.Sleep(w.errDelay)
				}
				res, err := rc.R().SetBody(body).SetHeader("Content-Type", contentType).Post(wc.URL())
				if err != nil {
					retries++
					if w.errCh != nil {
						w.errCh <- err
					}
					if w.enableLogging {
						w.logger.Error("failed to send webhook", slog.Any("err", err))
					}
					continue
				}
				resJ := gjson.ParseBytes(res.Body())
				if res.StatusCode() == 429 && strings.Contains(resJ.Get("message").Str, "rate limited") {
					rateLimitRetries++
					secs := resJ.Get("retry_after").Float()
					delay := time.Second*time.Duration(secs) + time.Millisecond*500
					time.Sleep(delay)
					if w.errCh != nil {
						w.errCh <- fmt.Errorf("rate limited, retrying in %s", delay)
					}
					if w.enableLogging {
						w.logger.Error("webhook rate limited", slog.Any("body", res.Body()))
					}
					continue
				} else if res.StatusCode() != 204 && res.StatusCode() != 200 {
					retries++
					if w.errCh != nil {
						w.errCh <- fmt.Errorf("failed to send webhook, status code: %d", res.StatusCode())
					}
					if w.enableLogging {
						w.logger.Error("failed to send webhook", slog.Any("body", res.Body()), slog.Any("status", res.StatusCode()))
					}
					continue
				}
				break
			}
			if w.enableLogging {
				w.logger.Info("sent webhook")
			}
		}()
	}
}

func ParseWebhook(webhook discord.WebhookMessageCreate) (body []byte, contentType string, err error) {
	webhookBody, err := webhook.ToBody()
	if err != nil {
		return body, contentType, err
	}
	switch webhookBody := webhookBody.(type) {
	case discord.WebhookMessageCreate:
		body, err = json.Marshal(webhookBody)
		if err != nil {
			return body, contentType, err
		}
		contentType = "application/json"
	case *discord.MultipartBuffer:
		body = webhookBody.Buffer.Bytes()
		contentType = webhookBody.ContentType
	}
	return
}
