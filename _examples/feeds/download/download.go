package main

import (
	"fmt"
	"log"
	"os"

	"github.com/synthient/go-synthient"
)

func main() {
	client := synthient.NewClient(os.Getenv("SYNTHIENT_API_KEY"))
	n, err := client.DownloadAnonymizersFeed(synthient.AnonymizersQuery{
		Provider:     "BIRDPROXIES",
		Type:         "RESIDENTIAL_PROXY",
		LastObserved: "7D",
		Format:       "CSV",
		CountryCode:  "US",
		Full:         false,
		Order:        "desc",
	}, "feed.csv", nil)
	if err != nil {
		log.Fatalf("failed to download feed: %s", err)
	}

	fmt.Println(n, "bytes downloaded")
}
