package main

import (
	"encoding/json"
	"log"
	"time"

	Message "4cus-guard/internal/message"
	"4cus-guard/internal/pubsub"

	"github.com/spf13/cobra"
)

func NewBlockCmd(pub pubsub.Publisher) *cobra.Command {
	cmd := &cobra.Command{
		Use: "block [url]",
		Run: func(cmd *cobra.Command, args []string) {
			action := "block"
			now := time.Now().Unix()
			url := args[0]
			msg := Message.Message{Action: action, Timestamp: now, URL: url}

			jsonMsg, err := json.Marshal(msg)
			if err != nil {
				log.Fatal(err)
			}

			ctx := cmd.Context()
			pub.Publish(ctx, "Blocker", string(jsonMsg))
		},
	}
	return cmd
}
