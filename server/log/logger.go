package log

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/config"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/http/data"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Logger *zap.SugaredLogger
)

func InitLogger() {
	encoder := getEncoder()
	var core zapcore.Core
	consoleCore := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), getLogLevel())
	if config.Config.LogConfig.FilePath == "" {
		core = consoleCore
	} else {
		writeSyncer := getLogWriter()
		fileCore := zapcore.NewCore(encoder, writeSyncer, getLogLevel())
		core = zapcore.NewTee(fileCore, consoleCore)
	}
	Logger = zap.New(core, zap.AddCaller()).Sugar()
}

type ctxKey struct{}

func WithContext(ctx context.Context) *zap.SugaredLogger {
	if ctx == nil {
		return Logger
	}
	if l, ok := ctx.Value(ctxKey{}).(*zap.SugaredLogger); ok && l != nil {
		fmt.Println("logger found in context, will use it")
		return l
	}
	return Logger
}

func TraceLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		span := trace.SpanFromContext(c.Request.Context())
		sc := span.SpanContext()

		logger := Logger
		if sc.IsValid() {
			fmt.Println("span valid, will log with trace")
			logger = Logger.With(
				"trace_id", sc.TraceID().String(),
				"span_id", sc.SpanID().String(),
				"service_name", data.ServiceName,
			)
		}

		ctx := context.WithValue(c.Request.Context(), ctxKey{}, logger)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func getLogLevel() zapcore.Level {
	level := zapcore.Level(0)
	if config.Config.LogConfig.Level != "" {
		if err := level.UnmarshalText([]byte(config.Config.LogConfig.Level)); err != nil {
			level = zapcore.DebugLevel // fallback to DebugLevel if there's an error
		}
	} else {
		level = zapcore.DebugLevel // default to DebugLevel if Level is not set
	}
	return level
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Local().Format("2006-01-02 15:04:05"))
	}
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter() zapcore.WriteSyncer {
	if config.Config.LogConfig.FilePath == "" {
		return zapcore.AddSync(os.Stdout)
	}
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to get current working directory: %v\n", err)
		panic(err)
	}
	logPath := filepath.Join(cwd, config.Config.LogConfig.FilePath)
	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		panic(fmt.Sprintf("Failed to create directories: %v", err))
	}
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	return zapcore.AddSync(file)
}
