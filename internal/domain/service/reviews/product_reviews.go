package reviews

import (
	"context"
	"errors"
	"fmt"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"go.uber.org/zap"
)

var (
	ErrProductNotFound = errors.New("product not found")
)

type ProductReviewRepository interface {
	AddReview(ctx context.Context, productID int, userID int, rating int, review string) error
	GetReviewsByProductID(ctx context.Context, productID int) ([]entity.ProductReview, error)
	GetProductByID(ctx context.Context, productID int) (entity.Product, error)
}

// ProductReviewsService defines the interface for product review business logic.
type ProductReviewsService interface {
	AddReview(ctx context.Context, productID int, userID int, rating int, review string) (int, error)
	GetReviewsByProductID(ctx context.Context, productID int) ([]entity.ProductReview, error)
}

type ProductReviewService struct {
	prr    ProductReviewRepository
	logger *zap.Logger
}

func NewProductReviewService(prr ProductReviewRepository, l *zap.Logger) ProductReviewsService {
	return &ProductReviewService{
		prr:    prr,
		logger: l,
	}
}

func (s *ProductReviewService) AddReview(
	ctx context.Context, productID int, userID int, rating int, review string,
) (
	int, error,
) {
	const op = "productReviewService.AddReviews()"
	log := s.logger.With(zap.String("op", op))

	log.Info("Existence check")
	_, err := s.prr.GetProductByID(ctx, productID)
	if err != nil {
		log.Warn("Product not found", zap.Int("productID", productID), zap.Error(err))
		return 0, fmt.Errorf("op: %s, err: %w", op, err)
	}

	log.Info("Adding a review")
	err = s.prr.AddReview(ctx, productID, userID, rating, review)
	if err != nil {
		log.Warn("Failed to add review", zap.Error(err))
		return 0, fmt.Errorf("op: %s, err: %w", op, err)
	}

	log.Info("Review added successfully")
	return productID, nil
}

func (s *ProductReviewService) GetReviewsByProductID(
	ctx context.Context, productID int,
) (
	[]entity.ProductReview, error,
) {
	const op = "productReviewService.GetReviewsByProductID()"
	log := s.logger.With(zap.String("op", op))

	log.Info("Existence check")
	_, err := s.prr.GetProductByID(ctx, productID)
	if err != nil {
		log.Warn("Product not found", zap.Int("productID", productID), zap.Error(err))
		return nil, err
	}

	log.Info("Receiving reviews")
	reviews, err := s.prr.GetReviewsByProductID(ctx, productID)
	if err != nil {
		log.Warn("Failed to get reviews", zap.Error(err))
		return nil, fmt.Errorf("op: %s, err: %w", op, err)
	}

	log.Info("Reviews received successfully")
	return reviews, nil
}
