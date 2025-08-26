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

// ProductReviewRepository defines the interface for product review data operations.
type ProductReviewRepository interface {
	AddReview(ctx context.Context, productID int, userID int, rating int, review string) error
	GetProductByID(ctx context.Context, productID int) (entity.Product, error)
	GetReviewsByProductID(ctx context.Context, productID int) ([]entity.ProductReview, error)
}

type productReviewsRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func NewProductReviewRepository(db *sqlx.DB, l *zap.Logger) ProductReviewRepository {
	return &productReviewsRepository{
		db:     db,
		logger: l,
	}
}

func (r *productReviewsRepository) AddReview(
	ctx context.Context, productID int, userID int, rating int, review string,
) error {
	const op = "productReviewsRepository.AddReview()"
	log := r.logger.With(zap.String("op", op))

	query, args, err := squirrel.Insert("product_reviews").
		Columns("product_id", "user_id", "rating", "review").
		Values(productID, userID, rating, review).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		log.Error("Failed to build query", zap.Error(err))
		return fmt.Errorf("op: %s, err: %w", op, err)
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		log.Error("Failed to execute query", zap.Error(err))
		return fmt.Errorf("op: %s, err: %w", op, err)
	}

	return nil
}

func (r *productReviewsRepository) GetProductByID(
	ctx context.Context, productID int,
) (
	entity.Product, error,
) {
	const op = "productReviewsRepository.GetProductByID()"
	log := r.logger.With(zap.String("op", op))

	query, args, err := squirrel.
		Select("id", "name", "description", "category_id as categoryid").
		From("products").
		Where("id = $1", productID).
		Limit(1).
		ToSql()
	if err != nil {
		log.Error("Failed to build query", zap.Error(err))
		return entity.Product{}, fmt.Errorf("op: %s, err: %w", op, err)
	}

	var product entity.Product
	err = r.db.GetContext(ctx, &product, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return entity.Product{}, apperror.NewReviewError(apperror.NotFound, "product not found")
		}
		log.Error("Failed to execute query", zap.Error(err))
		return entity.Product{}, fmt.Errorf("op: %s, err: %w", op, err)
	}

	return product, nil
}

func (r *productReviewsRepository) GetReviewsByProductID(
	ctx context.Context, productID int,
) (
	[]entity.ProductReview, error,
) {
	const op = "productReviewsRepository.GetReviewsByProductID()"
	log := r.logger.With(zap.String("op", op))

	query, args, err := squirrel.
		Select("id", "product_id as productid", "user_id as userid", "rating", "review", "created_at").
		From("product_reviews").
		Where("product_id = $1", productID).
		ToSql()
	if err != nil {
		log.Error("Failed to build query", zap.Error(err))
		return nil, fmt.Errorf("op: %s, err: %s", op, err)
	}

	var reviews []entity.ProductReview
	err = r.db.SelectContext(ctx, &reviews, query, args...)
	if err != nil {
		log.Error("Failed to execute query", zap.Error(err))
		return nil, fmt.Errorf("op: %s, err: %s", op, err)
	}

	return reviews, nil
}
