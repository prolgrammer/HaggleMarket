package handler

import (
	"net/http"
	"time"

	"github.com/EM-Stawberry/Stawberry/internal/handler/helpers"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now().Unix(),
	})
}

func (h *HealthHandler) authCheck(c *gin.Context) {
	userID, ok := helpers.UserIDContext(c)
	var status string
	if ok {
		status = "UserID found"
	} else {
		status = "UserID not found"
	}
	isStore, ok := helpers.UserIsStoreContext(c)

	c.JSON(http.StatusOK, gin.H{
		"userID":       userID,
		"status":       status,
		"isStore":      isStore,
		"isStoreFound": ok,
		"time":         time.Now().Unix(),
	})
}
