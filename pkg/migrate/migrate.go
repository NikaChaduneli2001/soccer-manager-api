package migrate

import (
	"database/sql"
	"errors"
	"io/fs"
	"sort"
	"strings"

	"github.com/nika/soccer-manager-api/migrations"
)

// Serializes migration runs when several API processes start at once.
const migrationAdvisoryLockID int64 = 0x736f636d6772

const ensureMigrationsTable = `
CREATE TABLE IF NOT EXISTS schema_migrations (
	version TEXT PRIMARY KEY,
	applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`

// Run applies pending SQL migration files in filename order. Files already
// recorded in schema_migrations are skipped.
func Run(db *sql.DB) error {
	if _, err := db.Exec(`SELECT pg_advisory_lock($1)`, migrationAdvisoryLockID); err != nil {
		return err
	}
	defer db.Exec(`SELECT pg_advisory_unlock($1)`, migrationAdvisoryLockID)

	if _, err := db.Exec(ensureMigrationsTable); err != nil {
		return err
	}

	entries, err := fs.ReadDir(migrations.FS, ".")
	if err != nil {
		return err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		var one int
		err := db.QueryRow(`SELECT 1 FROM schema_migrations WHERE version = $1`, name).Scan(&one)
		if err == nil {
			continue
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		body, err := fs.ReadFile(migrations.FS, name)
		if err != nil {
			return err
		}

		tx, err := db.Begin()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(string(body)); err != nil {
			tx.Rollback()
			return err
		}
		if _, err := tx.Exec(`INSERT INTO schema_migrations (version) VALUES ($1)`, name); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}
