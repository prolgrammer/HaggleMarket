package database

import (
	"context"
	"fmt"
	"strings"

	"github.com/EM-Stawberry/Stawberry/internal/repository/model"
	"github.com/EM-Stawberry/Stawberry/pkg/security"
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

func SeedDB() {
	pkgLog.Warn("clearing out all tables and seeding with test data")
	users, err := formDefaultUsers()
	if err != nil {
		return
	}

	tx, err := pkgDB.BeginTxx(context.Background(), nil)
	if err != nil {
		pkgLog.Error("Failed to start transaction, aborting", zap.Error(err))
		return
	}
	defer func() {
		_ = tx.Rollback()
	}()

	_, err = pkgDB.Exec("TRUNCATE users CASCADE")
	if err != nil {
		pkgLog.Error("Failed to truncate users table, aborting", zap.Error(err))
		return
	}

	_, err = pkgDB.Exec("ALTER SEQUENCE users_id_seq RESTART WITH 1")
	if err != nil {
		pkgLog.Error("Failed to reset users sequence, aborting", zap.Error(err))
		return
	}

	q := squirrel.Insert("users").
		Columns("name", "email", "phone_number", "password_hash", "is_store")

	for _, u := range users {
		q = q.Values(u.Name, u.Email, u.Phone, u.Password, u.IsStore)
	}

	sql, args, err := q.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		pkgLog.Error("Failed to generate SQL for seeding users, aborting", zap.Error(err))
		return
	}

	_, err = pkgDB.Exec(sql, args...)
	if err != nil {
		pkgLog.Error("Failed to seed users, aborting", zap.Error(err))
		return
	}

	_, err = sqlx.LoadFile(pkgDB, "./migrations/seed_data/seed_data.sql")
	if err != nil {
		pkgLog.Error("Failed to load seed data, aborting", zap.Error(err))
		return
	}

	err = tx.Commit()
	if err != nil {
		pkgLog.Error("Failed to commit transaction, aborting", zap.Error(err))
		return
	}

	DefaultAdminAcc()
}

func formDefaultUsers() ([]model.User, error) {
	passwords := []string{"shop1", "shop2", "user1", "user2"}
	users := make([]model.User, 0, len(passwords))
	for _, psw := range passwords {
		hash, err := security.HashArgon2id(psw)
		if err != nil {
			pkgLog.Error("Failed to hash password, aborting seeding", zap.Error(err))
			return nil, err
		}
		users = append(users, model.User{
			Name:     psw,
			Phone:    fmt.Sprintf("%sphone", psw),
			Email:    fmt.Sprintf("%s@%s.com", psw, psw),
			Password: hash,
			IsStore:  strings.Contains(psw, "shop"),
			// IsAdmin: false,
		})
	}
	return users, nil
}

func ClearDB() {
	pkgLog.Warn("clearing out all tables and resetting owned sequences")

	err := goose.Reset(pkgDB.DB, "./migrations")
	if err != nil {
		pkgLog.Error("Failed to reset migrations, aborting", zap.Error(err))
		return
	}

	err = goose.Up(pkgDB.DB, "./migrations")
	if err != nil {
		pkgLog.Error("Failed to run migrations, aborting", zap.Error(err))
		return
	}

	DefaultAdminAcc()
}

func DefaultAdminAcc() {
	hash, err := security.HashArgon2id(pkgCfg.DefAdmPswd)
	if err != nil {
		pkgLog.Error("Failed to hash default admin password", zap.Error(err))
		return
	}

	admin := model.User{
		Name:     "admin",
		Phone:    "adminphone",
		Email:    "admin@admin.com",
		Password: hash,
		IsStore:  false,
		// IsAdmin: true,
	}

	// PLEASE ADD ADMIN FLAG WHEN POSSIBLE
	q, args := squirrel.Insert("users").
		Columns("name", "email", "phone_number", "password_hash", "is_store").
		Values(admin.Name, admin.Email, admin.Phone, admin.Password, admin.IsStore).
		PlaceholderFormat(squirrel.Dollar).MustSql()

	_, err = pkgDB.Exec(q, args...)
	if err != nil && !strings.Contains(err.Error(), "duplicate key value") {
		pkgLog.Error("Failed to insert default admin account", zap.Error(err))
		return
	}
}
