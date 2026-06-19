package synthient

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"testing"
)

func TestFeedStreamPath(t *testing.T) {
	cases := map[string][]string{
		"proxies":        {"proxies"},
		"anonymizers":    {"anonymizers"},
		"torrents":       {"torrents"},
		"honeypot_http":  {"helio", "http"},
		"honeypot_https": {"helio", "https"},
		"honeypot_dns":   {"helio", "dns"},
		"honeypot_adb":   {"helio", "adb"},
		"custom":         {"custom"},
	}
	for stream, want := range cases {
		got := feedStreamPath(stream)
		if !slices.Equal(got, want) {
			t.Errorf("feedStreamPath(%q) = %v, want %v", stream, got, want)
		}
	}
}

// TestFeedRequestPaths guards the regression where honeypot streams were routed
// to /feeds/<name>/export instead of the helio/<protocol> path the API serves.
func TestFeedRequestPaths(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	base, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	client := Client{HttpClient: server.Client(), Token: "test", BaseAPI: *base}

	snapshots := []struct {
		stream string
		want   string
	}{
		{"proxies", "/feeds/proxies/export"},
		{"honeypot_https", "/feeds/helio/https/export"},
		{"honeypot_http", "/feeds/helio/http/export"},
		{"honeypot_dns", "/feeds/helio/dns/export"},
		{"honeypot_adb", "/feeds/helio/adb/export"},
	}
	for _, c := range snapshots {
		_, err := client.FeedSnapshots(c.stream, nil, nil)
		if err != nil {
			t.Fatalf("FeedSnapshots(%q): %v", c.stream, err)
		}
		if gotPath != c.want {
			t.Errorf("FeedSnapshots(%q) path = %q, want %q", c.stream, gotPath, c.want)
		}
	}

	_, err = client.FeedSnapshotMeta("honeypot_https", "latest", nil)
	if err != nil {
		t.Fatal(err)
	}
	if want := "/feeds/helio/https/export/latest/meta"; gotPath != want {
		t.Errorf("FeedSnapshotMeta path = %q, want %q", gotPath, want)
	}

	r, err := client.DownloadFeedSnapshot("honeypot_https", "latest", nil, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, r)
	_ = r.Close()
	if want := "/feeds/helio/https/export/latest"; gotPath != want {
		t.Errorf("DownloadFeedSnapshot path = %q, want %q", gotPath, want)
	}
}
