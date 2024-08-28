package webhooker

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/tidwall/gjson"
)

type ContentType string

const (
	ContentTypeJSON      ContentType = "application/json"
	ContentTypeMultipart ContentType = "multipart/form-data"
)

type WebhookOpts struct {
	rateLimitRetries int
	rateLimitDelay   time.Duration
	contentType      ContentType
}

type WebhookOptFunc func(opts *WebhookOpts)

func DefaultWebhookOpts() *WebhookOpts {
	return &WebhookOpts{
		rateLimitRetries: 0,
		rateLimitDelay:   0,
		contentType:      ContentTypeJSON,
	}
}

func RateLimitWebhookOpt(retries int, delay time.Duration) WebhookOptFunc {
	return func(opts *WebhookOpts) {
		opts.rateLimitRetries = retries
		opts.rateLimitDelay = delay
	}
}

func ContentTypeWebhookOpt(contentType ContentType) WebhookOptFunc {
	return func(opts *WebhookOpts) {
		opts.contentType = contentType
	}
}

func Send[T discord.WebhookMessageCreate | []byte | string](webhookURL string, payload T, opts ...WebhookOptFunc) (err error) {
	o := DefaultWebhookOpts()
	for _, opt := range opts {
		opt(o)
	}

	var bodyBytes []byte
	contentType := o.contentType

	switch payload := any(payload).(type) {
	case discord.WebhookMessageCreate:
		body, err := payload.ToBody()
		if err != nil {
			return err
		}
		switch body := body.(type) {
		case discord.WebhookMessageCreate:
			bodyBytes, err = json.Marshal(body)
			if err != nil {
				return err
			}
		case *discord.MultipartBuffer:
			bodyBytes = body.Buffer.Bytes()
			contentType = ContentTypeMultipart
		}
	case string:
		bodyBytes = []byte(payload)
	default:
		return fmt.Errorf("unsupported payload type")
	}

	retries := 0
	for {
		client := http.Client{}
		req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(bodyBytes))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", string(contentType))
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			if resp.StatusCode == 429 && o.rateLimitRetries > 0 && retries < o.rateLimitRetries {
				resJ := gjson.ParseBytes(body)
				if strings.Contains(resJ.Get("message").Str, "rate limited") {
					secs := resJ.Get("retry_after").Float()
					time.Sleep(time.Second*time.Duration(secs) + time.Millisecond*100 + o.rateLimitDelay)
					retries++
					continue
				} else {
					return errors.New("invalid rate limit response, status code")
				}
			}
			return fmt.Errorf("failed to send webhook, status code: %d, body: %s", resp.StatusCode, string(body))
		}
		return nil
	}
}

func SendWithWait[T discord.WebhookMessageCreate | []byte | string](webhookURL string, payload T, opts ...WebhookOptFunc) (m discord.Message, err error) {
	o := DefaultWebhookOpts()
	for _, opt := range opts {
		opt(o)
	}
	var bodyBytes []byte
	contentType := o.contentType

	switch payload := any(payload).(type) {
	case discord.WebhookMessageCreate:
		body, err := payload.ToBody()
		if err != nil {
			return m, err
		}
		switch body := body.(type) {
		case discord.WebhookMessageCreate:
			bodyBytes, err = json.Marshal(body)
			if err != nil {
				return m, err
			}
		case *discord.MultipartBuffer:
			bodyBytes = body.Buffer.Bytes()
			contentType = ContentTypeMultipart
		}
	case string:
		bodyBytes = []byte(payload)
	default:
		return m, fmt.Errorf("unsupported payload type")
	}

	retries := 0
	for {
		client := http.Client{}
		req, err := http.NewRequest("POST", webhookURL+"?wait=true", bytes.NewBuffer(bodyBytes))
		if err != nil {
			return m, err
		}
		req.Header.Set("Content-Type", string(contentType))
		resp, err := client.Do(req)
		if err != nil {
			return m, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return m, err
			}
			if resp.StatusCode == 429 && o.rateLimitRetries > 0 && retries < o.rateLimitRetries {
				resJ := gjson.ParseBytes(body)
				if strings.Contains(resJ.Get("message").Str, "rate limited") {
					secs := resJ.Get("retry_after").Float()
					time.Sleep(time.Second*time.Duration(secs) + time.Millisecond*100 + o.rateLimitDelay)
					retries++
					continue
				} else {
					return m, errors.New("invalid rate limit response, status code")
				}
			}
			return m, fmt.Errorf("failed to send webhook, status code: %d, body: %s", resp.StatusCode, string(body))
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return m, err
		}
		return m, json.Unmarshal(body, &m)
	}
}

func Edit[T discord.WebhookMessageUpdate | []byte | string](webhookURL string, msgID string, payload T, opts ...WebhookOptFunc) (err error) {
	o := DefaultWebhookOpts()
	for _, opt := range opts {
		opt(o)
	}
	var bodyBytes []byte
	contentType := o.contentType

	switch payload := any(payload).(type) {
	case discord.WebhookMessageUpdate:
		body, err := payload.ToBody()
		if err != nil {
			return err
		}
		switch body := body.(type) {
		case discord.WebhookMessageUpdate:
			bodyBytes, err = json.Marshal(body)
			if err != nil {
				return err
			}
		case *discord.MultipartBuffer:
			bodyBytes = body.Buffer.Bytes()
			contentType = ContentTypeMultipart
		}
	case string:
		bodyBytes = []byte(payload)
	default:
		return fmt.Errorf("unsupported payload type")
	}

	retries := 0
	for {
		client := http.Client{}
		req, err := http.NewRequest("PATCH", fmt.Sprintf("%s/messages/%s", webhookURL, msgID), bytes.NewBuffer(bodyBytes))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", string(contentType))
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			if resp.StatusCode == 429 && o.rateLimitRetries > 0 && retries < o.rateLimitRetries {
				resJ := gjson.ParseBytes(body)
				if strings.Contains(resJ.Get("message").Str, "rate limited") {
					secs := resJ.Get("retry_after").Float()
					time.Sleep(time.Second*time.Duration(secs) + time.Millisecond*100 + o.rateLimitDelay)
					retries++
					continue
				} else {
					return errors.New("invalid rate limit response, status code")
				}
			}
			return fmt.Errorf("failed to send webhook, status code: %d, body: %s", resp.StatusCode, string(body))
		}
		return nil
	}
}
