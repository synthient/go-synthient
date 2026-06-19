package synthient

import (
	"fmt"
	"net/http"
	"net/url"
)

// Domain represents the JSON response returned by the Synthient domain lookup endpoint.
//
// It aggregates activity observed for a single domain across Synthient's sensors:
// rolling event and unique-IP counts, an hourly time series, top ASNs, subdomains,
// and ports, geographic distribution, an hour-by-day-of-week activity heatmap, and a
// sample of recent events. Fields and nested structs map 1:1 to the API's JSON
// payload via struct tags.
//
// Commonly used fields include Domain.Status, Stats.Events24H, and UniqueIPs.Value24H.
type Domain struct {
	Domain string `json:"domain"`
	Status string `json:"status"`
	Stats  struct {
		Events24H      int `json:"events_24h"`
		TotalEvents30D int `json:"total_events_30d"`
	} `json:"stats"`
	TimeSeries []struct {
		Date      int `json:"date"`
		Events    int `json:"events"`
		UniqueIPs int `json:"unique_ips"`
	} `json:"time_series"`
	UniqueIPs struct {
		Value24H     int   `json:"value_24h"`
		Value30D     int   `json:"value_30d"`
		Sparkline24H []int `json:"sparkline_24h"`
	} `json:"unique_ips"`
	TopASN struct {
		ASN    int `json:"asn"`
		Events int `json:"events"`
	} `json:"top_asn"`
	TopSubdomains []struct {
		Subdomain string `json:"subdomain"`
		Count     int    `json:"count"`
	} `json:"top_subdomains"`
	TopPorts []struct {
		Port  int `json:"port"`
		Count int `json:"count"`
	} `json:"top_ports"`
	GeoDistribution []struct {
		CountryCode string `json:"country_code"`
		UniqueIPs   int    `json:"unique_ips"`
		Events      int    `json:"events"`
	} `json:"geo_distribution"`
	HourDowHeatmap [][]int `json:"hour_dow_heatmap"`
	ActivityStats  struct {
		PeakHour      int    `json:"peak_hour"`
		QuietHour     int    `json:"quiet_hour"`
		MedianPerHour int    `json:"median_per_hour"`
		P95PerHour    int    `json:"p95_per_hour"`
		Cadence       string `json:"cadence"`
	} `json:"activity_stats"`
	RecentEvents []struct {
		Timestamp       int    `json:"timestamp"`
		SourceIPMasked  string `json:"source_ip_masked"`
		TargetSubdomain string `json:"target_subdomain"`
		Port            int    `json:"port"`
		CountryCode     string `json:"country_code"`
	} `json:"recent_events"`
}

// GetDomain retrieves traffic statistics and recent activity for a single domain.
//
// It performs an HTTP GET request to the Synthient domain lookup endpoint and
// unmarshals the JSON response into a Domain value. The request is expected to
// return http.StatusOK; non-OK responses are returned as errors.
//
// options can be used to customize request behavior (timeouts, headers, etc.).
//
// Example:
//
//	domain, err := client.GetDomain("google.com", nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("%+v\n", domain)
func (client *Client) GetDomain(domain string, options *RequestOptions) (Domain, error) {
	path, err := url.JoinPath(client.BaseAPI.String(), "lookup", "domain", domain)
	if err != nil {
		return Domain{}, fmt.Errorf("creating path for domain request: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return Domain{}, fmt.Errorf("making request for domain (%s): %w", domain, err)
	}

	resp, err := requestJSON[Domain](options, client, req, http.StatusOK)
	if err != nil {
		return Domain{}, fmt.Errorf("requesting JSON data: %w", err)
	}

	return resp, nil
}
