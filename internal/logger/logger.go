package logger

import (
	"fmt"
	"time"

	"github.com/TheZeroSlave/zapsentry"
	"github.com/getsentry/sentry-go"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ConfigureLogger embed Sentry configuration on given zap.Logger.
func ConfigureLogger(client *sentry.Client, logger *zap.Logger) (logr.Logger, error) {
	cfg := zapsentry.Configuration{
		Tags:              nil,
		LoggerNameKey:     "",
		DisableStacktrace: false,
		Level:             zapcore.ErrorLevel, // when to send message to sentry
		EnableBreadcrumbs: true,               // enable sending breadcrumbs to Sentry
		BreadcrumbLevel:   zapcore.InfoLevel,  // at what level should we sent breadcrumbs to sentry
		FlushTimeout:      time.Second * 5,
		Hub:               nil, // use the default one
	}

	// to use breadcrumbs feature - create new scope explicitly
	logger = logger.With(zapsentry.NewScope())

	core, err := zapsentry.NewCore(cfg, zapsentry.NewSentryClientFromClient(client))
	if err != nil {
		return logr.Logger{}, fmt.Errorf("logger: failed to configure zap logger: %w", err)
	}

	// embed sentry into given logger
	logger, err = zapsentry.AttachCoreToLogger(core, logger), nil
	if err != nil {
		return logr.Logger{}, fmt.Errorf("logger: failed to configure zapsentry logger: %w", err)
	}

	return zapr.NewLogger(logger), nil
}
