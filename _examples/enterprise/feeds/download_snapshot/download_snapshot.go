package main

import (
	"log"
	"os"

	"github.com/synthient/go-synthient/v2"
)

func main() {
	client := synthient.NewClient(os.Getenv("SYNTHIENT_API_KEY"))

	_, err := client.DownloadFeedSnapshot("proxies", "latest", nil, "proxies-latest.parquet", nil)
	if err != nil {
		log.Fatalf("failed to download feed snapshot: %s", err)
	}
	log.Println("downloaded proxies-latest.parquet")
}
