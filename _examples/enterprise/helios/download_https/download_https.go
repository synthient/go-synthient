package main

import (
	"log"
	"os"

	"github.com/synthient/go-synthient/v2"
)

func main() {
	client := synthient.NewClient(os.Getenv("SYNTHIENT_API_KEY"))

	_, err := client.DownloadHeliosTLS("latest", nil, "helios-tls-latest.parquet", nil)
	if err != nil {
		log.Fatalf("failed to download Helios TLS snapshot: %s", err)
	}
	log.Println("downloaded helios-tls-latest.parquet")
}
