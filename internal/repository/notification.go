package repository

import (
	"github.com/jmoiron/sqlx"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
)

type NotificationRepository struct {
	db *sqlx.DB
}

func NewNotificationRepository(db *sqlx.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) SelectUserNotifications(
	id string,
	offset, limit int,
) ([]entity.Notification, int, error) {
	var total int64

	_ = id
	_ = offset
	_ = limit

	return nil, int(total), nil
}
