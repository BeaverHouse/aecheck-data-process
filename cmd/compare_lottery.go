package cmd

import (
	"aecheck-data-process/internal/logic/batch"

	"github.com/spf13/cobra"
)

var lotteryID string

var compareLotteryCmd = &cobra.Command{
	Use:   "compare-lottery",
	Short: "Compare lottery translation data between API and DB",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbService, closeDB := bootstrapDB()
		defer closeDB()

		batch.CompareLotteryTranslations(lotteryID, dbService)
		return nil
	},
}

func init() {
	compareLotteryCmd.Flags().StringVar(&lotteryID, "id", "f603d8fd25c8968aa45ac6c21fdbb797", "Lottery ID")
	rootCmd.AddCommand(compareLotteryCmd)
}
