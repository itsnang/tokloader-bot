package downloader

// Config holds all platform resolver configuration loaded from environment variables.
type Config struct {
	TikwmBaseURL string // TIKWM_BASE_URL (default: https://www.tikwm.com)
	CobaltBaseURL string // COBALT_BASE_URL (default: https://api.cobalt.tools)
	CobaltAPIKey  string // COBALT_API_KEY — optional, for hosted cobalt instances that require auth
}
