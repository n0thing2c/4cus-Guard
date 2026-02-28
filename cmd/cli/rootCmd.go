package main

import (
	"4cus-guard/internal/pubsub"

	"github.com/spf13/cobra"
)

func NewRootCmd(pub pubsub.Publisher) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "focus",
		Short: "Root command",
	}

	cmd.AddCommand(NewStartCmd(pub))
	cmd.AddCommand(NewStopCmd(pub))
	cmd.AddCommand(NewBlockCmd(pub))
	cmd.AddCommand(NewUnblockCmd(pub))
	return cmd
}
