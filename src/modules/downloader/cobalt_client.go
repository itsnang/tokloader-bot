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

// CobaltClient calls the cobalt.tools media resolver API.
// Shared by the Instagram and Facebook providers.
type CobaltClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewCobaltClient(cfg Config) *CobaltClient {
	baseURL := strings.TrimRight(cfg.CobaltBaseURL, "/")
	if baseURL == "" {
		baseURL = "https://api.cobalt.tools"
	}
	return &CobaltClient{
		baseURL:    baseURL,
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

// Fetch sends a request to cobalt and returns the parsed response.
// cookie is optional; when non-empty it is forwarded as a Cookie header (needed for stories).
func (c *CobaltClient) Fetch(ctx context.Context, mediaURL, mode, cookie string) (*cobaltResponse, error) {
	body, err := json.Marshal(cobaltRequest{URL: mediaURL, DownloadMode: mode})
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
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
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

// toMediaResponse converts a cobalt API response to a MediaResponse.
func (cr *cobaltResponse) toMediaResponse() *MediaResponse {
	if cr.Status == "picker" {
		images := make([]string, 0, len(cr.Picker))
		for _, item := range cr.Picker {
			images = append(images, item.URL)
		}
		return &MediaResponse{Images: images}
	}
	return &MediaResponse{VideoURL: cr.URL}
}
