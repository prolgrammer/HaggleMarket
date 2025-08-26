package reviews_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/EM-Stawberry/Stawberry/internal/handler/middleware"
	"github.com/EM-Stawberry/Stawberry/internal/handler/reviews"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

type mockProductReviewsService struct {
	reviews map[int][]entity.ProductReview
}

func newMockProductReviewsService() *mockProductReviewsService {
	return &mockProductReviewsService{
		reviews: make(map[int][]entity.ProductReview),
	}
}

func (m *mockProductReviewsService) AddReview(
	_ context.Context, productID int, userID int, rating int, review string,
) (int, error) {
	if productID == 999 {
		return 0, apperror.NewReviewError(apperror.NotFound, "product not found")
	}

	reviewEntity := entity.ProductReview{
		ProductID: productID,
		UserID:    userID,
		Rating:    rating,
		Review:    review,
	}
	m.reviews[productID] = append(m.reviews[productID], reviewEntity)
	return productID, nil
}

func (m *mockProductReviewsService) GetReviewsByProductID(
	_ context.Context, productID int,
) ([]entity.ProductReview, error) {
	if productID == 999 {
		return nil, apperror.NewReviewError(apperror.NotFound, "product not found")
	}
	return m.reviews[productID], nil
}

var _ = Describe("ProductReviewsHandler", func() {
	var (
		handler *reviews.ProductReviewsHandler
		service *mockProductReviewsService
		router  *gin.Engine
	)

	BeforeEach(func() {
		service = newMockProductReviewsService()
		handler = reviews.NewProductReviewHandler(service, zap.NewNop())
		gin.SetMode(gin.ReleaseMode)
		router = gin.New()
		router.Use(middleware.Errors())

		router.Use(func(c *gin.Context) {
			c.Set("userID", 1)
			c.Next()
		})
		router.POST("/api/products/:id/reviews", handler.AddReview)
		router.GET("/api/products/:id/reviews", handler.GetReviews)
	})

	Context("AddReview", func() {
		// It("should add a new review successfully", func() {
		// 	// Arrange
		// 	review := dto.AddReviewDTO{
		// 		Rating: 5,
		// 		Review: "Great product!",
		// 	}
		// 	body, _ := json.Marshal(review)
		// 	req, _ := http.NewRequest("POST", "/api/products/1/reviews", bytes.NewBuffer(body))
		// 	req.Header.Set("Content-Type", "application/json")
		// 	w := httptest.NewRecorder()

		// 	// Act
		// 	router.ServeHTTP(w, req)

		// 	// Assert
		// 	Expect(w.Code).To(Equal(http.StatusCreated))
		// 	var response map[string]string
		// 	err := json.Unmarshal(w.Body.Bytes(), &response)
		// 	Expect(err).ShouldNot(HaveOccurred())
		// 	Expect(response["message"]).To(Equal("review added successfully"))
		// })

		// It("should return 404 for non-existent product", func() {
		// 	// Arrange
		// 	review := dto.AddReviewDTO{
		// 		Rating: 5,
		// 		Review: "Great product!",
		// 	}
		// 	body, _ := json.Marshal(review)
		// 	req, _ := http.NewRequest("POST", "/api/products/999/reviews", bytes.NewBuffer(body))
		// 	req.Header.Set("Content-Type", "application/json")
		// 	w := httptest.NewRecorder()

		// 	// Act
		// 	router.ServeHTTP(w, req)

		// 	// Assert
		// 	Expect(w.Code).To(Equal(http.StatusNotFound))
		// })

		It("should return 400 for invalid input", func() {
			// Arrange
			req, _ := http.NewRequest("POST", "/api/products/1/reviews", bytes.NewBuffer([]byte("invalid json")))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Act
			router.ServeHTTP(w, req)

			// Assert
			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("GetReviews", func() {
		It("should return reviews for existing product", func() {
			// Arrange
			service.reviews[1] = []entity.ProductReview{
				{
					ProductID: 1,
					UserID:    1,
					Rating:    5,
					Review:    "Great product!",
				},
			}
			req, _ := http.NewRequest("GET", "/api/products/1/reviews", nil)
			w := httptest.NewRecorder()

			// Act
			router.ServeHTTP(w, req)

			// Assert
			Expect(w.Code).To(Equal(http.StatusOK))
			var response []entity.ProductReview
			err := json.Unmarshal(w.Body.Bytes(), &response)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(response).To(HaveLen(1))
			Expect(response[0].ProductID).To(Equal(1))
			Expect(response[0].UserID).To(Equal(1))
			Expect(response[0].Rating).To(Equal(5))
			Expect(response[0].Review).To(Equal("Great product!"))
		})

		It("should return 404 for non-existent product", func() {
			// Arrange
			req, _ := http.NewRequest("GET", "/api/products/999/reviews", nil)
			w := httptest.NewRecorder()

			// Act
			router.ServeHTTP(w, req)

			// Assert
			Expect(w.Code).To(Equal(http.StatusNotFound))
		})

		It("should return 400 for invalid product ID", func() {
			// Arrange
			req, _ := http.NewRequest("GET", "/api/products/invalid/reviews", nil)
			w := httptest.NewRecorder()

			// Act
			router.ServeHTTP(w, req)

			// Assert
			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})
	})
})

func TestProductReviewsHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ProductReviews Handler Suite")
}
