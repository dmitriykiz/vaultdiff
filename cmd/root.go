package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourorg/vaultdiff/internal/diff"
	"github.com/yourorg/vaultdiff/internal/output"
	"github.com/yourorg/vaultdiff/internal/vault"
)

var (
	vaultAddr   string
	vaultToken  string
	formatFlag  string
	prefixes    []string
	excludePaths []string
)

var rootCmd = &cobra.Command{
	Use:   "vaultdiff <path-a> <path-b>",
	Short: "Compare two Vault secret paths and output structured diffs",
	Args:  cobra.ExactArgs(2),
	RunE:  runDiff,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&vaultAddr, "address", "", "Vault server address (overrides VAULT_ADDR)")
	rootCmd.PersistentFlags().StringVar(&vaultToken, "token", "", "Vault token (overrides VAULT_TOKEN)")
	rootCmd.PersistentFlags().StringVar(&formatFlag, "format", "text", "Output format: text or json")
	rootCmd.PersistentFlags().StringArrayVar(&prefixes, "prefix", nil, "Include only keys matching these prefixes (repeatable)")
	rootCmd.PersistentFlags().StringArrayVar(&excludePaths, "exclude", nil, "Exclude keys matching these prefixes (repeatable)")
}

func runDiff(cmd *cobra.Command, args []string) error {
	client, err := vault.NewClient(vaultAddr, vaultToken)
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	filter := vault.NewFilter(prefixes, excludePaths)

	secretsA, err := vault.RecurseSecrets(cmd.Context(), client, args[0], filter)
	if err != nil {
		return fmt.Errorf("reading %s: %w", args[0], err)
	}

	secretsB, err := vault.RecurseSecrets(cmd.Context(), client, args[1], filter)
	if err != nil {
		return fmt.Errorf("reading %s: %w", args[1], err)
	}

	entries := diff.Compare(secretsA, secretsB)

	fmt, err := output.ParseFormat(formatFlag)
	if err != nil {
		return err
	}

	formatter := output.NewFormatter(os.Stdout, fmt)
	return formatter.Write(entries)
}
