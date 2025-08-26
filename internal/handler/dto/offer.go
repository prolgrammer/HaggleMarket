package dto

import (
	"time"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
)

type PostOfferReq struct {
	ProductID uint    `json:"product_id" binding:"required"`
	ShopID    uint    `json:"shop_id" binding:"required"`
	Price     float64 `json:"price" binding:"required,gte=0"`
	Currency  string  `json:"currency" binding:"required,iso4217"`
}

type PostOfferResp struct {
	ID uint `json:"id"`
}

func (po *PostOfferReq) ConvertToEntity() entity.Offer {
	return entity.Offer{
		Price:     po.Price,
		Currency:  po.Currency,
		ShopID:    po.ShopID,
		ProductID: po.ProductID,
	}
}

func ConvertToPostOfferResp(o entity.Offer) PostOfferResp {
	return PostOfferResp{ID: o.ID}
}

type PatchOfferStatusReq struct {
	Status string `json:"status" binding:"required"`
}

type PatchOfferStatusResp struct {
	NewStatus string `json:"new_status"`
}

func (p *PatchOfferStatusReq) ConvertToEntity() entity.Offer {
	return entity.Offer{
		Status: p.Status,
	}
}

func ConvertToPatchOfferStatusResp(o entity.Offer) PatchOfferStatusResp {
	return PatchOfferStatusResp{NewStatus: o.Status}
}

type OfferResp struct {
	ID        uint
	Price     float64
	Currency  string
	Status    string
	CreatedAt time.Time
	ExpiresAt time.Time
	ShopID    uint
	ProductID uint
}

type GetUserOffersResp struct {
	Data []OfferResp `json:"data"`
	Meta struct {
		CurrentPage int `json:"current_page"`
		PerPage     int `json:"per_page"`
		TotalItems  int `json:"total_items"`
		TotalPages  int `json:"total_pages"`
	}
}

func FormUserOffers(ofrs []entity.Offer, page, limit, total, totalPages int) GetUserOffersResp {
	data := make([]OfferResp, 0, len(ofrs))

	for _, ofr := range ofrs {
		data = append(data, OfferResp{
			ID:        ofr.ID,
			Price:     ofr.Price,
			Currency:  ofr.Currency,
			Status:    ofr.Status,
			CreatedAt: ofr.CreatedAt,
			ExpiresAt: ofr.ExpiresAt,
			ShopID:    ofr.ShopID,
			ProductID: ofr.ProductID,
		})
	}

	return GetUserOffersResp{
		Data: data,
		Meta: struct {
			CurrentPage int `json:"current_page"`
			PerPage     int `json:"per_page"`
			TotalItems  int `json:"total_items"`
			TotalPages  int `json:"total_pages"`
		}{
			CurrentPage: page,
			PerPage:     limit,
			TotalItems:  total,
			TotalPages:  totalPages,
		},
	}
}
