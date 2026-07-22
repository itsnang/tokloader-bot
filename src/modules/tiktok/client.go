package tiktok

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{},
	}
}

type resolverAuthor struct {
	UniqueID string `json:"unique_id"`
}

type resolverData struct {
	ID       string         `json:"id"`
	Title    string         `json:"title"`
	Author   resolverAuthor `json:"author"`
	Duration int            `json:"duration"`
	Cover    string         `json:"cover"`
	HDPlay   string         `json:"hdplay"`
	Play     string         `json:"play"`
	WMPlay   string         `json:"wmplay"`
	Music    string         `json:"music"`
	Images   []string       `json:"images"`
}

type resolverResponse struct {
	Code int          `json:"code"`
	Msg  string       `json:"msg"`
	Data resolverData `json:"data"`
}

// Info resolves a TikTok URL and returns post metadata.
func (c *Client) Info(ctx context.Context, tikTokURL string) (*InfoResponse, error) {
	form := url.Values{}
	form.Set("url", tikTokURL)
	form.Set("hd", "1")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tikwm request: %w", err)
	}
	defer resp.Body.Close()

	var raw resolverResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if raw.Code != 0 {
		return nil, fmt.Errorf("tikwm error: %s", raw.Msg)
	}

	noWatermark := raw.Data.HDPlay
	if noWatermark == "" {
		noWatermark = raw.Data.Play
	}

	return &InfoResponse{
		ID:          raw.Data.ID,
		Title:       raw.Data.Title,
		Author:      raw.Data.Author.UniqueID,
		Duration:    raw.Data.Duration,
		Cover:       raw.Data.Cover,
		NoWatermark: noWatermark,
		Watermark:   raw.Data.WMPlay,
		Music:       raw.Data.Music,
		Images:      raw.Data.Images,
	}, nil
}
