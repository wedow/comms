package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const version = "v0.1.0"

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "comms",
		Short: "Unified message relay for AI agents",
		Run: func(cmd *cobra.Command, args []string) {
			v, _ := cmd.Flags().GetBool("version")
			if v {
				fmt.Fprintln(cmd.OutOrStdout(), "comms "+version)
				return
			}
			cmd.Help()
		},
	}
	cmd.Flags().Bool("version", false, "print version")
	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newChannelsCmd())
	return cmd
}

// Execute runs the root command.
func Execute() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
