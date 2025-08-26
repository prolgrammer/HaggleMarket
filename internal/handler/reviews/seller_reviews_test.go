package reviews_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/EM-Stawberry/Stawberry/internal/handler/reviews"
	"github.com/EM-Stawberry/Stawberry/internal/handler/reviews/dto"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

type mockSellerReviewsService struct {
	reviews map[int][]entity.SellerReview
}

func newMockSellerReviewsService() *mockSellerReviewsService {
	return &mockSellerReviewsService{
		reviews: make(map[int][]entity.SellerReview),
	}
}

func (m *mockSellerReviewsService) AddReview(
	_ context.Context, sellerID int, userID int, rating int, review string,
) (int, error) {
	if sellerID == 999 {
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

func (m *mockSellerReviewsService) GetReviewsByID(
	_ context.Context, sellerID int,
) ([]entity.SellerReview, error) {
	if sellerID == 999 {
		return nil, apperror.NewReviewError(apperror.NotFound, "seller not found")
	}
	return m.reviews[sellerID], nil
}

var _ = Describe("SellerReviewsHandler", func() {
	var (
		handler *reviews.SellerReviewsHandler
		service *mockSellerReviewsService
		router  *gin.Engine
	)

	BeforeEach(func() {
		service = newMockSellerReviewsService()
		handler = reviews.NewSellerReviewsHandler(service, zap.NewNop())
		gin.SetMode(gin.ReleaseMode)
		router = gin.New()

		router.Use(func(c *gin.Context) {
			c.Set("userID", 1)
			c.Next()
		})
		router.POST("/api/sellers/:id/reviews", handler.AddReview)
		router.GET("/api/sellers/:id/reviews", handler.GetReviews)
	})

	Context("AddReview", func() {
		It("should add a new review successfully", func() {
			// Arrange
			review := dto.AddReviewDTO{
				Rating: 5,
				Review: "Great seller!",
			}
			body, _ := json.Marshal(review)
			req, _ := http.NewRequest("POST", "/api/sellers/1/reviews", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Act
			router.ServeHTTP(w, req)

			// Assert
			Expect(w.Code).To(Equal(http.StatusCreated))
			var response map[string]string
			err := json.Unmarshal(w.Body.Bytes(), &response)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(response["message"]).To(Equal("review added successfully"))
		})

		It("should return 404 for non-existent seller", func() {
			// Arrange
			review := dto.AddReviewDTO{
				Rating: 5,
				Review: "Great seller!",
			}
			body, _ := json.Marshal(review)
			req, _ := http.NewRequest("POST", "/api/sellers/999/reviews", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Act
			router.ServeHTTP(w, req)

			// Assert
			Expect(w.Code).To(Equal(http.StatusNotFound))
		})

		It("should return 400 for invalid input", func() {
			// Arrange
			req, _ := http.NewRequest("POST", "/api/sellers/1/reviews", bytes.NewBuffer([]byte("invalid json")))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Act
			router.ServeHTTP(w, req)

			// Assert
			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("GetReviews", func() {
		It("should return reviews for existing seller", func() {
			// Arrange
			service.reviews[1] = []entity.SellerReview{
				{
					SellerID: 1,
					UserID:   1,
					Rating:   5,
					Review:   "Great seller!",
				},
			}
			req, _ := http.NewRequest("GET", "/api/sellers/1/reviews", nil)
			w := httptest.NewRecorder()

			// Act
			router.ServeHTTP(w, req)

			// Assert
			Expect(w.Code).To(Equal(http.StatusOK))
			var response []entity.SellerReview
			err := json.Unmarshal(w.Body.Bytes(), &response)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(response).To(HaveLen(1))
			Expect(response[0].SellerID).To(Equal(1))
			Expect(response[0].UserID).To(Equal(1))
			Expect(response[0].Rating).To(Equal(5))
			Expect(response[0].Review).To(Equal("Great seller!"))
		})

		It("should return 404 for non-existent seller", func() {
			// Arrange
			req, _ := http.NewRequest("GET", "/api/sellers/999/reviews", nil)
			w := httptest.NewRecorder()

			// Act
			router.ServeHTTP(w, req)

			// Assert
			Expect(w.Code).To(Equal(http.StatusNotFound))
		})

		It("should return 400 for invalid seller ID", func() {
			// Arrange
			req, _ := http.NewRequest("GET", "/api/sellers/invalid/reviews", nil)
			w := httptest.NewRecorder()

			// Act
			router.ServeHTTP(w, req)

			// Assert
			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})
	})
})
