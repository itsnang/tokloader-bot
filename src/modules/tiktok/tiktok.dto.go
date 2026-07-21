package tiktok

// InfoResponse holds resolved TikTok post data.
type InfoResponse struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Author      string   `json:"author"`
	Duration    int      `json:"duration"`
	Cover       string   `json:"cover"`
	NoWatermark string   `json:"no_watermark"`
	Watermark   string   `json:"watermark"`
	Music       string   `json:"music"`
	Images      []string `json:"images"`
}

// IsImage reports whether the post is a photo slideshow rather than a video.
func (r *InfoResponse) IsImage() bool {
	return len(r.Images) > 0
}
