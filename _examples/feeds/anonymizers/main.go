package main

import (
	"fmt"
	"log"
	"os"

	"github.com/synthient/go-synthient"
)

func main() {
	client := synthient.NewClient(os.Getenv("SYNTHIENT_API_KEY"))
	n, err := client.DownloadAnonymizersFeed(synthient.AnonymizersQuery{}, "feed.csv", nil)
	if err != nil {
		log.Fatalf("failed to download anonymizer feed: %s", err)
	}
	fmt.Printf("%d bytes downloaded\n", n)
}
