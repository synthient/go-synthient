package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/synthient/go-synthient/v2"
)

func main() {
	client := synthient.NewClient(os.Getenv("SYNTHIENT_API_KEY"))

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	result, err := client.GRPCSchema(ctx, nil)
	if err != nil {
		if msg := synthient.ExplainGRPCError(err); msg != "" {
			log.Fatal(msg)
		}
		log.Fatal(err)
	}

	fmt.Printf("endpoint:  %s\n", result.Endpoint)
	fmt.Printf("services:  %d\n", len(result.Symbols))
	fmt.Printf("files:     %d\n\n", len(result.DescriptorSet.File))

	for _, svc := range result.Symbols {
		fmt.Printf("service  %s\n", svc)
	}
	fmt.Println()
	for _, file := range result.DescriptorSet.File {
		fmt.Printf("file     %-50s  %s\n", file.GetName(), file.GetPackage())
	}
}
