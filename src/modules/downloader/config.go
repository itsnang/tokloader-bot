package downloader

// Config holds all platform resolver configuration loaded from environment variables.
type Config struct {
	TikwmBaseURL    string // TIKWM_BASE_URL (default: https://www.tikwm.com)
	CobaltBaseURL   string // COBALT_BASE_URL (default: https://api.cobalt.tools)
	InstagramCookie string // INSTAGRAM_COOKIE — required for Instagram stories
	FacebookCookie  string // FACEBOOK_COOKIE — required for Facebook stories
}
