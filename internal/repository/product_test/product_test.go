package repository_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/EM-Stawberry/Stawberry/internal/repository"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ProductRepository", func() {
	var (
		db     *sql.DB
		sqlxDB *sqlx.DB
		mock   sqlmock.Sqlmock
		repo   *repository.ProductRepository
		ctx    context.Context
	)

	BeforeEach(func() {
		var err error
		db, mock, err = sqlmock.New()
		Expect(err).ToNot(HaveOccurred())

		sqlxDB = sqlx.NewDb(db, "sqlmock")
		repo = &repository.ProductRepository{Db: sqlxDB}
		ctx = context.Background()
	})

	Describe("GetProductAttributesByID", func() {
		It("should return attributes when product exists", func() {
			// Arrange
			productID := "1"
			attributesMap := map[string]interface{}{
				"color": "blue",
				"size":  "XL",
			}
			attributesJSON, _ := json.Marshal(attributesMap)

			rows := sqlmock.NewRows([]string{"attributes"}).AddRow(attributesJSON)

			mock.ExpectQuery(`SELECT attributes FROM product_attributes WHERE product_id = \$1`).
				WithArgs(productID).
				WillReturnRows(rows)

			// Act
			attrs, err := repo.GetAttributesByID(ctx, productID)

			// Assert
			Expect(err).ToNot(HaveOccurred())
			Expect(attrs).To(HaveKeyWithValue("color", "blue"))
			Expect(attrs).To(HaveKeyWithValue("size", "XL"))
			Expect(mock.ExpectationsWereMet()).To(Succeed())
		})

		It("should return nil and no error when no rows", func() {
			productID := "missing-id"

			mock.ExpectQuery(`SELECT attributes FROM product_attributes WHERE product_id = \$1`).
				WithArgs(productID).
				WillReturnError(sql.ErrNoRows)

			attrs, err := repo.GetAttributesByID(ctx, productID)

			Expect(err).To(BeNil())
			Expect(attrs).To(BeNil())
		})

		It("should return error when query fails", func() {
			productID := "error-id"

			mock.ExpectQuery(`SELECT attributes FROM product_attributes WHERE product_id = \$1`).
				WithArgs(productID).
				WillReturnError(sql.ErrConnDone)

			attrs, err := repo.GetAttributesByID(ctx, productID)

			Expect(err).To(HaveOccurred())
			Expect(attrs).To(BeNil())
		})
	})

	Describe("GetPriceRangeByProductID", func() {
		It("should return min and max price when data exists", func() {
			productID := 123
			rows := sqlmock.NewRows([]string{"min", "max"}).
				AddRow(1005, 9909)

			mock.ExpectQuery(`SELECT CAST\(MIN\(price\) \* 100 AS BIGINT\) AS min,` +
				` CAST\(MAX\(price\) \* 100 AS BIGINT\) AS max FROM shop_inventory WHERE product_id = \$1`).
				WithArgs(productID).
				WillReturnRows(rows)

			min, max, err := repo.GetPriceRangeByProductID(ctx, productID)

			Expect(err).ToNot(HaveOccurred())
			Expect(min).To(Equal(1005))
			Expect(max).To(Equal(9909))
			Expect(mock.ExpectationsWereMet()).To(Succeed())
		})

		It("should return zeros when no valid min/max prices", func() {
			productID := 123
			rows := sqlmock.NewRows([]string{"min", "max"}).
				AddRow(nil, nil)

			mock.ExpectQuery(`SELECT CAST\(MIN\(price\) \* 100 AS BIGINT\) AS min,` +
				` CAST\(MAX\(price\) \* 100 AS BIGINT\) AS max FROM shop_inventory WHERE product_id = \$1`).
				WithArgs(productID).
				WillReturnRows(rows)

			min, max, err := repo.GetPriceRangeByProductID(ctx, productID)

			Expect(err).ToNot(HaveOccurred())
			Expect(min).To(Equal(0))
			Expect(max).To(Equal(0))
		})

		It("should return error on query failure", func() {
			productID := 123

			mock.ExpectQuery(`SELECT CAST\(MIN\(price\) \* 100 AS BIGINT\) AS min,` +
				` CAST\(MAX\(price\) \* 100 AS BIGINT\) AS max FROM shop_inventory WHERE product_id = \$1`).
				WithArgs(productID).
				WillReturnError(errors.New("some db error"))

			min, max, err := repo.GetPriceRangeByProductID(ctx, productID)

			Expect(err).To(HaveOccurred())
			Expect(min).To(Equal(0))
			Expect(max).To(Equal(0))
		})
	})

	Describe("GetAverageRatingByProductID", func() {

		It("should return average rating and count when data exists", func() {
			productID := 123
			rows := sqlmock.NewRows([]string{"average", "count"}).
				AddRow(4.5, 10)

			mock.ExpectQuery(`SELECT AVG\(rating\) average, COUNT\(\*\) count FROM product_reviews WHERE product_id = \$1`).
				WithArgs(productID).
				WillReturnRows(rows)

			avg, count, err := repo.GetAverageRatingByProductID(ctx, productID)

			Expect(err).ToNot(HaveOccurred())
			Expect(avg).To(Equal(4.5))
			Expect(count).To(Equal(10))
			Expect(mock.ExpectationsWereMet()).To(Succeed())
		})

		It("should return zeros when no reviews (NULL values)", func() {
			productID := 123
			rows := sqlmock.NewRows([]string{"average", "count"}).
				AddRow(nil, nil)

			mock.ExpectQuery(`SELECT AVG\(rating\) average, COUNT\(\*\) count FROM product_reviews WHERE product_id = \$1`).
				WithArgs(productID).
				WillReturnRows(rows)

			avg, count, err := repo.GetAverageRatingByProductID(ctx, productID)

			Expect(err).ToNot(HaveOccurred())
			Expect(avg).To(Equal(0.0))
			Expect(count).To(Equal(0))
			Expect(mock.ExpectationsWereMet()).To(Succeed())
		})

		It("should return error when query fails", func() {
			productID := 123
			mock.ExpectQuery(`SELECT AVG\(rating\) average, COUNT\(\*\) count FROM product_reviews WHERE product_id = \$1`).
				WithArgs(productID).
				WillReturnError(sql.ErrConnDone)

			avg, count, err := repo.GetAverageRatingByProductID(ctx, productID)

			Expect(err).To(HaveOccurred())
			Expect(avg).To(Equal(0.0))
			Expect(count).To(Equal(0))
		})
	})
})
