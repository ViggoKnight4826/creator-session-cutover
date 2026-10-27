package main

import (
	"context"
	"fmt"
	"os"
)

func main() {
	client, err := NewSessionClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	job := DeliveryJob{AssetID: "asset-17", SubscriberID: "subscriber-4", SessionID: "session-9"}
	refreshToken := os.Getenv("CREATOR_REFRESH_TOKEN")
	if refreshToken == "" {
		fmt.Fprintln(os.Stderr, "CREATOR_REFRESH_TOKEN is required")
		os.Exit(1)
	}
	result, err := deliverAfterRotation(context.Background(), client, job, refreshToken)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(result)
}
