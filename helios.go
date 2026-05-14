package synthient

import (
	"io"
	"iter"
)

// HeliosHTTPEvent is a single HTTP capture delivered by the Helios HTTP sensor stream.
type HeliosHTTPEvent struct {
	Timestamp int64  `json:"timestamp"`
	Domain    string `json:"domain"`
	Port      int    `json:"port"`
	TunnelID  int64  `json:"tunnel_id"`
	Protocol  string `json:"protocol"`
	Details   struct {
		Method  string            `json:"method"`
		URI     string            `json:"uri"`
		Version string            `json:"version"`
		Headers map[string]string `json:"headers"`
	} `json:"details"`
	Raw  string `json:"raw"`
	Meta struct {
		PoolID   string `json:"pool_id"`
		Provider string `json:"provider"`
		ProxyIP  string `json:"proxy_ip"`
		Server   string `json:"server"`
	} `json:"meta"`
}

// HeliosTLSDetails holds the parsed TLS ClientHello from a Helios TLS capture event.
// It is nil when the sensor was unable to parse the handshake record.
type HeliosTLSDetails struct {
	RecordVersion    string `json:"record_version"`
	HandshakeVersion string `json:"handshake_version"`
	ClientRandom     string `json:"client_random"`
	SessionID        string `json:"session_id"`
	SessionIDLength  int    `json:"session_id_length"`
	CipherSuites     []struct {
		Code int    `json:"code"`
		Name string `json:"name"`
	} `json:"cipher_suites"`
	CompressionMethods []int    `json:"compression_methods"`
	SNI                string   `json:"sni"`
	SupportedVersions  []string `json:"supported_versions"`
	SupportedGroups    []struct {
		Code int    `json:"code"`
		Name string `json:"name"`
	} `json:"supported_groups"`
	ECPointFormats      []string `json:"ec_point_formats"`
	SignatureAlgorithms []struct {
		Code int    `json:"code"`
		Name string `json:"name"`
	} `json:"signature_algorithms"`
	Extensions []struct {
		Code   int    `json:"code"`
		Name   string `json:"name"`
		Length int    `json:"length"`
	} `json:"extensions"`
	KeyShareGroups      []string `json:"key_share_groups"`
	PSKKeyExchangeModes []string `json:"psk_key_exchange_modes"`

	ExtendedMasterSecret        bool `json:"extended_master_secret"`
	RenegotiationInfo           bool `json:"renegotiation_info"`
	StatusRequest               bool `json:"status_request"`
	SignedCertificateTimestamps bool `json:"signed_certificate_timestamps"`
	HasGREASE                   bool `json:"has_grease"`
	EncryptThenMAC              bool `json:"encrypt_then_mac"`
	PostHandshakeAuth           bool `json:"post_handshake_auth"`
	DelegatedCredentials        bool `json:"delegated_credentials"`
	ApplicationSettings         bool `json:"application_settings"`
}

// HeliosTLSEvent is a single TLS ClientHello capture delivered by the Helios HTTPS sensor stream.
type HeliosTLSEvent struct {
	Timestamp int64  `json:"timestamp"`
	Domain    string `json:"domain"`
	Port      int    `json:"port"`
	TunnelID  int64  `json:"tunnel_id"`
	Protocol  string `json:"protocol"`
	Meta      struct {
		PoolID   string `json:"pool_id"`
		Provider string `json:"provider"`
		ProxyIP  string `json:"proxy_ip"`
		Server   string `json:"server"`
	} `json:"meta"`
	Details *HeliosTLSDetails `json:"details"`
}

// StreamHeliosTLS connects to the real-time Helios TLS capture stream and returns an
// iterator that yields one HeliosTLSEvent per newline-delimited JSON event. Each event
// carries the fully parsed TLS ClientHello including cipher suites, extensions, supported
// groups, signature algorithms, and handshake flags. Details is nil when parsing failed.
// The stream runs until the connection is closed, the context in requestOptions is
// cancelled, or a decode error occurs.
//
// Example:
//
//	for event, err := range client.StreamHeliosTLS(nil) {
//		if err != nil {
//			log.Fatal(err)
//		}
//		if event.Details != nil {
//			fmt.Printf("%s  %s  suites=%d\n", event.Domain, event.Details.HandshakeVersion,
//
// len(event.Details.CipherSuites))
//
//		}
//	}
func (client *Client) StreamHeliosTLS(
	requestOptions *RequestOptions,
) iter.Seq2[HeliosTLSEvent, error] {
	return streamFeed[HeliosTLSEvent](client, requestOptions, "feeds", "helio", "https", "stream")
}

// // HeliosDNSEvent is a single DNS resolution observation delivered by the Helios DNS sensor
// stream.
// type HeliosDNSEvent struct {
// 	Timestamp int64  `json:"timestamp"`
// 	TunnelID  int64  `json:"tunnel_id"`
// 	Domain    string `json:"domain"`
// 	Port      int    `json:"port"`
// 	Meta      struct {
// 		PoolID   string `json:"pool_id"`
// 		Provider string `json:"provider"`
// 		ProxyIP  string `json:"proxy_ip"`
// 		Server   string `json:"server"`
// 	} `json:"meta"`
// }

// // HeliosADBEvent is a single Android Debug Bridge command capture delivered by the Helios
// // ADB sensor stream.
// type HeliosADBEvent struct {
// 	Session      string `json:"session"`
// 	SequentialID int64  `json:"sequential_id"`
// 	Command      string `json:"command"`
// 	Hash         string `json:"hash"`
// }

// // StreamHeliosADB connects to the real-time Helios ADB capture stream and returns an
// // iterator that yields one HeliosADBEvent per newline-delimited JSON event. Each event
// // contains the raw shell command an attacker executed, the session hash grouping commands
// // from the same connection, and a SHA-256 of the command bytes for deduplication across
// // sessions. The stream runs until the connection is closed, the context in requestOptions
// // is cancelled, or a decode error occurs.
// //
// // Example:
// //
// //	for event, err := range client.StreamHeliosADB(nil) {
// //		if err != nil {
// //			log.Fatal(err)
// //		}
// //		fmt.Printf("[%s #%d] %s\n", event.Session, event.SequentialID, event.Command)
// //	}
// func (client *Client) StreamHeliosADB(requestOptions *RequestOptions) iter.Seq2[HeliosADBEvent,
// error] {
// 	return streamFeed[HeliosADBEvent](client, requestOptions, "feeds", "helio", "adb", "stream")
// }

// // StreamHeliosDNS connects to the real-time Helios DNS capture stream and returns an
// // iterator that yields one HeliosDNSEvent per newline-delimited JSON event. Each event
// // records the hostname an inbound flow resolved and the destination port, useful for
// // detecting C2 lookups and fast-flux infrastructure. TunnelID joins back to matching
// // HTTP and TLS captures from the same flow. The stream runs until the connection is
// // closed, the context in requestOptions is cancelled, or a decode error occurs.
// //
// // Example:
// //
// //	for event, err := range client.StreamHeliosDNS(nil) {
// //		if err != nil {
// //			log.Fatal(err)
// //		}
// //		fmt.Printf("%s  port=%d  (via %s)\n", event.Domain, event.Port, event.Meta.ProxyIP)
// //	}
// func (client *Client) StreamHeliosDNS(requestOptions *RequestOptions) iter.Seq2[HeliosDNSEvent,
// error] {
// 	return streamFeed[HeliosDNSEvent](client, requestOptions, "feeds", "helio", "dns", "stream")
// }

// StreamHeliosHTTP connects to the real-time Helios HTTP capture stream and returns an
// iterator that yields one HeliosHTTPEvent per newline-delimited JSON event. Each event
// includes the HTTP method, URI, headers, raw request bytes, and source metadata. The
// stream runs until the connection is closed, the context in requestOptions is cancelled,
// or a decode error occurs.
//
// Example:
//
//	for event, err := range client.StreamHeliosHTTP(nil) {
//		if err != nil {
//			log.Fatal(err)
//		}
//		fmt.Printf("%s %s %s\n", event.Details.Method, event.Details.URI, event.Domain)
//	}
func (client *Client) StreamHeliosHTTP(
	requestOptions *RequestOptions,
) iter.Seq2[HeliosHTTPEvent, error] {
	return streamFeed[HeliosHTTPEvent](client, requestOptions, "feeds", "helio", "http", "stream")
}

// DownloadHeliosHTTP downloads a Helios HTTP capture Parquet snapshot. If filename is
// non-empty the snapshot is written to that file and the returned reader is nil. If
// filename is empty the caller receives the raw reader and must close it. The API issues
// a 307 redirect to a presigned URL; this method follows it automatically.
//
// date accepts "latest" for the most recent hourly snapshot, or a YYYY-MM-DD string for
// a daily rollup. For a specific hourly within the current UTC day, set hour to a non-nil
// pointer in the range 0–23.
//
// Example (write to file):
//
//	_, err := client.DownloadHeliosHTTP("latest", nil, "helios-http.parquet", nil)
//
// Example (stream reader):
//
//	r, err := client.DownloadHeliosHTTP("latest", nil, "", nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer r.Close()
func (client *Client) DownloadHeliosHTTP(
	date string,
	hour *int,
	filename string,
	requestOptions *RequestOptions,
) (io.ReadCloser, error) {
	return downloadFeed(client, requestOptions, date, hour, filename, "feeds", "helio", "http")
}

// DownloadHeliosTLS downloads a Helios TLS capture Parquet snapshot. If filename is
// non-empty the snapshot is written to that file and the returned reader is nil. If
// filename is empty the caller receives the raw reader and must close it. The API issues
// a 307 redirect to a presigned URL; this method follows it automatically.
//
// date accepts "latest" for the most recent hourly snapshot, or a YYYY-MM-DD string for
// a daily rollup. For a specific hourly within the current UTC day, set hour to a non-nil
// pointer in the range 0–23.
//
// Example (write to file):
//
//	_, err := client.DownloadHeliosTLS("latest", nil, "helios-tls.parquet", nil)
//
// Example (stream reader):
//
//	r, err := client.DownloadHeliosTLS("latest", nil, "", nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer r.Close()
func (client *Client) DownloadHeliosTLS(
	date string,
	hour *int,
	filename string,
	requestOptions *RequestOptions,
) (io.ReadCloser, error) {
	return downloadFeed(client, requestOptions, date, hour, filename, "feeds", "helio", "https")
}
