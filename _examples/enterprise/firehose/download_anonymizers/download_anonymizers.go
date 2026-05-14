package main

import (
	"log"
	"os"

	"github.com/synthient/go-synthient/v2"
)

func main() {
	client := synthient.NewClient(os.Getenv("SYNTHIENT_API_KEY"))

	_, err := client.DownloadAnonymizer("latest", nil, "anonymizers-latest.parquet", nil)
	if err != nil {
		log.Fatalf("failed to download anonymizer snapshot: %s", err)
	}
	log.Println("downloaded anonymizers-latest.parquet")
}
