package user

import (
	"context"
	"fmt"
	"time"

	"github.com/EM-Stawberry/Stawberry/pkg/email"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
)

//go:generate mockgen -source=$GOFILE -destination=user_mock_test.go -package=user Repository, TokenService

type Repository interface {
	InsertUser(ctx context.Context, user User) (uint, error)
	GetUser(ctx context.Context, email string) (entity.User, error)
	GetUserByID(ctx context.Context, id uint) (entity.User, error)
}

// PasswordManager выполняет операции с паролями, такие как хеширование и проверка
type PasswordManager interface {
	Hash(password string) (string, error)
	Compare(password, hash string) (bool, error)
}

type TokenService interface {
	GenerateTokens(ctx context.Context, fingerprint string, userID uint) (string, entity.RefreshToken, error)
	InsertToken(ctx context.Context, token entity.RefreshToken) error
	RevokeActivesByUserID(ctx context.Context, userID uint) error
	GetByUUID(ctx context.Context, uuid string) (entity.RefreshToken, error)
	Update(ctx context.Context, refresh entity.RefreshToken) (entity.RefreshToken, error)
	CleanUpExpiredByUserID(ctx context.Context, userID uint) error
}

type Service struct {
	userRepository  Repository
	tokenService    TokenService
	passwordManager PasswordManager
	mailer          email.MailerService
}

func NewService(userRepo Repository,
	tokenService TokenService,
	passwordManager PasswordManager,
	mailer email.MailerService,
) *Service {
	return &Service{
		userRepository:  userRepo,
		tokenService:    tokenService,
		passwordManager: passwordManager,
		mailer:          mailer,
	}
}

// CreateUser создает пользователя, хэшируя его пароль, используя HashArgon2id
// генерирует access токен и uuid refresh uuid.
func (us *Service) CreateUser(
	ctx context.Context,
	user User,
	fingerprint string,
) (string, string, error) {
	hash, err := us.passwordManager.Hash(user.Password)
	if err != nil {
		err := apperror.ErrFailedToGeneratePassword
		err.WrappedErr = fmt.Errorf("failed to generate password %w", err)
		return "", "", err
	}
	user.Password = hash

	id, err := us.userRepository.InsertUser(ctx, user)
	if err != nil {
		return "", "", err
	}

	accessToken, refreshToken, err := us.tokenService.GenerateTokens(ctx, fingerprint, id)
	if err != nil {
		return "", "", err
	}

	if err = us.tokenService.InsertToken(ctx, refreshToken); err != nil {
		return "", "", err
	}

	us.mailer.Registered(user.Name, user.Email)

	return accessToken, refreshToken.UUID.String(), nil
}

// Authenticate аутентифицирует пользователя по email и паролю, создавая новые токены.
func (us *Service) Authenticate(
	ctx context.Context,
	email,
	password,
	fingerprint string,
) (string, string, error) {
	user, err := us.userRepository.GetUser(ctx, email)
	if err != nil {
		return "", "", apperror.ErrUserNotFound
	}

	compared, err := us.passwordManager.Compare(password, user.Password)
	if err != nil {
		return "", "", err
	}

	if !compared {
		return "", "", apperror.ErrIncorrectPassword
	}

	if err := us.tokenService.RevokeActivesByUserID(ctx, user.ID); err != nil {
		return "", "", err
	}

	err = us.tokenService.CleanUpExpiredByUserID(ctx, user.ID)
	if err != nil {
		return "", "", err
	}

	accessToken, refreshToken, err := us.tokenService.GenerateTokens(ctx, fingerprint, user.ID)
	if err != nil {
		return "", "", err
	}

	if err = us.tokenService.InsertToken(ctx, refreshToken); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken.UUID.String(), nil
}

// Refresh обновляет пару токенов аутентификации.
func (us *Service) Refresh(
	ctx context.Context,
	refreshToken,
	fingerprint string,
) (string, string, error) {
	refresh, err := us.tokenService.GetByUUID(ctx, refreshToken)
	if err != nil {
		return "", "", err
	}

	if !refresh.IsValid() {
		return "", "", apperror.ErrInvalidToken
	}

	if refresh.Fingerprint != fingerprint {
		return "", "", apperror.ErrInvalidFingerprint
	}

	now := time.Now()
	refresh.RevokedAt = &now

	refresh, err = us.tokenService.Update(ctx, refresh)
	if err != nil {
		return "", "", err
	}

	user, err := us.userRepository.GetUserByID(ctx, refresh.UserID)
	if err != nil {
		return "", "", err
	}

	access, refresh, err := us.tokenService.GenerateTokens(ctx, fingerprint, user.ID)
	if err != nil {
		return "", "", err
	}

	err = us.tokenService.CleanUpExpiredByUserID(ctx, user.ID)
	if err != nil {
		return "", "", err
	}

	err = us.tokenService.InsertToken(ctx, refresh)
	if err != nil {
		return "", "", err
	}

	return access, refresh.UUID.String(), nil
}

func (us *Service) Logout(
	ctx context.Context,
	refreshToken,
	fingerprint string,
) error {
	refresh, err := us.tokenService.GetByUUID(ctx, refreshToken)
	if err != nil {
		return apperror.ErrInvalidToken
	}

	if !refresh.IsValid() {
		return apperror.ErrInvalidToken
	}

	if refresh.Fingerprint != fingerprint {
		return apperror.ErrInvalidFingerprint
	}

	now := time.Now()
	refresh.RevokedAt = &now

	_, err = us.tokenService.Update(ctx, refresh)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}

func (us *Service) GetUserByID(ctx context.Context, id uint) (entity.User, error) {
	return us.userRepository.GetUserByID(ctx, id)
}
