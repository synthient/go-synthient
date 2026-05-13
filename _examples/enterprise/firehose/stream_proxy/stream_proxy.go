package main

import (
	"fmt"
	"log"
	"os"

	"github.com/synthient/go-synthient/v2"
)

func main() {
	client := synthient.NewClient(os.Getenv("SYNTHIENT_API_KEY"))

	for event, err := range client.StreamProxy(nil) {
		if err != nil {
			log.Fatalf("proxy stream error: %s", err)
		}
		fmt.Printf("%-15s  %-12s  %-22s  asn=%-8d  %s\n",
			event.IP, event.CountryCode, event.Type, event.ASN, event.Provider)
	}
}
