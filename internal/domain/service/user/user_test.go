package user

import (
	"context"
	"errors"
	"time"

	"github.com/EM-Stawberry/Stawberry/pkg/email/mock_email"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("UserService", func() {
	var (
		ctrl                *gomock.Controller
		mockRepo            *MockRepository
		mockTokenService    *MockTokenService
		mockPasswordManager *MockPasswordManager
		mockEmailService    *mock_email.MockMailerService
		userService         *Service
		ctx                 context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockRepo = NewMockRepository(ctrl)
		mockTokenService = NewMockTokenService(ctrl)
		mockPasswordManager = NewMockPasswordManager(ctrl)
		mockEmailService = mock_email.NewMockMailerService(ctrl)
		userService = NewService(mockRepo, mockTokenService, mockPasswordManager, mockEmailService)
		ctx = context.Background()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("CreateUser", func() {
		var (
			testUser       User
			hashedPassword string
			fingerprint    string
		)

		BeforeEach(func() {
			testUser = User{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "password123",
			}
			hashedPassword = "hashed-password"
			fingerprint = "test-fingerprint"
		})

		Context("when user creation is successful", func() {
			It("should create user and return tokens", func() {
				mockPasswordManager.EXPECT().Hash(testUser.Password).Return(hashedPassword, nil)
				mockRepo.EXPECT().InsertUser(ctx, gomock.Any()).Return(uint(1), nil)
				mockTokenService.EXPECT().
					GenerateTokens(ctx, fingerprint, uint(1)).
					Return("access-token", entity.RefreshToken{UUID: uuid.New()}, nil)
				mockTokenService.EXPECT().InsertToken(ctx, gomock.Any()).Return(nil)
				mockEmailService.EXPECT().Registered(testUser.Name, testUser.Email)

				accessToken, refreshToken, err := userService.CreateUser(ctx, testUser, fingerprint)

				Expect(err).ToNot(HaveOccurred())
				Expect(accessToken).ToNot(BeEmpty())
				Expect(refreshToken).ToNot(BeEmpty())
			})
		})

		Context("when password hashing fails", func() {
			It("should return error", func() {
				mockPasswordManager.EXPECT().Hash(testUser.Password).Return("", errors.New("failed to generate password"))

				accessToken, refreshToken, err := userService.CreateUser(ctx, testUser, fingerprint)

				Expect(err).To(HaveOccurred())
				Expect(accessToken).To(BeEmpty())
				Expect(refreshToken).To(BeEmpty())
			})
		})

		Context("when user insertion fails", func() {
			It("should return error", func() {
				mockPasswordManager.EXPECT().Hash(testUser.Password).Return(hashedPassword, nil)
				mockRepo.EXPECT().InsertUser(ctx, gomock.Any()).Return(uint(0), errors.New("db error"))

				accessToken, refreshToken, err := userService.CreateUser(ctx, testUser, fingerprint)

				Expect(err).To(HaveOccurred())
				Expect(accessToken).To(BeEmpty())
				Expect(refreshToken).To(BeEmpty())
			})
		})

		Context("when token generation fails", func() {
			It("should return error", func() {
				mockPasswordManager.EXPECT().Hash(testUser.Password).Return(hashedPassword, nil)
				mockRepo.EXPECT().InsertUser(ctx, gomock.Any()).Return(uint(1), nil)
				mockTokenService.EXPECT().
					GenerateTokens(ctx, fingerprint, uint(1)).
					Return("", entity.RefreshToken{}, errors.New("token generation error"))

				accessToken, refreshToken, err := userService.CreateUser(ctx, testUser, fingerprint)

				Expect(err).To(HaveOccurred())
				Expect(accessToken).To(BeEmpty())
				Expect(refreshToken).To(BeEmpty())
			})
		})

		Context("when token insertion fails", func() {
			It("should return error", func() {
				mockPasswordManager.EXPECT().Hash(testUser.Password).Return(hashedPassword, nil)
				mockRepo.EXPECT().InsertUser(ctx, gomock.Any()).Return(uint(1), nil)
				mockTokenService.EXPECT().
					GenerateTokens(ctx, fingerprint, uint(1)).
					Return("access-token", entity.RefreshToken{}, nil)
				mockTokenService.EXPECT().
					InsertToken(ctx, gomock.Any()).
					Return(errors.New("token insertion error"))

				accessToken, refreshToken, err := userService.CreateUser(ctx, testUser, fingerprint)

				Expect(err).To(HaveOccurred())
				Expect(accessToken).To(BeEmpty())
				Expect(refreshToken).To(BeEmpty())
			})
		})
	})

	Describe("Authenticate", func() {
		var (
			email          string
			password       string
			fingerprint    string
			hashedPassword string
			testUser       entity.User
		)

		BeforeEach(func() {
			email = "test@example.com"
			password = "password123"
			fingerprint = "test-fingerprint"
			hashedPassword = "hashed-password"
			testUser = entity.User{
				ID:       1,
				Email:    email,
				Password: hashedPassword,
			}
		})

		Context("when authentication is successful", func() {
			It("should authenticate user and return tokens", func() {
				mockRepo.EXPECT().GetUser(ctx, email).Return(testUser, nil)
				mockPasswordManager.EXPECT().Compare(password, hashedPassword).Return(true, nil)
				mockTokenService.EXPECT().RevokeActivesByUserID(ctx, testUser.ID).Return(nil)
				mockTokenService.EXPECT().CleanUpExpiredByUserID(ctx, testUser.ID).Return(nil)
				mockTokenService.EXPECT().
					GenerateTokens(ctx, fingerprint, testUser.ID).
					Return("access-token", entity.RefreshToken{UUID: uuid.New()}, nil)
				mockTokenService.EXPECT().InsertToken(ctx, gomock.Any()).Return(nil)

				accessToken, refreshToken, err := userService.Authenticate(ctx, email, password, fingerprint)

				Expect(err).ToNot(HaveOccurred())
				Expect(accessToken).ToNot(BeEmpty())
				Expect(refreshToken).ToNot(BeEmpty())
			})
		})

		Context("when user is not found", func() {
			It("should return user not found error", func() {
				mockRepo.EXPECT().GetUser(ctx, email).Return(entity.User{}, apperror.ErrUserNotFound)

				accessToken, refreshToken, err := userService.Authenticate(ctx, email, password, fingerprint)

				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(apperror.ErrUserNotFound))
				Expect(accessToken).To(BeEmpty())
				Expect(refreshToken).To(BeEmpty())
			})
		})

		Context("when password is incorrect", func() {
			It("should return incorrect password error", func() {
				mockRepo.EXPECT().GetUser(ctx, email).Return(testUser, nil)
				mockPasswordManager.EXPECT().Compare("wrong_password", hashedPassword).Return(false, nil)

				accessToken, refreshToken, err := userService.Authenticate(ctx, email, "wrong_password", fingerprint)

				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(apperror.ErrIncorrectPassword))
				Expect(accessToken).To(BeEmpty())
				Expect(refreshToken).To(BeEmpty())
			})
		})

		Context("when password validation fails", func() {
			It("should return error", func() {
				mockRepo.EXPECT().GetUser(ctx, email).Return(testUser, nil)
				mockPasswordManager.EXPECT().Compare(password, hashedPassword).Return(false, errors.New("invalid password"))

				accessToken, refreshToken, err := userService.Authenticate(ctx, email, password, fingerprint)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid password"))
				Expect(accessToken).To(BeEmpty())
				Expect(refreshToken).To(BeEmpty())
			})
		})

		Context("when revoking active tokens fails", func() {
			It("should return error", func() {
				mockRepo.EXPECT().GetUser(ctx, email).Return(testUser, nil)
				mockPasswordManager.EXPECT().Compare(password, hashedPassword).Return(true, nil)
				mockTokenService.EXPECT().RevokeActivesByUserID(ctx, testUser.ID).Return(errors.New("revoke error"))

				accessToken, refreshToken, err := userService.Authenticate(ctx, email, password, fingerprint)

				Expect(err).To(HaveOccurred())
				Expect(accessToken).To(BeEmpty())
				Expect(refreshToken).To(BeEmpty())
			})
		})

		Context("when cleaning up expired tokens fails", func() {
			It("should return error", func() {
				mockRepo.EXPECT().GetUser(ctx, email).Return(testUser, nil)
				mockPasswordManager.EXPECT().Compare(password, hashedPassword).Return(true, nil)
				mockTokenService.EXPECT().RevokeActivesByUserID(ctx, testUser.ID).Return(nil)
				mockTokenService.EXPECT().CleanUpExpiredByUserID(ctx, testUser.ID).Return(errors.New("cleanup error"))

				accessToken, refreshToken, err := userService.Authenticate(ctx, email, password, fingerprint)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cleanup error"))
				Expect(accessToken).To(BeEmpty())
				Expect(refreshToken).To(BeEmpty())
			})
		})

		Context("when token generation fails", func() {
			It("should return error", func() {
				mockRepo.EXPECT().GetUser(ctx, email).Return(testUser, nil)
				mockPasswordManager.EXPECT().Compare(password, hashedPassword).Return(true, nil)
				mockTokenService.EXPECT().RevokeActivesByUserID(ctx, testUser.ID).Return(nil)
				mockTokenService.EXPECT().CleanUpExpiredByUserID(ctx, testUser.ID).Return(nil)
				mockTokenService.EXPECT().
					GenerateTokens(ctx, fingerprint, testUser.ID).
					Return("", entity.RefreshToken{}, errors.New("token generation error"))

				accessToken, refreshToken, err := userService.Authenticate(ctx, email, password, fingerprint)

				Expect(err).To(HaveOccurred())
				Expect(accessToken).To(BeEmpty())
				Expect(refreshToken).To(BeEmpty())
			})
		})

		Context("when token insertion fails", func() {
			It("should return error", func() {
				mockRepo.EXPECT().GetUser(ctx, email).Return(testUser, nil)
				mockPasswordManager.EXPECT().Compare(password, hashedPassword).Return(true, nil)
				mockTokenService.EXPECT().RevokeActivesByUserID(ctx, testUser.ID).Return(nil)
				mockTokenService.EXPECT().CleanUpExpiredByUserID(ctx, testUser.ID).Return(nil)
				mockTokenService.EXPECT().
					GenerateTokens(ctx, fingerprint, testUser.ID).
					Return("access-token", entity.RefreshToken{}, nil)
				mockTokenService.EXPECT().InsertToken(ctx, gomock.Any()).Return(errors.New("insert error"))

				accessToken, refreshToken, err := userService.Authenticate(ctx, email, password, fingerprint)

				Expect(err).To(HaveOccurred())
				Expect(accessToken).To(BeEmpty())
				Expect(refreshToken).To(BeEmpty())
			})
		})
	})

	Describe("Refresh", func() {
		var (
			refreshTokenStr   string
			fingerprint       string
			userID            uint
			validRefreshToken entity.RefreshToken
		)

		BeforeEach(func() {
			refreshTokenStr = uuid.New().String()
			fingerprint = "test-fingerprint"
			userID = uint(1)
			validRefreshToken = entity.RefreshToken{
				UUID:        uuid.New(),
				ExpiresAt:   time.Now().Add(time.Hour),
				Fingerprint: fingerprint,
				UserID:      userID,
			}
		})

		Context("when token refresh is successful", func() {
			It("should refresh tokens and return new ones", func() {
				mockTokenService.EXPECT().GetByUUID(ctx, refreshTokenStr).Return(validRefreshToken, nil)
				mockTokenService.EXPECT().Update(ctx, gomock.Any()).Return(validRefreshToken, nil)
				mockRepo.EXPECT().GetUserByID(ctx, userID).Return(entity.User{ID: userID}, nil)
				mockTokenService.EXPECT().
					GenerateTokens(ctx, fingerprint, userID).
					Return("new-access-token", entity.RefreshToken{UUID: uuid.New()}, nil)
				mockTokenService.EXPECT().CleanUpExpiredByUserID(ctx, userID).Return(nil)
				mockTokenService.EXPECT().InsertToken(ctx, gomock.Any()).Return(nil)

				accessToken, newRefreshToken, err := userService.Refresh(ctx, refreshTokenStr, fingerprint)

				Expect(err).ToNot(HaveOccurred())
				Expect(accessToken).ToNot(BeEmpty())
				Expect(newRefreshToken).ToNot(BeEmpty())
			})
		})

		Context("when refresh token is invalid (expired)", func() {
			It("should return invalid token error", func() {
				invalidToken := validRefreshToken
				invalidToken.ExpiresAt = time.Now().Add(-time.Hour)

				mockTokenService.EXPECT().GetByUUID(ctx, refreshTokenStr).Return(invalidToken, nil)

				accessToken, newRefreshToken, err := userService.Refresh(ctx, refreshTokenStr, fingerprint)

				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(apperror.ErrInvalidToken))
				Expect(accessToken).To(BeEmpty())
				Expect(newRefreshToken).To(BeEmpty())
			})
		})

		Context("when fingerprint is invalid", func() {
			It("should return invalid fingerprint error", func() {
				mockTokenService.EXPECT().GetByUUID(ctx, refreshTokenStr).Return(validRefreshToken, nil)

				accessToken, newRefreshToken, err := userService.Refresh(ctx, refreshTokenStr, "wrong-fingerprint")

				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(apperror.ErrInvalidFingerprint))
				Expect(accessToken).To(BeEmpty())
				Expect(newRefreshToken).To(BeEmpty())
			})
		})

		Context("when getting refresh token by UUID fails", func() {
			It("should return error", func() {
				mockTokenService.EXPECT().
					GetByUUID(ctx, refreshTokenStr).
					Return(entity.RefreshToken{}, errors.New("database error"))

				accessToken, newRefreshToken, err := userService.Refresh(ctx, refreshTokenStr, fingerprint)

				Expect(err).To(HaveOccurred())
				Expect(accessToken).To(BeEmpty())
				Expect(newRefreshToken).To(BeEmpty())
			})
		})

		Context("when updating refresh token fails", func() {
			It("should return error", func() {
				mockTokenService.EXPECT().GetByUUID(ctx, refreshTokenStr).Return(validRefreshToken, nil)
				mockTokenService.EXPECT().Update(ctx, gomock.Any()).Return(entity.RefreshToken{}, errors.New("update error"))

				accessToken, newRefreshToken, err := userService.Refresh(ctx, refreshTokenStr, fingerprint)

				Expect(err).To(HaveOccurred())
				Expect(accessToken).To(BeEmpty())
				Expect(newRefreshToken).To(BeEmpty())
			})
		})

		Context("when getting user by ID fails", func() {
			It("should return error", func() {
				mockTokenService.EXPECT().GetByUUID(ctx, refreshTokenStr).Return(validRefreshToken, nil)
				mockTokenService.EXPECT().Update(ctx, gomock.Any()).Return(validRefreshToken, nil)
				mockRepo.EXPECT().GetUserByID(ctx, userID).Return(entity.User{}, errors.New("user not found"))

				accessToken, newRefreshToken, err := userService.Refresh(ctx, refreshTokenStr, fingerprint)

				Expect(err).To(HaveOccurred())
				Expect(accessToken).To(BeEmpty())
				Expect(newRefreshToken).To(BeEmpty())
			})
		})

		Context("when generating new tokens fails", func() {
			It("should return error", func() {
				mockTokenService.EXPECT().GetByUUID(ctx, refreshTokenStr).Return(validRefreshToken, nil)
				mockTokenService.EXPECT().Update(ctx, gomock.Any()).Return(validRefreshToken, nil)
				mockRepo.EXPECT().GetUserByID(ctx, userID).Return(entity.User{ID: userID}, nil)
				mockTokenService.EXPECT().
					GenerateTokens(ctx, fingerprint, userID).
					Return("", entity.RefreshToken{}, errors.New("token generation error"))

				accessToken, newRefreshToken, err := userService.Refresh(ctx, refreshTokenStr, fingerprint)

				Expect(err).To(HaveOccurred())
				Expect(accessToken).To(BeEmpty())
				Expect(newRefreshToken).To(BeEmpty())
			})
		})

		Context("when cleaning up expired tokens fails", func() {
			It("should return error", func() {
				mockTokenService.EXPECT().GetByUUID(ctx, refreshTokenStr).Return(validRefreshToken, nil)
				mockTokenService.EXPECT().Update(ctx, gomock.Any()).Return(validRefreshToken, nil)
				mockRepo.EXPECT().GetUserByID(ctx, userID).Return(entity.User{ID: userID}, nil)
				mockTokenService.EXPECT().
					GenerateTokens(ctx, fingerprint, userID).
					Return("new-access-token", entity.RefreshToken{}, nil)
				mockTokenService.EXPECT().CleanUpExpiredByUserID(ctx, userID).Return(errors.New("cleanup error"))

				accessToken, newRefreshToken, err := userService.Refresh(ctx, refreshTokenStr, fingerprint)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cleanup error"))
				Expect(accessToken).To(BeEmpty())
				Expect(newRefreshToken).To(BeEmpty())
			})
		})

		Context("when inserting new refresh token fails", func() {
			It("should return error", func() {
				mockTokenService.EXPECT().GetByUUID(ctx, refreshTokenStr).Return(validRefreshToken, nil)
				mockTokenService.EXPECT().Update(ctx, gomock.Any()).Return(validRefreshToken, nil)
				mockRepo.EXPECT().GetUserByID(ctx, userID).Return(entity.User{ID: userID}, nil)
				mockTokenService.EXPECT().
					GenerateTokens(ctx, fingerprint, userID).
					Return("new-access-token", entity.RefreshToken{}, nil)
				mockTokenService.EXPECT().CleanUpExpiredByUserID(ctx, userID).Return(nil)
				mockTokenService.EXPECT().InsertToken(ctx, gomock.Any()).Return(errors.New("insert error"))

				accessToken, newRefreshToken, err := userService.Refresh(ctx, refreshTokenStr, fingerprint)

				Expect(err).To(HaveOccurred())
				Expect(accessToken).To(BeEmpty())
				Expect(newRefreshToken).To(BeEmpty())
			})
		})
	})

	Describe("Logout", func() {
		var (
			refreshTokenStr   string
			fingerprint       string
			validRefreshToken entity.RefreshToken
		)

		BeforeEach(func() {
			refreshTokenStr = uuid.New().String()
			fingerprint = "test-fingerprint"
			validRefreshToken = entity.RefreshToken{
				UUID:        uuid.New(),
				ExpiresAt:   time.Now().Add(time.Hour),
				Fingerprint: fingerprint,
			}
		})

		Context("when logout is successful", func() {
			It("should logout user successfully", func() {
				mockTokenService.EXPECT().GetByUUID(ctx, refreshTokenStr).Return(validRefreshToken, nil)
				mockTokenService.EXPECT().Update(ctx, gomock.Any()).Return(entity.RefreshToken{}, nil)

				err := userService.Logout(ctx, refreshTokenStr, fingerprint)

				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when token is not found", func() {
			It("should return invalid token error", func() {
				mockTokenService.EXPECT().GetByUUID(ctx, refreshTokenStr).Return(entity.RefreshToken{}, apperror.ErrInvalidToken)

				err := userService.Logout(ctx, refreshTokenStr, fingerprint)

				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(apperror.ErrInvalidToken))
			})
		})

		Context("when token is expired", func() {
			It("should return invalid token error", func() {
				expiredToken := entity.RefreshToken{
					UUID:        uuid.New(),
					ExpiresAt:   time.Now().Add(-time.Hour),
					Fingerprint: fingerprint,
				}

				mockTokenService.EXPECT().GetByUUID(ctx, refreshTokenStr).Return(expiredToken, nil)

				err := userService.Logout(ctx, refreshTokenStr, fingerprint)

				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(apperror.ErrInvalidToken))
			})
		})

		Context("when token is already revoked", func() {
			It("should return invalid token error", func() {
				revokedTime := time.Now().Add(-time.Hour)
				revokedToken := entity.RefreshToken{
					UUID:        uuid.New(),
					ExpiresAt:   time.Now().Add(time.Hour),
					RevokedAt:   &revokedTime,
					Fingerprint: fingerprint,
				}

				mockTokenService.EXPECT().GetByUUID(ctx, refreshTokenStr).Return(revokedToken, nil)

				err := userService.Logout(ctx, refreshTokenStr, fingerprint)

				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(apperror.ErrInvalidToken))
			})
		})

		Context("when fingerprint is invalid", func() {
			It("should return invalid fingerprint error", func() {
				mockTokenService.EXPECT().GetByUUID(ctx, refreshTokenStr).Return(validRefreshToken, nil)

				err := userService.Logout(ctx, refreshTokenStr, "wrong-fingerprint")

				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(apperror.ErrInvalidFingerprint))
			})
		})

		Context("when token update fails", func() {
			It("should return error", func() {
				mockTokenService.EXPECT().GetByUUID(ctx, refreshTokenStr).Return(validRefreshToken, nil)
				mockTokenService.EXPECT().Update(ctx, gomock.Any()).Return(entity.RefreshToken{}, errors.New("update error"))

				err := userService.Logout(ctx, refreshTokenStr, fingerprint)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to revoke refresh token"))
			})
		})
	})

	Describe("GetUserByID", func() {
		var (
			userID       uint
			expectedUser entity.User
		)

		BeforeEach(func() {
			userID = uint(1)
			expectedUser = entity.User{
				ID:    userID,
				Email: "test@example.com",
			}
		})

		Context("when getting user is successful", func() {
			It("should return user", func() {
				mockRepo.EXPECT().GetUserByID(ctx, userID).Return(expectedUser, nil)

				user, err := userService.GetUserByID(ctx, userID)

				Expect(err).ToNot(HaveOccurred())
				Expect(user).To(Equal(expectedUser))
			})
		})

		Context("when user is not found", func() {
			It("should return user not found error", func() {
				mockRepo.EXPECT().GetUserByID(ctx, userID).Return(entity.User{}, apperror.ErrUserNotFound)

				user, err := userService.GetUserByID(ctx, userID)

				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(apperror.ErrUserNotFound))
				Expect(user).To(Equal(entity.User{}))
			})
		})
	})
})
