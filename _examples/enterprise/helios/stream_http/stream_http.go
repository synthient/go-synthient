package main

import (
	"fmt"
	"log"
	"os"

	"github.com/synthient/go-synthient/v2"
)

func main() {
	client := synthient.NewClient(os.Getenv("SYNTHIENT_API_KEY"))

	for event, err := range client.StreamHeliosHTTP(nil) {
		if err != nil {
			log.Fatalf("helios http stream error: %s", err)
		}
		fmt.Printf("%-4s %-40s %s  (via %s)\n",
			event.Details.Method, event.Details.URI, event.Domain, event.Meta.ProxyIP)
	}
}
