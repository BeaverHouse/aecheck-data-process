package common

import (
	"log"

	"github.com/BeaverHouse/go-common/logger"
)

// Log is the process-wide logger selected from the environment.
var Log = newLogger()

func newLogger() logger.Logger {
	built, err := logger.NewLogger()
	if err != nil {
		log.Fatalf("logger init failed: %v", err)
	}
	return built
}
