package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wedow/comms/internal/provider/telegram"
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
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newDaemonCmd())
	cmd.AddCommand(newUnreadCmd())
	cmd.AddCommand(newAckCmd())
	cmd.AddCommand(newSendCmd(telegram.NewBot))
	cmd.AddCommand(newReactCmd(telegram.NewBot))
	cmd.AddCommand(newAllowCmd())
	return cmd
}

// Execute runs the root command.
func Execute() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
