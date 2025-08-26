package middleware

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	colorGreen   = "\033[38;5;82m"    // Bright green for GET and success codes
	colorBlue    = "\033[38;5;39m"    // Bright blue for POST
	colorYellow  = "\033[38;5;220m"   // Bright yellow for PUT and warning codes
	colorRed     = "\033[38;5;196m"   // Bright red for DELETE and error codes
	colorCyan    = "\033[38;5;51m"    // Cyan for PATCH and arrows
	colorMagenta = "\033[38;5;201m"   // Magenta for OPTIONS
	colorPink    = "\033[1;38;5;219m" // Bold pink for metadata and handler names
	colorReset   = "\033[0m"          // Reset color
)

// methodColor returns method string with appropriate color based on HTTP method
func methodColor(method string) string {
	switch method {
	case "GET":
		return colorGreen + method + colorReset
	case "POST":
		return colorBlue + method + colorReset
	case "PUT":
		return colorYellow + method + colorReset
	case "DELETE":
		return colorRed + method + colorReset
	case "PATCH":
		return colorCyan + method + colorReset
	case "OPTIONS":
		return colorMagenta + method + colorReset
	default:
		return colorReset + method + colorReset
	}
}

// statusCodeColor returns status code with appropriate color based on HTTP status
func statusCodeColor(code int) string {
	switch {
	case code >= 200 && code < 300:
		return colorGreen + fmt.Sprintf("%d", code) + colorReset
	case code >= 300 && code < 400:
		return colorBlue + fmt.Sprintf("%d", code) + colorReset
	case code >= 400 && code < 500:
		return colorYellow + fmt.Sprintf("%d", code) + colorReset
	default:
		return colorRed + fmt.Sprintf("%d", code) + colorReset
	}
}

// ginRoutesRegex matches Gin's debug route definitions
var ginRoutesRegex = regexp.MustCompile(
	`(GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS|CONNECT|TRACE)\s+(.+)\s+--> (.+) \((\d+) handlers\)`,
)

func formatGinDebugMessage(s string) string {
	if matches := ginRoutesRegex.FindStringSubmatch(s); len(matches) == 5 {
		method := matches[1]
		path := matches[2]
		handler := matches[3]

		handlerParts := strings.Split(handler, ".")
		shortHandler := handlerParts[len(handlerParts)-1]

		prefix := fmt.Sprintf("Route: %s ", methodColor(method))

		visiblePrefix := fmt.Sprintf("Route: %s ", method)
		visibleLength := len(visiblePrefix) + len(path)
		arrowPosition := 50
		padding := arrowPosition - visibleLength
		if padding < 1 {
			padding = 1
		}

		return colorPink + "[Strawberry]" + colorPink + " " + prefix + path +
			strings.Repeat(" ", padding) + colorCyan + "→" + colorReset + " " +
			colorReset + shortHandler + colorReset
	}

	s = strings.TrimPrefix(s, "[GIN-debug] ")

	if strings.HasPrefix(s, "Listening and serving HTTP") {
		return colorGreen + "Server started: " + s + colorReset
	}

	if strings.HasPrefix(s, "redirecting request") {
		return colorBlue + "Redirect: " + s + colorReset
	}

	if strings.HasPrefix(s, "Loading HTML Templates") {
		return colorCyan + "Templates loaded: " + s + colorReset
	}

	if strings.Contains(s, "router") {
		return colorYellow + "Router: " + s + colorReset
	}

	return s
}

// zapWriter implements io.Writer interface to redirect Gin logs to Zap
type zapWriter struct {
	logger *zap.Logger
}

func (w zapWriter) Write(p []byte) (n int, err error) {
	s := strings.TrimSpace(string(p))

	if strings.Contains(s, "[GIN-debug]") {
		message := formatGinDebugMessage(s)
		w.logger.Debug(message, zap.String("component", "gin"))
	} else if strings.Contains(s, "[GIN]") {
		message := strings.TrimSpace(strings.Replace(s, "[GIN]", "", 1))
		w.logger.Debug(message, zap.String("component", "gin"))
	} else {
		w.logger.Debug(s)
	}

	return len(p), nil
}

// SetupGinWithZap configures Gin to use Zap as its logger
func SetupGinWithZap(logger *zap.Logger) {
	gin.DefaultWriter = &zapWriter{logger: logger}
	gin.DefaultErrorWriter = &zapWriter{logger: logger.WithOptions(zap.IncreaseLevel(zapcore.ErrorLevel))}
}

// ZapLogger returns a gin.HandlerFunc middleware that logs requests using Zap
func ZapLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start)
		status := c.Writer.Status()
		method := c.Request.Method
		ip := c.ClientIP()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		if len(query) > 0 {
			path = path + "?" + query
		}

		message := fmt.Sprintf("%s %s %s%s%s %s",
			methodColor(method),
			statusCodeColor(status),
			colorReset,
			latency.String(),
			colorReset,
			path,
		)

		fields := []zap.Field{zap.String("ip", ip)}

		if errorMessage != "" {
			fields = append(fields, zap.String("error", errorMessage))
		}

		switch {
		case status >= 500:
			logger.Error(message, fields...)
		case status >= 400:
			logger.Warn(message, fields...)
		default:
			logger.Info(message, fields...)
		}
	}
}

// ZapRecovery returns a gin.HandlerFunc middleware that recovers from panics
func ZapRecovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("ip", c.ClientIP()),
				)
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}
