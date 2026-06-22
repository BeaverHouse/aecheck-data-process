package batch

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"aecheck-data-process/internal/logic/common"
	"aecheck-data-process/internal/logic/database"
	"aecheck-data-process/internal/logic/parse"

	"github.com/BeaverHouse/go-common/logger"
)

type recentLinkProcessor struct {
	label   string
	prepare func(string, *database.Service) (*recentUpdateTask, error)
}

type recentUpdateTask struct {
	label  string
	name   string
	update func() error
}

func UpdateRecent(dbService *database.Service) error {
	groups, err := parse.GetRecentUpdateGroups()
	if err != nil {
		return err
	}

	var failures []error
	for _, group := range groups {
		processor := processorForRecentUpdate(group.Kind)
		common.Log.Info("Processing recent update group",
			logger.Field{Key: "kind", Value: group.Kind},
			logger.Field{Key: "label", Value: group.Label},
			logger.Field{Key: "count", Value: len(group.Links)},
		)
		for _, link := range group.Links {
			common.Log.Info("Processing recent "+processor.label,
				logger.Field{Key: "name", Value: link.Text},
				logger.Field{Key: "url", Value: link.URL},
			)
			task, err := processor.prepare(link.URL, dbService)
			if err != nil {
				common.Log.Error("Failed to process recent "+processor.label,
					logger.Field{Key: "name", Value: link.Text},
					logger.Field{Key: "error", Value: err},
				)
				failures = append(failures, fmt.Errorf("%s %q: %w", processor.label, link.Text, err))
				continue
			}
			task.name = link.Text
			if !confirmRecentUpdate(task) {
				common.Log.Info("Skipped recent "+task.label, logger.Field{Key: "name", Value: task.name})
				continue
			}
			if err := task.update(); err != nil {
				common.Log.Error("Failed to update recent "+task.label,
					logger.Field{Key: "name", Value: task.name},
					logger.Field{Key: "error", Value: err},
				)
				failures = append(failures, fmt.Errorf("%s %q: %w", task.label, task.name, err))
			}
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("recent update completed with failures: %w", errors.Join(failures...))
	}
	return nil
}

func processorForRecentUpdate(kind parse.RecentUpdateKind) recentLinkProcessor {
	if kind == parse.RecentUpdateSidekick {
		return recentLinkProcessor{label: "buddy", prepare: prepareRecentBuddy}
	}
	return recentLinkProcessor{label: "character", prepare: prepareRecentCharacter}
}

func prepareRecentCharacter(url string, dbService *database.Service) (task *recentUpdateTask, err error) {
	err = captureRecentPanic(func() {
		ctx := ResolveCharacter(url, dbService)
		CompareCharacter(ctx, dbService)
		task = &recentUpdateTask{
			label: "character",
			update: func() error {
				return captureRecentPanic(func() {
					UpdateCharacter(ctx, false, dbService)
				})
			},
		}
	})
	return task, err
}

func prepareRecentBuddy(url string, dbService *database.Service) (task *recentUpdateTask, err error) {
	err = captureRecentPanic(func() {
		ctx := ResolveBuddy(url, dbService)
		CompareBuddy(ctx, dbService)
		task = &recentUpdateTask{
			label: "buddy",
			update: func() error {
				return captureRecentPanic(func() {
					UpdateBuddy(ctx, false, dbService)
				})
			},
		}
	})
	return task, err
}

func captureRecentPanic(run func()) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("panic: %v", recovered)
		}
	}()
	run()
	return nil
}

func confirmRecentUpdate(task *recentUpdateTask) bool {
	fmt.Printf("이 recent %s(%s) 정보를 DB/스토리지에 반영할까요? (y/N): ", task.label, task.name)
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes"
}
