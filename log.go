// Package tpl provides a template engine with string interpolation and control structures.
package tpl

import (
	"context"
	"log/slog"
)

// TplCtxValue is used as a context key for template-related values.
type TplCtxValue int

const (
	// TplCtxLog is the context key for logging interface
	TplCtxLog TplCtxValue = iota + 1
)

// CtxLog defines an interface for context-aware logging.
type CtxLog interface {
	// LogWarn logs a warning message with the given arguments
	LogWarn(msg string, arg ...any)
}

// LogWarn logs a warning message using the logger from context if available,
// otherwise falls back to the default structured logger.
func LogWarn(ctx context.Context, msg string, arg ...any) {
	if ctx == nil {
		// If context is nil, use default logger
		slog.Warn(msg, arg...)
		return
	}

	if c, ok := ctx.Value(TplCtxLog).(CtxLog); ok {
		// Use context-specific logger
		c.LogWarn(msg, arg...)
	} else {
		// Use default structured logger
		slog.Warn(msg, append([]any{"component", "tpl"}, arg...)...)
	}
}

// LogError logs an error with context information.
// It's designed to be used with template execution errors.
func LogError(ctx context.Context, err error, msg string, arg ...any) {
	// Use default logger if context is nil
	if ctx == nil {
		slog.Error(msg, append([]any{"error", err, "component", "tpl"}, arg...)...)
		return
	}

	// Try to use context-specific logger first
	if c, ok := ctx.Value(TplCtxLog).(interface{ LogError(string, error, ...any) }); ok {
		c.LogError(msg, err, arg...)
		return
	}

	// Fall back to default structured logger
	attrs := []any{"error", err, "component", "tpl"}

	// Add error details if available
	if tplErr, ok := err.(*Error); ok {
		attrs = append(attrs,
			"template", tplErr.Template,
			"line", tplErr.Line,
			"position", tplErr.Char,
		)
	}

	slog.Error(msg, append(attrs, arg...)...)
}

// LogDebug logs a debug message with source information.
func LogDebug(ctx context.Context, msg string, arg ...any) {
	// Use default logger if context is nil
	if ctx == nil {
		slog.Debug(msg, append([]any{"component", "tpl"}, arg...)...)
		return
	}

	// Try to use context-specific logger first
	if c, ok := ctx.Value(TplCtxLog).(interface{ LogDebug(string, ...any) }); ok {
		c.LogDebug(msg, arg...)
		return
	}

	// Fall back to default structured logger
	slog.Debug(msg, append([]any{"component", "tpl"}, arg...)...)
}
