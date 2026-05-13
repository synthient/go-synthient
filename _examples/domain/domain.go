package main

import (
	"fmt"
	"log"
	"os"

	"github.com/synthient/go-synthient/v2"
)

func main() {
	client := synthient.NewClient(os.Getenv("SYNTHIENT_API_KEY"))
	resp, err := client.GetDomain("google.com", nil)
	if err != nil {
		log.Fatalf("failed to get domain: %s", err)
	}
	fmt.Println(resp)
}
