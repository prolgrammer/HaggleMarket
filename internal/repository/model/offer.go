package model

import (
	"time"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
)

type Offer struct {
	ID        uint      `db:"id"`
	Price     float64   `db:"offer_price"`
	Currency  string    `db:"currency"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	ExpiresAt time.Time `db:"expires_at"`
	ShopID    uint      `db:"shop_id"`
	UserID    uint      `db:"user_id"`
	ProductID uint      `db:"product_id"`
}

type OfferWithCount struct {
	Offer
	TotalCount int `db:"total_count"`
}

func (o *Offer) ConvertToEntity() entity.Offer {
	return entity.Offer{
		ID:        o.ID,
		Price:     o.Price,
		Currency:  o.Currency,
		Status:    o.Status,
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
		ExpiresAt: o.ExpiresAt,
		ShopID:    o.ShopID,
		UserID:    o.UserID,
		ProductID: o.ProductID,
	}
}

func ConvertOfferEntityToModel(offer entity.Offer) Offer {
	return Offer{
		ID:        offer.ID,
		Price:     offer.Price,
		Currency:  offer.Currency,
		Status:    offer.Status,
		CreatedAt: offer.CreatedAt,
		UpdatedAt: offer.UpdatedAt,
		ExpiresAt: offer.ExpiresAt,
		ShopID:    offer.ShopID,
		UserID:    offer.UserID,
		ProductID: offer.ProductID,
	}
}
