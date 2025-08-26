package entity

type Product struct {
	ID            int                    `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	CategoryID    int                    `json:"category_id"`
	MinimalPrice  int                    `json:"minimal_price"`
	MaximalPrice  int                    `json:"maximal_price"`
	AverageRating float64                `json:"average_rating"`
	CountReviews  int                    `json:"count_reviews"`
	Attributes    map[string]interface{} `json:"product_attributes"`
}

type NewProduct struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CategoryID  int    `json:"category_id"`
}
