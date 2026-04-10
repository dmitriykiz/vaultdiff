package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/example/vaultdiff/internal/diff"
	"github.com/example/vaultdiff/internal/output"
	"github.com/example/vaultdiff/internal/vault"
)

var (
	vaultAddr  string
	vaultToken string
	format     string
	mount      string
)

var rootCmd = &cobra.Command{
	Use:   "vaultdiff <path1> <path2>",
	Short: "Compare two HashiCorp Vault secret paths",
	Long:  `vaultdiff compares secrets at two Vault paths and outputs a structured diff.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runDiff,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&vaultAddr, "address", "", "Vault server address (overrides VAULT_ADDR)")
	rootCmd.PersistentFlags().StringVar(&vaultToken, "token", "", "Vault token (overrides VAULT_TOKEN)")
	rootCmd.PersistentFlags().StringVar(&format, "format", "text", "Output format: text or json")
	rootCmd.PersistentFlags().StringVar(&mount, "mount", "secret", "KV mount path")
}

func runDiff(cmd *cobra.Command, args []string) error {
	path1, path2 := args[0], args[1]

	client, err := vault.NewClient(vaultAddr, vaultToken)
	if err != nil {
		return fmt.Errorf("failed to create vault client: %w", err)
	}

	secrets1, err := client.ReadSecret(mount, path1)
	if err != nil {
		return fmt.Errorf("failed to read path %q: %w", path1, err)
	}

	secrets2, err := client.ReadSecret(mount, path2)
	if err != nil {
		return fmt.Errorf("failed to read path %q: %w", path2, err)
	}

	entries := diff.Compare(secrets1, secrets2)

	fmt, err := output.ParseFormat(format)
	if err != nil {
		return err
	}

	f := output.NewFormatter(os.Stdout, fmt)
	return f.Write(entries)
}
