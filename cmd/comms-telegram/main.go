package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/wedow/comms/providers/telegram"
)

func main() {
	root := &cobra.Command{
		Use:           "comms-telegram",
		Short:         "Telegram provider for comms",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.AddCommand(telegram.NewSendCmd())
	root.AddCommand(telegram.NewReactCmd())
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
