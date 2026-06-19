package synthient

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

// FeedSnapshot represents a single Parquet snapshot entry returned by the feeds export endpoint.
type FeedSnapshot struct {
	Kind         string `json:"kind"`
	Date         string `json:"date"`
	Hour         *int   `json:"hour,omitempty"`
	SizeBytes    int64  `json:"size_bytes"`
	RowCount     int64  `json:"row_count"`
	Checksum     string `json:"checksum"`
	ID           string `json:"id"`
	CreatedAt    int64  `json:"created_at"`
	DownloadPath string `json:"download_path"`
}

// FeedSnapshotsPage is a single page of results from FeedSnapshots.
type FeedSnapshotsPage struct {
	Stream     string         `json:"stream"`
	Feeds      []FeedSnapshot `json:"feeds"`
	NextCursor string         `json:"next_cursor"`
}

// FeedSnapshotsOptions controls pagination for the FeedSnapshots call.
type FeedSnapshotsOptions struct {
	// Limit is the page size. Defaults to 100; values above 500 are clamped by the API.
	Limit int
	// Cursor is the opaque pagination token from the previous page's NextCursor field.
	// Leave empty on the first call.
	Cursor string
}

// FeedSnapshots returns one page of available daily and hourly Parquet snapshots for the
// given stream. Pages are ordered newest-first and capped at 500 rows by the API.
//
// stream must be one of: proxies, anonymizers, torrents, honeypot_http, honeypot_https,
// honeypot_dns, or honeypot_adb.
//
// Pass FeedSnapshotsPage.NextCursor back via opts.Cursor to fetch the next page.
// NextCursor is empty on the final page.
//
// Example:
//
//	page, err := client.FeedSnapshots("proxies", &synthient.FeedSnapshotsOptions{Limit: 50}, nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	for _, snap := range page.Feeds {
//		fmt.Printf("%s %s %d bytes\n", snap.Kind, snap.ID, snap.SizeBytes)
//	}
func (client *Client) FeedSnapshots(
	stream string,
	options *FeedSnapshotsOptions,
	requestOptions *RequestOptions,
) (FeedSnapshotsPage, error) {
	segments := append([]string{"feeds"}, feedStreamPath(stream)...)
	segments = append(segments, "export")
	path, err := url.JoinPath(client.BaseAPI.String(), segments...)
	if err != nil {
		return FeedSnapshotsPage{}, fmt.Errorf("creating path for feed snapshots request: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return FeedSnapshotsPage{}, fmt.Errorf(
			"making request for feed snapshots (%s): %w",
			stream,
			err,
		)
	}

	if options != nil {
		q := req.URL.Query()
		if options.Limit > 0 {
			q.Set("limit", strconv.Itoa(options.Limit))
		}
		if options.Cursor != "" {
			q.Set("cursor", options.Cursor)
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := requestJSON[FeedSnapshotsPage](requestOptions, client, req, http.StatusOK)
	if err != nil {
		return FeedSnapshotsPage{}, fmt.Errorf("requesting JSON data: %w", err)
	}

	return resp, nil
}

// FeedSnapshotMeta holds the metadata returned for a single Parquet snapshot.
type FeedSnapshotMeta struct {
	Stream    string `json:"stream"`
	Kind      string `json:"kind"`
	Hour      *int   `json:"hour,omitempty"`
	ID        string `json:"id"`
	Format    string `json:"format"`
	Date      int64  `json:"date"`
	CreatedAt int64  `json:"created_at"`
	Size      int64  `json:"size"`
	Rows      int64  `json:"rows"`
	Checksum  string `json:"checksum"`
	Schema    struct {
		Fields []struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"fields"`
	} `json:"schema"`
}

// FeedSnapshotMeta returns JSON metadata for a single Parquet snapshot, including its
// SHA-256 checksum, byte size, row count, parquet schema, and canonical date.
//
// stream must be one of: proxies, anonymizers, torrents, honeypot_http, honeypot_https,
// honeypot_dns, or honeypot_adb.
//
// date is the snapshot identifier: YYYY-MM-DD for daily rollups, YYYY-MM-DD/HH for past
// hourlies, or "latest" for the most recent hourly snapshot.
//
// Example:
//
//	meta, err := client.FeedSnapshotMeta("proxies", "latest", nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("rows=%d checksum=%s\n", meta.Rows, meta.Checksum)
func (client *Client) FeedSnapshotMeta(
	stream string,
	date string,
	requestOptions *RequestOptions,
) (FeedSnapshotMeta, error) {
	segments := append([]string{"feeds"}, feedStreamPath(stream)...)
	segments = append(segments, "export", date, "meta")
	path, err := url.JoinPath(client.BaseAPI.String(), segments...)
	if err != nil {
		return FeedSnapshotMeta{}, fmt.Errorf("creating path for feed snapshot meta request: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return FeedSnapshotMeta{}, fmt.Errorf(
			"making request for feed snapshot meta (%s, %s): %w",
			stream,
			date,
			err,
		)
	}

	resp, err := requestJSON[FeedSnapshotMeta](requestOptions, client, req, http.StatusOK)
	if err != nil {
		return FeedSnapshotMeta{}, fmt.Errorf("requesting JSON data: %w", err)
	}

	return resp, nil
}

// feedStreamPath maps a public stream name to its API path segments. Honeypot
// streams are served under helio/<protocol>; every other stream uses its name.
func feedStreamPath(stream string) []string {
	switch stream {
	case "honeypot_http":
		return []string{"helio", "http"}
	case "honeypot_https":
		return []string{"helio", "https"}
	case "honeypot_dns":
		return []string{"helio", "dns"}
	case "honeypot_adb":
		return []string{"helio", "adb"}
	default:
		return []string{stream}
	}
}

func downloadFeed(
	client *Client,
	requestOptions *RequestOptions,
	date string,
	hour *int,
	filename string,
	pathPrefixSegments ...string,
) (io.ReadCloser, error) {
	label := pathPrefixSegments[len(pathPrefixSegments)-1]
	segments := append(pathPrefixSegments, "export", date)
	if hour != nil {
		segments = append(segments, strconv.Itoa(*hour))
	}

	path, err := url.JoinPath(client.BaseAPI.String(), segments...)
	if err != nil {
		return nil, fmt.Errorf("creating path for %s download request: %w", label, err)
	}

	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("making request for %s download (%s): %w", label, date, err)
	}

	body, err := request(requestOptions, client, req, http.StatusOK)
	if err != nil {
		return nil, fmt.Errorf("requesting %s snapshot: %w", label, err)
	}

	if filename == "" {
		return body, nil
	}

	defer func() { _ = body.Close() }()
	f, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("creating file %s: %w", filename, err)
	}
	defer func() { _ = f.Close() }()
	_, err = io.Copy(f, body)
	if err != nil {
		return nil, fmt.Errorf("writing %s snapshot to %s: %w", label, filename, err)
	}
	return nil, nil
}

// DownloadFeedSnapshot downloads a Parquet snapshot and returns a streaming reader for its
// contents. The API issues a 307 redirect to a presigned URL valid for 24 hours; this
// method follows the redirect automatically.
//
// stream must be one of: proxies, anonymizers, torrents, honeypot_http, honeypot_https,
// honeypot_dns, or honeypot_adb.
//
// date accepts "latest" for the most recent hourly snapshot, or a YYYY-MM-DD string for
// a daily rollup. For a specific hourly within the current UTC day, set hour to a non-nil
// pointer in the range 0–23.
//
// If filename is non-empty the snapshot is written to that file and the returned reader is
// nil. If filename is empty the caller receives the raw reader and must close it.
//
// Example (write to file):
//
//	hour := 21
//	_, err := client.DownloadFeedSnapshot("proxies", "2026-05-07", &hour, "proxies.parquet", nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//
// Example (stream reader):
//
//	r, err := client.DownloadFeedSnapshot("proxies", "2026-05-07", nil, "", nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer r.Close()
func (client *Client) DownloadFeedSnapshot(
	stream string,
	date string,
	hour *int,
	filename string,
	requestOptions *RequestOptions,
) (io.ReadCloser, error) {
	segments := append([]string{"feeds"}, feedStreamPath(stream)...)
	return downloadFeed(client, requestOptions, date, hour, filename, segments...)
}
