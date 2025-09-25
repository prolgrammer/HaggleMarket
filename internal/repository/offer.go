package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/repository/model"
	"github.com/Masterminds/squirrel"

	"github.com/jmoiron/sqlx"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
)

type OfferRepository struct {
	db *sqlx.DB
}

func NewOfferRepository(db *sqlx.DB) *OfferRepository {
	return &OfferRepository{db: db}
}

func (r *OfferRepository) InsertOffer(
	ctx context.Context,
	offer entity.Offer,
) (uint, error) {
	offerModel := model.ConvertOfferEntityToModel(offer)

	checkExistingOfferQuery, args := squirrel.Select("count(*)").
		From("offers").
		Where(squirrel.Eq{"status": "pending",
			"product_id": offerModel.ProductID,
			"shop_id":    offerModel.ShopID,
			"user_id":    offerModel.UserID}).
		PlaceholderFormat(squirrel.Dollar).
		MustSql()
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, apperror.New(apperror.DatabaseError, "failed to begin transaction", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var existingOfferCount int
	err = tx.QueryRowxContext(ctx, checkExistingOfferQuery, args...).Scan(&existingOfferCount)
	if err != nil {
		return 0, apperror.New(apperror.DatabaseError, "error checking existing offer", err)
	}

	if existingOfferCount > 0 {
		return 0, apperror.New(apperror.Conflict,
			"user already has an active offer for this product in this shop", nil)
	}

	insertOfferQuery, args := squirrel.Insert("offers").
		Columns("offer_price", "currency", "status", "created_at", "updated_at", "expires_at",
			"shop_id", "user_id", "product_id").
		Values(offerModel.Price, offerModel.Currency, offerModel.Status,
			offerModel.CreatedAt, offerModel.UpdatedAt, offerModel.ExpiresAt,
			offerModel.ShopID, offerModel.UserID, offerModel.ProductID).
		Suffix("returning id").
		PlaceholderFormat(squirrel.Dollar).
		MustSql()

	var offerID uint
	err = tx.QueryRowxContext(ctx, insertOfferQuery, args...).Scan(&offerID)
	if err != nil {
		return 0, apperror.New(apperror.DatabaseError, "error inserting offer into database", err)
	}

	err = tx.Commit()
	if err != nil {
		return 0, apperror.New(apperror.DatabaseError, "failed to commit transaction", err)
	}

	return offerID, nil
}

func (r *OfferRepository) GetOfferByID(
	ctx context.Context,
	offerID uint,
) (entity.Offer, error) {
	var offer entity.Offer

	_ = ctx
	_ = offerID

	return offer, nil
}

func (r *OfferRepository) SelectUserOffers(
	ctx context.Context,
	userID uint,
	limit, offset int,
) ([]entity.Offer, int, error) {
	var total int

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, 0, apperror.New(apperror.DatabaseError,
			"failed to begin transaction", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// lazy update
	if err = r.updateExpiredOffers(ctx, userID, tx); err != nil {
		return nil, 0, err
	}

	selectUserOffersQuery, args := squirrel.Select("id, offer_price, currency, status, " +
		"created_at, updated_at, expires_at, shop_id, product_id, user_id," +
		"COUNT (*) OVER() as total_count").
		From("offers").
		Where(squirrel.Eq{"status": "pending", "user_id": userID}).
		OrderBy("created_at desc").
		Offset(uint64(offset)).
		Limit(uint64(limit)).
		PlaceholderFormat(squirrel.Dollar).
		MustSql()

	offersWithCount := make([]model.OfferWithCount, 0, limit)

	err = tx.SelectContext(ctx, &offersWithCount, selectUserOffersQuery, args...)
	if err != nil {
		return nil, 0, apperror.New(apperror.DatabaseError, "error selecting user offers after lazy update", err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, 0, apperror.New(apperror.DatabaseError, "failed to commit transaction for lazy update", err)
	}

	if len(offersWithCount) == 0 {
		return []entity.Offer{}, 0, nil
	}

	total = offersWithCount[0].TotalCount

	offers := make([]entity.Offer, len(offersWithCount))
	for i, offerModel := range offersWithCount {
		offers[i] = offerModel.ConvertToEntity()
	}

	return offers, total, nil
}

// SelectUserOffers lazy update helper function
func (r *OfferRepository) updateExpiredOffers(ctx context.Context, userID uint, tx *sqlx.Tx) error {
	updateExpiredQuery, args := squirrel.Update("offers").
		Set("status", "cancelled").
		Set("updated_at", time.Now()).
		Where(squirrel.Lt{"expires_at": time.Now()}).
		Where(squirrel.Eq{"status": "pending"}).
		Where(squirrel.Eq{"user_id": userID}).
		PlaceholderFormat(squirrel.Dollar).
		MustSql()

	_, err := tx.ExecContext(ctx, updateExpiredQuery, args...)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return apperror.New(apperror.DatabaseError, "error updating expired offers", err)
	}
	return nil
}

func (r *OfferRepository) UpdateOfferStatus(
	ctx context.Context,
	offerEntity entity.Offer,
	userID uint,
	isStore bool,
) (entity.Offer, error) {
	offer := model.ConvertOfferEntityToModel(offerEntity)

	updateOfferStatusQuery, args := squirrel.Update("offers").
		Set("status", offer.Status).
		Set("updated_at", time.Now()).
		Where(squirrel.Eq{"id": offer.ID}).
		Suffix("returning status").
		PlaceholderFormat(squirrel.Dollar).
		MustSql()

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return entity.Offer{}, apperror.New(apperror.DatabaseError, "failed to begin transaction", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	err = isPendingOffer(ctx, offer.ID, tx)
	if err != nil {
		return entity.Offer{}, err
	}

	if isStore {
		err = isUserShopOwner(ctx, offer.ID, userID, tx)
		if err != nil {
			return entity.Offer{}, err
		}

	} else {
		// Если запрос на обновление статуса отправляет НЕ магазин, то добавляем проверку user_id
		// в квери, чтобы убедиться, что пользователь является создателем оффера.
		updateOfferStatusQuery, args = squirrel.Update("offers").
			Set("status", offer.Status).
			Set("updated_at", time.Now()).
			Where(squirrel.Eq{"id": offer.ID, "user_id": userID}).
			Suffix("returning status").
			PlaceholderFormat(squirrel.Dollar).
			MustSql()
	}

	var offerResp model.Offer
	err = tx.QueryRowxContext(ctx, updateOfferStatusQuery, args...).StructScan(&offerResp)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Offer{}, apperror.New(apperror.Unauthorized,
				"unauthorized to update offer status", nil)
		}
		return entity.Offer{}, apperror.New(apperror.DatabaseError, "error scanning into struct", err)
	}

	err = tx.Commit()
	if err != nil {
		return entity.Offer{}, apperror.New(apperror.DatabaseError, "failed to commit transaction", err)
	}

	return offerResp.ConvertToEntity(), nil
}

func isUserShopOwner(ctx context.Context, offerID, userID uint, tx *sqlx.Tx) error {
	validateShopOwnerIDQuery, args := squirrel.Select("users.id").
		From("users").
		InnerJoin("shops on users.id = shops.user_id").
		InnerJoin("offers on shops.id = offers.shop_id").
		Where(squirrel.Eq{"offers.id": offerID}).
		PlaceholderFormat(squirrel.Dollar).
		MustSql()

	var requiredID uint
	err := tx.QueryRowxContext(ctx, validateShopOwnerIDQuery, args...).Scan(&requiredID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.ErrUserNotFound
		}
		return apperror.New(apperror.InternalError, "error scanning into uint", err)
	}

	if userID != requiredID {
		return apperror.New(apperror.Unauthorized, "unauthorized to update offer status", nil)
	}

	return nil
}

func isPendingOffer(ctx context.Context, offerID uint, tx *sqlx.Tx) error {
	getOfferStatusQuery, args := squirrel.Select("offers.status = 'pending'").
		From("offers").
		Where(squirrel.Eq{"offers.id": offerID}).
		PlaceholderFormat(squirrel.Dollar).
		MustSql()

	var ok bool
	err := tx.QueryRowxContext(ctx, getOfferStatusQuery, args...).Scan(&ok)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.ErrOfferNotFound
		}
		return apperror.New(apperror.InternalError, "error scanning offer status", err)
	}

	if !ok {
		return apperror.New(apperror.Conflict, "offer is not in a pending status", nil)
	}

	return nil
}

func (r *OfferRepository) DeleteOffer(
	ctx context.Context,
	offerID uint,
) (entity.Offer, error) {
	var offer entity.Offer

	_ = ctx
	_ = offerID

	return offer, nil
}
