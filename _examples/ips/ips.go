package main

import (
	"fmt"
	"log"
	"os"

	"github.com/synthient/go-synthient/v2"
)

func main() {
	client := synthient.NewClient(os.Getenv("SYNTHIENT_API_KEY"))
	results, err := client.GetIPs([]string{"8.8.8.8", "1.1.1.1", "101.53.218.152"}, nil)
	if err != nil {
		log.Fatalf("failed to get ip addresses: %s", err)
	}
	for _, ip := range results {
		fmt.Println(ip)
	}
}
