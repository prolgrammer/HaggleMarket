package entity

import "time"

type AuditEntry struct {
	Method     string
	Url        string
	RespStatus int
	UserID     uint
	IP         string
	UserRole   string
	ReceivedAt time.Time
	ReqBody    map[string]interface{}
	RespBody   map[string]interface{}
}
