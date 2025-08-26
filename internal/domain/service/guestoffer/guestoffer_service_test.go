package guestoffer_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	guestofferservice "github.com/EM-Stawberry/Stawberry/internal/domain/service/guestoffer"
	repomocks "github.com/EM-Stawberry/Stawberry/internal/repository/guestoffer"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gomock "go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

var _ = Describe("GuestOfferService", func() {
	var (
		ctrl                   *gomock.Controller
		mockStoreInfoGetter    *repomocks.MockStoreInfoGetter
		mockNotificationSender *guestofferservice.MockNotificationSender
		service                guestofferservice.Service
		ctx                    context.Context
		log                    *zap.Logger
		offerData              entity.GuestOfferData
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockStoreInfoGetter = repomocks.NewMockStoreInfoGetter(ctrl)
		mockNotificationSender = guestofferservice.NewMockNotificationSender(ctrl)
		log = zaptest.NewLogger(GinkgoT())

		service = guestofferservice.NewService(mockStoreInfoGetter, mockNotificationSender, log)
		ctx = context.Background()

		offerData = entity.GuestOfferData{
			ProductID:  1,
			StoreID:    101,
			Price:      100.50,
			Currency:   "USD",
			GuestName:  "John Doe",
			GuestEmail: "john.doe@example.com",
			GuestPhone: "123-456-7890",
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("ProcessGuestOffer", func() {
		Context("when getting store owner email is successful", func() {
			It("should send a notification and return nil", func() {
				expectedEmail := "owner@example.com"

				mockStoreInfoGetter.EXPECT().
					GetStoreOwnerEmailByStoreID(ctx, offerData.StoreID).
					Return(expectedEmail, nil).
					Times(1)

				expectedSubject := "New guest offer received"
				expectedBody := fmt.Sprintf(
					"A new guest offer has been received:\n\n"+
						"Product ID: %d\n"+
						"Store ID: %d\n"+
						"Proposed Price: %.2f %s\n"+
						"Guest Name: %s\n"+
						"Guest Email: %s\n"+
						"Guest Phone: %s",
					offerData.ProductID,
					offerData.StoreID,
					offerData.Price,
					offerData.Currency,
					offerData.GuestName,
					offerData.GuestEmail,
					offerData.GuestPhone,
				)

				mockNotificationSender.EXPECT().
					SendGuestOfferNotification(expectedEmail, expectedSubject, expectedBody).
					Times(1)

				err := service.ProcessGuestOffer(ctx, offerData)

				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when getting store owner email returns StoreNotFound error", func() {
			It("should return a GuestOfferStoreNotFound error", func() {
				repoError := apperror.NewGuestOfferError(apperror.GuestOfferStoreNotFound, "store not found")
				mockStoreInfoGetter.EXPECT().
					GetStoreOwnerEmailByStoreID(ctx, offerData.StoreID).
					Return("", repoError).
					Times(1)

				mockNotificationSender.EXPECT().SendGuestOfferNotification(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

				err := service.ProcessGuestOffer(ctx, offerData)

				Expect(err).To(HaveOccurred())
				Expect(err).To(BeAssignableToTypeOf(&apperror.GuestOfferError{}))
				guestOfferErr, ok := err.(*apperror.GuestOfferError)
				Expect(ok).To(BeTrue())
				Expect(guestOfferErr.Code).To(Equal(apperror.GuestOfferStoreNotFound))
			})
		})

		Context("when getting store owner email returns a different error", func() {
			It("should return a wrapped error", func() {
				repoError := errors.New("some database error")
				mockStoreInfoGetter.EXPECT().
					GetStoreOwnerEmailByStoreID(ctx, offerData.StoreID).
					Return("", repoError).
					Times(1)

				mockNotificationSender.EXPECT().SendGuestOfferNotification(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

				err := service.ProcessGuestOffer(ctx, offerData)

				Expect(err).To(HaveOccurred())
				Expect(errors.Is(err, repoError)).To(BeTrue())
			})
		})
	})
})
