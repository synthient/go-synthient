package synthient

import (
	"net/http"
	"net/url"
)

type Client struct {
	HttpClient *http.Client
	Token      string
	BaseAPI    url.URL
	BaseFeeds  url.URL
}

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
