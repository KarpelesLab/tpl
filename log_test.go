package tpl_test

import (
	"context"
	"errors"
	"testing"

	"github.com/KarpelesLab/tpl"
)

type mockLogger struct {
	warnCalled  bool
	warnMessage string
	warnArgs    []any
}

func (m *mockLogger) LogWarn(msg string, arg ...any) {
	m.warnCalled = true
	m.warnMessage = msg
	m.warnArgs = arg
}

func TestLogWarn(t *testing.T) {
	// Test with nil context (use context.TODO as recommended)
	tpl.LogWarn(context.TODO(), "test message", "arg1", 123)

	// Test with context but no logger
	ctx := context.Background()
	tpl.LogWarn(ctx, "test message", "arg1", 123)

	// Test with context and logger
	logger := &mockLogger{}
	ctxWithLogger := context.WithValue(ctx, tpl.TplCtxLog, logger)
	tpl.LogWarn(ctxWithLogger, "test message", "arg1", 123)

	if !logger.warnCalled {
		t.Errorf("LogWarn with logger did not call logger.LogWarn")
	}
	if logger.warnMessage != "test message" {
		t.Errorf("LogWarn message = %v, want %v", logger.warnMessage, "test message")
	}
	if len(logger.warnArgs) != 2 || logger.warnArgs[0] != "arg1" || logger.warnArgs[1] != 123 {
		t.Errorf("LogWarn args = %v, want %v", logger.warnArgs, []any{"arg1", 123})
	}
}

func TestLogError(t *testing.T) {
	err := errors.New("test error")

	// Test with nil context (use context.TODO as recommended)
	tpl.LogError(context.TODO(), err, "error occurred", "arg1", 123)

	// Test with context, no logger
	ctx := context.Background()
	tpl.LogError(ctx, err, "error occurred", "arg1", 123)

	// Test with tpl.Error
	tplErr := &tpl.Error{
		Message:  "template error",
		Template: "test.tpl",
		Line:     10,
		Char:     20,
	}
	tpl.LogError(ctx, tplErr, "template error occurred", "arg1", 123)
}

func TestLogDebug(t *testing.T) {
	// Test with nil context (use context.TODO as recommended)
	tpl.LogDebug(context.TODO(), "debug message", "arg1", 123)

	// Test with context, no logger
	ctx := context.Background()
	tpl.LogDebug(ctx, "debug message", "arg1", 123)
}
