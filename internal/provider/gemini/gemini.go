package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"vinylquoter/internal/catalog"
	"vinylquoter/internal/provider/jsonparse"
	"vinylquoter/internal/provider/visionpayload"
)

type Client struct {
	APIKey, Model string
	HTTPClient    *http.Client
	MaxRetries    int
	RetryDelay    time.Duration
}

func (c Client) Identify(ctx context.Context, imagePath string) (catalog.Identification, error) {
	mimeType, encodedImage, err := visionpayload.InlineImage(imagePath)
	if err != nil {
		return catalog.Identification{}, err
	}
	payload := map[string]any{"contents": []any{map[string]any{"parts": []any{map[string]string{"text": visionpayload.Prompt()}, map[string]any{"inline_data": map[string]string{"mime_type": mimeType, "data": encodedImage}}}}}, "generationConfig": map[string]any{"temperature": 0, "responseMimeType": "application/json"}}
	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", c.Model, c.APIKey)
	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	retries := c.MaxRetries
	delay := c.RetryDelay
	if delay == 0 {
		delay = 7 * time.Second
	}
	var lastErr error
	for attempt := 0; attempt <= retries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			return catalog.Identification{}, err
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			return catalog.Identification{}, err
		}
		respBody, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return ParseResponse(respBody)
		}
		lastErr = fmt.Errorf("gemini HTTP %d: %s", resp.StatusCode, string(respBody))
		if resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode != http.StatusServiceUnavailable {
			break
		}
		time.Sleep(delay)
	}
	return catalog.Identification{}, lastErr
}

func ParseResponse(body []byte) (catalog.Identification, error) {
	var payload struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return catalog.Identification{}, err
	}
	if len(payload.Candidates) == 0 || len(payload.Candidates[0].Content.Parts) == 0 {
		return catalog.Identification{}, fmt.Errorf("gemini returned no text")
	}
	return jsonparse.IdentificationFromText(payload.Candidates[0].Content.Parts[0].Text)
}
