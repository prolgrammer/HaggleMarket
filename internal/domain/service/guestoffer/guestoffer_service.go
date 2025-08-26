package guestoffer

import (
	"context"
	"errors"
	"fmt"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	guestofferrepo "github.com/EM-Stawberry/Stawberry/internal/repository/guestoffer"
	"go.uber.org/zap"
)

// NotificationSender interface for sending guest offer notifications
type NotificationSender interface {
	SendGuestOfferNotification(email string, subject string, body string)
}

// Service describes the interface for the guest offer service
type Service interface {
	ProcessGuestOffer(ctx context.Context, offerData entity.GuestOfferData) error
}

// GuestOfferService implements the Service interface
type GuestOfferService struct {
	storeInfoGetter    guestofferrepo.StoreInfoGetter
	notificationSender NotificationSender
	log                *zap.Logger
}

// NewService creates a new instance of GuestOfferService and returns the Service interface
func NewService(
	storeInfoGetter guestofferrepo.StoreInfoGetter,
	notificationSender NotificationSender,
	log *zap.Logger,
) Service {
	return &GuestOfferService{
		storeInfoGetter:    storeInfoGetter,
		notificationSender: notificationSender,
		log:                log,
	}
}

// ProcessGuestOffer handles the guest offer business logic
func (s *GuestOfferService) ProcessGuestOffer(ctx context.Context, offerData entity.GuestOfferData) error {
	shopOwnerEmail, err := s.storeInfoGetter.GetStoreOwnerEmailByStoreID(ctx, offerData.StoreID)
	if err != nil {
		s.log.Error("Failed to get shop owner email", zap.Error(err), zap.Uint("store_id", offerData.StoreID))

		var guestOfferErr *apperror.GuestOfferError
		if errors.As(err, &guestOfferErr) && guestOfferErr.Code == apperror.GuestOfferStoreNotFound {
			return err
		}

		return fmt.Errorf("failed to get store owner email from repository: %w", err)
	}

	emailSubject := "New guest offer received"
	emailBody := fmt.Sprintf(
		"A new guest offer has been received:\n\n"+
			"Product ID: %d\n"+
			"Store ID: %d\n"+
			"Proposed Price: %.2f %s\n"+
			"Guest Name: %s\n"+
			"Guest Email: %s\n"+
			"Guest Phone: %s",
		offerData.ProductID,
		offerData.StoreID,
		offerData.Price,
		offerData.Currency,
		offerData.GuestName,
		offerData.GuestEmail,
		offerData.GuestPhone,
	)

	s.notificationSender.SendGuestOfferNotification(shopOwnerEmail, emailSubject, emailBody)

	return nil
}
