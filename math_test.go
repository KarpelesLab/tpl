package tpl_test

import (
	"context"
	"testing"

	"github.com/KarpelesLab/tpl"
)

func TestIfCondition(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"if_true", `{{if {{_val}}}}yes{{/if}}`, map[string]any{"_val": true}, "yes"},
		{"if_false", `{{if {{_val}}}}yes{{/if}}`, map[string]any{"_val": false}, ""},
		{"if_else_true", `{{if {{_val}}}}yes{{else}}no{{/if}}`, map[string]any{"_val": true}, "yes"},
		{"if_else_false", `{{if {{_val}}}}yes{{else}}no{{/if}}`, map[string]any{"_val": false}, "no"},
		{"if_int_nonzero", `{{if {{_val}}}}yes{{/if}}`, map[string]any{"_val": 1}, "yes"},
		{"if_int_zero", `{{if {{_val}}}}yes{{/if}}`, map[string]any{"_val": 0}, ""},
		{"if_string_nonempty", `{{if {{_val}}}}yes{{/if}}`, map[string]any{"_val": "hello"}, "yes"},
		{"if_string_empty", `{{if {{_val}}}}yes{{/if}}`, map[string]any{"_val": ""}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := tpl.New()
			engine.Raw.TemplateData["main"] = tt.template

			ctx := context.Background()
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

			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSetVariable(t *testing.T) {
	engine := tpl.New()
	engine.Raw.TemplateData["main"] = `{{set _x= "hello"}}{{_x}}{{/set}}`

	ctx := context.Background()

	if err := engine.Compile(ctx); err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	result, err := engine.ParseAndReturn(ctx, "main")
	if err != nil {
		t.Fatalf("ParseAndReturn failed: %v", err)
	}

	if result != "hello" {
		t.Errorf("got %q, want %q", result, "hello")
	}
}
