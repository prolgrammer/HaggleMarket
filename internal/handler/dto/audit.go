package dto

import (
	"time"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
)

type AuditEntry struct {
	Method     string                 `json:"method"`
	Url        string                 `json:"url"`
	RespStatus int                    `json:"resp_status"`
	UserID     uint                   `json:"user_id"`
	IP         string                 `json:"user_ip"`
	UserRole   string                 `json:"user_role"`
	ReceivedAt time.Time              `json:"received_at"`
	ReqBody    map[string]interface{} `json:"req_body"`
	RespBody   map[string]interface{} `json:"resp_body"`
}

func FormResponse(entities []entity.AuditEntry) []AuditEntry {
	resp := make([]AuditEntry, len(entities))
	for i, e := range entities {
		resp[i] = AuditEntry{
			Method:     e.Method,
			Url:        e.Url,
			RespStatus: e.RespStatus,
			UserID:     e.UserID,
			IP:         e.IP,
			UserRole:   e.UserRole,
			ReceivedAt: e.ReceivedAt,
			ReqBody:    e.ReqBody,
			RespBody:   e.RespBody,
		}
	}
	return resp
}
