package synthient

import (
	"fmt"
	"net/http"
	"net/url"
)

// IP represents the API response for an IP lookup.
type IP struct {
	IP      string `json:"ip"`
	Network struct {
		Asn        int    `json:"asn"`
		Isp        string `json:"isp"`
		Type       string `json:"type"`
		Org        string `json:"org"`
		AbuseEmail string `json:"abuse_email"`
		AbusePhone string `json:"abuse_phone"`
		Domain     string `json:"domain"`
	} `json:"network"`
	Location struct {
		Country   string  `json:"country"`
		State     string  `json:"state"`
		City      string  `json:"city"`
		Timezone  string  `json:"timezone"`
		Longitude float64 `json:"longitude"`
		Latitude  float64 `json:"latitude"`
		GeoHash   string  `json:"geo_hash"`
	} `json:"location"`
	IPData struct {
		Devices []struct {
			OS      string `json:"os"`
			Version string `json:"version"`
		} `json:"devices"`
		DeviceCount int      `json:"device_count"`
		Behavior    []string `json:"behavior"`
		Categories  []string `json:"categories"`
		Enriched    []struct {
			Provider string `json:"provider"`
			Type     string `json:"type"`
			LastSeen string `json:"last_seen"`
		} `json:"enriched"`
		IPRisk int `json:"ip_risk"`
	} `json:"ip_data"`
}

// GetIP looks up enrichment details for a single IPv4/IPv6 address.
//
// It performs an HTTP GET request to the API endpoint:
//
//	{BaseAPI}/lookup/ip/{ip}
//
// The provided ip is inserted into the request path (and should be a valid IP
// string). Request behavior such as context can be configured via options. A successful response
// (HTTP 200 OK) is decoded as JSON into an IP value and returned.
func (client *Client) GetIP(ip string, options *RequestOptions) (IP, error) {
	path, err := url.JoinPath(client.BaseAPI.String(), "lookup", "ip", ip)
	if err != nil {
		return IP{}, fmt.Errorf("creating path for ip request: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return IP{}, fmt.Errorf("making request for IP (%s): %w", ip, err)
	}

	resp, err := requestJSON[IP](options, client, req, http.StatusOK)
	if err != nil {
		return IP{}, fmt.Errorf("requesting JSON data: %w", err)
	}

	return resp, nil
}
