package notification

import "github.com/EM-Stawberry/Stawberry/internal/domain/entity"

type Repository interface {
	SelectUserNotifications(id string, offset, limit int) ([]entity.Notification, int, error)
}

type Service struct {
	notificationRepository Repository
}

func NewService(notificationRepository Repository) *Service {
	return &Service{notificationRepository}
}

func (ns *Service) GetNotification(id string, offset int, limit int) ([]entity.Notification, int, error) {
	return ns.notificationRepository.SelectUserNotifications(id, offset, limit)
}
