# go-synthient

[![godoc](https://pkg.go.dev/badge/github.com/synthient/go-synthient/v2?utm_source=godoc)](https://pkg.go.dev/github.com/synthient/go-synthient/v2)
![go.mod version](https://img.shields.io/github/go-mod/go-version/synthient/go-synthient)
[![report card](https://goreportcard.com/badge/github.com/synthient/go-synthient/v2)](https://goreportcard.com/report/github.com/synthient/go-synthient/v2)

Synthient SDK for Go.

## Installation

```bash
go get -u github.com/synthient/go-synthient/v2
```

Requires Go 1.23 or later (real-time streams use `iter.Seq2`).

## Getting started

All requests are made through a `synthient.Client`. Create one with your API key:

```go
client := synthient.NewClient(os.Getenv("SYNTHIENT_API_KEY"))
```

Pass `nil` as the last argument to any method to use default request options, or supply a `*synthient.RequestOptions` to attach a context for cancellation and timeouts:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
opts := &synthient.RequestOptions{Context: ctx}
```

## IP lookup

[`client.GetIP`](https://pkg.go.dev/github.com/synthient/go-synthient/v2#Client.GetIP) returns enrichment data for a single address:

```go
ip, err := client.GetIP("213.149.183.127", nil)
if err != nil {
    log.Fatal(err)
}
fmt.Println(ip.Intelligence.RiskScore, ip.Network.Isp, ip.Location.Country)
```

[`client.GetIPs`](https://pkg.go.dev/github.com/synthient/go-synthient/v2#Client.GetIPs) looks up multiple addresses in one request:

```go
results, err := client.GetIPs([]string{"8.8.8.8", "1.1.1.1"}, nil)
if err != nil {
    log.Fatal(err)
}
for _, ip := range results {
    fmt.Println(ip.IP, ip.Intelligence.RiskScore)
}
```

## Domain lookup

[`client.GetDomain`](https://pkg.go.dev/github.com/synthient/go-synthient/v2#Client.GetDomain) returns traffic statistics, geo distribution, and recent events for a domain:

```go
domain, err := client.GetDomain("google.com", nil)
if err != nil {
    log.Fatal(err)
}
fmt.Println(domain.Stats.Events24H, domain.Status)
```

## Account

[`client.GetAccount`](https://pkg.go.dev/github.com/synthient/go-synthient/v2#Client.GetAccount) returns profile and quota details for the authenticated user:

```go
account, err := client.GetAccount(nil)
if err != nil {
    log.Fatal(err)
}
fmt.Println(account.Organization.Name, account.LookupQuota.Credits)
```

## Parquet snapshot feeds

Stream identifiers: `proxies`, `anonymizers`, `torrents`, `honeypot_http`, `honeypot_https`, `honeypot_dns`, `honeypot_adb`.

### List snapshots

[`client.FeedSnapshots`](https://pkg.go.dev/github.com/synthient/go-synthient/v2#Client.FeedSnapshots) returns a paginated list of available daily and hourly snapshots, newest-first:

```go
var cursor string
for {
    page, err := client.FeedSnapshots("proxies", &synthient.FeedSnapshotsOptions{
        Limit:  50,
        Cursor: cursor,
    }, nil)
    if err != nil {
        log.Fatal(err)
    }
    for _, snap := range page.Feeds {
        fmt.Println(snap.Kind, snap.ID, snap.SizeBytes)
    }
    if page.NextCursor == "" {
        break
    }
    cursor = page.NextCursor
}
```

### Snapshot metadata

[`client.FeedSnapshotMeta`](https://pkg.go.dev/github.com/synthient/go-synthient/v2#Client.FeedSnapshotMeta) returns the checksum, row count, byte size, and parquet schema for a snapshot. `date` accepts `"latest"`, `"YYYY-MM-DD"`, or `"YYYY-MM-DD/HH"`:

```go
meta, err := client.FeedSnapshotMeta("proxies", "latest", nil)
if err != nil {
    log.Fatal(err)
}
fmt.Println(meta.Rows, meta.Checksum)
for _, field := range meta.Schema.Fields {
    fmt.Println(field.Name, field.Type)
}
```

### Download a snapshot

[`client.DownloadFeedSnapshot`](https://pkg.go.dev/github.com/synthient/go-synthient/v2#Client.DownloadFeedSnapshot) follows the API's 307 redirect and returns a streaming reader for the parquet file. The caller must close it.

Pass a non-nil `hour` pointer (0–23) to address a specific hourly snapshot within the current UTC day:

```go
// latest hourly
r, err := client.DownloadFeedSnapshot("proxies", "latest", nil, nil)
if err != nil {
    log.Fatal(err)
}
defer r.Close()

f, _ := os.Create("proxies-latest.parquet")
defer f.Close()
io.Copy(f, r)
```

```go
// specific hour
hour := 21
r, err := client.DownloadFeedSnapshot("proxies", "2026-05-07", &hour, nil)
```

## Real-time firehose streams

All stream methods return an [`iter.Seq2[T, error]`](https://pkg.go.dev/iter#Seq2) that yields one event per NDJSON line. Break out of the loop or cancel the context to stop the stream.

### Proxies

```go
for event, err := range client.StreamProxy(nil) {
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(event.IP, event.Provider, event.CountryCode)
}
```

[`ProxyEvent`](https://pkg.go.dev/github.com/synthient/go-synthient/v2#ProxyEvent) fields: `IP`, `Provider`, `Type`, `Timestamp`, `CountryCode`, `ASN`.

### Anonymizers

```go
for event, err := range client.StreamAnonymizer(nil) {
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(event.RangeStart, event.RangeEnd, event.Type, event.Provider)
}
```

[`AnonymizerEvent`](https://pkg.go.dev/github.com/synthient/go-synthient/v2#AnonymizerEvent) fields: `RangeStart`, `RangeEnd`, `Provider`, `Type`, `Timestamp`.

### Torrents

```go
for event, err := range client.StreamTorrent(nil) {
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%s %s peers=%d\n", event.InfoHash, event.Name, len(event.Peers))
}
```

[`TorrentEvent`](https://pkg.go.dev/github.com/synthient/go-synthient/v2#TorrentEvent) fields: `InfoHash`, `Name`, `MagnetURI`, `TotalSize`, `PieceLength`, `FileCount`, `Files`, `Peers`, `Timestamp`.

## Helios sensor streams

### HTTP captures

```go
for event, err := range client.StreamHeliosHTTP(nil) {
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(event.Details.Method, event.Details.URI, event.Domain)
}
```

[`HeliosHTTPEvent`](https://pkg.go.dev/github.com/synthient/go-synthient/v2#HeliosHTTPEvent) fields: `Timestamp`, `Domain`, `Port`, `TunnelID`, `Protocol`, `Details` (method, URI, version, headers map), `Raw`, `Meta` (pool ID, provider, proxy IP, server).

### TLS ClientHello captures

```go
for event, err := range client.StreamHeliosTLS(nil) {
    if err != nil {
        log.Fatal(err)
    }
    if event.Details == nil {
        continue // parse failed
    }
    fmt.Println(event.Domain, event.Details.HandshakeVersion, len(event.Details.CipherSuites))
}
```

[`HeliosTLSEvent`](https://pkg.go.dev/github.com/synthient/go-synthient/v2#HeliosTLSEvent) carries the fully parsed ClientHello in `Details` ([`*HeliosTLSDetails`](https://pkg.go.dev/github.com/synthient/go-synthient/v2#HeliosTLSDetails)), which is `nil` when the sensor could not parse the handshake. Details includes cipher suites, extensions, supported groups, signature algorithms, key share groups, supported versions, and boolean handshake flags (`extended_master_secret`, `renegotiation_info`, `has_grease`, etc.).

## gRPC schema introspection

[`client.GRPCSchema`](https://pkg.go.dev/github.com/synthient/go-synthient/v2#Client.GRPCSchema) uses gRPC server reflection to fetch protobuf file descriptors from `grpc.synthient.com:443`. Pass `nil` to resolve all services, or supply a list of fully-qualified service names:

```go
ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
defer cancel()

// all services
result, err := client.GRPCSchema(ctx, nil)
if err != nil {
    if msg := synthient.ExplainGRPCError(err); msg != "" {
        log.Fatal(msg)
    }
    log.Fatal(err)
}
for _, svc := range result.Symbols {
    fmt.Println(svc)
}
for _, f := range result.DescriptorSet.File {
    fmt.Println(f.GetName(), f.GetPackage())
}
```

```go
// specific symbol
result, err := client.GRPCSchema(ctx, &synthient.GRPCSchemaOptions{
    Symbols: []string{"synthient.lookup.v1.LookupService"},
})
```

[`GRPCSchemaResult`](https://pkg.go.dev/github.com/synthient/go-synthient/v2#GRPCSchemaResult) fields:

| Field | Type | Description |
|---|---|---|
| `Endpoint` | `string` | Normalized address that was queried |
| `Symbols` | `[]string` | Service symbols that were resolved |
| `DescriptorSet` | `*descriptorpb.FileDescriptorSet` | All file descriptors in topological order |

[`NormalizeGRPCEndpoint`](https://pkg.go.dev/github.com/synthient/go-synthient/v2#NormalizeGRPCEndpoint) and [`ExplainGRPCError`](https://pkg.go.dev/github.com/synthient/go-synthient/v2#ExplainGRPCError) are exported for use in CLI applications that need custom endpoint handling or human-readable error messages.

> **Note:** `GRPCSchema` adds `google.golang.org/grpc` and `google.golang.org/protobuf` to your module's dependency graph. If you only need REST API access these are still pulled in transitively, but no gRPC connections are made unless you call `GRPCSchema`.

## Client customization

Override `BaseAPI` to point at a self-hosted endpoint:

```go
client := synthient.NewClient("SECRET TOKEN")
client.BaseAPI.Host = "synthient.myserver.com"
```

Set a custom HTTP client for timeouts or proxies:

```go
client.HttpClient = &http.Client{Timeout: 30 * time.Second}
```
