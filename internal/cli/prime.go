package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wedow/comms/internal/embeddocs"
)

func newPrimeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "prime",
		Short: "Print bootstrap guide for AI agents",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprint(cmd.OutOrStdout(), string(embeddocs.PrimeDoc))
			return err
		},
	}
}
