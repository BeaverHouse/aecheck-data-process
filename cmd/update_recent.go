package cmd

import (
	"aecheck-data-process/internal/logic/batch"

	"github.com/spf13/cobra"
)

var updateRecentCmd = &cobra.Command{
	Use:          "update-recent",
	Short:        "Update recent released, sidekick, and stellar entries from AEWiki home",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		dbService, closeDB := bootstrapDB()
		defer closeDB()

		return batch.UpdateRecent(dbService)
	},
}

func init() {
	rootCmd.AddCommand(updateRecentCmd)
}
