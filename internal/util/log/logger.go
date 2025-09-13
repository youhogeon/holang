package log

import (
	"fmt"
	"os"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger   *zap.Logger
	logLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
)

type Field = zap.Field

func init() {
	f, err := os.OpenFile("log.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}

	encCfg := zap.NewDevelopmentEncoderConfig()
	encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	encCfg.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString("\x1b[36m" + t.Format("15:04:05.000") + "\x1b[0m")
	}

	encCfg.EncodeCaller = func(c zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		const width = 30
		path := c.TrimmedPath()

		if len(path) > width {
			path = path[len(path)-width:]
		}

		enc.AppendString("\x1b[2m" + fmt.Sprintf("%-*s", width, path) + "\x1b[0m")
	}
	consoleEnc := zapcore.NewConsoleEncoder(encCfg)

	encCfg2 := zap.NewDevelopmentEncoderConfig()
	fileEncoder := zapcore.NewJSONEncoder(encCfg2)

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEnc, zapcore.AddSync(os.Stdout), logLevel),
		zapcore.NewCore(fileEncoder, zapcore.AddSync(f), logLevel),
	)

	logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
}

var fieldsStore, _ = lru.New[int64, []Field](1024 * 8)

func With(uid int64, fields ...Field) {
	fieldsStore.Add(uid, fields)
}

func Debug(msg string, fields ...Field) {
	logger.Debug(msg, fields...)
}

func DebugIfEnabled(msg string, fieldsFunc func() []Field) {
	if logger.Core().Enabled(zap.DebugLevel) {
		fields := fieldsFunc()
		logger.Debug(msg, fields...)
	}
}

func InfoIfEnabled(msg string, fieldsFunc func() []Field) {
	if logger.Core().Enabled(zap.InfoLevel) {
		fields := fieldsFunc()
		logger.Info(msg, fields...)
	}
}

func Info(msg string, fields ...Field) {
	logger.Info(msg, fields...)
}

func Warn(msg string, fields ...Field) {
	logger.Warn(msg, fields...)
}

func Error(msg string, fields ...Field) {
	logger.Error(msg, fields...)
}

func Fatal(msg string, fields ...Field) {
	logger.Fatal(msg, fields...)

	panic(msg)
}

func DebugWith(uid int64, msg string, fields ...Field) {
	if _fields, ok := fieldsStore.Get(uid); ok {
		logger.Debug(msg, append(fields, _fields...)...)
	} else {
		logger.Debug(msg, fields...)
	}
}

func InfoWith(uid int64, msg string, fields ...Field) {
	if _fields, ok := fieldsStore.Get(uid); ok {
		logger.Info(msg, append(fields, _fields...)...)
	} else {
		logger.Info(msg, fields...)
	}
}

func WarnWith(uid int64, msg string, fields ...Field) {
	if _fields, ok := fieldsStore.Get(uid); ok {
		logger.Warn(msg, append(fields, _fields...)...)
	} else {
		logger.Warn(msg, fields...)
	}
}

func ErrorWith(uid int64, msg string, fields ...Field) {
	if _fields, ok := fieldsStore.Get(uid); ok {
		logger.Error(msg, append(fields, _fields...)...)
	} else {
		logger.Error(msg, fields...)
	}
}

func Sync() error {
	return logger.Sync()
}

// EnableDebug switches log level to debug at runtime; call early (e.g., after parsing --debug flag).
func EnableDebug() { logLevel.SetLevel(zap.DebugLevel) }

// DisableDebug switches log level back to info.
func DisableDebug() { logLevel.SetLevel(zap.InfoLevel) }

func S(key string, value string) Field {
	return zap.String(key, value)
}

func I(key string, value int) Field {
	return zap.Int(key, value)
}

func I64(key string, value int64) Field {
	return zap.Int64(key, value)
}

func E(err error) Field {
	return zap.Error(err)
}

func A(key string, value any) Field {
	return zap.Any(key, value)
}

func F(key string, value float64) Field {
	return zap.Float64(key, value)
}

func D(key string, value time.Duration) Field {
	return zap.Duration(key, value)
}

func T(key string, value time.Time) Field {
	return zap.Time(key, value)
}
