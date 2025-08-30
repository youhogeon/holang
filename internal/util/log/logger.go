package log

import (
	"fmt"
	"os"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

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
		zapcore.NewCore(consoleEnc, zapcore.AddSync(os.Stdout), zap.DebugLevel),
		zapcore.NewCore(fileEncoder, zapcore.AddSync(f), zap.DebugLevel),
	)

	logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
}

var fieldsStore, _ = lru.New[int64, []zap.Field](1024 * 8)

func With(uid int64, fields ...zap.Field) {
	fieldsStore.Add(uid, fields)
}

func Debug(msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}

func DebugIfEnabled(msg string, fieldsFunc func() []zap.Field) {
	if logger.Core().Enabled(zap.DebugLevel) {
		fields := fieldsFunc()
		logger.Debug(msg, fields...)
	}
}

func Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	logger.Fatal(msg, fields...)

	panic(msg)
}

func DebugWith(uid int64, msg string, fields ...zap.Field) {
	if _fields, ok := fieldsStore.Get(uid); ok {
		logger.Debug(msg, append(fields, _fields...)...)
	} else {
		logger.Debug(msg, fields...)
	}
}

func InfoWith(uid int64, msg string, fields ...zap.Field) {
	if _fields, ok := fieldsStore.Get(uid); ok {
		logger.Info(msg, append(fields, _fields...)...)
	} else {
		logger.Info(msg, fields...)
	}
}

func WarnWith(uid int64, msg string, fields ...zap.Field) {
	if _fields, ok := fieldsStore.Get(uid); ok {
		logger.Warn(msg, append(fields, _fields...)...)
	} else {
		logger.Warn(msg, fields...)
	}
}

func ErrorWith(uid int64, msg string, fields ...zap.Field) {
	if _fields, ok := fieldsStore.Get(uid); ok {
		logger.Error(msg, append(fields, _fields...)...)
	} else {
		logger.Error(msg, fields...)
	}
}

func Sync() error {
	return logger.Sync()
}

func Common(botId string, uid int64, code string) []zap.Field {
	return []zap.Field{
		S("botId", botId),
		I("uid", uid),
		S("code", code),
	}
}

func S(key string, value string) zap.Field {
	return zap.String(key, value)
}

func I(key string, value int64) zap.Field {
	return zap.Int64(key, value)
}

func Int(key string, value int) zap.Field {
	return zap.Int(key, value)
}

func E(err error) zap.Field {
	return zap.Error(err)
}

func A(key string, value any) zap.Field {
	return zap.Any(key, value)
}

func F(key string, value float64) zap.Field {
	return zap.Any(key, value)
}

func D(key string, value time.Duration) zap.Field {
	return zap.Duration(key, value)
}

func T(key string, value time.Time) zap.Field {
	return zap.Time(key, value)
}
