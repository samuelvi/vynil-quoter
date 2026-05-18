package lmstudio

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"vinylquoter/internal/catalog"
	"vinylquoter/internal/provider"
	"vinylquoter/internal/provider/jsonparse"
	"vinylquoter/internal/provider/visionpayload"
)

type Client struct {
	BaseURL, Model string
	HTTPClient     *http.Client
	Timeout        time.Duration
}

func (c Client) Identify(ctx context.Context, request provider.RecognitionRequest) (catalog.Identification, error) {
	imageURL, err := visionpayload.DataURL(request.ImagePath)
	if err != nil {
		return catalog.Identification{}, err
	}
	payload := map[string]any{"model": c.Model, "temperature": 0, "messages": []any{map[string]any{"role": "user", "content": []any{map[string]string{"type": "text", "text": visionpayload.PromptForCondition(request.MediaCondition, request.SleeveCondition)}, map[string]any{"type": "image_url", "image_url": map[string]string{"url": imageURL}}}}}}
	body, _ := json.Marshal(payload)
	baseURL := strings.TrimRight(c.BaseURL, "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return catalog.Identification{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return catalog.Identification{}, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return catalog.Identification{}, fmt.Errorf("lm studio HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	return ParseResponse(respBody)
}

func ParseResponse(body []byte) (catalog.Identification, error) {
	var payload struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return catalog.Identification{}, err
	}
	if len(payload.Choices) == 0 || strings.TrimSpace(payload.Choices[0].Message.Content) == "" {
		return catalog.Identification{}, fmt.Errorf("lm studio returned no text")
	}
	return jsonparse.IdentificationFromText(payload.Choices[0].Message.Content)
}
