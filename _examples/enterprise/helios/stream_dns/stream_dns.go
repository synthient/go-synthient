package main

import (
	"fmt"
	"log"
	"os"

	"github.com/synthient/go-synthient"
)

func main() {
	client := synthient.NewClient(os.Getenv("SYNTHIENT_API_KEY"))

	for event, err := range client.StreamHeliosDNS(nil) {
		if err != nil {
			log.Fatalf("helios dns stream error: %s", err)
		}
		fmt.Printf("%-40s  port=%-5d  tunnel=%d  (via %s)\n",
			event.Domain, event.Port, event.TunnelID, event.Meta.ProxyIP)
	}
}
