package synthient

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

// AnonymizersQuery defines the set of filters and output options used when
// requesting the Synthient anonymizers feed.
//
// Fields are translated into HTTP query parameters by StreamAnonymizersFeed /
// DownloadAnonymizersFeed. Leave string fields empty to omit that filter.
//
// Typical values include:
//   - Provider: feed source/provider identifier (e.g. "BIRDPROXIES").
//   - Type: anonymizer category/type (e.g. "RESIDENTIAL_PROXY").
//   - LastObserved: recency window for when an entry was last observed
//     (API-specific, e.g. "7D").
//   - CountryCode: ISO 3166-1 alpha-2 country code (e.g. "US").
//   - Format: output format (e.g. "CSV").
//   - Full: when true, request the “full” dataset if supported by the API.
//   - Order: sort order.
type AnonymizersQuery struct {
	Provider     string
	Type         string
	LastObserved string
	CountryCode  string
	Format       string
	Full         bool
	Order        string
}

// StreamAnonymizersFeed starts a streaming HTTP GET request for the Synthient
// “anonymizers” feed and returns the response body as an io.ReadCloser.
//
// The returned reader contains the raw feed payload (for example, CSV when
// query.Format is "CSV"). Callers MUST ALWAYS close the returned ReadCloser.
// For large feeds, prefer streaming consumption (io.Copy, bufio.Scanner, or a
// CSV reader) instead of reading the entire body into memory.
//
// Query fields are translated into request parameters:
//   - Provider     -> provider
//   - Type         -> type
//   - LastObserved -> last_observed
//   - CountryCode  -> country_code
//   - Full         -> full
//   - Format       -> format
//   - Order        -> order
//
// Request behavior (timeouts, headers, etc.) can be customized via options.
// The request is expected to return http.StatusOK; non-OK responses are
// returned as errors.
//
// Example:
//
//	stream, err := client.StreamAnonymizersFeed(synthient.AnonymizersQuery{
//		Provider:     "BIRDPROXIES",
//		Type:         "RESIDENTIAL_PROXY",
//		LastObserved: "7D",
//		Format:       "CSV",
//		CountryCode:  "US",
//		Full:         false,
//		Order:        "desc",
//	}, nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer stream.Close()
func (client *Client) StreamAnonymizersFeed(
	query AnonymizersQuery,
	options *RequestOptions,
) (io.ReadCloser, error) {
	params := url.Values{}
	if query.Provider != "" {
		params.Add("provider", query.Provider)
	}
	if query.Type != "" {
		params.Add("type", query.Type)
	}
	if query.LastObserved != "" {
		params.Add("last_observed", query.LastObserved)
	}
	if query.CountryCode != "" {
		params.Add("country_code", query.CountryCode)
	}
	params.Add("full", strconv.FormatBool(query.Full))
	params.Add("format", query.Format)
	params.Add("order", query.Order)

	path, err := url.JoinPath(client.BaseFeeds.String(), "feeds", "anonymizers")
	if err != nil {
		return nil, fmt.Errorf("creating url for anonymizer feed: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s?%s", path, params.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request for anonymizer data: %w", err)
	}

	reader, err := request(options, client, req, http.StatusOK)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	return reader, nil
}

// DownloadAnonymizersFeed downloads the Synthient “anonymizers” feed to a file.
//
// This is a convenience wrapper around StreamAnonymizersFeed that streams the
// HTTP response body directly to disk (via io.Copy) to avoid buffering the
// entire feed in memory. It returns the number of bytes written.
//
// filepath must not already exist. If it does, DownloadAnonymizersFeed returns
// ErrFileExists (wrapped) and does not modify the filesystem. On success, the
// file is created, written, and fsynced (file.Sync) before returning.
//
// Request behavior (timeouts, headers, etc.) can be customized via options.
// The query is interpreted the same way as StreamAnonymizersFeed.
//
// Example:
//
//	n, err := client.DownloadAnonymizersFeed(synthient.AnonymizersQuery{
//		Provider:     "BIRDPROXIES",
//		Type:         "RESIDENTIAL_PROXY",
//		LastObserved: "7D",
//		Format:       "CSV",
//		CountryCode:  "US",
//		Full:         false,
//		Order:        "desc",
//	}, "anonymizers.csv", nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	log.Printf("wrote %d bytes\n", n)
func (client *Client) DownloadAnonymizersFeed(
	query AnonymizersQuery,
	filepath string,
	options *RequestOptions,
) (int64, error) {
	_, err := os.Stat(filepath)
	if !errors.Is(err, fs.ErrNotExist) {
		return 0, fmt.Errorf("creating file at %s: %w", filepath, ErrFileExists)
	}

	body, err := client.StreamAnonymizersFeed(query, options)
	if err != nil {
		return 0, fmt.Errorf("making request: %w", err)
	}
	defer func() { _ = body.Close() }()

	file, err := os.Create(filepath)
	if err != nil {
		return 0, fmt.Errorf("creating output file (path: %s): %w", filepath, err)
	}
	defer func() { _ = file.Close() }()

	bytes, err := io.Copy(file, body)
	if err != nil {
		return 0, fmt.Errorf("streaming response to file: %w", err)
	}

	err = file.Sync()
	if err != nil {
		return 0, fmt.Errorf("syncing output to file: %w", err)
	}

	return bytes, nil
}
