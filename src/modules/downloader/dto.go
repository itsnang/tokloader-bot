package downloader

// MediaResponse holds resolved media data from any supported platform.
type MediaResponse struct {
	ID       string
	Title    string
	Author   string
	Duration int
	Cover    string
	VideoURL string   // watermark-free video URL; empty for image posts
	AudioURL string   // background audio URL; empty if unavailable
	Images   []string // non-empty = photo slideshow
}

// IsImage reports whether the post is a photo slideshow rather than a video.
func (r *MediaResponse) IsImage() bool {
	return len(r.Images) > 0
}
