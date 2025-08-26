package reviews

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// SellerReviewsRepository defines the interface for seller review data operations.
type SellerReviewsRepository interface {
	AddReview(ctx context.Context, sellerID int, userID int, rating int, review string) (int, error)
	GetReviewsBySellerID(ctx context.Context, sellerID int) ([]entity.SellerReview, error)
	GetSellerByID(ctx context.Context, sellerID int) (entity.SellerReview, error)
}

type sellerReviewsRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func NewSellerReviewRepository(db *sqlx.DB, l *zap.Logger) SellerReviewsRepository {
	return &sellerReviewsRepository{
		db:     db,
		logger: l,
	}
}

func (r *sellerReviewsRepository) AddReview(
	ctx context.Context, sellerID int, userID int, rating int, review string,
) (int, error) {
	const op = "sellerReviewsRepository.AddReview()"
	log := r.logger.With(zap.String("op", op))

	var id int
	query, args, err := squirrel.Insert("seller_reviews").
		Columns("seller_id", "user_id", "rating", "review").
		Values(sellerID, userID, rating, review).
		PlaceholderFormat(squirrel.Dollar).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		log.Error("Failed to build query", zap.Error(err))
		return 0, fmt.Errorf("op: %s, err: %w", op, err)
	}

	err = r.db.QueryRowContext(ctx, query, args...).Scan(&id)
	if err != nil {
		log.Error("Failed to execute query", zap.Error(err))
		return 0, fmt.Errorf("op: %s, err: %w", op, err)
	}

	return id, nil
}

func (r *sellerReviewsRepository) GetReviewsBySellerID(
	ctx context.Context, sellerID int,
) (
	[]entity.SellerReview, error,
) {
	const op = "sellerReviewsRepository.GetReviewsBySellerID()"
	log := r.logger.With(zap.String("op", op))

	query, args, err := squirrel.
		Select("id", "seller_id", "user_id", "rating", "review", "created_at").
		From("seller_reviews").
		Where("seller_id = $1", sellerID).ToSql()
	if err != nil {
		log.Error("Failed to build query", zap.Error(err))
		return nil, fmt.Errorf("op: %s, err: %s", op, err)
	}

	var reviews []entity.SellerReview
	err = r.db.SelectContext(ctx, &reviews, query, args...)
	if err != nil {
		log.Error("Failed to execute query", zap.Error(err))
		return nil, fmt.Errorf("op: %s, err: %s", op, err)
	}

	return reviews, nil
}

func (r *sellerReviewsRepository) GetSellerByID(
	ctx context.Context, sellerID int,
) (
	entity.SellerReview, error,
) {
	const op = "sellerReviewsRepository.GetSellerByID()"
	log := r.logger.With(zap.String("op", op))

	query, args, err := squirrel.
		Select("id", "seller_id", "user_id", "rating", "review", "created_at").
		From("seller_reviews").
		Where("seller_id = $1", sellerID).
		Limit(1).
		ToSql()
	if err != nil {
		log.Error("Failed to build query", zap.Error(err))
		return entity.SellerReview{}, fmt.Errorf("op: %s, err: %w", op, err)
	}

	var review entity.SellerReview
	err = r.db.GetContext(ctx, &review, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return entity.SellerReview{}, apperror.NewReviewError(apperror.NotFound, "seller not found")
		}
		log.Error("Failed to execute query", zap.Error(err))
		return entity.SellerReview{}, fmt.Errorf("op: %s, err: %w", op, err)
	}

	return review, nil
}
