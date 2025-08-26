package reviews_test

import (
	"context"
	"errors"
	"testing"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	servicereview "github.com/EM-Stawberry/Stawberry/internal/domain/service/reviews"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type mockProductReviewRepository struct {
	products map[int]entity.Product
	reviews  map[int][]entity.ProductReview
}

func newMockProductReviewRepository() *mockProductReviewRepository {
	return &mockProductReviewRepository{
		products: make(map[int]entity.Product),
		reviews:  make(map[int][]entity.ProductReview),
	}
}

func (m *mockProductReviewRepository) AddReview(
	_ context.Context, productID int, userID int, rating int, review string,
) error {
	if _, exists := m.products[productID]; !exists {
		return apperror.NewReviewError(apperror.NotFound, "product not found")
	}

	reviewEntity := entity.ProductReview{
		ProductID: productID,
		UserID:    userID,
		Rating:    rating,
		Review:    review,
	}
	m.reviews[productID] = append(m.reviews[productID], reviewEntity)
	return nil
}

func (m *mockProductReviewRepository) GetProductByID(
	_ context.Context, productID int,
) (entity.Product, error) {
	product, exists := m.products[productID]
	if !exists {
		return entity.Product{}, apperror.NewReviewError(apperror.NotFound, "product not found")
	}
	return product, nil
}

func (m *mockProductReviewRepository) GetReviewsByProductID(
	_ context.Context, productID int,
) ([]entity.ProductReview, error) {
	return m.reviews[productID], nil
}

var _ = Describe("ProductReviewsRepository", func() {
	var (
		repository servicereview.ProductReviewRepository
		ctx        context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		repository = newMockProductReviewRepository()
	})

	Context("AddReview", func() {
		It("should add a new review successfully", func() {
			// Arrange
			productID := 1
			userID := 2
			rating := 5
			review := "Great product!"

			// Добавляем тестовый продукт
			mockRepo := repository.(*mockProductReviewRepository)
			mockRepo.products[productID] = entity.Product{
				ID:          productID,
				Name:        "Test Product",
				Description: "Test Description",
				CategoryID:  1,
			}

			// Act
			err := repository.AddReview(ctx, productID, userID, rating, review)

			// Assert
			Expect(err).NotTo(HaveOccurred())
			reviews, err := repository.GetReviewsByProductID(ctx, productID)
			Expect(err).NotTo(HaveOccurred())
			Expect(reviews).To(HaveLen(1))
			Expect(reviews[0].ProductID).To(Equal(productID))
			Expect(reviews[0].UserID).To(Equal(userID))
			Expect(reviews[0].Rating).To(Equal(rating))
			Expect(reviews[0].Review).To(Equal(review))
		})

		It("should return error for non-existent product", func() {
			// Act
			err := repository.AddReview(ctx, 999, 1, 5, "Review")

			// Assert
			Expect(err).To(HaveOccurred())
			var reviewErr *apperror.ReviewError
			Expect(errors.As(err, &reviewErr)).To(BeTrue())
			Expect(reviewErr.Code).To(Equal(apperror.NotFound))
		})
	})

	Context("GetProductByID", func() {
		It("should return product if exists", func() {
			// Arrange
			productID := 1
			mockRepo := repository.(*mockProductReviewRepository)
			mockRepo.products[productID] = entity.Product{
				ID:          productID,
				Name:        "Test Product",
				Description: "Test Description",
				CategoryID:  1,
			}

			// Act
			product, err := repository.GetProductByID(ctx, productID)

			// Assert
			Expect(err).NotTo(HaveOccurred())
			Expect(product.ID).To(Equal(productID))
			Expect(product.Name).To(Equal("Test Product"))
			Expect(product.Description).To(Equal("Test Description"))
			Expect(product.CategoryID).To(Equal(1))
		})

		It("should return NotFound error for non-existent product", func() {
			// Act
			_, err := repository.GetProductByID(ctx, 999)

			// Assert
			Expect(err).To(HaveOccurred())
			var reviewErr *apperror.ReviewError
			Expect(errors.As(err, &reviewErr)).To(BeTrue())
			Expect(reviewErr.Code).To(Equal(apperror.NotFound))
		})
	})

	Context("GetReviewsByProductID", func() {
		It("should return reviews for existing product", func() {
			// Arrange
			productID := 1
			userID := 2
			rating := 5
			review := "Great product!"

			mockRepo := repository.(*mockProductReviewRepository)
			mockRepo.products[productID] = entity.Product{
				ID:          productID,
				Name:        "Test Product",
				Description: "Test Description",
				CategoryID:  1,
			}
			mockRepo.reviews[productID] = []entity.ProductReview{
				{
					ProductID: productID,
					UserID:    userID,
					Rating:    rating,
					Review:    review,
				},
			}

			// Act
			reviews, err := repository.GetReviewsByProductID(ctx, productID)

			// Assert
			Expect(err).NotTo(HaveOccurred())
			Expect(reviews).To(HaveLen(1))
			Expect(reviews[0].ProductID).To(Equal(productID))
			Expect(reviews[0].UserID).To(Equal(userID))
			Expect(reviews[0].Rating).To(Equal(rating))
			Expect(reviews[0].Review).To(Equal(review))
		})

		It("should return empty slice for non-existent product", func() {
			// Act
			reviews, err := repository.GetReviewsByProductID(ctx, 999)

			// Assert
			Expect(err).NotTo(HaveOccurred())
			Expect(reviews).To(BeEmpty())
		})
	})
})

func TestProductReviews(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ProductReviews Repository Suite")
}
