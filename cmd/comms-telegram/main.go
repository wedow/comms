package main

import (
	"fmt"
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
	root.AddCommand(&cobra.Command{
		Use:    "subprocess",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			configJSON := os.Getenv("COMMS_PROVIDER_CONFIG")
			if configJSON == "" {
				return fmt.Errorf("COMMS_PROVIDER_CONFIG env var is required")
			}
			return telegram.RunSubprocess(cmd.Context(), os.Stdin, os.Stdout, configJSON)
		},
	})
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
