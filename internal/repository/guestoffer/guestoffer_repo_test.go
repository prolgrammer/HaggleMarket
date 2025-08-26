package guestoffer_test

import (
	"context"
	"database/sql"
	"errors"

	go_sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	guestofferrepo "github.com/EM-Stawberry/Stawberry/internal/repository/guestoffer"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GuestOfferRepository", func() {
	var (
		db      *sql.DB
		mock    go_sqlmock.Sqlmock
		repo    guestofferrepo.StoreInfoGetter
		ctx     context.Context
		storeID uint
	)

	BeforeEach(func() {
		var err error
		db, mock, err = go_sqlmock.New()
		Expect(err).NotTo(HaveOccurred())

		sqlxDB := sqlx.NewDb(db, "sqlmock")

		repo = guestofferrepo.NewRepository(sqlxDB)
		ctx = context.Background()
		storeID = 123
	})

	AfterEach(func() {
		Expect(mock.ExpectationsWereMet()).To(Succeed())
	})

	Describe("GetStoreOwnerEmailByStoreID", func() {
		Context("when the store exists", func() {
			It("should return the store owner's email", func() {
				expectedEmail := "owner@example.com"
				rows := go_sqlmock.NewRows([]string{"email"}).AddRow(expectedEmail)

				mock.ExpectQuery("SELECT users.email FROM users JOIN shops ON users.id = shops.user_id WHERE shops.id = \\$1").
					WithArgs(storeID).
					WillReturnRows(rows)

				email, err := repo.GetStoreOwnerEmailByStoreID(ctx, storeID)

				Expect(err).NotTo(HaveOccurred())
				Expect(email).To(Equal(expectedEmail))
			})
		})

		Context("when the store does not exist", func() {
			It("should return a GuestOfferStoreNotFound error", func() {
				mock.ExpectQuery("SELECT users.email FROM users JOIN shops ON users.id = shops.user_id WHERE shops.id = \\$1").
					WithArgs(storeID).
					WillReturnError(sql.ErrNoRows)

				_, err := repo.GetStoreOwnerEmailByStoreID(ctx, storeID)

				Expect(err).To(HaveOccurred())
				Expect(err).To(BeAssignableToTypeOf(&apperror.GuestOfferError{}))
				Expect(err.(*apperror.GuestOfferError).Code).To(Equal(apperror.GuestOfferStoreNotFound))
			})
		})

		Context("when a database error occurs", func() {
			It("should return a GuestOfferDatabaseError", func() {
				dbError := errors.New("database connection error")
				mock.ExpectQuery("SELECT users.email FROM users JOIN shops ON users.id = shops.user_id WHERE shops.id = \\$1").
					WithArgs(storeID).
					WillReturnError(dbError)

				_, err := repo.GetStoreOwnerEmailByStoreID(ctx, storeID)

				Expect(err).To(HaveOccurred())
				Expect(err).To(BeAssignableToTypeOf(&apperror.GuestOfferError{}))
				guestOfferErr, ok := err.(*apperror.GuestOfferError)
				Expect(ok).To(BeTrue())
				Expect(guestOfferErr.Code).To(Equal(apperror.GuestOfferDatabaseError))
				Expect(guestOfferErr.Message).To(Equal("failed to get store owner email"))
			})
		})
	})
})
