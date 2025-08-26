package model

import (
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/user"
)

type User struct {
	ID            uint   `db:"id"`
	Name          string `db:"name"`
	Email         string `db:"email"`
	Phone         string `db:"phone_number"`
	Password      string `db:"password_hash"`
	IsStore       bool   `db:"is_store"`
	Notifications []Notification
}

func ConvertUserFromSvc(u user.User) User {
	return User{
		Name:     u.Name,
		Email:    u.Email,
		Phone:    u.Phone,
		Password: u.Password,
		IsStore:  u.IsStore,
	}
}

func ConvertUserToEntity(u User) entity.User {
	return entity.User{
		ID:       u.ID,
		Name:     u.Name,
		Email:    u.Email,
		Phone:    u.Phone,
		Password: u.Password,
		IsStore:  u.IsStore,
	}
}
