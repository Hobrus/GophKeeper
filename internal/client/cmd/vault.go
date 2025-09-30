package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"gophkeeper/internal/client/vault"
)

func newVaultCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "vault", Short: "Manage local vault key"}
	cmd.AddCommand(&cobra.Command{Use: "init", Short: "Generate local vault key", RunE: func(cmd *cobra.Command, args []string) error {
		_, err := vault.Generate()
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Vault key generated at", vault.Path())
		return nil
	}})
	cmd.AddCommand(&cobra.Command{Use: "status", Short: "Show vault status", Run: func(cmd *cobra.Command, args []string) {
		if vault.Exists() {
			fmt.Fprintln(cmd.OutOrStdout(), "Vault: ready")
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "Vault: not initialized")
		}
	}})
	return cmd
}
