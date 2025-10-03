package cmd

import (
	"github.com/spf13/cobra"
)

func NewRootCmd(version, buildDate string) *cobra.Command {
	var serverURL string
	root := &cobra.Command{
		Use:   "gophkeeper",
		Short: "GophKeeper CLI",
	}
	root.PersistentFlags().StringVar(&serverURL, "server", "http://localhost:8080", "Server base URL")

	root.AddCommand(newVersionCmd(version, buildDate))
	root.AddCommand(newAuthCmd(&serverURL))
	root.AddCommand(newRecordsCmd(&serverURL))
	root.AddCommand(newVaultCmd())
	return root
}
