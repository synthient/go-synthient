package main

import (
	"fmt"
	"log"
	"os"

	"github.com/synthient/go-synthient/v2"
)

func main() {
	client := synthient.NewClient(os.Getenv("SYNTHIENT_API_KEY"))

	meta, err := client.FeedSnapshotMeta("proxies", "latest", nil)
	if err != nil {
		log.Fatalf("failed to get feed snapshot meta: %s", err)
	}

	fmt.Printf("stream:    %s\n", meta.Stream)
	fmt.Printf("kind:      %s\n", meta.Kind)
	fmt.Printf("id:        %s\n", meta.ID)
	fmt.Printf("size:      %d bytes\n", meta.Size)
	fmt.Printf("rows:      %d\n", meta.Rows)
	fmt.Printf("checksum:  %s\n", meta.Checksum)
	fmt.Printf("schema:\n")
	for _, field := range meta.Schema.Fields {
		fmt.Printf("  %-20s %s\n", field.Name, field.Type)
	}
}
