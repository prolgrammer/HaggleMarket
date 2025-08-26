package dto

type AddReviewDTO struct {
	Rating int    `json:"rating" binding:"required,min=1,max=5"`
	Review string `json:"review" binding:"required"`
}
