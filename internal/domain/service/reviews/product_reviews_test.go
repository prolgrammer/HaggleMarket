package reviews_test

import (
	"context"
	"testing"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/reviews"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
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

var _ = Describe("ProductReviewService", func() {
	var (
		service reviews.ProductReviewsService
		repo    *mockProductReviewRepository
		ctx     context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		repo = newMockProductReviewRepository()
		service = reviews.NewProductReviewService(repo, zap.NewNop())
	})

	Context("AddReview", func() {
		It("should add a new review successfully", func() {
			// Arrange
			productID := 1
			userID := 2
			rating := 5
			review := "Great product!"

			// Добавляем тестовый продукт
			repo.products[productID] = entity.Product{
				ID:          productID,
				Name:        "Test Product",
				Description: "Test Description",
				CategoryID:  1,
			}

			// Act
			id, err := service.AddReview(ctx, productID, userID, rating, review)

			// Assert
			Expect(err).NotTo(HaveOccurred())
			Expect(id).To(Equal(productID))
			reviews, err := repo.GetReviewsByProductID(ctx, productID)
			Expect(err).NotTo(HaveOccurred())
			Expect(reviews).To(HaveLen(1))
			Expect(reviews[0].ProductID).To(Equal(productID))
			Expect(reviews[0].UserID).To(Equal(userID))
			Expect(reviews[0].Rating).To(Equal(rating))
			Expect(reviews[0].Review).To(Equal(review))
		})

		It("should return error for non-existent product", func() {
			// Act
			_, err := service.AddReview(ctx, 999, 1, 5, "Review")

			// Assert
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("product not found"))
		})
	})

	Context("GetReviewsByProductID", func() {
		It("should return reviews for existing product", func() {
			// Arrange
			productID := 1
			userID := 2
			rating := 5
			review := "Great product!"

			repo.products[productID] = entity.Product{
				ID:          productID,
				Name:        "Test Product",
				Description: "Test Description",
				CategoryID:  1,
			}
			repo.reviews[productID] = []entity.ProductReview{
				{
					ProductID: productID,
					UserID:    userID,
					Rating:    rating,
					Review:    review,
				},
			}

			// Act
			reviews, err := service.GetReviewsByProductID(ctx, productID)

			// Assert
			Expect(err).NotTo(HaveOccurred())
			Expect(reviews).To(HaveLen(1))
			Expect(reviews[0].ProductID).To(Equal(productID))
			Expect(reviews[0].UserID).To(Equal(userID))
			Expect(reviews[0].Rating).To(Equal(rating))
			Expect(reviews[0].Review).To(Equal(review))
		})

		It("should return error for non-existent product", func() {
			// Act
			_, err := service.GetReviewsByProductID(ctx, 999)

			// Assert
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("product not found"))
		})
	})
})

func TestProductReviewService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ProductReview Service Suite")
}
