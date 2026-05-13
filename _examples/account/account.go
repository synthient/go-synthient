package main

import (
	"fmt"
	"log"
	"os"

	"github.com/synthient/go-synthient"
)

func main() {
	client := synthient.NewClient(os.Getenv("SYNTHIENT_API_KEY"))
	resp, err := client.GetAccount(nil)
	if err != nil {
		log.Fatalf("failed to get account: %s", err)
	}
	fmt.Println(resp)
}
