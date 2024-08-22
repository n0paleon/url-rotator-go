package logger

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	L *zap.SugaredLogger
)

func NewLogger(cfg *viper.Viper) *zap.SugaredLogger {
	logPath := []string{"stdout"}
	errPath := []string{"stderr"}

	options := zap.Config{
		Level:       zap.NewAtomicLevelAt(zapcore.Level(cfg.GetInt("log.level"))),
		Development: cfg.GetBool("app.development"),
		Encoding:    cfg.GetString("log.encoding"),
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},

		OutputPaths:      logPath,
		ErrorOutputPaths: errPath,
	}

	logger, err := options.Build()
	if err != nil {
		panic(err)
	}

	defer logger.Sync()

	L = logger.Sugar()

	return logger.Sugar()
}
