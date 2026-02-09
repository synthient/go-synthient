package main

import (
	"io"
	"log"
	"os"

	"github.com/synthient/go-synthient"
)

func main() {
	client := synthient.NewClient(os.Getenv("SYNTHIENT_API_KEY"))
	stream, err := client.StreamAnonymizersFeed(synthient.AnonymizersQuery{
		Provider:     "BIRDPROXIES",
		Type:         "RESIDENTIAL_PROXY",
		LastObserved: "7D",
		Format:       "CSV",
		CountryCode:  "US",
		Full:         false,
		Order:        "desc",
	}, nil)
	if err != nil {
		log.Fatalf("failed to stream feed: %s", err)
	}
	defer func() { _ = stream.Close() }() // important! make sure to close stream

	_, err = io.Copy(os.Stdout, stream)
	if err != nil {
		log.Fatalf("failed to read stream: %s", err)
	}
}
