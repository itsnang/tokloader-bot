package downloader

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type CobaltClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewCobaltClient(cfg Config) *CobaltClient {
	return &CobaltClient{
		baseURL:    strings.TrimRight(cfg.CobaltBaseURL, "/"),
		apiKey:     cfg.CobaltAPIKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

type cobaltRequest struct {
	URL          string `json:"url"`
	DownloadMode string `json:"downloadMode"`
}

type cobaltPickerItem struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type cobaltError struct {
	Code string `json:"code"`
}

type cobaltResponse struct {
	Status string             `json:"status"`
	URL    string             `json:"url"`
	Picker []cobaltPickerItem `json:"picker"`
	Error  *cobaltError       `json:"error"`
}

func (c *CobaltClient) Fetch(ctx context.Context, mediaURL string) (*cobaltResponse, error) {
	body, err := json.Marshal(cobaltRequest{URL: mediaURL, DownloadMode: "auto"})
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "tokloader-bot/1.0")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Api-Key "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cobalt request: %w", err)
	}
	defer resp.Body.Close()

	var cr cobaltResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if cr.Status == "error" {
		code := "unknown"
		if cr.Error != nil {
			code = cr.Error.Code
		}
		return nil, fmt.Errorf("cobalt error: %s", code)
	}
	return &cr, nil
}

func (cr *cobaltResponse) toMediaResponse() *MediaResponse {
	if cr.Status == "picker" {
		images := make([]string, len(cr.Picker))
		for i, item := range cr.Picker {
			images[i] = item.URL
		}
		return &MediaResponse{Images: images}
	}
	return &MediaResponse{VideoURL: cr.URL}
}
