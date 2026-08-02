package downloader

type MediaResponse struct {
	ID       string
	Title    string
	Author   string
	Duration int
	Cover    string
	VideoURL string
	AudioURL string
	Images   []string
}

func (r *MediaResponse) IsImage() bool {
	return len(r.Images) > 0
}
