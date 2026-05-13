package synthient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// IP represents the JSON response returned by the Synthient IP lookup endpoint.
//
// It groups data into three major sections:
//
//   - Network: ASN/ISP and ownership/abuse contacts for the IP’s network.
//   - Location: coarse geolocation attributes associated with the IP.
//   - IPData: device/behavior/category/enrichment signals and an overall risk score.
//
// Fields and nested structs map 1:1 to the API’s JSON payload via struct tags.
// Note that values (especially geolocation and “risk”) are provider-derived and
// may be approximate.
//
// Commonly used fields include IP.IP, Network.Asn/Network.Isp, Location.Country,
// and IPData.IPRisk.
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
	Intelligence struct {
		RiskScore  int      `json:"risk_score"`
		Behavior   []string `json:"behavior"`
		Categories []string `json:"categories"`
		Devices    []struct {
			OS      string `json:"os"`
			Version string `json:"version"`
		} `json:"devices"`
		Providers []struct {
			Provider string `json:"provider"`
			Type     string `json:"type"`
			LastSeen int64  `json:"last_seen"`
		} `json:"providers"`
	} `json:"intelligence"`
}

// GetIP looks up enrichment data for a single IP address.
//
// It performs an HTTP GET request to the Synthient IP lookup endpoint and
// unmarshals the JSON response into an IP value. The request is expected to
// return http.StatusOK; non-OK responses are returned as errors.
//
// options can be used to customize request behavior (timeouts, headers, etc.).
//
// Example:
//
//	info, err := client.GetIP("8.8.8.8", nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("%+v\n", info)
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

// GetIPs looks up enrichment data for multiple IP addresses in a single request.
//
// It performs an HTTP POST request to the Synthient bulk IP lookup endpoint,
// sending the provided IPs as a JSON body, and returns the results in the same
// order as the input slice. The request is expected to return http.StatusOK;
// non-OK responses are returned as errors.
//
// options can be used to customize request behavior (timeouts, headers, etc.).
//
// Example:
//
//	results, err := client.GetIPs([]string{"8.8.8.8", "1.1.1.1"}, nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	for _, info := range results {
//		fmt.Printf("%s: risk=%d\n", info.IP, info.Intelligence.RiskScore)
//	}
func (client *Client) GetIPs(ips []string, options *RequestOptions) ([]IP, error) {
	path, err := url.JoinPath(client.BaseAPI.String(), "lookup", "ips")
	if err != nil {
		return []IP{}, fmt.Errorf("creating path for ips request: %w", err)
	}

	body, err := json.Marshal(struct {
		IPs []string `json:"ips"`
	}{IPs: ips})
	if err != nil {
		return []IP{}, fmt.Errorf("encoding ips request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	if err != nil {
		return []IP{}, fmt.Errorf("making request for ips: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := requestJSON[struct {
		Results []IP `json:"results"`
	}](options, client, req, http.StatusOK)
	if err != nil {
		return []IP{}, fmt.Errorf("requesting JSON data: %w", err)
	}

	return resp.Results, nil
}
