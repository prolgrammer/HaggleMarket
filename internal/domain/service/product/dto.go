package product

type Product struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CategoryID  uint   `json:"category_id"`
}

type UpdateProduct struct {
	StoreID     *uint    `json:"store_id,omitempty"`
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Price       *float64 `json:"price,omitempty"`
	Category    *string  `json:"category,omitempty"`
	InStock     *bool    `json:"in_stock,omitempty"`
}
