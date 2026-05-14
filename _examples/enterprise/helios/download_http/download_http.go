package main

import (
	"log"
	"os"

	"github.com/synthient/go-synthient/v2"
)

func main() {
	client := synthient.NewClient(os.Getenv("SYNTHIENT_API_KEY"))

	_, err := client.DownloadHeliosHTTP("latest", nil, "helios-http-latest.parquet", nil)
	if err != nil {
		log.Fatalf("failed to download Helios HTTP snapshot: %s", err)
	}
	log.Println("downloaded helios-http-latest.parquet")
}
