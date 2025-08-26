package guestoffer

import (
	"context"
	"database/sql"
	"errors"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

// StoreInfoGetter interface for getting store information specific to guest offers
type StoreInfoGetter interface {
	GetStoreOwnerEmailByStoreID(ctx context.Context, storeID uint) (string, error)
}

type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new instance of Repository and returns the StoreInfoGetter interface
func NewRepository(db *sqlx.DB) StoreInfoGetter {
	return &Repository{db: db}
}

// GetStoreOwnerEmailByStoreID retrieves the email of the store owner by store ID.
// It implements the StoreInfoGetter interface.
func (r *Repository) GetStoreOwnerEmailByStoreID(ctx context.Context, storeID uint) (string, error) {
	query, args, err := squirrel.Select("users.email").
		From("users").
		Join("shops ON users.id = shops.user_id").
		Where(squirrel.Eq{"shops.id": storeID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return "", apperror.NewGuestOfferError(
			apperror.GuestOfferDatabaseError,
			"failed to build query for store owner email",
		)
	}

	var email string
	err = r.db.GetContext(ctx, &email, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", apperror.NewGuestOfferError(apperror.GuestOfferStoreNotFound, "store not found for guest offer")
		}
		return "", apperror.NewGuestOfferError(apperror.GuestOfferDatabaseError, "failed to get store owner email")
	}

	return email, nil
}
