package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/user"
	"github.com/EM-Stawberry/Stawberry/internal/repository/model"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// InsertUser вставляет пользователя в БД
func (r *UserRepository) InsertUser(
	ctx context.Context,
	user user.User,
) (uint, error) {
	userModel := model.ConvertUserFromSvc(user)

	stmt := sq.Insert("users").
		Columns("name", "email", "phone_number", "password_hash", "is_store").
		Values(user.Name, user.Email, user.Phone, user.Password, user.IsStore).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar)

	query, args := stmt.MustSql()

	err := r.db.QueryRowxContext(ctx, query, args...).Scan(&userModel.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr); pgErr.Code == pgerrcode.UniqueViolation {
			return 0, apperror.New(apperror.DuplicateError, "user with this email already exists", err)
		}
		return 0, apperror.New(apperror.DuplicateError, "failed to create user", err)
	}

	return userModel.ID, nil
}

// GetUser получает пользователя по почте
func (r *UserRepository) GetUser(
	ctx context.Context,
	email string,
) (entity.User, error) {
	var userModel model.User

	stmt := sq.Select("id", "name", "email", "phone_number", "password_hash", "is_store").
		From("users").
		Where(sq.Eq{"email": email}).
		PlaceholderFormat(sq.Dollar)

	query, args := stmt.MustSql()

	err := r.db.QueryRowxContext(ctx, query, args...).StructScan(&userModel)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.User{}, apperror.ErrUserNotFound
		}
		return entity.User{}, apperror.New(apperror.DatabaseError, "failed to fetch user by email", err)
	}

	return model.ConvertUserToEntity(userModel), nil
}

// GetUserByID получает пользователя по айди
func (r *UserRepository) GetUserByID(
	ctx context.Context,
	id uint,
) (entity.User, error) {
	var userModel model.User

	stmt := sq.Select("id", "name", "email", "phone_number", "password_hash", "is_store").
		From("users").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar)

	query, args := stmt.MustSql()

	err := r.db.QueryRowxContext(ctx, query, args...).StructScan(&userModel)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.User{}, apperror.ErrUserNotFound
		}
		return entity.User{}, apperror.New(apperror.DatabaseError, "failed to fetch user by ID", err)
	}

	return model.ConvertUserToEntity(userModel), nil
}
