package synthient

import (
	"fmt"
	"net/http"
	"net/url"
)

type IP struct {
	IP      string `json:"ip"`
	Network struct {
		Asn  int    `json:"asn"`
		Isp  string `json:"isp"`
		Type string `json:"type"`
		// Org        interface{} `json:"org"`
		// AbuseEmail interface{} `json:"abuse_email"`
		// AbusePhone interface{} `json:"abuse_phone"`
		// Domain     interface{} `json:"domain"`
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
		// Devices     []interface{} `json:"devices"`
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

func (client *Client) GetIP(ip string, options *RequestOptions) (IP, error) {
	path, err := url.JoinPath(client.Base.Path, "lookup", "ip", ip)
	if err != nil {
		return IP{}, fmt.Errorf("%w failed to create path for resource request", err)
	}
	client.Base.Path = path
	req, err := http.NewRequest(
		http.MethodGet,
		client.Base.String(),
		nil,
	)
	if err != nil {
		return IP{}, fmt.Errorf("%w failed to make request for ip \"%s\"", err, ip)
	}

	resp, err := request[IP](options, client, req, http.StatusOK)
	if err != nil {
		return IP{}, err
	}

	return resp, nil
}
