package entity

// GuestOfferData represents the data of a guest offer
type GuestOfferData struct {
	ProductID  uint
	StoreID    uint
	Price      float64
	Currency   string
	GuestName  string
	GuestEmail string
	GuestPhone string
}
