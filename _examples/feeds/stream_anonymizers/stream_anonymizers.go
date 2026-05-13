package main

import (
	"fmt"
	"log"
	"os"

	"github.com/synthient/go-synthient"
)

func main() {
	client := synthient.NewClient(os.Getenv("SYNTHIENT_API_KEY"))

	for event, err := range client.StreamAnonymizer(nil) {
		if err != nil {
			log.Fatalf("anonymizer stream error: %s", err)
		}
		fmt.Printf("%s-%s  %-20s  %s\n",
			event.RangeStart, event.RangeEnd, event.Type, event.Provider)
	}
}
