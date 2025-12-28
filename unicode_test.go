package tpl_test

import (
	"context"
	"testing"

	"github.com/KarpelesLab/tpl"
)

func TestUnicodeFilter(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"unicode_length", `{{_val|unicode()|length()}}`, map[string]any{"_val": "ABC"}, "3"},
		{"unicode_single", `{{_val|unicode()|length()}}`, map[string]any{"_val": "A"}, "1"},
		{"unicode_multibyte", `{{_val|unicode()|length()}}`, map[string]any{"_val": "日本"}, "2"},
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
