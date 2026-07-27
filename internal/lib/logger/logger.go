package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/artrctx/shuffle-core/internal/env"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/log/global"
	otellog "go.opentelemetry.io/otel/sdk/log"
)

type ContextKey string

const LoggerCtxKey ContextKey = "slogger"

type LoggerProvider struct {
	provider *otellog.LoggerProvider
}

// if app env is development will use stdio
func Initialize(ctx context.Context) (*LoggerProvider, error) {
	if env.AppEnv == "development" {
		return &LoggerProvider{}, nil
	}

	// Create OTLP HTTP exporter
	exporter, err := otlploghttp.New(ctx,
		otlploghttp.WithEndpoint("us.i.posthog.com"),
		otlploghttp.WithURLPath("/i/v1/logs"),
		otlploghttp.WithHeaders(map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", env.PostHogToken),
		}),
	)
	if err != nil {
		return nil, err
	}

	os.Setenv("OTEL_SERVICE_NAME", env.AppName)

	stdoutExporter, err := stdoutlog.New()
	if err != nil {
		return nil, err
	}

	// Create logger provider
	loggerProvider := otellog.NewLoggerProvider(
		otellog.WithProcessor(otellog.NewBatchProcessor(exporter)),
		otellog.WithProcessor(otellog.NewSimpleProcessor(stdoutExporter)),
	)

	// set defaults
	global.SetLoggerProvider(loggerProvider)
	slog.SetDefault(otelslog.NewLogger(fmt.Sprintf("%s-logger", env.AppName)))

	return &LoggerProvider{loggerProvider}, nil
}

func (lp *LoggerProvider) Shutdown(ctx context.Context) error {
	if lp.provider == nil {
		return nil
	}
	return lp.provider.Shutdown(ctx)
}

func FromCtx(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(LoggerCtxKey).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

func CtxWithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, LoggerCtxKey, l)
}
