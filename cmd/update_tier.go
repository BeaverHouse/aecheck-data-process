package cmd

import (
	"aecheck-data-process/internal/logic/batch"

	"github.com/spf13/cobra"
)

var updateTierCmd = &cobra.Command{
	Use:   "update-tier",
	Short: "Update character tier data from JSON",
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonPath, _ := cmd.Flags().GetString("json")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		dbService, closeDB := bootstrapDB()
		defer closeDB()

		batch.UpdateTierFromJSON(jsonPath, dbService, dryRun)
		return nil
	},
}

func init() {
	updateTierCmd.Flags().String("json", "tier.json", "Path to tier JSON file")
	updateTierCmd.Flags().Bool("dry-run", false, "Run without applying changes to DB")
	rootCmd.AddCommand(updateTierCmd)
}
