package token

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/google/uuid"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

func TestTokenServiceSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Token Service Suite")
}

var _ = Describe("TokenService", func() {
	var (
		ctrl        *gomock.Controller
		repo        *MockRepository
		jwtManager  *MockJWTManager
		service     *Service
		refreshLife time.Duration
		accessLife  time.Duration
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())

		refreshLife = 24 * time.Hour
		accessLife = time.Hour

		repo = NewMockRepository(ctrl)
		jwtManager = NewMockJWTManager(ctrl)
		service = NewService(repo, jwtManager, refreshLife, accessLife)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("NewService", func() {

		It("should return a new Service instance", func() {
			Expect(service).NotTo(BeNil())
		})
	})

	Describe("GenerateTokens", func() {

		Context("when generating tokens", func() {
			When("JWT generation is successful", func() {
				It("should return valid access and refresh tokens", func() {
					fingerprint := "test-fingerprint"
					userID := uint(1)
					mockJWT := "mock.jwt.token"

					jwtManager.EXPECT().
						Generate(userID, accessLife).
						Return(mockJWT, nil).
						Times(1)

					accessToken, refreshToken, err := service.GenerateTokens(context.Background(), fingerprint, userID)

					Expect(err).NotTo(HaveOccurred())
					Expect(accessToken).To(Equal(mockJWT))
					Expect(refreshToken.UUID).NotTo(BeEmpty())
					Expect(refreshToken.Fingerprint).To(Equal(fingerprint))
					Expect(refreshToken.UserID).To(Equal(userID))
				})
			})

			When("JWT generation fails", func() {
				It("should return an error", func() {
					fingerprint := "test-fingerprint"
					userID := uint(1)
					mockJWTErr := fmt.Errorf("jwt error")

					jwtManager.EXPECT().
						Generate(userID, accessLife).
						Return("", mockJWTErr).
						Times(1)

					accessToken, refreshToken, err := service.GenerateTokens(context.Background(), fingerprint, userID)

					Expect(err).To(HaveOccurred())

					Expect(accessToken).To(BeEmpty())
					Expect(refreshToken).To(Equal(entity.RefreshToken{}))
				})
			})
		})
	})

	Describe("ValidateToken", func() {

		Context("when validating a token", func() {
			When("the token is valid", func() {
				It("should return the access token entity", func() {
					validToken := "valid-token"
					expectedAccessToken := entity.AccessToken{
						UserID:    1,
						IssuedAt:  time.Now(),
						ExpiresAt: time.Now().Add(time.Hour),
					}

					jwtManager.EXPECT().Parse(validToken).Return(expectedAccessToken, nil).Times(1)

					accessToken, err := service.ValidateToken(context.Background(), validToken)

					Expect(err).NotTo(HaveOccurred())
					Expect(accessToken).To(Equal(expectedAccessToken))
				})
			})

			When("the token is expired", func() {
				It("should return ErrInvalidToken", func() {
					expiredToken := "expired-token"
					expiredAccessToken := entity.AccessToken{
						UserID:    1,
						IssuedAt:  time.Now().Add(-2 * time.Hour),
						ExpiresAt: time.Now().Add(-1 * time.Hour),
					}

					jwtManager.EXPECT().Parse(expiredToken).Return(expiredAccessToken, nil).Times(1)

					accessToken, err := service.ValidateToken(context.Background(), expiredToken)

					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError(apperror.ErrInvalidToken))
					Expect(accessToken).To(Equal(entity.AccessToken{}))
				})
			})

			When("the token is invalid (parsing failed)", func() {
				It("should return ErrInvalidToken", func() {
					invalidToken := "invalid.token.string"

					jwtManager.EXPECT().Parse(invalidToken).Return(entity.AccessToken{}, apperror.ErrInvalidToken).Times(1)

					accessToken, err := service.ValidateToken(context.Background(), invalidToken)

					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError(apperror.ErrInvalidToken))
					Expect(accessToken).To(Equal(entity.AccessToken{}))
				})
			})
		})
	})

	Describe("Repository Methods", func() {
		var (
			ctx          context.Context
			refreshToken entity.RefreshToken
		)

		BeforeEach(func() {
			ctx = context.Background()

			refreshToken = entity.RefreshToken{
				UUID:        uuid.New(),
				CreatedAt:   time.Now(),
				ExpiresAt:   time.Now().Add(24 * time.Hour),
				Fingerprint: "test-fingerprint",
				UserID:      1,
			}
		})

		Describe("InsertToken", func() {
			It("should call the repository's InsertToken method and return no error", func() {
				repo.EXPECT().InsertToken(ctx, refreshToken).Return(nil).Times(1)

				err := service.InsertToken(ctx, refreshToken)

				Expect(err).NotTo(HaveOccurred())
			})

			When("the repository returns an error", func() {
				It("should return the error", func() {
					mockErr := fmt.Errorf("db insert error")
					repo.EXPECT().InsertToken(ctx, refreshToken).Return(mockErr).Times(1)

					err := service.InsertToken(ctx, refreshToken)

					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError(mockErr))
				})
			})
		})

		Describe("GetActivesTokenByUserID", func() {
			It("should call the repository's GetActivesTokenByUserID method and return tokens", func() {
				expected := []entity.RefreshToken{refreshToken}
				userID := uint(1)
				repo.EXPECT().GetActivesTokenByUserID(ctx, userID).Return(expected, nil).Times(1)

				tokens, err := service.GetActivesTokenByUserID(ctx, userID)

				Expect(err).NotTo(HaveOccurred())
				Expect(tokens).To(Equal(expected))
			})

			When("the repository returns an error", func() {
				It("should return the error and empty tokens", func() {
					mockErr := fmt.Errorf("db query error")
					userID := uint(1)
					repo.EXPECT().GetActivesTokenByUserID(ctx, userID).Return(nil, mockErr).Times(1)

					tokens, err := service.GetActivesTokenByUserID(ctx, userID)

					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError(mockErr))
					Expect(tokens).To(BeNil())
				})
			})
		})

		Describe("RevokeActivesByUserID", func() {
			It("should call the repository's RevokeActivesByUserID method and return no error", func() {
				userID := uint(1)
				repo.EXPECT().RevokeActivesByUserID(ctx, userID, uint(5)).Return(nil).Times(1)

				Expect(service.RevokeActivesByUserID(ctx, userID)).To(Succeed())
			})

			When("the repository returns an error", func() {
				It("should return the error", func() {
					mockErr := fmt.Errorf("db revoke error")
					userID := uint(1)
					repo.EXPECT().RevokeActivesByUserID(ctx, userID, uint(5)).Return(mockErr).Times(1)

					err := service.RevokeActivesByUserID(ctx, userID)

					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError(mockErr))
				})
			})
		})

		Describe("GetByUUID", func() {
			It("should call the repository's GetByUUID method and return the token", func() {
				uuidStr := refreshToken.UUID.String()
				repo.EXPECT().GetByUUID(ctx, uuidStr).Return(refreshToken, nil).Times(1)

				Expect(service.GetByUUID(ctx, uuidStr)).To(Equal(refreshToken))
			})

			When("the repository returns an error", func() {
				It("should return the error and zero value token", func() {
					mockErr := fmt.Errorf("db get by uuid error")
					uuidStr := refreshToken.UUID.String()
					repo.EXPECT().GetByUUID(ctx, uuidStr).Return(entity.RefreshToken{}, mockErr).Times(1)

					token, err := service.GetByUUID(ctx, uuidStr)

					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError(mockErr))
					Expect(token).To(Equal(entity.RefreshToken{}))
				})
			})
		})

		Describe("Update", func() {
			It("should call the repository's Update method and return the updated token", func() {
				repo.EXPECT().Update(ctx, refreshToken).Return(refreshToken, nil).Times(1)

				Expect(service.Update(ctx, refreshToken)).To(Equal(refreshToken))
			})

			When("the repository returns an error", func() {
				It("should return the error and zero value token", func() {
					mockErr := fmt.Errorf("db update error")
					repo.EXPECT().Update(ctx, refreshToken).Return(entity.RefreshToken{}, mockErr).Times(1)

					token, err := service.Update(ctx, refreshToken)

					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError(mockErr))
					Expect(token).To(Equal(entity.RefreshToken{}))
				})
			})
		})
	})
})
