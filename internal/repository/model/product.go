package model

import (
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
)

type Product struct {
	ID          int    `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	CategoryID  int    `db:"category_id"`
}

type ProductFilter struct {
	CategoryID *int    `form:"category_id"`
	ShopID     *int    `form:"shop_id"`
	MinPrice   *int    `form:"min_price"`
	MaxPrice   *int    `form:"max_price"`
	Name       *string `form:"name"`
	Attributes map[string]string
}

func ConvertProductToEntity(p Product) entity.Product {
	return entity.Product{
		ID:           int(p.ID),
		Name:         p.Name,
		Description:  p.Description,
		CategoryID:   p.CategoryID,
		MinimalPrice: 0,
		MaximalPrice: 0,
		Attributes:   make(map[string]interface{}),
	}
}
