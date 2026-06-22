package cmd

import (
	"aecheck-data-process/internal/logic/batch"
	"aecheck-data-process/internal/logic/common"
	"context"
	"fmt"

	"github.com/BeaverHouse/go-common/logger"
	"github.com/spf13/cobra"
)

var updateCharacterCmd = &cobra.Command{
	Use:   "update-character",
	Short: "Update character data from AEWiki",
	RunE: func(cmd *cobra.Command, args []string) error {
		url, _ := cmd.Flags().GetString("url")
		random, _ := cmd.Flags().GetBool("random")

		if random && url != "" {
			return fmt.Errorf("--random and --url are mutually exclusive")
		}
		if !random && url == "" {
			return fmt.Errorf("--url is required (or use --random)")
		}

		dbService, closeDB := bootstrapDB()
		defer closeDB()

		if random {
			randomURL, err := dbService.GetRandomCharacterWikiURL(context.Background())
			if err != nil {
				return fmt.Errorf("failed to pick random character: %w", err)
			}
			url = randomURL
			common.Log.Info("Random character selected", logger.Field{Key: "url", Value: url})
		}

		ctx := batch.ResolveCharacter(url, dbService)
		batch.CompareCharacter(ctx, dbService)
		if !confirmApply("이 character 정보를 DB/스토리지에 반영할까요? (y/N): ") {
			common.Log.Info("취소되었습니다")
			return nil
		}
		batch.UpdateCharacter(ctx, false, dbService)
		return nil
	},
}

func init() {
	updateCharacterCmd.Flags().String("url", "", "AEWiki character URL")
	updateCharacterCmd.Flags().Bool("random", false, "Pick a random existing character from DB")
	rootCmd.AddCommand(updateCharacterCmd)
}
