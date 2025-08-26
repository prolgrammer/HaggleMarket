package audit

import (
	"context"
	"time"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
)

type AuditRepository interface {
	LogStore([]entity.AuditEntry) error
	GetLogs(context.Context, time.Time, time.Time, uint, int, int) ([]entity.AuditEntry, int, error)
}

type AuditService struct {
	auditRepository AuditRepository
}

func NewAuditService(ar AuditRepository) *AuditService {
	return &AuditService{
		auditRepository: ar,
	}
}

func (as *AuditService) Log(entries []entity.AuditEntry) error {
	return as.auditRepository.LogStore(entries)
}

func (as *AuditService) DisplayLogs(
	ctx context.Context,
	fromT,
	toT time.Time,
	uid uint,
	limit,
	offset int,
) (
	[]entity.AuditEntry,
	int,
	error,
) {
	return as.auditRepository.GetLogs(ctx, fromT, toT, uid, limit, offset)
}
