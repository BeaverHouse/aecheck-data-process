package cmd

import (
	"fmt"
	"os"

	"github.com/BeaverHouse/go-common/env"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "aecheck",
	Short: "AE Check data processing CLI",
	Long:  `CLI tool for processing and managing Another Eden game data (characters, buddies, tiers, etc.).`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if env.IsGoEnv(env.LocalEnv) {
			if err := godotenv.Load(); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to load .env file: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
