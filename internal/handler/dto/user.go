package dto

import (
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/user"
)

type RegistrationUserReq struct {
	Name        string `json:"name" binding:"required"`
	Password    string `json:"password" binding:"required"`
	Email       string `json:"email" binding:"required"`
	Phone       string `json:"phone" binding:"required"`
	Fingerprint string `json:"fingerprint" binding:"required"`
}

type RegistrationUserResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (ru *RegistrationUserReq) ConvertToSvc() user.User {
	return user.User{
		Name:     ru.Name,
		Password: ru.Password,
		Email:    ru.Email,
		Phone:    ru.Phone,
	}
}

type LoginUserReq struct {
	Email       string `json:"email" binding:"required"`
	Password    string `json:"password" binding:"required"`
	Fingerprint string `json:"fingerprint" binding:"required"`
}

type RefreshReq struct {
	RefreshToken string `json:"refresh_token,omitempty"`
	Fingerprint  string `json:"fingerprint" binding:"required"`
}

type RefreshResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type LoginUserResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type LogoutReq struct {
	RefreshToken string `json:"refresh_token,omitempty"`
	Fingerprint  string `json:"fingerprint" validate:"required"`
}
