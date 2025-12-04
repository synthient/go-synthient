package synthient

import (
	"net/http"
	"net/url"
)

type Client struct {
	HttpClient *http.Client
	Token      string
	Base       *url.URL
}

func NewClient(token string) Client {
	return Client{
		HttpClient: &http.Client{},
		Token:      token,
		Base: &url.URL{
			Scheme: "https",
			Host:   "v3api.synthient.com",
			Path:   "/api/v3",
		},
	}
}
