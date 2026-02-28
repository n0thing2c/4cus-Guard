package main

import (
	"context"
	"log"

	"4cus-guard/internal/config"
	"4cus-guard/internal/pubsub"
)

func main() {
	conf := config.LoadConfig()

	//Init Infrastructure
	ctx := context.Background()
	rdb, err := pubsub.NewRedisBroker(ctx, conf.RedisAddr, conf.RedisPass)
	if err != nil {
		log.Fatalf("Fail to init: %v", err)
	}

	//init Cobra
	rootCmd := NewRootCmd(rdb)
	rootCmd.Execute()

}
