package tpl_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/KarpelesLab/tpl"
)

func TestParseErrors(t *testing.T) {
	e := tpl.New()
	ctx := context.Background()
	
	// Test ParseAndWrite with non-existent template
	outBuf := &bytes.Buffer{}
	err := e.ParseAndWrite(ctx, "nonexistent", outBuf)
	if !errors.Is(err, tpl.ErrTplNotFound) {
		t.Errorf("ParseAndWrite with non-existent template error = %v, want %v", err, tpl.ErrTplNotFound)
	}
	
	// Test ParseAndReturn with non-existent template
	_, err = e.ParseAndReturn(ctx, "nonexistent")
	if !errors.Is(err, tpl.ErrTplNotFound) {
		t.Errorf("ParseAndReturn with non-existent template error = %v, want %v", err, tpl.ErrTplNotFound)
	}
	
	// Test with canceled context
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel() // Cancel the context immediately
	
	err = e.ParseAndWrite(cancelCtx, "nonexistent", outBuf)
	if err == nil || !errors.Is(err, context.Canceled) {
		t.Errorf("ParseAndWrite with canceled context error = %v, want context.Canceled", err)
	}
	
	_, err = e.ParseAndReturn(cancelCtx, "nonexistent")
	if err == nil || !errors.Is(err, context.Canceled) {
		t.Errorf("ParseAndReturn with canceled context error = %v, want context.Canceled", err)
	}
}

func TestCompileErrors(t *testing.T) {
	e := tpl.New()
	ctx := context.Background()
	
	// Test Compile with invalid template (missing main)
	err := e.Compile(ctx)
	if err == nil {
		t.Errorf("Compile with invalid template should return error")
	}
	
	// Test Compile with canceled context
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel() // Cancel the context immediately
	
	err = e.Compile(cancelCtx)
	if err == nil || !errors.Is(err, context.Canceled) {
		t.Errorf("Compile with canceled context error = %v, want context.Canceled", err)
	}
	
	// Test Compile with syntax error in template
	e.Raw.TemplateData["main"] = "{{foreach}}" // Missing required parameters
	err = e.Compile(ctx)
	if err == nil {
		t.Errorf("Compile with syntax error should return error")
	}
	
	// Check if it's a tpl.Error
	var tplErr *tpl.Error
	if !errors.As(err, &tplErr) {
		t.Errorf("Error should be of type *tpl.Error, got %T", err)
	}
}