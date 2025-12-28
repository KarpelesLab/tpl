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
