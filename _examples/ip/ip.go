package main

import (
	"fmt"
	"log"
	"os"

	"github.com/synthient/go-synthient"
)

func main() {
	client := synthient.NewClient(os.Getenv("SYNTHIENT_API_KEY"))
	resp, err := client.GetIP("213.149.183.127", nil)
	if err != nil {
		log.Fatalf("failed to get ip address: %s", err)
	}
	fmt.Println(resp)
}
