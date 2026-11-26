package cmd

import (
	"aecheck-data-process/internal/logic/seed"

	"github.com/spf13/cobra"
)

var seedCmd = &cobra.Command{
	Use:   "seed-test",
	Short: "Copy the aecheck tables an empty test database cannot rebuild",
}

var seedExportCmd = &cobra.Command{
	Use:          "export <dir>",
	Short:        "Read the aecheck tables into CSV files (run against production)",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		return seed.Export(cmd.Context(), args[0], dryRun)
	},
}

var seedImportCmd = &cobra.Command{
	Use:          "import <dir>",
	Short:        "Replace the aecheck tables in the local test database with exported files",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		return seed.Import(cmd.Context(), args[0], dryRun)
	},
}

func init() {
	seedCmd.PersistentFlags().Bool("dry-run", false, "Print the plan without connecting or writing")
	seedCmd.AddCommand(seedExportCmd, seedImportCmd)
	rootCmd.AddCommand(seedCmd)
}
