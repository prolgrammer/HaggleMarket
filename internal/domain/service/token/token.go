package token

import (
	"context"
	"time"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/google/uuid"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
)

//go:generate mockgen -source=$GOFILE -destination=token_mock_test.go -package=token Repository

//go:generate mockgen -source=$GOFILE -destination=token_mock_test.go -package=token Repository

type Repository interface {
	InsertToken(ctx context.Context, token entity.RefreshToken) error
	GetActivesTokenByUserID(ctx context.Context, userID uint) ([]entity.RefreshToken, error)
	RevokeActivesByUserID(ctx context.Context, userID uint, retain uint) error
	GetByUUID(ctx context.Context, uuid string) (entity.RefreshToken, error)
	Update(ctx context.Context, refresh entity.RefreshToken) (entity.RefreshToken, error)
	CleanExpired(ctx context.Context, userID uint, retain uint) error
}

type JWTManager interface {
	Generate(userID uint, duration time.Duration) (string, error)
	Parse(token string) (entity.AccessToken, error)
}

type Service struct {
	tokenRepository Repository
	jwtManager      JWTManager
	refreshLife     time.Duration
	accessLife      time.Duration
}

func NewService(tokenRepo Repository, jwtManager JWTManager, refreshLife, accessLife time.Duration) *Service {
	return &Service{
		tokenRepository: tokenRepo,
		jwtManager:      jwtManager,
		refreshLife:     refreshLife,
		accessLife:      accessLife,
	}
}

// GenerateTokens генерирует новый токен доступа и токен обновления для пользователя.
func (ts *Service) GenerateTokens(
	ctx context.Context,
	fingerprint string,
	userID uint,
) (string, entity.RefreshToken, error) {

	if ctx.Err() != nil {
		return "", entity.RefreshToken{}, ctx.Err()
	}

	accessToken, err := ts.jwtManager.Generate(userID, ts.accessLife)
	if err != nil {
		return "", entity.RefreshToken{}, err
	}

	if ctx.Err() != nil {
		return "", entity.RefreshToken{}, ctx.Err()
	}

	entityRefreshToken, err := generateRefresh(fingerprint, userID, ts.refreshLife)
	if err != nil {
		return "", entity.RefreshToken{}, err
	}

	return accessToken, entityRefreshToken, nil
}

// ValidateToken проверяет access токен и возвращает расшифрованную информацию, если она действительна.
func (ts *Service) ValidateToken(
	ctx context.Context,
	token string,
) (entity.AccessToken, error) {

	if ctx.Err() != nil {
		return entity.AccessToken{}, ctx.Err()
	}

	accessToken, err := ts.jwtManager.Parse(token)
	if err != nil {
		return entity.AccessToken{}, err
	}

	if time.Now().After(accessToken.ExpiresAt) {
		return entity.AccessToken{}, apperror.ErrInvalidToken
	}

	return accessToken, nil
}

func (ts *Service) InsertToken(
	ctx context.Context,
	token entity.RefreshToken,
) error {
	return ts.tokenRepository.InsertToken(ctx, token)
}

// GetActivesTokenByUserID извлекает все активные токены обновления для конкретного пользователя.
func (ts *Service) GetActivesTokenByUserID(
	ctx context.Context,
	userID uint,
) ([]entity.RefreshToken, error) {
	return ts.tokenRepository.GetActivesTokenByUserID(ctx, userID)
}

// retainActive определяет количество последних активных сессий, которые сохраняются при зачистке
const retainActive = 5

// RevokeActivesByUserID аннулирует все активные токены обновления для определенного пользователя.
func (ts *Service) RevokeActivesByUserID(
	ctx context.Context,
	userID uint,
) error {
	return ts.tokenRepository.RevokeActivesByUserID(ctx, userID, retainActive)
}

// retainExpired определяет количество отозванных и устаревших токенов, которые
// сохраняются в базе при вызове CleanUpExpiredByUserID
const retainExpired = 5

// CleanExpiredByUserID удаляет все устаревшие и отозванные токены обновления для определённого пользователя
func (ts *Service) CleanUpExpiredByUserID(
	ctx context.Context,
	userID uint,
) error {
	return ts.tokenRepository.CleanExpired(ctx, userID, retainExpired)
}

func (ts *Service) GetByUUID(
	ctx context.Context,
	uuid string,
) (entity.RefreshToken, error) {
	return ts.tokenRepository.GetByUUID(ctx, uuid)
}

func (ts *Service) Update(
	ctx context.Context,
	refresh entity.RefreshToken,
) (entity.RefreshToken, error) {
	return ts.tokenRepository.Update(ctx, refresh)
}

// generateRefresh создает новый refresh токен обновления с указанным
// userID, fingerprint и сроком действия.
func generateRefresh(fingerprint string, userID uint, refreshLife time.Duration) (entity.RefreshToken, error) {
	now := time.Now()

	refreshUUID, err := uuid.NewRandom()
	if err != nil {
		return entity.RefreshToken{}, err
	}

	return entity.RefreshToken{
		UUID:        refreshUUID,
		CreatedAt:   now,
		ExpiresAt:   now.Add(refreshLife),
		RevokedAt:   nil,
		Fingerprint: fingerprint,
		UserID:      userID,
	}, nil
}
