package main

import (
	"log"
	"os"

	"github.com/synthient/go-synthient/v2"
)

func main() {
	client := synthient.NewClient(os.Getenv("SYNTHIENT_API_KEY"))

	_, err := client.DownloadTorrent("latest", nil, "torrents-latest.parquet", nil)
	if err != nil {
		log.Fatalf("failed to download torrent snapshot: %s", err)
	}
	log.Println("downloaded torrents-latest.parquet")
}
