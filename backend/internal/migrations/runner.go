package migrations

import (
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

//go:embed sql/*.sql
var sqlFS embed.FS

// Run applies all pending SQL migrations to the database behind the given
// GORM handle. Safe to call on every boot: golang-migrate tracks the applied
// version in `schema_migrations` and skips versions already applied.
//
// The SQL files are embedded into the binary at compile time (//go:embed),
// so a fresh environment only needs the backend image — no external
// migrations volume or filesystem layout is required at runtime.
//
// Why embed rather than reading from the container's filesystem?
//   - The Go binary is self-contained; we don't need to copy the migrations
//     directory into the runtime image.
//   - The same set of migrations ships with every binary; no risk of the
//     binary and the migrations volume drifting apart.
//   - Works identically for `go run`, `docker-compose up`, and CI.
//
// Order of operations on boot (see bootstrap.NewApp):
//  1. repository.InitDB — opens the GORM connection and runs AutoMigrate
//     for any tables that the GORM models declare but the SQL migrations
//     don't yet cover.
//  2. migrations.Run — applies the embedded SQL migrations. After this,
//     every table the SQL files describe is guaranteed to exist with the
//     authoritative schema, and the RBAC + reference data is seeded.
//  3. The Go idempotent seeders (SeedUnits, SeedChartOfAccounts,
//     SeedAIUser, SeedRoles) fill in data that lives in Go rather than
//     SQL (units, chart of accounts, AI agent user, role list).
func Run(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("migrations: get underlying *sql.DB: %w", err)
	}

	// source.iofs reads the migration files from the embedded FS.
	source, err := iofs.New(sqlFS, "sql")
	if err != nil {
		return fmt.Errorf("migrations: build iofs source: %w", err)
	}
	// We close the source only on fatal errors; the migrate library will
	// close it itself when Close() is called on the migrate instance.
	defer func() {
		_ = source.Close()
	}()

	driver, err := migratepg.WithInstance(sqlDB, &migratepg.Config{})
	if err != nil {
		return fmt.Errorf("migrations: build postgres driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return fmt.Errorf("migrations: init: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrations: up: %w", err)
	}

	version, dirty, vErr := m.Version()
	if vErr != nil && !errors.Is(vErr, migrate.ErrNilVersion) {
		// Non-fatal — we already applied whatever was pending.
		logrus.WithError(vErr).Warn("migrations: could not read current version")
	}
	if dirty {
		return fmt.Errorf("migrations: database is in a DIRTY state at version %d — manual intervention required", version)
	}
	logrus.WithFields(logrus.Fields{
		"version": version,
		"dirty":   dirty,
	}).Info("SQL migrations applied")
	return nil
}

// Version returns the current applied migration version. Used by the
// readiness probe to fail loud if a fresh DB somehow hasn't been migrated.
func Version(db *gorm.DB) (uint, bool, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return 0, false, err
	}
	driver, err := migratepg.WithInstance(sqlDB, &migratepg.Config{})
	if err != nil {
		return 0, false, err
	}
	source, err := iofs.New(sqlFS, "sql")
	if err != nil {
		return 0, false, err
	}
	defer func() { _ = source.Close() }()
	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return 0, false, err
	}
	version, dirty, err := m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		return 0, false, nil
	}
	return version, dirty, err
}
