package reviews_test

import (
	"context"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/reviews"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

type mockSellerReviewRepository struct {
	sellers map[int]entity.SellerReview
	reviews map[int][]entity.SellerReview
}

func newMockSellerReviewRepository() *mockSellerReviewRepository {
	return &mockSellerReviewRepository{
		sellers: make(map[int]entity.SellerReview),
		reviews: make(map[int][]entity.SellerReview),
	}
}

func (m *mockSellerReviewRepository) AddReview(
	_ context.Context, sellerID int, userID int, rating int, review string,
) (int, error) {
	if _, exists := m.sellers[sellerID]; !exists {
		return 0, apperror.NewReviewError(apperror.NotFound, "seller not found")
	}

	reviewEntity := entity.SellerReview{
		SellerID: sellerID,
		UserID:   userID,
		Rating:   rating,
		Review:   review,
	}
	m.reviews[sellerID] = append(m.reviews[sellerID], reviewEntity)
	return sellerID, nil
}

func (m *mockSellerReviewRepository) GetSellerByID(
	_ context.Context, sellerID int,
) (entity.SellerReview, error) {
	seller, exists := m.sellers[sellerID]
	if !exists {
		return entity.SellerReview{}, apperror.NewReviewError(apperror.NotFound, "seller not found")
	}
	return seller, nil
}

func (m *mockSellerReviewRepository) GetReviewsBySellerID(
	_ context.Context, sellerID int,
) ([]entity.SellerReview, error) {
	return m.reviews[sellerID], nil
}

var _ = Describe("SellerReviewService", func() {
	var (
		service reviews.SellerReviewsService
		repo    *mockSellerReviewRepository
		ctx     context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		repo = newMockSellerReviewRepository()
		service = reviews.NewSellerReviewService(repo, zap.NewNop())
	})

	Context("AddReview", func() {
		It("should add a new review successfully", func() {
			// Arrange
			sellerID := 1
			userID := 2
			rating := 5
			review := "Great seller!"

			// Добавляем тестового продавца
			repo.sellers[sellerID] = entity.SellerReview{
				SellerID: sellerID,
				UserID:   userID,
				Rating:   rating,
				Review:   review,
			}

			// Act
			id, err := service.AddReview(ctx, sellerID, userID, rating, review)

			// Assert
			Expect(err).NotTo(HaveOccurred())
			Expect(id).To(Equal(sellerID))
			reviews, err := repo.GetReviewsBySellerID(ctx, sellerID)
			Expect(err).NotTo(HaveOccurred())
			Expect(reviews).To(HaveLen(1))
			Expect(reviews[0].SellerID).To(Equal(sellerID))
			Expect(reviews[0].UserID).To(Equal(userID))
			Expect(reviews[0].Rating).To(Equal(rating))
			Expect(reviews[0].Review).To(Equal(review))
		})

		It("should return error for non-existent seller", func() {
			// Act
			_, err := service.AddReview(ctx, 999, 1, 5, "Review")

			// Assert
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("seller not found"))
		})
	})

	Context("GetReviewsByID", func() {
		It("should return reviews for existing seller", func() {
			// Arrange
			sellerID := 1
			userID := 2
			rating := 5
			review := "Great seller!"

			repo.sellers[sellerID] = entity.SellerReview{
				SellerID: sellerID,
				UserID:   userID,
				Rating:   rating,
				Review:   review,
			}
			repo.reviews[sellerID] = []entity.SellerReview{
				{
					SellerID: sellerID,
					UserID:   userID,
					Rating:   rating,
					Review:   review,
				},
			}

			// Act
			reviews, err := service.GetReviewsByID(ctx, sellerID)

			// Assert
			Expect(err).NotTo(HaveOccurred())
			Expect(reviews).To(HaveLen(1))
			Expect(reviews[0].SellerID).To(Equal(sellerID))
			Expect(reviews[0].UserID).To(Equal(userID))
			Expect(reviews[0].Rating).To(Equal(rating))
			Expect(reviews[0].Review).To(Equal(review))
		})

		It("should return error for non-existent seller", func() {
			// Act
			_, err := service.GetReviewsByID(ctx, 999)

			// Assert
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("seller not found"))
		})
	})
})
