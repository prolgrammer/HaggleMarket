package repository

import (
	"log"

	"github.com/EM-Stawberry/Stawberry/config"

	// Импорт драйвера для sqlx
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func InitDB(cfg *config.Config) *sqlx.DB {
	db, err := sqlx.Connect("pgx", cfg.DB.GetDBConnString())
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	return db
}
