package synthient

import (
	"fmt"
	"net/http"
	"net/url"
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
	path, err := url.JoinPath(client.BaseAPI.String(), "feeds", stream, "export")
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
