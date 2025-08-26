package handler

import (
	"context"
	"net/http"

	"github.com/EM-Stawberry/Stawberry/config"
	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/user"
	"github.com/EM-Stawberry/Stawberry/internal/handler/dto"
	"github.com/gin-gonic/gin"
)

//go:generate mockgen -source=$GOFILE -destination=user_mock_test.go -package=handler UserService

type UserService interface {
	CreateUser(ctx context.Context, user user.User, fingerprint string) (string, string, error)
	Authenticate(ctx context.Context, email, password, fingerprint string) (string, string, error)
	Refresh(ctx context.Context, refreshToken, fingerprint string) (string, string, error)
	Logout(ctx context.Context, refreshToken, fingerprint string) error
	GetUserByID(ctx context.Context, id uint) (entity.User, error)
}

type UserHandler struct {
	userService UserService
	refreshLife int
	basePath    string
	domain      string
}

func NewUserHandler(
	cfg *config.Config,
	userService UserService,
) *UserHandler {
	return &UserHandler{
		userService: userService,
		refreshLife: int(cfg.Token.RefreshTokenDuration),
		domain:      cfg.Server.Domain,
	}
}

// Registration godoc
//
//	@Summary		Регистрация нового пользователя
//	@Description	Регистрирует нового пользователя и возвращает токены доступа/обновления
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			user	body		dto.RegistrationUserReq	true	"Данные для регистрации пользователя"
//	@Success		200		{object}	dto.RegistrationUserResp
//	@Failure		400		{object}	apperror.AppError
//	@Router			/auth/reg [post]
func (h *UserHandler) Registration(c *gin.Context) {
	var regUserDTO dto.RegistrationUserReq
	if err := c.ShouldBindJSON(&regUserDTO); err != nil {
		_ = c.Error(apperror.New(apperror.BadRequest, "Invalid user data", err))
		return
	}

	accessToken, refreshToken, err := h.userService.CreateUser(
		c.Request.Context(),
		regUserDTO.ConvertToSvc(),
		regUserDTO.Fingerprint,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response := dto.RegistrationUserResp{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	setRefreshCookie(c, refreshToken, h.basePath, h.domain, h.refreshLife)

	c.JSON(http.StatusOK, response)
}

// Login godoc
//
//	@Summary		Аутентификация пользователя
//	@Description	Аутентифицирует пользователя и возвращает токены access/refresh
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			user	body		dto.LoginUserReq	true	"Учетные данные пользователя"
//	@Success		200		{object}	dto.LoginUserResp
//	@Failure		400		{object}	apperror.AppError
//	@Router			/auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var loginUserDTO dto.LoginUserReq
	if err := c.ShouldBindJSON(&loginUserDTO); err != nil {
		_ = c.Error(apperror.New(apperror.BadRequest, "Invalid user data", err))
		return
	}

	accessToken, refreshToken, err := h.userService.Authenticate(
		c.Request.Context(),
		loginUserDTO.Email,
		loginUserDTO.Password,
		loginUserDTO.Fingerprint,
	)

	if err != nil {
		_ = c.Error(err)
		return
	}

	response := dto.LoginUserResp{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	setRefreshCookie(c, refreshToken, h.basePath, h.domain, h.refreshLife)

	c.JSON(http.StatusOK, response)
}

// Refresh godoc
//
//	@Summary		Обновление токенов
//	@Description	Обновляет токены access и refresh
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			refresh	body		dto.RefreshReq	true	"Данные токена refresh"
//	@Success		200		{object}	dto.RefreshResp
//	@Failure		400		{object}	apperror.AppError
//	@Router			/auth/refresh [post]
func (h *UserHandler) Refresh(c *gin.Context) {
	var refreshDTO dto.RefreshReq
	if err := c.ShouldBindJSON(&refreshDTO); err != nil {
		_ = c.Error(apperror.New(apperror.BadRequest, "Invalid refresh data", err))
		return
	}

	if refreshDTO.RefreshToken == "" {
		refresh, err := c.Cookie("refresh_token")
		if err != nil {
			_ = c.Error(apperror.New(apperror.BadRequest, "Invalid refresh data", err))
			return
		}
		refreshDTO.RefreshToken = refresh
	}

	accessToken, refreshToken, err := h.userService.Refresh(
		c.Request.Context(),
		refreshDTO.RefreshToken,
		refreshDTO.Fingerprint,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response := dto.RefreshResp{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	setRefreshCookie(c, refreshToken, h.basePath, h.domain, h.refreshLife)

	c.JSON(http.StatusOK, response)
}

// Logout godoc
//
//	@Summary		Выход из системы
//	@Description	Выход пользователя и инвалидация токена обновления
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			logout	body	dto.LogoutReq	true	"Данные для выхода"
//	@Success		200
//	@Failure		400	{object}	apperror.AppError
//	@Router			/auth/logout [post]
func (h *UserHandler) Logout(c *gin.Context) {
	var logoutDTO dto.LogoutReq
	if err := c.ShouldBindJSON(&logoutDTO); err != nil {
		_ = c.Error(apperror.New(apperror.BadRequest, "Invalid refresh data", err))
		return
	}

	if logoutDTO.RefreshToken == "" {
		refresh, err := c.Cookie("refresh_token")
		if err != nil {
			_ = c.Error(apperror.New(apperror.BadRequest, "Invalid refresh data", err))
			return
		}
		logoutDTO.RefreshToken = refresh
	}

	if err := h.userService.Logout(
		c.Request.Context(),
		logoutDTO.RefreshToken,
		logoutDTO.Fingerprint,
	); err != nil {
		_ = c.Error(err)
		return
	}

	c.Status(http.StatusOK)
}

func setRefreshCookie(c *gin.Context, refreshToken, basePath, domain string, maxAge int) {
	jwtCookie := http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     basePath + "/auth",
		Domain:   domain,
		MaxAge:   maxAge,
		Secure:   true,
		HttpOnly: true,
	}

	c.SetCookie(
		jwtCookie.Name,
		jwtCookie.Value,
		jwtCookie.MaxAge,
		jwtCookie.Path,
		jwtCookie.Domain,
		jwtCookie.Secure,
		jwtCookie.HttpOnly,
	)

	c.SetSameSite(http.SameSiteStrictMode)
}
