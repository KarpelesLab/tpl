package tpl_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/KarpelesLab/tpl"
)

func TestQueryEscapeAny(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		value    any
		expected string
	}{
		{"string", tpl.NewValue("hello world"), "hello+world"},
		{"special_chars", tpl.NewValue("a=b&c=d"), "a%3Db%26c%3Dd"},
		{"unicode", tpl.NewValue("日本語"), "%E6%97%A5%E6%9C%AC%E8%AA%9E"},
		{"bytes", tpl.NewValue([]byte("hello")), "hello"},
		{"buffer", tpl.NewValue(bytes.NewBufferString("test")), "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tpl.QueryEscapeAny(ctx, tt.value.(tpl.Value))
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestUrlEncodeFilter(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"simple", `{{_val|urlencode()}}`, map[string]any{"_val": "hello world"}, "hello+world"},
		{"special", `{{_val|urlencode()}}`, map[string]any{"_val": "a=1&b=2"}, "a%3D1%26b%3D2"},
		{"unicode", `{{_val|urlencode()}}`, map[string]any{"_val": "日本"}, "%E6%97%A5%E6%9C%AC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := tpl.New()
			engine.Raw.TemplateData["main"] = tt.template

			ctx := context.Background()
			ctx = tpl.ValuesCtx(ctx, tt.vars)

			if err := engine.Compile(ctx); err != nil {
				t.Fatalf("Compile failed: %v", err)
			}

			result, err := engine.ParseAndReturn(ctx, "main")
			if err != nil {
				t.Fatalf("ParseAndReturn failed: %v", err)
			}

			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestRawUrlEncodeFilter(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"simple", `{{_val|rawurlencode()}}`, map[string]any{"_val": "hello world"}, "hello%20world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := tpl.New()
			engine.Raw.TemplateData["main"] = tt.template

			ctx := context.Background()
			ctx = tpl.ValuesCtx(ctx, tt.vars)

			if err := engine.Compile(ctx); err != nil {
				t.Fatalf("Compile failed: %v", err)
			}

			result, err := engine.ParseAndReturn(ctx, "main")
			if err != nil {
				t.Fatalf("ParseAndReturn failed: %v", err)
			}

			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestQueryEscapeAnyDirectTypes(t *testing.T) {
	ctx := context.Background()

	// Test raw string
	if result := tpl.QueryEscapeAny(ctx, "hello world"); result != "hello+world" {
		t.Errorf("string: got %q, want %q", result, "hello+world")
	}

	// Test raw []byte
	if result := tpl.QueryEscapeAny(ctx, []byte("test data")); result != "test+data" {
		t.Errorf("[]byte: got %q, want %q", result, "test+data")
	}

	// Test bytes.Buffer
	buf := bytes.Buffer{}
	buf.WriteString("buffer content")
	if result := tpl.QueryEscapeAny(ctx, buf); result != "buffer+content" {
		t.Errorf("bytes.Buffer: got %q, want %q", result, "buffer+content")
	}

	// Test nil
	if result := tpl.QueryEscapeAny(ctx, nil); result != "" {
		t.Errorf("nil: got %q, want %q", result, "")
	}

	// Test int
	if result := tpl.QueryEscapeAny(ctx, 42); result != "42" {
		t.Errorf("int: got %q, want %q", result, "42")
	}

	// Test int64
	if result := tpl.QueryEscapeAny(ctx, int64(123)); result != "123" {
		t.Errorf("int64: got %q, want %q", result, "123")
	}

	// Test uint64
	if result := tpl.QueryEscapeAny(ctx, uint64(456)); result != "456" {
		t.Errorf("uint64: got %q, want %q", result, "456")
	}

	// Test map[string]interface{}
	m := map[string]interface{}{"a": "1", "b": "2"}
	result := tpl.QueryEscapeAny(ctx, m)
	// Map order is not guaranteed, but it should contain both key-value pairs
	if result != "a=1&b=2" && result != "b=2&a=1" {
		t.Errorf("map: got %q, want 'a=1&b=2' or 'b=2&a=1'", result)
	}
}
