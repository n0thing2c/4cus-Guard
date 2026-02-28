package main

import (
	"encoding/json"
	"log"
	"time"

	Message "4cus-guard/internal/message"
	"4cus-guard/internal/pubsub"

	"github.com/spf13/cobra"
)

func NewStopCmd(pub pubsub.Publisher) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop focus mode and timer",
		Run: func(cmd *cobra.Command, args []string) {
			action := "stop"
			now := time.Now().Unix()
			msg := Message.Message{Action: action, Timestamp: now, URL: ""}

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
