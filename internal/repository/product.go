package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"

	"database/sql"

	"github.com/jmoiron/sqlx"

	sq "github.com/Masterminds/squirrel"

	"github.com/EM-Stawberry/Stawberry/internal/repository/model"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
)

type ProductRepository struct {
	Db *sqlx.DB
}

func NewProductRepository(Db *sqlx.DB) *ProductRepository {
	return &ProductRepository{Db: Db}
}

// GetProductByID позволяет получить продукт по его ID
func (r *ProductRepository) GetProductByID(
	ctx context.Context,
	id string,
) (entity.Product, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	queryBuilder := psql.
		Select("id", "name", "description", "category_id").
		From("products").
		Where(sq.Eq{"id": id})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return entity.Product{}, apperror.New(apperror.DatabaseError, "failed to build SQL query", err)
	}

	var productModel model.Product
	if err := r.Db.GetContext(ctx, &productModel, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Product{}, apperror.ErrProductNotFound
		}
		return entity.Product{}, apperror.New(apperror.DatabaseError, "failed to fetch product", err)
	}

	return model.ConvertProductToEntity(productModel), nil
}

func (r *ProductRepository) GetFilteredProducts(
	ctx context.Context,
	filter model.ProductFilter,
	limit, offset int) ([]entity.Product, error) {
	args := []interface{}{}
	categoryID := 0
	if filter.CategoryID != nil {
		categoryID = *filter.CategoryID
	}
	args = append(args, categoryID)

	selectBuilder := sq.StatementBuilder.
		PlaceholderFormat(sq.Dollar).
		Select("DISTINCT ON (p.id) p.*").
		From("products p").
		LeftJoin("shop_inventory si ON si.product_id = p.id").
		OrderBy("p.id")

	if filter.CategoryID != nil {
		selectBuilder = selectBuilder.Where("p.category_id IN (SELECT id FROM subcategories)")
	}

	if filter.MinPrice != nil {
		selectBuilder = selectBuilder.Where(
			sq.Expr("CAST(si.price * 100 AS BIGINT) >= ?", *filter.MinPrice),
		)
	}
	if filter.MaxPrice != nil {
		selectBuilder = selectBuilder.Where(
			sq.Expr("CAST(si.price * 100 AS BIGINT) <= ?", *filter.MaxPrice),
		)
	}
	if filter.ShopID != nil {
		selectBuilder = selectBuilder.Where(sq.Eq{"si.shop_id": *filter.ShopID})
	}
	if filter.Name != nil {
		selectBuilder = selectBuilder.Where(sq.ILike{"p.name": "%" + *filter.Name + "%"})
	}

	if len(filter.Attributes) > 0 {
		selectBuilder = selectBuilder.Join("product_attributes pa ON p.id = pa.product_id")
		for attr, val := range filter.Attributes {
			condition := fmt.Sprintf("pa.attributes ->> '%s' = ?", attr)
			strVal := fmt.Sprintf("%v", val)
			selectBuilder = selectBuilder.Where(sq.Expr(condition, strVal))
		}
	}

	selectSQL, queryArgs, err := selectBuilder.ToSql()
	if err != nil {
		fmt.Println("Ошибка в билде запроса")
		return nil, apperror.New(apperror.DatabaseError, "failed to build SQL", err)
	}

	selectSQL = shiftPlaceholders(selectSQL, 1)

	recursivePart := `
		WITH RECURSIVE subcategories AS (
			SELECT id FROM categories WHERE id = $1
			UNION ALL
			SELECT c.id FROM categories c
			JOIN subcategories sc ON c.parent_id = sc.id
		)
	`

	limitStr := fmt.Sprintf(" LIMIT %d ", limit)
	offsetStr := fmt.Sprintf("OFFSET %d", offset)

	fullSQL := recursivePart + selectSQL + limitStr + offsetStr

	args = append(args, queryArgs...)

	var productModels []model.Product
	err = r.Db.SelectContext(ctx, &productModels, fullSQL, args...)
	if err != nil {
		fmt.Println(fullSQL)
		fmt.Println(args...)
		return nil, apperror.New(apperror.DatabaseError, "failed to fetch filtered products", err)
	}
	products := make([]entity.Product, len(productModels))
	for i, pm := range productModels {
		products[i] = model.ConvertProductToEntity(pm)
	}

	return products, nil
}

func (r *ProductRepository) GetFilteredProductsCount(ctx context.Context,
	filter model.ProductFilter) (int, error) {
	args := []interface{}{}
	categoryID := 0
	if filter.CategoryID != nil {
		categoryID = *filter.CategoryID
	}
	args = append(args, categoryID)

	selectBuilder := sq.StatementBuilder.
		PlaceholderFormat(sq.Dollar).
		Select("COUNT(DISTINCT p.id)").
		From("products p").
		LeftJoin("shop_inventory si ON si.product_id = p.id")

	if filter.CategoryID != nil {
		selectBuilder = selectBuilder.Where("p.category_id IN (SELECT id FROM subcategories)")
	}

	if filter.MinPrice != nil {
		selectBuilder = selectBuilder.Where(
			sq.Expr("CAST(si.price * 100 AS BIGINT) >= ?", *filter.MinPrice),
		)
	}
	if filter.MaxPrice != nil {
		selectBuilder = selectBuilder.Where(
			sq.Expr("CAST(si.price * 100 AS BIGINT) <= ?", *filter.MaxPrice),
		)
	}
	if filter.ShopID != nil {
		selectBuilder = selectBuilder.Where(sq.Eq{"si.shop_id": *filter.ShopID})
	}
	if filter.Name != nil {
		selectBuilder = selectBuilder.Where(sq.ILike{"p.name": "%" + *filter.Name + "%"})
	}

	if len(filter.Attributes) > 0 {
		selectBuilder = selectBuilder.Join("product_attributes pa ON p.id = pa.product_id")
		for attr, val := range filter.Attributes {
			condition := fmt.Sprintf("pa.attributes ->> '%s' = ?", attr)
			strVal := fmt.Sprintf("%v", val)
			selectBuilder = selectBuilder.Where(sq.Expr(condition, strVal))
		}
	}

	selectSQL, queryArgs, err := selectBuilder.ToSql()
	if err != nil {
		fmt.Println("Ошибка в билде запроса")
		return 0, apperror.New(apperror.DatabaseError, "failed to build SQL", err)
	}

	selectSQL = shiftPlaceholders(selectSQL, 1)

	recursivePart := `
		WITH RECURSIVE subcategories AS (
			SELECT id FROM categories WHERE id = $1
			UNION ALL
			SELECT c.id FROM categories c
			JOIN subcategories sc ON c.parent_id = sc.id
		)
	`

	fullSQL := recursivePart + selectSQL

	args = append(args, queryArgs...)

	var count int
	err = r.Db.GetContext(ctx, &count, fullSQL, args...)
	if err != nil {
		fmt.Println("Ошибка при запросе количества")
		fmt.Println(fullSQL)
		fmt.Println(args...)
		return 0, apperror.New(apperror.DatabaseError, "failed to fetch filtered products", err)
	}
	return count, nil
}

// GetAttributesByID получает аттрибуты продукта по его ID
func (r *ProductRepository) GetAttributesByID(ctx context.Context,
	productID string) (map[string]interface{}, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	queryBuilder := psql.
		Select("attributes").
		From("product_attributes").
		Where(sq.Eq{"product_id": productID}).
		Limit(1)

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, apperror.New(apperror.DatabaseError, "failed to build query", err)
	}

	var attributesJSONb []byte

	err = r.Db.GetContext(ctx, &attributesJSONb, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, apperror.New(apperror.DatabaseError, "failed to fetch product attributes", err)
	}

	var attributes map[string]interface{}
	if err := json.Unmarshal(attributesJSONb, &attributes); err != nil {
		return nil, apperror.New(apperror.DatabaseError, "failed to unmarshal product attributes", err)
	}
	return attributes, nil
}

// GetPriceRangeByProductID получает минимальную и максимальную цену на продукт
func (r *ProductRepository) GetPriceRangeByProductID(ctx context.Context,
	productID int) (int, int, error) {
	var priceRange struct {
		Min sql.NullInt64 `db:"min"`
		Max sql.NullInt64 `db:"max"`
	}

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	queryBuilder := psql.
		Select(
			"CAST(MIN(price) * 100 AS BIGINT) AS min",
			"CAST(MAX(price) * 100 AS BIGINT) AS max",
		).
		From("shop_inventory").
		Where(sq.Eq{"product_id": productID})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return 0, 0, apperror.New(apperror.DatabaseError, "failed to build price range query", err)
	}

	err = r.Db.GetContext(ctx, &priceRange, query, args...)
	if err != nil {
		return 0, 0, apperror.New(apperror.DatabaseError, "failed to calculate min/max price", err)
	}

	minPrice := 0
	maxPrice := 0
	if priceRange.Min.Valid {
		minPrice = int(priceRange.Min.Int64)
	}
	if priceRange.Max.Valid {
		maxPrice = int(priceRange.Max.Int64)
	}

	return minPrice, maxPrice, nil
}

// GetAverageRatingByProductID получает средний рейтинг и количество отзывов на продукт
func (r *ProductRepository) GetAverageRatingByProductID(ctx context.Context,
	productID int) (float64, int, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	queryBuilder := psql.
		Select("AVG(rating) average", "COUNT(*) count").
		From("product_reviews").
		Where(sq.Eq{"product_id": productID})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return 0, 0, apperror.New(apperror.DatabaseError, "failed to build query", err)
	}

	var reviewStats struct {
		Average sql.NullFloat64 `db:"average"`
		Count   sql.NullInt64   `db:"count"`
	}

	if err := r.Db.GetContext(ctx, &reviewStats, query, args...); err != nil {
		return 0, 0, apperror.New(apperror.DatabaseError, "failed to calculate average rating/count of reviews", err)
	}

	avg := 0.0
	count := 0
	if reviewStats.Average.Valid {
		avg = reviewStats.Average.Float64
	}
	if reviewStats.Count.Valid {
		count = int(reviewStats.Count.Int64)
	}

	return avg, count, nil
}

func shiftPlaceholders(sql string, offset int) string {
	re := regexp.MustCompile(`\$(\d+)`)
	return re.ReplaceAllStringFunc(sql, func(match string) string {
		num, _ := strconv.Atoi(match[1:])
		return "$" + strconv.Itoa(num+offset)
	})
}
