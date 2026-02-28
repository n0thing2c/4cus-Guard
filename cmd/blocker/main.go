package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"4cus-guard/internal/config"
	Message "4cus-guard/internal/message"
	"4cus-guard/internal/pubsub"
	"4cus-guard/internal/services"
)

func main() {
	conf := config.LoadConfig()

	ctx := context.Background()
	rdb, err := pubsub.NewRedisBroker(ctx, conf.RedisAddr, conf.RedisPass)
	if err != nil {
		log.Fatalf("Fail to init: %v", err)
	}

	msgChannel, error := rdb.Subscribe(ctx, "Blocker")
	if error != nil {
		log.Fatal(error)
	}
	fmt.Println("Blocker is active")

	for payload := range msgChannel {
		var msg Message.Message
		json.Unmarshal([]byte(payload), &msg)

		url := msg.URL
		action := msg.Action

		switch action {
		case "block":
			services.BlockURL(url)
			fmt.Printf("Blocked successfully: %v\n", url)
		case "unblock":
			services.UnblockURL(url)
			fmt.Printf("Unblocked successfully: %v\n", url)
		}
	}
}
