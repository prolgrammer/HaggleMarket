package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
)

// jumping through hoops to scan JSONB into a map[string]interface{}
type JSONBMap map[string]interface{}

func (j *JSONBMap) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &j)
}

func (j *JSONBMap) Value() (driver.Value, error) {
	return json.Marshal(j)
}

type AuditEntry struct {
	Method     string    `db:"method"`
	Url        string    `db:"url"`
	RespStatus int       `db:"resp_status"`
	UserID     uint      `db:"user_id"`
	IP         string    `db:"user_ip"`
	UserRole   string    `db:"user_role"`
	ReceivedAt time.Time `db:"received_at"`
	ReqBody    JSONBMap  `db:"req_body"`
	RespBody   JSONBMap  `db:"resp_body"`
}

func ConvertEntityToAuditEntries(entries []entity.AuditEntry) []AuditEntry {
	models := make([]AuditEntry, len(entries))
	for i, entry := range entries {
		models[i] = AuditEntry{
			Method:     entry.Method,
			Url:        entry.Url,
			RespStatus: entry.RespStatus,
			UserID:     entry.UserID,
			IP:         entry.IP,
			UserRole:   entry.UserRole,
			ReceivedAt: entry.ReceivedAt,
			ReqBody:    JSONBMap(entry.ReqBody),
			RespBody:   JSONBMap(entry.RespBody),
		}
	}
	return models
}

type AuditEntryMeta struct {
	AuditEntry
	TotalCount int `db:"total_count"`
}

func ConvertAuditEntriesToEntity(entries []AuditEntryMeta) ([]entity.AuditEntry, int) {
	if len(entries) == 0 {
		return nil, 0
	}

	entities := make([]entity.AuditEntry, len(entries))
	for i, entry := range entries {
		entities[i] = entity.AuditEntry{
			Method:     entry.Method,
			Url:        entry.Url,
			RespStatus: entry.RespStatus,
			UserID:     entry.UserID,
			IP:         entry.IP,
			UserRole:   entry.UserRole,
			ReceivedAt: entry.ReceivedAt,
			ReqBody:    map[string]interface{}(entry.ReqBody),
			RespBody:   map[string]interface{}(entry.RespBody),
		}
	}
	return entities, entries[0].TotalCount
}
