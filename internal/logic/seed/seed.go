package seed

import (
	"aecheck-data-process/internal/db/postgres"
	"aecheck-data-process/internal/logic/common"
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gopostgres "github.com/BeaverHouse/go-common/database/postgres"
	"github.com/BeaverHouse/go-common/env"
	"github.com/BeaverHouse/go-common/errorhandle"
	"github.com/BeaverHouse/go-common/logger"
	"github.com/jackc/pgx/v5"
)

// The tables that hold data scraped by hand over time, which a fresh schema
// cannot rebuild.
var tables = []string{"aecheck.characters", "aecheck.buddies", "aecheck.dungeons", "aecheck.translations"}

// Export writes each table to dir as CSV with a header row inside a read-only
// transaction, so it is the step that runs against production.
func Export(ctx context.Context, dir string, dryRun bool) error {
	if err := refuseRepoDir(dir); err != nil {
		return err
	}
	if dryRun {
		fmt.Println("Target host:", env.GetEnv("POSTGRES_HOST", ""))
		for _, table := range tables {
			fmt.Printf("%s -> %s\n", table, filepath.Join(dir, table+".csv"))
		}
		return nil
	}

	pool := gopostgres.InitFromEnv()
	defer pool.Close()
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return errorhandle.ErrDBOperation("begin export", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // a read-only transaction has nothing to keep

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return errorhandle.ErrInternal(err)
	}
	for _, table := range tables {
		file, err := os.Create(filepath.Join(dir, table+".csv"))
		if err != nil {
			return errorhandle.ErrInternal(err)
		}
		_, err = tx.Conn().PgConn().CopyTo(ctx, file, "COPY "+table+" TO STDOUT WITH (FORMAT csv, HEADER)")
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
		if err != nil {
			return errorhandle.ErrDBOperation("export "+table, err)
		}
	}
	common.Log.Info("Exported tables", logger.Field{Key: "dir", Value: dir})
	return nil
}

// Reset drops the aecheck schema of the local test database and rebuilds it
// empty from this repository's schema file.
func Reset(ctx context.Context, dryRun bool) error {
	host, err := localHost()
	if err != nil {
		return err
	}
	if dryRun {
		fmt.Println("Target host:", host)
		fmt.Println(postgres.ResetSQL)
		return nil
	}

	pool := gopostgres.InitFromEnv()
	defer pool.Close()
	if err := postgres.ResetSchema(ctx, pool); err != nil {
		return errorhandle.ErrDBOperation("reset schema", err)
	}
	common.Log.Info("Rebuilt the aecheck schema")
	return nil
}

// localHost refuses any database but a local one, since these commands erase.
func localHost() (string, error) {
	host := env.GetEnv("POSTGRES_HOST", "")
	if host != "127.0.0.1" && host != "localhost" {
		return "", errorhandle.New(errorhandle.KindInvalid, "SEED_TARGET_NOT_LOCAL",
			fmt.Sprintf("seed-test writes only to a local database, not %q", host))
	}
	return host, nil
}

// Import replaces the tables in the local test database with the files Export
// wrote, in one transaction.
func Import(ctx context.Context, dir string, dryRun bool) error {
	host, err := localHost()
	if err != nil {
		return err
	}
	if err := refuseRepoDir(dir); err != nil {
		return err
	}
	if dryRun {
		fmt.Println("Target host:", host)
		fmt.Println("TRUNCATE", strings.Join(tables, ", "))
		for _, table := range tables {
			sql, err := copyStatement(dir, table)
			if err != nil {
				return err
			}
			fmt.Println(sql)
		}
		return nil
	}

	pool := gopostgres.InitFromEnv()
	defer pool.Close()
	tx, err := pool.Begin(ctx)
	if err != nil {
		return errorhandle.ErrDBOperation("begin import", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // a no-op once Commit succeeds

	if err := postgres.New(tx).TruncateSeedTables(ctx); err != nil {
		return errorhandle.ErrDBOperation("truncate seed tables", err)
	}
	for _, table := range tables {
		sql, err := copyStatement(dir, table)
		if err != nil {
			return err
		}
		file, err := os.Open(filepath.Join(dir, table+".csv"))
		if err != nil {
			return errorhandle.ErrInternal(err)
		}
		_, err = tx.Conn().PgConn().CopyFrom(ctx, file, sql)
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
		if err != nil {
			return errorhandle.ErrDBOperation("import "+table, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return errorhandle.ErrDBOperation("commit import", err)
	}
	common.Log.Info("Imported tables", logger.Field{Key: "dir", Value: dir})
	return nil
}

// copyStatement names the columns from the file's header, so a column order
// that differs between the two databases does not matter.
func copyStatement(dir, table string) (string, error) {
	file, err := os.Open(filepath.Join(dir, table+".csv"))
	if err != nil {
		return "", errorhandle.ErrInternal(err)
	}
	defer file.Close()

	header, err := csv.NewReader(file).Read()
	if err != nil {
		return "", errorhandle.ErrInternal(err)
	}
	columns := make([]string, len(header))
	for i, name := range header {
		columns[i] = pgx.Identifier{name}.Sanitize()
	}
	return "COPY " + table + " (" + strings.Join(columns, ", ") + ") FROM STDIN WITH (FORMAT csv, HEADER)", nil
}

// refuseRepoDir keeps production dumps out of the working tree, where the CLI
// runs from.
func refuseRepoDir(dir string) error {
	root, err := os.Getwd()
	if err != nil {
		return errorhandle.ErrInternal(err)
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return errorhandle.ErrInternal(err)
	}
	rel, err := filepath.Rel(root, abs)
	if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return errorhandle.New(errorhandle.KindInvalid, "SEED_DIR_IN_REPO",
			fmt.Sprintf("%s is inside the repository; choose a directory outside it", dir))
	}
	return nil
}
