package main

import (
	"fmt"
	"log"
	"os"

	"github.com/synthient/go-synthient"
)

func main() {
	client := synthient.NewClient(os.Getenv("SYNTHIENT_API_KEY"))

	for event, err := range client.StreamTorrent(nil) {
		if err != nil {
			log.Fatalf("torrent stream error: %s", err)
		}
		fmt.Printf("%s  %-50s  files=%-4d  peers=%d\n",
			event.InfoHash, event.Name, event.FileCount, len(event.Peers))
	}
}
