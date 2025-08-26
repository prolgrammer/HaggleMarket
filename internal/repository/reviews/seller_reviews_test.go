package reviews_test

import (
	"context"
	"errors"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	servicereview "github.com/EM-Stawberry/Stawberry/internal/domain/service/reviews"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type mockSellerReviewRepository struct {
	reviews map[int][]entity.SellerReview
	nextID  int
}

func newMockSellerReviewRepository() *mockSellerReviewRepository {
	return &mockSellerReviewRepository{
		reviews: make(map[int][]entity.SellerReview),
		nextID:  1,
	}
}

func (m *mockSellerReviewRepository) AddReview(
	_ context.Context, sellerID int, userID int, rating int, review string,
) (int, error) {
	reviewEntity := entity.SellerReview{
		ID:       m.nextID,
		SellerID: sellerID,
		UserID:   userID,
		Rating:   rating,
		Review:   review,
	}
	m.reviews[sellerID] = append(m.reviews[sellerID], reviewEntity)
	m.nextID++
	return reviewEntity.ID, nil
}

func (m *mockSellerReviewRepository) GetReviewsBySellerID(
	_ context.Context, sellerID int,
) ([]entity.SellerReview, error) {
	return m.reviews[sellerID], nil
}

func (m *mockSellerReviewRepository) GetSellerByID(
	_ context.Context, sellerID int,
) (entity.SellerReview, error) {
	reviews, exists := m.reviews[sellerID]
	if !exists || len(reviews) == 0 {
		return entity.SellerReview{}, apperror.NewReviewError(apperror.NotFound, "seller not found")
	}
	return reviews[0], nil
}

var _ = Describe("SellerReviewsRepository", func() {
	var (
		repository servicereview.SellerReviewRepository
		ctx        context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		repository = newMockSellerReviewRepository()
	})

	Context("AddReview", func() {
		It("should add a new review successfully", func() {
			// Arrange
			sellerID := 1
			userID := 2
			rating := 5
			review := "Great seller!"

			// Act
			id, err := repository.AddReview(ctx, sellerID, userID, rating, review)

			// Assert
			Expect(err).NotTo(HaveOccurred())
			Expect(id).To(BeNumerically(">", 0))
		})
	})

	Context("GetReviewsBySellerID", func() {
		It("should return reviews for existing seller", func() {
			// Arrange
			sellerID := 1
			userID := 2
			rating := 5
			review := "Great seller!"

			// Добавляем тестовый отзыв
			_, err := repository.AddReview(ctx, sellerID, userID, rating, review)
			Expect(err).NotTo(HaveOccurred())

			// Act
			reviews, err := repository.GetReviewsBySellerID(ctx, sellerID)

			// Assert
			Expect(err).NotTo(HaveOccurred())
			Expect(reviews).To(HaveLen(1))
			Expect(reviews[0].SellerID).To(Equal(sellerID))
			Expect(reviews[0].UserID).To(Equal(userID))
			Expect(reviews[0].Rating).To(Equal(rating))
			Expect(reviews[0].Review).To(Equal(review))
		})

		It("should return empty slice for non-existent seller", func() {
			// Act
			reviews, err := repository.GetReviewsBySellerID(ctx, 999)

			// Assert
			Expect(err).NotTo(HaveOccurred())
			Expect(reviews).To(BeEmpty())
		})
	})

	Context("GetSellerByID", func() {
		It("should return seller review if exists", func() {
			// Arrange
			sellerID := 1
			userID := 2
			rating := 5
			review := "Great seller!"

			// Добавляем тестовый отзыв
			_, err := repository.AddReview(ctx, sellerID, userID, rating, review)
			Expect(err).NotTo(HaveOccurred())

			// Act
			sellerReview, err := repository.GetSellerByID(ctx, sellerID)

			// Assert
			Expect(err).NotTo(HaveOccurred())
			Expect(sellerReview.SellerID).To(Equal(sellerID))
			Expect(sellerReview.UserID).To(Equal(userID))
			Expect(sellerReview.Rating).To(Equal(rating))
			Expect(sellerReview.Review).To(Equal(review))
		})

		It("should return NotFound error for non-existent seller", func() {
			// Act
			_, err := repository.GetSellerByID(ctx, 999)

			// Assert
			Expect(err).To(HaveOccurred())
			var reviewErr *apperror.ReviewError
			Expect(errors.As(err, &reviewErr)).To(BeTrue())
			Expect(reviewErr.Code).To(Equal(apperror.NotFound))
		})
	})
})
