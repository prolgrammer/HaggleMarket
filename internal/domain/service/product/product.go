package product

import (
	"context"
	"fmt"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"

	"github.com/EM-Stawberry/Stawberry/internal/repository/model"
)

type Repository interface {
	GetFilteredProducts(ctx context.Context, filter model.ProductFilter, limit, offset int) ([]entity.Product, error)
	GetFilteredProductsCount(ctx context.Context, filter model.ProductFilter) (int, error)
	GetProductByID(ctx context.Context, id string) (entity.Product, error)
	GetAttributesByID(ctx context.Context, productID string) (map[string]interface{}, error)
	GetPriceRangeByProductID(ctx context.Context, productID int) (int, int, error)
	GetAverageRatingByProductID(ctx context.Context, productID int) (float64, int, error)
}

type Service struct {
	ProductRepository Repository
}

func NewService(productRepo Repository) *Service {
	return &Service{ProductRepository: productRepo}
}

// GetProductByID получает продукт по его ID
func (ps *Service) GetProductByID(
	ctx context.Context,
	id string,
) (entity.Product, error) {
	product, err := ps.ProductRepository.GetProductByID(ctx, id)
	if err != nil {
		return entity.Product{}, err
	}
	attrs, err := ps.ProductRepository.GetAttributesByID(ctx, id)
	if err != nil {
		return entity.Product{}, err
	}
	product.Attributes = attrs

	enrichedProduct, err := ps.enrichProducts(ctx, product)
	if err != nil {
		return entity.Product{}, err
	}

	return enrichedProduct, nil
}

func (ps *Service) GetFilteredProducts(ctx context.Context,
	filter model.ProductFilter,
	limit, offset int) ([]entity.Product, int, error) {
	products, err := ps.ProductRepository.GetFilteredProducts(ctx, filter, limit, offset)
	if err != nil {
		fmt.Println("Ошибка при получении продуктов")
		return nil, 0, err
	}

	count, err := ps.ProductRepository.GetFilteredProductsCount(ctx, filter)
	if err != nil {
		fmt.Println("Ошибка при получении количества")
		return nil, 0, err
	}
	for i := range products {
		products[i], err = ps.enrichProducts(ctx, products[i])
		if err != nil {
			fmt.Println("Ошибка при обогащении продуктов")
			return nil, 0, err
		}
	}

	return products, count, nil
}

// EnrichProducts выполняет обогащение продукта информацией о диапазоне цены, средней оценке и количестве отзывов
func (ps *Service) enrichProducts(
	ctx context.Context,
	product entity.Product,
) (entity.Product, error) {

	minPrice, maxPrice, err := ps.ProductRepository.GetPriceRangeByProductID(ctx, product.ID)
	if err != nil {
		return entity.Product{}, err
	}

	avgRating, countReviews, err := ps.ProductRepository.GetAverageRatingByProductID(ctx, product.ID)
	if err != nil {
		return entity.Product{}, err
	}

	product.MinimalPrice = minPrice
	product.MaximalPrice = maxPrice
	product.AverageRating = avgRating
	product.CountReviews = countReviews

	return product, nil
}
