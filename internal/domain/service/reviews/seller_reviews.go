package reviews

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"go.uber.org/zap"
)

type SellerReviewRepository interface {
	AddReview(ctx context.Context, sellerID int, userID int, rating int, review string) (int, error)
	GetReviewsBySellerID(ctx context.Context, sellerID int) ([]entity.SellerReview, error)
	GetSellerByID(ctx context.Context, sellerID int) (entity.SellerReview, error)
}

// SellerReviewsService defines the interface for seller review business logic.
type SellerReviewsService interface {
	AddReview(ctx context.Context, sellerID int, userID int, rating int, review string) (int, error)
	GetReviewsByID(ctx context.Context, sellerID int) ([]entity.SellerReview, error)
}

type SellerReviewService struct {
	srs    SellerReviewRepository
	logger *zap.Logger
}

func NewSellerReviewService(srr SellerReviewRepository, l *zap.Logger) SellerReviewsService {
	return &SellerReviewService{
		srs:    srr,
		logger: l,
	}
}

func (s *SellerReviewService) AddReview(
	ctx context.Context, sellerID int, userID int, rating int, review string,
) (
	int, error,
) {
	const op = "sellerReviewService.AddReview()"
	log := s.logger.With(zap.String("op", op))

	log.Info("Existence check")
	_, err := s.srs.GetSellerByID(ctx, sellerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, apperror.NewReviewError(apperror.NotFound, "seller not found")
		}
		log.Warn("Seller not found", zap.Int("sellerID", sellerID), zap.Error(err))
		return 0, fmt.Errorf("op: %s, err: %w", op, err)
	}

	log.Info("Adding a review")
	sellerID, err = s.srs.AddReview(ctx, sellerID, userID, rating, review)
	if err != nil {
		log.Warn("Failed to add review", zap.Error(err))
		return 0, fmt.Errorf("op: %s, err: %w", op, err)
	}

	log.Info("Review added successfully")
	return sellerID, nil
}

func (s *SellerReviewService) GetReviewsByID(
	ctx context.Context, sellerID int,
) (
	[]entity.SellerReview, error,
) {
	const op = "sellerReviewService.GetReviewsBySellerID()"
	log := s.logger.With(zap.String("op", op))

	log.Info("Existence check")
	_, err := s.srs.GetSellerByID(ctx, sellerID)
	if err != nil {
		var reviewErr *apperror.ReviewError
		if errors.As(err, &reviewErr) {
			return nil, err
		}
		log.Error("Failed to execute query", zap.Error(err))
		return nil, fmt.Errorf("op: %s, err: %w", op, err)
	}

	log.Info("Getting reviews")
	reviews, err := s.srs.GetReviewsBySellerID(ctx, sellerID)
	if err != nil {
		log.Error("Failed to fetch reviews", zap.Error(err))
		return nil, fmt.Errorf("op: %s, err: %w", op, err)
	}

	log.Info("Reviews gets successfully")
	return reviews, nil
}
