package guestoffer

import (
	"errors"
	"net/http"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	guestofferservice "github.com/EM-Stawberry/Stawberry/internal/domain/service/guestoffer"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler handles requests related to guest offers
type Handler struct {
	service guestofferservice.Service
	log     *zap.Logger
}

// NewHandler creates a new Handler for guest offers
func NewHandler(service guestofferservice.Service, log *zap.Logger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

// PostGuestOffer handles the guest offer creation request
// @Summary Send a guest offer
// @Description Allows sending an offer for a product on behalf of a guest
// @Tags guest
// @Accept json
// @Produce json
// @Param offer body GuestPostOfferReq true "Guest offer data"
// @Success 202 {object} map[string]string "Offer accepted and forwarded"
// @Failure 400 {object} apperror.Error "Invalid guest offer data"
// @Failure 404 {object} apperror.Error "Store or product not found"
// @Failure 500 {object} apperror.Error "Internal server error"
// @Router /guest/offers [post]
func (h *Handler) PostGuestOffer(c *gin.Context) {
	var guestOfferReq GuestPostOfferReq
	if err := c.ShouldBindJSON(&guestOfferReq); err != nil {
		_ = c.Error(apperror.NewGuestOfferError(apperror.GuestOfferInvalidData, "invalid guest offer data"))
		return
	}

	offerData := entity.GuestOfferData{
		ProductID:  guestOfferReq.ProductID,
		StoreID:    guestOfferReq.StoreID,
		Price:      guestOfferReq.Price,
		Currency:   guestOfferReq.Currency,
		GuestName:  guestOfferReq.GuestName,
		GuestEmail: guestOfferReq.GuestEmail,
		GuestPhone: guestOfferReq.GuestPhone,
	}

	err := h.service.ProcessGuestOffer(c.Request.Context(), offerData)
	if err != nil {
		h.log.Error("Failed to process guest offer", zap.Error(err))
		var guestOfferErr *apperror.GuestOfferError
		if errors.As(err, &guestOfferErr) {
			switch guestOfferErr.Code {
			case apperror.GuestOfferStoreNotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": guestOfferErr.Error()})
				_ = c.Error(guestOfferErr)
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process guest offer"})
				_ = c.Error(guestOfferErr)
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process guest offer"})
			_ = c.Error(apperror.NewGuestOfferError(apperror.GuestOfferProcessFailed, "failed to process guest offer"))
		}
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Offer accepted and forwarded"})
}
