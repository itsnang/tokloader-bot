package downloader

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type TikTokClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewTikTokClient(cfg Config) *TikTokClient {
	return &TikTokClient{
		baseURL:    strings.TrimRight(cfg.TikwmBaseURL, "/"),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

type tikwmAuthor struct {
	UniqueID string `json:"unique_id"`
}

type tikwmData struct {
	ID       string      `json:"id"`
	Title    string      `json:"title"`
	Author   tikwmAuthor `json:"author"`
	Duration int         `json:"duration"`
	Cover    string      `json:"cover"`
	HDPlay   string      `json:"hdplay"`
	Play     string      `json:"play"`
	Music    string      `json:"music"`
	Images   []string    `json:"images"`
}

type tikwmResponse struct {
	Code int       `json:"code"`
	Msg  string    `json:"msg"`
	Data tikwmData `json:"data"`
}

func (c *TikTokClient) Info(ctx context.Context, tikTokURL string) (*MediaResponse, error) {
	form := url.Values{}
	form.Set("url", tikTokURL)
	form.Set("hd", "1")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "tokloader-bot/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tikwm request: %w", err)
	}
	defer resp.Body.Close()

	var raw tikwmResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if raw.Code != 0 {
		return nil, fmt.Errorf("tikwm error: %s", raw.Msg)
	}

	videoURL := raw.Data.HDPlay
	if videoURL == "" {
		videoURL = raw.Data.Play
	}
	if videoURL == "" && len(raw.Data.Images) == 0 {
		return nil, fmt.Errorf("resolver returned no usable video URL")
	}

	return &MediaResponse{
		ID:       raw.Data.ID,
		Title:    raw.Data.Title,
		Author:   raw.Data.Author.UniqueID,
		Duration: raw.Data.Duration,
		Cover:    raw.Data.Cover,
		VideoURL: videoURL,
		AudioURL: raw.Data.Music,
		Images:   raw.Data.Images,
	}, nil
}
