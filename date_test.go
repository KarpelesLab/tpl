package tpl_test

import (
	"context"
	"testing"

	"github.com/KarpelesLab/tpl"
	"golang.org/x/text/language"
)

func TestDateFilter(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		check    func(string) bool
	}{
		{
			"unix_timestamp",
			`{{_val|date("%Y")}}`,
			map[string]any{"_val": "@0"},
			func(s string) bool { return s == "1970" },
		},
		{
			"now",
			`{{_val|date("%Y")}}`,
			map[string]any{"_val": "now"},
			func(s string) bool { return len(s) == 4 },
		},
		{
			"unix_with_decimals",
			`{{_val|date("%Y")}}`,
			map[string]any{"_val": "@0.123456"},
			func(s string) bool { return s == "1970" },
		},
		{
			"empty_string",
			`{{_val|date()}}`,
			map[string]any{"_val": ""},
			func(s string) bool { return s == "N/A" },
		},
		{
			"default_format",
			`{{_val|date()}}`,
			map[string]any{"_val": "@0"},
			func(s string) bool { return len(s) > 0 },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := tpl.New()
			engine.Raw.TemplateData["main"] = tt.template

			ctx := context.Background()
			ctx = context.WithValue(ctx, "_language", language.English) //lint:ignore SA1029 template variables use string keys by design
			if tt.vars != nil {
				ctx = tpl.ValuesCtx(ctx, tt.vars)
			}

			if err := engine.Compile(ctx); err != nil {
				t.Fatalf("Compile failed: %v", err)
			}

			result, err := engine.ParseAndReturn(ctx, "main")
			if err != nil {
				t.Fatalf("ParseAndReturn failed: %v", err)
			}

			if !tt.check(result) {
				t.Errorf("check failed for result %q", result)
			}
		})
	}
}

func TestDurationFilter(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected string
	}{
		{"zero", 0, "00:00"},
		{"seconds", 45, "00:45"},
		{"minute", 60, "01:00"},
		{"minutes_seconds", 125, "02:05"},
		{"hour", 3600, "01:00:00"},
		{"hours_minutes", 3725, "01:02:05"},
		{"day", 86400, "1:00:00:00"},
		{"days_hours", 90061, "1:01:01:01"},
		{"negative", -65, "-01:05"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := tpl.New()
			engine.Raw.TemplateData["main"] = `{{_val|duration()}}`

			ctx := context.Background()
			ctx = tpl.ValuesCtx(ctx, map[string]any{
				"_val": tt.input,
			})

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
