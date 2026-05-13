package synthient

import (
	"encoding/json"
	"fmt"
	"iter"
	"net/http"
	"net/url"
)

// ProxyEvent is a single observation delivered by the proxies real-time stream.
type ProxyEvent struct {
	IP          string `json:"ip"`
	Provider    string `json:"provider"`
	Type        string `json:"type"`
	Timestamp   int64  `json:"timestamp"`
	CountryCode string `json:"country_code"`
	ASN         int    `json:"asn"`
}

// TorrentEvent is a single observation delivered by the torrents real-time stream.
type TorrentEvent struct {
	InfoHash    string `json:"info_hash"`
	Name        string `json:"name"`
	MagnetURI   string `json:"magnet_uri"`
	TotalSize   int64  `json:"total_size"`
	PieceLength int64  `json:"piece_length"`
	FileCount   int    `json:"file_count"`
	Files       []struct {
		Path   string `json:"path"`
		Length int64  `json:"length"`
	} `json:"files"`
	Peers []struct {
		IP        string `json:"ip"`
		Port      int    `json:"port"`
		Source    string `json:"source"`
		Encrypted bool   `json:"encrypted"`
	} `json:"peers"`
	Timestamp int64 `json:"timestamp"`
}

// AnonymizerEvent is a single observation delivered by the anonymizers real-time stream.
// Anonymizer events describe IP ranges rather than individual addresses.
type AnonymizerEvent struct {
	RangeStart string `json:"range_start"`
	RangeEnd   string `json:"range_end"`
	Provider   string `json:"provider"`
	Type       string `json:"type"`
	Timestamp  int64  `json:"timestamp"`
}

func streamFeed[T any](client *Client, feedSegment string, requestOptions *RequestOptions) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		var zero T

		path, err := url.JoinPath(client.BaseAPI.String(), "feeds", feedSegment, "stream")
		if err != nil {
			yield(zero, fmt.Errorf("creating path for %s stream request: %w", feedSegment, err))
			return
		}

		req, err := http.NewRequest(http.MethodGet, path, nil)
		if err != nil {
			yield(zero, fmt.Errorf("making request for %s stream: %w", feedSegment, err))
			return
		}

		body, err := request(requestOptions, client, req, http.StatusOK)
		if err != nil {
			yield(zero, fmt.Errorf("connecting to %s stream: %w", feedSegment, err))
			return
		}
		defer func() { _ = body.Close() }()

		dec := json.NewDecoder(body)
		for dec.More() {
			var event T
			err = dec.Decode(&event)
			if err != nil {
				yield(zero, fmt.Errorf("decoding %s stream event: %w", feedSegment, err))
				return
			}
			if !yield(event, nil) {
				return
			}
		}
	}
}

// StreamProxy connects to the real-time proxy stream and returns an iterator that yields
// one ProxyEvent per newline-delimited JSON event. The stream runs until the connection is
// closed, the context in requestOptions is cancelled, or a decode error occurs.
//
// Example:
//
//	for event, err := range client.StreamProxy(nil) {
//		if err != nil {
//			log.Fatal(err)
//		}
//		fmt.Printf("%s %s %s\n", event.IP, event.Provider, event.CountryCode)
//	}
func (client *Client) StreamProxy(requestOptions *RequestOptions) iter.Seq2[ProxyEvent, error] {
	return streamFeed[ProxyEvent](client, "proxies", requestOptions)
}

// StreamAnonymizer connects to the real-time anonymizer stream and returns an iterator
// that yields one AnonymizerEvent per newline-delimited JSON event. Events describe IP
// ranges (VPNs, Tor exits, relay-class detections) rather than individual addresses. The
// stream runs until the connection is closed, the context in requestOptions is cancelled,
// or a decode error occurs.
//
// Example:
//
//	for event, err := range client.StreamAnonymizer(nil) {
//		if err != nil {
//			log.Fatal(err)
//		}
//		fmt.Printf("%s-%s %s %s\n", event.RangeStart, event.RangeEnd, event.Type, event.Provider)
//	}
func (client *Client) StreamAnonymizer(requestOptions *RequestOptions) iter.Seq2[AnonymizerEvent, error] {
	return streamFeed[AnonymizerEvent](client, "anonymizers", requestOptions)
}

// StreamTorrent connects to the real-time torrent stream and returns an iterator that
// yields one TorrentEvent per newline-delimited JSON event. Each event includes the
// info hash, metadata, per-file details, and observed peers from DHT, PEX, or trackers.
// The stream runs until the connection is closed, the context in requestOptions is
// cancelled, or a decode error occurs.
//
// Example:
//
//	for event, err := range client.StreamTorrent(nil) {
//		if err != nil {
//			log.Fatal(err)
//		}
//		fmt.Printf("%s %s %d peers\n", event.InfoHash, event.Name, len(event.Peers))
//	}
func (client *Client) StreamTorrent(requestOptions *RequestOptions) iter.Seq2[TorrentEvent, error] {
	return streamFeed[TorrentEvent](client, "torrents", requestOptions)
}
