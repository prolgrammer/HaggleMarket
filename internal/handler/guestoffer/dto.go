package guestoffer

// GuestPostOfferReq DTO for the guest offer creation request
type GuestPostOfferReq struct {
	ProductID  uint    `json:"product_id" binding:"required"`
	StoreID    uint    `json:"store_id" binding:"required"`
	Price      float64 `json:"offer_price" binding:"required"`
	Currency   string  `json:"currency" binding:"required"`
	GuestName  string  `json:"guest_name" binding:"required"`
	GuestEmail string  `json:"guest_email" binding:"required,email"`
	GuestPhone string  `json:"guest_phone" binding:"required"`
}
