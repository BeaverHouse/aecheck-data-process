package cmd

import (
	"aecheck-data-process/internal/logic/database"
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/BeaverHouse/go-common/database/postgres"
)

// bootstrapDB initializes a postgres pool and wraps it in a database.Service.
// The returned closer must be deferred by the caller.
func bootstrapDB() (*database.Service, func()) {
	pool := postgres.InitFromEnv()
	return database.NewService(pool), func() { pool.Close() }
}

func confirmApply(prompt string) bool {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes"
}
