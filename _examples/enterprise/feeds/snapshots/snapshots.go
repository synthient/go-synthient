package main

import (
	"fmt"
	"log"
	"os"

	"github.com/synthient/go-synthient"
)

func main() {
	client := synthient.NewClient(os.Getenv("SYNTHIENT_API_KEY"))

	var cursor string
	for {
		page, err := client.FeedSnapshots("proxies", &synthient.FeedSnapshotsOptions{
			Limit:  50,
			Cursor: cursor,
		}, nil)
		if err != nil {
			log.Fatalf("failed to get feed snapshots: %s", err)
		}

		for _, snap := range page.Feeds {
			fmt.Printf("kind=%-7s id=%-20s date=%s size=%d bytes\n", snap.Kind, snap.ID, snap.Date, snap.SizeBytes)
		}

		if page.NextCursor == "" {
			break
		}
		cursor = page.NextCursor
	}
}
