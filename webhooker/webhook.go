package webhooker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/disgoorg/disgo/discord"
)

type ContentType string

const (
	ContentTypeJSON      ContentType = "application/json"
	ContentTypeMultipart ContentType = "multipart/form-data"
)

func Send[T discord.WebhookMessageCreate | []byte | string](webhookURL string, payload T, contentType ...ContentType) (err error) {
	var bodyBytes []byte
	var cType ContentType = ContentTypeJSON

	if len(contentType) == 1 {
		cType = contentType[0]
	}
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
			cType = ContentTypeMultipart
		}
	case string:
		bodyBytes = []byte(payload)
	default:
		return fmt.Errorf("unsupported payload type")
	}

	client := http.Client{}
	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", string(cType))

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send webhook, status code: %d", resp.StatusCode)
	}
	return nil
}

func SendWithWait[T discord.WebhookMessageCreate | []byte | string](webhookURL string, payload T, contentType ...ContentType) (m discord.Message, err error) {
	var bodyBytes []byte
	var cType ContentType = ContentTypeJSON

	if len(contentType) == 1 {
		cType = contentType[0]
	}
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
			cType = ContentTypeMultipart
		}
	case string:
		bodyBytes = []byte(payload)
	default:
		return m, fmt.Errorf("unsupported payload type")
	}

	client := http.Client{}
	req, err := http.NewRequest("POST", webhookURL+"?wait=true", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return m, err
	}
	req.Header.Set("Content-Type", string(cType))

	resp, err := client.Do(req)
	if err != nil {
		return m, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return m, fmt.Errorf("failed to send webhook, status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return m, err
	}
	return m, json.Unmarshal(body, &m)
}

func Edit[T discord.WebhookMessageUpdate | []byte | string](webhookURL string, msgID string, payload T, contentType ...ContentType) (err error) {
	var bodyBytes []byte
	var cType ContentType = ContentTypeJSON

	if len(contentType) == 1 {
		cType = contentType[0]
	}

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
			cType = ContentTypeMultipart
		}
	case string:
		bodyBytes = []byte(payload)
	default:
		return fmt.Errorf("unsupported payload type")
	}

	client := http.Client{}
	req, err := http.NewRequest("PATCH", fmt.Sprintf("%s/messages/%s", webhookURL, msgID), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", string(cType))

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update webhook, status code: %d", resp.StatusCode)
	}

	return
}
