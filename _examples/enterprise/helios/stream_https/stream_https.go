package main

import (
	"fmt"
	"log"
	"os"

	"github.com/synthient/go-synthient"
)

func main() {
	client := synthient.NewClient(os.Getenv("SYNTHIENT_API_KEY"))

	for event, err := range client.StreamHeliosTLS(nil) {
		if err != nil {
			log.Fatalf("helios tls stream error: %s", err)
		}
		if event.Details == nil {
			fmt.Printf("%-30s  (parse failed)\n", event.Domain)
			continue
		}
		fmt.Printf("%-30s  %-8s  suites=%-3d  exts=%d\n",
			event.Domain, event.Details.HandshakeVersion,
			len(event.Details.CipherSuites), len(event.Details.Extensions))
	}
}
