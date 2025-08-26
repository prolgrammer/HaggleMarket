package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/EM-Stawberry/Stawberry/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	ColorGray   = "\033[38;5;246m" // Gray for timestamps and filenames
	ColorCyan   = "\033[36m"       // Cyan for debug level
	ColorGreen  = "\033[32m"       // Green for info level
	ColorYellow = "\033[33m"       // Yellow for warn level
	ColorRed    = "\033[31m"       // Red for error level
	ColorReset  = "\033[0m"        // Reset to default color
)

// DisabledCore wraps a zapcore.Core and disables JSON fields in development mode
type DisabledCore struct {
	zapcore.Core
}

// With overrides the With method to ignore all fields
func (c DisabledCore) With(fields []zapcore.Field) zapcore.Core {
	return c.Core.With(fields)
}

// Check verifies if the entry should be logged and returns a CheckedEntry with empty fields
func (c DisabledCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

// Write overrides the Write method to ignore all fields
func (c DisabledCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	return c.Core.Write(ent, fields)
}

// coloredTimeEncoder formats timestamps with gray color
func coloredTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	timestamp := t.Format("2006-01-02 15:04:05.000")
	coloredTime := fmt.Sprintf("%s%s%s", ColorGray, timestamp, ColorReset)
	enc.AppendString(coloredTime)
}

// coloredCallerEncoder formats caller information with gray color
func coloredCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	if !caller.Defined {
		enc.AppendString("undefined")
		return
	}
	path := caller.TrimmedPath()
	coloredCaller := fmt.Sprintf("%s%s%s", ColorGray, path, ColorReset)
	enc.AppendString(coloredCaller)
}

func getEncoder(isJSON bool) zapcore.Encoder {
	if isJSON {
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		return zapcore.NewJSONEncoder(encoderConfig)
	}

	// Настройка консольного энкодера с цветами
	encoderConfig := zap.NewDevelopmentEncoderConfig()

	// Используем кастомные цветные энкодеры
	encoderConfig.EncodeTime = coloredTimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoderConfig.EncodeCaller = coloredCallerEncoder
	encoderConfig.EncodeName = zapcore.FullNameEncoder
	encoderConfig.ConsoleSeparator = "\t"

	// Создаем оригинальный консольный энкодер
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func SetupLogger(env string) *zap.Logger {
	var level zapcore.Level
	var isJSON bool

	switch env {
	case config.EnvDev, config.EnvTest:
		level = zap.DebugLevel
		isJSON = false
	case config.EnvProd:
		level = zap.InfoLevel
		isJSON = true
	default:
		level = zap.InfoLevel
		isJSON = false
	}

	if env == config.EnvDev || env == config.EnvTest {
		isJSON = false
	}

	core := zapcore.NewCore(
		getEncoder(isJSON),
		zapcore.AddSync(os.Stdout),
		level,
	)

	if env == config.EnvDev || env == config.EnvTest {
		core = DisabledCore{Core: core}
	}

	zap.ReplaceGlobals(zap.New(core, zap.AddCaller()))

	return zap.New(core, zap.AddCaller())
}
