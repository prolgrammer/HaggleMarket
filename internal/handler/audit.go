package handler

import (
	"context"
	"math"
	"strconv"
	"time"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/EM-Stawberry/Stawberry/internal/handler/dto"
	"github.com/gin-gonic/gin"
)

type AuditQueryParams struct {
	From   time.Time
	To     time.Time
	UID    uint
	Limit  int
	Page   int
	Offset int
}

type AuditService interface {
	DisplayLogs(context.Context, time.Time, time.Time, uint, int, int) ([]entity.AuditEntry, int, error)
}

type AuditHandler struct {
	auditService AuditService
}

func NewAuditHandler(as AuditService) *AuditHandler {
	return &AuditHandler{
		auditService: as,
	}
}

// DisplayLogs retrieves audit logs with filtering and pagination
// @Summary Get audit logs
// @Description Retrieve audit trail entries with time range filtering and pagination
// @Tags Audit
// @Accept  json
// @Produce  json
// @Param from query string false "Start time in RFC3339 format (default: 24h ago)" Format(date-time)
// @Param to query string false "End time in RFC3339 format (default: now)" Format(date-time)
// @Param uid query integer false "Filter by user ID"
// @Param limit query integer false "Items per page (default 100)" minimum(1) maximum(500)
// @Param page query integer false "Page number (default 1)" minimum(1)
// @Success 200 {object} map[string]interface{} "Returns paginated audit logs"
// @Failure 400 {object} apperror.AppError "Invalid request parameters"
// @Failure 500 {object} apperror.AppError "Internal server error"
// @Router /audit/logs [get]
func (h *AuditHandler) DisplayLogs(c *gin.Context) {
	params, err := parseAuditQueryParams(c)
	if err != nil {
		c.Error(err)
		return
	}

	logsEnt, total, err := h.auditService.DisplayLogs(
		c.Request.Context(),
		params.From,
		params.To,
		params.UID,
		params.Limit,
		params.Offset,
	)
	if err != nil {
		c.Error(apperror.New(apperror.InternalError, err.Error(), err))
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(params.Limit)))

	c.JSON(200, gin.H{
		"total_logs":   total,
		"current_page": params.Page,
		"per_page":     params.Limit,
		"total_pages":  totalPages,
		"data":         dto.FormResponse(logsEnt),
	})
}

func parseAuditQueryParams(c *gin.Context) (*AuditQueryParams, error) {
	from := c.DefaultQuery("from", time.Now().AddDate(0, 0, -1).Format(time.RFC3339))
	fromT, err := time.Parse(time.RFC3339, from)
	if err != nil {
		return nil, apperror.New(apperror.BadRequest, "invalid from format", err)
	}

	to := c.DefaultQuery("to", time.Now().Format(time.RFC3339))
	toT, err := time.Parse(time.RFC3339, to)
	if err != nil {
		return nil, apperror.New(apperror.BadRequest, "invalid to format", err)
	}

	var uid uint
	if uidStr := c.Query("uid"); uidStr != "" {
		id, err := strconv.Atoi(uidStr)
		if err != nil {
			return nil, apperror.New(apperror.BadRequest, "invalid uid format", err)
		}
		uid = uint(id)
	}

	if fromT.After(toT) {
		return nil, apperror.New(apperror.BadRequest, "'from' must be before 'to'", nil)
	}

	if uid == 0 && toT.Sub(fromT) > 365*24*time.Hour { // potentially reduce this from a year to smth more reasonable
		return nil, apperror.New(
			apperror.BadRequest, "time window cannot exceed 365 days without providing a UID", nil,
		)
	}

	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 500 {
		limit = 100
	}

	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	return &AuditQueryParams{
		From:   fromT,
		To:     toT,
		UID:    uid,
		Limit:  limit,
		Page:   page,
		Offset: (page - 1) * limit,
	}, nil
}
