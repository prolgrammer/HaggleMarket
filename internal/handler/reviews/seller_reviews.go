package reviews

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/EM-Stawberry/Stawberry/internal/handler/reviews/dto"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SellerReviewsService interface {
	AddReview(ctx context.Context, sellerID int, userID int, rating int, review string) (int, error)
	GetReviewsByID(ctx context.Context, sellerID int) ([]entity.SellerReview, error)
}

type SellerReviewsHandler struct {
	srs    SellerReviewsService
	logger *zap.Logger
}

func NewSellerReviewsHandler(srs SellerReviewsService, l *zap.Logger) *SellerReviewsHandler {
	return &SellerReviewsHandler{
		srs:    srs,
		logger: l,
	}
}

// AddReviews godoc
// @Summary Добавление отзыва о продавце
// @Description Добавляет новый отзыв о продавце
// @Tags reviews
// @Accept json
// @Produce json
// @Param id path int true "Seller ID"
// @Param review body dto.AddReviewDTO true "Данные отзыва"
// @Security BearerAuth
// @Success 201 {object} map[string]string "Отзыв успешно добавлен"
// @Failure 400 {object} map[string]string "Некорректный ввод"
// @Failure 401 {object} map[string]string "Неавторизованный доступ"
// @Failure 404 {object} map[string]string "Продавец не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /sellers/{id}/reviews [post]
func (h *SellerReviewsHandler) AddReview(c *gin.Context) {
	const op = "sellerReviewsHandler.AddReviews()"
	log := h.logger.With(zap.String("op", op))

	sellerID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sellerID"})
		log.Warn("Failed to parse productID", zap.Error(err))
		return
	}

	var addReview dto.AddReviewDTO
	if err := c.ShouldBindJSON(&addReview); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		log.Warn("Failed to bind JSON", zap.Error(err))
		return
	}

	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		log.Warn("Failed to get userID from context", zap.Error(err))
		return
	}

	uid, ok := userID.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid userID type"})
		log.Warn("Invalid userID type")
		return
	}

	id, err := h.srs.AddReview(c.Request.Context(), sellerID, uid, addReview.Rating, addReview.Review)
	if err != nil {
		var reviewErr *apperror.ReviewError
		if errors.As(err, &reviewErr) {
			c.JSON(http.StatusNotFound, gin.H{"error": reviewErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add review"})
		log.Warn("Failed to add review", zap.Int("id", id), zap.Error(err))
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "review added successfully"})
}

// GetReviews godoc
// @Summary Получение списка отзывов о продавце
// @Description Получает все отзывы о продавце по его ID
// @Tags reviews
// @Accept json
// @Produce json
// @Param id path int true "Seller ID"
// @Success 200 {array} entity.SellerReview "Список отзывов"
// @Failure 400 {object} map[string]string "Некорректный ID продавца"
// @Failure 404 {object} map[string]string "Продавец не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /sellers/{id}/reviews [get]
func (h *SellerReviewsHandler) GetReviews(c *gin.Context) {
	const op = "sellerReviewsHandler.GetReviews()"
	log := h.logger.With(zap.String("op", op))

	sellerID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sellerID"})
		log.Warn("Failed to parse sellerID", zap.Error(err))
		return
	}

	reviews, err := h.srs.GetReviewsByID(c.Request.Context(), sellerID)
	if err != nil {
		var reviewErr *apperror.ReviewError
		if errors.As(err, &reviewErr) {
			c.JSON(http.StatusNotFound, gin.H{"error": reviewErr.Message})
			return
		}

		if err.Error() == "sql: no rows in result set" {
			c.JSON(http.StatusNotFound, gin.H{"error": "seller not found"})
			log.Warn("Seller not found", zap.Int("sellerID", sellerID))
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch reviews"})
		log.Warn("Failed to fetch reviews", zap.Error(err))
		return
	}

	c.JSON(http.StatusOK, reviews)
}
