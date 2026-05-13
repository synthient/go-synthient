package main

import (
	"io"
	"log"
	"os"

	"github.com/synthient/go-synthient/v2"
)

func main() {
	client := synthient.NewClient(os.Getenv("SYNTHIENT_API_KEY"))

	// download the latest hourly snapshot (no hour argument needed)
	r, err := client.DownloadFeedSnapshot("proxies", "latest", nil, nil)
	if err != nil {
		log.Fatalf("failed to download feed snapshot: %s", err)
	}
	defer func() { _ = r.Close() }()

	filename := "proxies-latest.parquet"
	f, err := os.Create(filename)
	if err != nil {
		log.Fatalf("failed to create output file: %s", err)
	}
	defer func() { _ = f.Close() }()

	_, err = io.Copy(f, r)
	if err != nil {
		log.Fatalf("failed to write snapshot: %s", err)
	}
	log.Println("downloaded", filename)
}
