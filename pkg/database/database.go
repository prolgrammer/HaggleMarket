package database

import (
	"github.com/EM-Stawberry/Stawberry/config"
	"go.uber.org/zap"

	// Import pgx driver to enable database connection via database/sql
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

var (
	pkgDB  *sqlx.DB
	pkgLog *zap.Logger
	pkgCfg *config.DBConfig
)

func InitDB(cfg *config.DBConfig, log *zap.Logger) (*sqlx.DB, func()) {
	db, err := sqlx.Connect("pgx", cfg.GetDBConnString())
	if err != nil {
		log.Fatal("Failed to connect to database:", zap.Error(err))
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)

	closer := func() {
		if err := db.Close(); err != nil {
			log.Error("error closing database", zap.Error(err))
		}
	}

	pkgDB = db
	pkgLog = log
	pkgCfg = cfg

	return db, closer
}
