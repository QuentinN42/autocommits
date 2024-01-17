package logger

import (
	"context"
	"fmt"
	"os"

	"github.com/sagikazarmark/slog-shim"
)

type handler struct {
	slog.Handler
}

func (h handler) handle(ctx context.Context, r slog.Record) error {
	if traceID, ok := ctx.Value("trace_id").(string); ok {
		r.Add("trace_id", slog.StringValue(traceID))
	}

	return h.Handler.Handle(ctx, r)
}

const (
	LevelTrace   = slog.Level(-8)
	LevelDebug   = slog.LevelDebug
	LevelInfo    = slog.LevelInfo
	LevelWarning = slog.LevelWarn
	LevelError   = slog.LevelError
	LevelFatal   = slog.Level(12)
)

func replaceAttr(groups []string, a slog.Attr) slog.Attr {
	if a.Key != slog.LevelKey {
		return a
	}

	level := a.Value.Any().(slog.Level)
	switch {
	case level < LevelDebug:
		a.Value = slog.StringValue("TRACE")
	case level < LevelInfo:
		a.Value = slog.StringValue("DEBUG")
	case level < LevelWarning:
		a.Value = slog.StringValue("INFO")
	case level < LevelError:
		a.Value = slog.StringValue("WARNING")
	case level < LevelFatal:
		a.Value = slog.StringValue("ERROR")
	default:
		a.Value = slog.StringValue("FATAL")
	}
	return a
}

var Logger = slog.New(slog.NewTextHandler(
	os.Stdout,
	&slog.HandlerOptions{
		Level:       LevelTrace,
		ReplaceAttr: replaceAttr,
	},
))

var (
	Trace = func(ctx context.Context, format string, args ...interface{}) {
		Logger.Log(ctx, LevelTrace, fmt.Sprintf(format, args...))
	}
	Debug = func(ctx context.Context, format string, args ...interface{}) {
		Logger.Log(ctx, LevelDebug, fmt.Sprintf(format, args...))
	}
	Info = func(ctx context.Context, format string, args ...interface{}) {
		Logger.Log(ctx, LevelInfo, fmt.Sprintf(format, args...))
	}
	Warning = func(ctx context.Context, format string, args ...interface{}) {
		Logger.Log(ctx, LevelWarning, fmt.Sprintf(format, args...))
	}
	Error = func(ctx context.Context, format string, args ...interface{}) {
		Logger.Log(ctx, LevelError, fmt.Sprintf(format, args...))
	}
	Fatal = func(ctx context.Context, format string, args ...interface{}) {
		Logger.Log(ctx, LevelFatal, fmt.Sprintf(format, args...))
	}
)
