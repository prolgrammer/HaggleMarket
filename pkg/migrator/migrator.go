package migrator

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

// RunMigrations applies database migrations using *sqlx.DB.
func RunMigrations(db *sqlx.DB, migrationsDir string) {
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal(err)
	}

	if err := goose.Up(db.DB, migrationsDir); err != nil {
		log.Fatal(err)
	}

	log.Println("Migrations applied successfully!")
}

// RunMigrationsWithZap applies database migrations using *sqlx.DB with Zap logger.
func RunMigrationsWithZap(db *sqlx.DB, migrationsDir string, logger *zap.Logger) {
	if err := goose.SetDialect("postgres"); err != nil {
		logger.Fatal("Failed to set database dialect", zap.Error(err))
	}

	if err := goose.Up(db.DB, migrationsDir); err != nil {
		logger.Fatal("Failed to apply migrations", zap.Error(err))
	}

	logger.Info("Migrations applied successfully")
}
