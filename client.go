package synthient

import (
	"net/http"
	"net/url"
)

// Client is a Synthient API client.
//
// It holds the HTTP transport and configuration required to make API and feed
// requests.
//
// Fields:
//   - HttpClient is the underlying HTTP client used for all requests. If nil,
//     the package may fall back to http.DefaultClient (depending on request
//     helpers).
//   - Token is the API token used for authentication.
//   - BaseAPI is the base URL for JSON API endpoints (e.g. lookups).
//   - BaseFeeds is the base URL for feed endpoints that may return large,
//     streamable payloads (e.g. CSV feeds).
type Client struct {
	HttpClient *http.Client
	Token      string
	BaseAPI    url.URL
	BaseFeeds  url.URL
}

// NewClient constructs a Client configured for the Synthient v3 API.
//
// The returned client is initialized with:
//   - a new *http.Client as the underlying transport,
//   - the provided token for authentication, and
//   - default base URLs for the JSON API (BaseAPI) and feeds service (BaseFeeds).
//
// If you need custom timeouts, proxies, or transports, modify c.HttpClient after
// construction. If Synthient endpoints differ for your environment, you may also
// override BaseAPI and/or BaseFeeds.
//
// Example:
//
//	c := synthient.NewClient(os.Getenv("SYNTHIENT_API_KEY"))
//	c.HttpClient.Timeout = 30 * time.Second
func NewClient(token string) Client {
	return Client{
		HttpClient: &http.Client{},
		Token:      token,
		BaseAPI: url.URL{
			Scheme: "https",
			Host:   "v3api.synthient.com",
			Path:   "/api/v3",
		},
		BaseFeeds: url.URL{
			Scheme: "https",
			Host:   "feeds.synthient.com",
			Path:   "/v3",
		},
	}
}
