package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newVersionCmd(version, buildDate string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "gophkeeper %s (%s)\n", version, buildDate)
		},
	}
}
