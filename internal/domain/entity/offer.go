package entity

import "time"

type Offer struct {
	ID        uint
	Price     float64
	Currency  string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
	ExpiresAt time.Time
	ShopID    uint
	UserID    uint
	ProductID uint
}
