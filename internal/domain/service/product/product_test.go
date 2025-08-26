package product

import (
	"context"
	"errors"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/product/mocks"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("EnrichProducts", func() {
	var (
		mockCtrl *gomock.Controller
		mockRepo *mocks.MockRepository
		svc      *Service
		ctx      context.Context
		product  entity.Product
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockRepo = mocks.NewMockRepository(mockCtrl)
		svc = &Service{ProductRepository: mockRepo}
		ctx = context.Background()

		product = entity.Product{ID: 1, Name: "Product A"}

	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("successfully enriches products", func() {
		mockRepo.EXPECT().GetPriceRangeByProductID(ctx, 1).Return(1000, 2000, nil)
		mockRepo.EXPECT().GetAverageRatingByProductID(ctx, 1).Return(4.5, 10, nil)

		result, err := svc.enrichProducts(ctx, product)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.MinimalPrice).To(Equal(1000))
		Expect(result.MaximalPrice).To(Equal(2000))
		Expect(result.AverageRating).To(Equal(4.5))
		Expect(result.CountReviews).To(Equal(10))
	})

	It("returns error if GetPriceRangeByProductID fails", func() {
		mockRepo.EXPECT().GetPriceRangeByProductID(ctx, 1).Return(0, 0, errors.New("db error"))

		_, err := svc.enrichProducts(ctx, product)
		Expect(err).To(MatchError("db error"))
	})

	It("returns error if GetAverageRatingByProductID fails", func() {
		mockRepo.EXPECT().GetPriceRangeByProductID(ctx, 1).Return(1000, 2000, nil)
		mockRepo.EXPECT().GetAverageRatingByProductID(ctx, 1).Return(0.0, 0, errors.New("rating error"))

		_, err := svc.enrichProducts(ctx, product)
		Expect(err).To(MatchError("rating error"))
	})
})
