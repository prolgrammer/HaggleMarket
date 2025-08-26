package entity

import "time"

type ProductReview struct {
	ID        int       `json:"id" db:"id"`
	ProductID int       `json:"product_id" db:"productid"`
	UserID    int       `json:"user_id" db:"userid"`
	Rating    int       `json:"rating" db:"rating"`
	Review    string    `json:"review" db:"review"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type SellerReview struct {
	ID        int       `json:"id" db:"id"`
	SellerID  int       `json:"seller_id" db:"seller_id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Rating    int       `json:"rating" db:"rating"`
	Review    string    `json:"review" db:"review"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
