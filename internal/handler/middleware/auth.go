package middleware

import (
	"context"
	"strings"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/EM-Stawberry/Stawberry/internal/handler/helpers"

	"github.com/gin-gonic/gin"
)

type UserGetter interface {
	GetUserByID(ctx context.Context, id uint) (entity.User, error)
}

type TokenValidator interface {
	ValidateToken(ctx context.Context, token string) (entity.AccessToken, error)
}

const (
	authorizationHeader = "Authorization"
	bearerSchema        = "Bearer"
)

// AuthMiddleware валидирует access token,
// достает из него userID и проверяет существование пользователя
func AuthMiddleware(userGetter UserGetter, validator TokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHead := c.GetHeader(authorizationHeader)
		if authHead == "" {
			_ = c.Error(apperror.New(apperror.Unauthorized, "Authorization header is missing", nil))
			c.Abort()
			return
		}

		parts := strings.Split(authHead, " ")
		if len(parts) != 2 || parts[0] != bearerSchema {
			_ = c.Error(apperror.New(apperror.Unauthorized, "Invalid authorization format", nil))
			c.Abort()
			return
		}

		access, err := validator.ValidateToken(c.Request.Context(), parts[1])
		if err != nil {
			_ = c.Error(apperror.New(apperror.Unauthorized, "Invalid or expired token", err))
			c.Abort()
			return
		}

		user, err := userGetter.GetUserByID(c.Request.Context(), access.UserID)
		if err != nil {
			_ = c.Error(apperror.New(apperror.Unauthorized, "User not found", err))
			c.Abort()
			return
		}

		c.Set(helpers.UserIDKey, user.ID)
		c.Set(helpers.UserIsStoreKey, user.IsStore)
		c.Set(helpers.UserName, user.Name)
		c.Set(helpers.UserEmail, user.Email)

		c.Set(helpers.UserIsAdminKey, false)
		c.Next()
	}
}
