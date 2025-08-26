package offer

import (
	"context"
	"time"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/EM-Stawberry/Stawberry/pkg/email"
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
	mailer          email.MailerService
}

func NewService(offerRepository Repository, mailer email.MailerService) *Service {
	return &Service{offerRepository: offerRepository, mailer: mailer}
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

	os.mailer.Registered(user.Name, user.Email)

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

	return offerResp, err
}

func (os *Service) DeleteOffer(
	ctx context.Context,
	offerID uint,
) (entity.Offer, error) {
	return os.offerRepository.DeleteOffer(ctx, offerID)
}
