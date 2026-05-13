package main

import (
	"fmt"
	"log"
	"os"

	"github.com/synthient/go-synthient"
)

func main() {
	client := synthient.NewClient(os.Getenv("SYNTHIENT_API_KEY"))

	for event, err := range client.StreamHeliosADB(nil) {
		if err != nil {
			log.Fatalf("helios adb stream error: %s", err)
		}
		fmt.Printf("[%s #%d] %s\n", event.Session, event.SequentialID, event.Command)
	}
}
