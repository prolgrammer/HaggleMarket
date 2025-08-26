// Package middleware implements audit logging with buffering and batching to:
// 1. Protect against traffic spikes overwhelming the storage
// 2. Minimize contention between HTTP handlers and storage system
// 3. Ensure request processing isn't blocked by audit writes
package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/EM-Stawberry/Stawberry/config"
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/EM-Stawberry/Stawberry/internal/handler/helpers"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	maxBodySize = 10 * 1024
	tickrate    = time.Second * 5
)

type AuditMiddleware struct {
	cfg             *config.AuditConfig
	toFlushChan     chan []entity.AuditEntry // Pass batches to flusher (unbuffered to apply backpressure)
	logChan         chan entity.AuditEntry   // Main intake channel for audit entries
	closeSignalChan chan struct{}            // Coordinates graceful shutdown
	service         AuditService
	log             *zap.Logger
	wg              *sync.WaitGroup     // Tracks worker goroutines
	mutex           *sync.Mutex         // Protects double buffer swap
	buffer          []entity.AuditEntry // Active buffer being filled
	backupBuffer    []entity.AuditEntry // Buffer being flushed, then reused
}

// bodyLogWriter это обертка над ResponseWriter, которая записывает тело запроса в буфер
// для сохранения в логах.
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

type AuditService interface {
	Log(entries []entity.AuditEntry) error
}

func NewAuditMiddleware(cfg *config.AuditConfig, as AuditService, log *zap.Logger) *AuditMiddleware {
	am := &AuditMiddleware{
		cfg:             cfg,
		toFlushChan:     make(chan []entity.AuditEntry), // unbuffered on purpose
		logChan:         make(chan entity.AuditEntry, cfg.QueueSize),
		closeSignalChan: make(chan struct{}, 1),
		service:         as,
		log:             log,
		wg:              &sync.WaitGroup{},
		mutex:           &sync.Mutex{},
		buffer:          make([]entity.AuditEntry, 0, cfg.BatchSize),
		backupBuffer:    make([]entity.AuditEntry, 0, cfg.BatchSize),
	}

	am.wg.Add(cfg.WorkerPoolSize)
	for range cfg.WorkerPoolSize {
		go am.worker()
	}

	go am.flusher()

	return am
}

// Меняет местами буферы, очищает активный. Таким образом активный буфер доступен worker'ам для записи
// с минимальными задержками, в то время как flusher пишет логи в бд из бэкап буфера.
func (am *AuditMiddleware) swapAndFlush() {
	am.buffer, am.backupBuffer = am.backupBuffer[:0], am.buffer
	am.toFlushChan <- am.backupBuffer
}

func (am *AuditMiddleware) storeLogs(entries []entity.AuditEntry) {
	if err := am.service.Log(entries); err != nil {
		am.log.Error("Failed to log audit entries, retrying...", zap.Error(err))
		delay := time.Second * 250
		for range 3 {
			time.Sleep(delay)
			delay = delay * 2
			if err = am.service.Log(entries); err != nil {
				am.log.Error("Failed to log audit entries, retrying...", zap.Error(err))
			}
		}
		if err != nil {
			am.log.Error("Failed to log audit entries, giving up", zap.Error(err))
		}
	}
}

// flusher периодически сбрасывает логи в бд, триггерится либо по тикеру,
// либо по загруженности буфера на 90%.
// Запускается в единственном потоке, это позволяет сохранить очередность логов при
// их сбросе в бд и жонглировать двумя буферами без глубокого копирования.
func (am *AuditMiddleware) flusher() {
	ticker := time.NewTicker(tickrate)
	defer ticker.Stop()

	for {
		select {
		case toFlush := <-am.toFlushChan:
			am.storeLogs(toFlush)
		case <-ticker.C:
			if len(am.buffer) > 0 {
				am.buffer, am.backupBuffer = am.backupBuffer[:0], am.buffer
				am.storeLogs(am.backupBuffer)
			}
		case <-am.closeSignalChan:
			return
		}
	}
}

// worker постоянно принимает логи из канала, вычищает их от конфиденциальных данных
// и пихает в буфер, если там есть место.
// Если буфер заполнен - свапает два буфера местами с очисткой старого, и отправляет в flusher.
// Если flusher ещё занят записью в бд - виснет на записи в небуферизированный toFlushChan в swapAndFlush,
// это может случиться если бд медленная. (Если один из воркеров висит на канале, остальные повиснут на мьютексе)
func (am *AuditMiddleware) worker() {
	defer am.wg.Done()
	for entry := range am.logChan {
		sanitizeSensitiveData(entry.ReqBody) // Sanitize before buffering to avoid holding sensitive data
		sanitizeSensitiveData(entry.RespBody)

		am.mutex.Lock()
		if len(am.buffer) > int(0.9*float64(am.cfg.QueueSize)) {
			am.log.Warn("Audit log buffer full", zap.Int("size", len(am.buffer)))
			am.swapAndFlush()
		} else {
			am.buffer = append(am.buffer, entry)
		}
		am.mutex.Unlock()
	}
}

func (am *AuditMiddleware) Close() {
	close(am.logChan)
	am.wg.Wait()
	close(am.closeSignalChan)

	if len(am.buffer) > 0 {
		am.storeLogs(am.buffer)
	}
	if len(am.backupBuffer) > 0 {
		am.storeLogs(am.backupBuffer)
	}
}

func (am *AuditMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead {
			c.Next()
			return
		}

		receivedAt := time.Now()

		bodyBytes, _ := c.GetRawData()
		if len(bodyBytes) > maxBodySize {
			bodyBytes = append(bodyBytes[:maxBodySize], []byte("... [TRUNCATED]")...)
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		usrID, _ := helpers.UserIDContext(c)

		reqBody := make(map[string]interface{})
		if len(bodyBytes) > 0 {
			if err := json.Unmarshal(bodyBytes, &reqBody); err != nil {
				am.log.Error("Failed to unmarshal request body", zap.Error(err))
			}
		}

		respBodyBytes := blw.body.Bytes()
		if len(respBodyBytes) > maxBodySize {
			respBodyBytes = append(respBodyBytes[:maxBodySize], []byte("... [TRUNCATED]")...)
		}

		respBody := make(map[string]interface{})
		if len(respBodyBytes) > 0 {
			if err := json.Unmarshal(respBodyBytes, &respBody); err != nil {
				am.log.Error("Failed to unmarshal response body", zap.Error(err))
			}
		}

		logE := entity.AuditEntry{
			Method:     c.Request.Method,
			Url:        c.Request.URL.Path,
			RespStatus: c.Writer.Status(),
			UserID:     usrID,
			IP:         c.ClientIP(),
			UserRole:   getRole(c),
			ReceivedAt: receivedAt,
			ReqBody:    reqBody,
			RespBody:   respBody,
		}

		select {
		case am.logChan <- logE:
		default:
			am.log.Warn("Audit log channel full, dropping audit log entry",
				zap.String("path", logE.Url),
				zap.Int("status", logE.RespStatus))
		}
	}
}

func getRole(c *gin.Context) string {
	isStore, _ := helpers.UserIsStoreContext(c)
	if isStore {
		return "shop"
	}
	isAdmin, _ := helpers.UserIsAdminContext(c)
	if isAdmin {
		return "admin"
	}
	return "user"
}

func sanitizeSensitiveData(data map[string]interface{}) {
	if data == nil {
		return
	}
	sensitiveFields := []string{"password", "fingerprint", "refresh_token", "access_token"}
	for _, field := range sensitiveFields {
		if _, ok := data[field]; ok {
			data[field] = "[REDACTED]"
		}
	}
}
