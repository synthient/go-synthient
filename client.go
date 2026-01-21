package synthient

import (
	"net/http"
	"net/url"
)

// Client is the primary API client for interacting with the Synthient v3 API.
//
// HttpClient is used to execute requests (defaults to a zero-config http.Client in NewClient).
// Token is sent as the Authorization header value for authenticated endpoints.
// BaseAPI is the base URL for JSON API endpoints.
// BaseFeeds is the base URL for feed download endpoints.
type Client struct {
	HttpClient *http.Client
	Token      string
	BaseAPI    url.URL
	BaseFeeds  url.URL
}

// NewClient constructs a Client configured for the Synthient v3 API.
//
// The returned client uses the provided token for authentication and sets the
// default BaseAPI and BaseFeeds endpoints. Callers may override HttpClient and/or
// base URLs after construction if needed.
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
