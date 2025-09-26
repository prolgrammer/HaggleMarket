package offer

import (
	"context"
	"errors"
	"time"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/user"
	"github.com/EM-Stawberry/Stawberry/pkg/email"
	"go.uber.org/zap"
)

type Repository interface {
	InsertOffer(ctx context.Context, offer entity.Offer) (uint, error)
	GetOfferByID(ctx context.Context, offerID uint) (entity.Offer, error)
	SelectUserOffers(ctx context.Context, userID uint, limit, offset int) ([]entity.Offer, int, error)
	UpdateOfferStatus(ctx context.Context, offer entity.Offer, userID uint, isStore bool) (entity.Offer, error)
	DeleteOffer(ctx context.Context, offerID uint) (entity.Offer, error)
}

const (
	statusAccepted  = "accepted"
	statusDeclined  = "declined"
	statusCancelled = "cancelled"
	statusPending   = "pending"

	offerLifetime = 7 * 24 * time.Hour
)

type Service struct {
	offerRepository Repository
	userRepository  user.Repository
	mailer          email.MailerService
}

func NewService(offerRepository Repository, userRepository user.Repository, mailer email.MailerService) *Service {
	return &Service{
		offerRepository: offerRepository,
		userRepository:  userRepository,
		mailer:          mailer,
	}
}

func (os *Service) CreateOffer(
	ctx context.Context,
	offer entity.Offer,
	user entity.User,
) (uint, error) {

	t := time.Now()
	offer.Status = statusPending
	offer.CreatedAt = t
	offer.UpdatedAt = t
	offer.ExpiresAt = t.Add(offerLifetime)

	offerID, err := os.offerRepository.InsertOffer(ctx, offer)
	if err != nil {
		return 0, err
	}

	os.mailer.OfferCreated(offerID, user.Email)
	shopOwnerEmail, err := os.userRepository.GetShopOwnerEmail(ctx, offerID)
	if err != nil {
		if errors.Is(err, apperror.ErrOfferNotFound) {
			zap.L().Warn("offer has no associated shop owner — email cannot be sent",
				zap.Uint("offerID", offerID),
				zap.Error(err),
			)
		} else {
			zap.L().Error("failed to get shop owner email due to database error",
				zap.Uint("offer_id", offerID),
				zap.Error(err),
			)
		}
	} else {
		os.mailer.OfferReceived(offerID, shopOwnerEmail)
	}

	return offerID, nil
}

func (os *Service) GetOffer(
	ctx context.Context,
	offerID uint,
) (entity.Offer, error) {
	return os.offerRepository.GetOfferByID(ctx, offerID)
}

func (os *Service) GetUserOffers(
	ctx context.Context,
	userID uint,
	page,
	limit int,
) ([]entity.Offer, int, error) {
	offset := (page - 1) * limit

	offers, total, err := os.offerRepository.SelectUserOffers(ctx, userID, limit, offset)

	return offers, total, err

}

func (os *Service) UpdateOfferStatus(
	ctx context.Context,
	offer entity.Offer,
	userID uint,
	isStore bool,
) (entity.Offer, error) {
	if isStore {
		validStatusesShop := map[string]struct{}{
			statusAccepted: {},
			statusDeclined: {},
		}
		if _, ok := validStatusesShop[offer.Status]; !ok {
			return entity.Offer{}, apperror.New(apperror.BadRequest, "invalid status field value", nil)
		}

	} else {
		validStatusesBuyer := map[string]struct{}{
			statusCancelled: {},
		}
		if _, ok := validStatusesBuyer[offer.Status]; !ok {
			return entity.Offer{}, apperror.New(apperror.BadRequest, "invalid status field value", nil)
		}
	}

	offerResp, err := os.offerRepository.UpdateOfferStatus(ctx, offer, userID, isStore)
	if err != nil {
		return entity.Offer{}, err
	}

	if isStore {
		buyerEmail, err := os.userRepository.GetBuyerEmail(ctx, offer.ID)
		if err != nil {
			if errors.Is(err, apperror.ErrOfferNotFound) {
				zap.L().Warn("offer has no associated buyer — email cannot be sent",
					zap.Uint("offerID", offer.ID),
					zap.Error(err),
				)
			} else {
				zap.L().Error("failed to get buyer email due to database error",
					zap.Uint("offer_id", offer.ID),
					zap.Error(err),
				)
			}
		} else {
			os.mailer.StatusUpdate(offer.ID, offer.Status, buyerEmail)
		}

	} else {
		shopOwnerEmail, err := os.userRepository.GetShopOwnerEmail(ctx, offer.ID)
		if err != nil {
			if errors.Is(err, apperror.ErrOfferNotFound) {
				zap.L().Warn("offer has no associated shop owner — email cannot be sent",
					zap.Uint("offerID", offer.ID),
					zap.Error(err),
				)
			} else {
				zap.L().Error("failed to get shop owner email due to database error",
					zap.Uint("offer_id", offer.ID),
					zap.Error(err),
				)
			}
		} else {
			os.mailer.StatusUpdate(offer.ID, offer.Status, shopOwnerEmail)
		}

	}
	return offerResp, nil
}

func (os *Service) DeleteOffer(
	ctx context.Context,
	offerID uint,
) (entity.Offer, error) {
	return os.offerRepository.DeleteOffer(ctx, offerID)
}
