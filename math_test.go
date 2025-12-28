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

func TestTryCatch(t *testing.T) {
	engine := tpl.New()
	engine.Raw.TemplateData["main"] = `{{try}}{{@error("test error")}}{{catch}}caught{{/try}}`

	ctx := context.Background()

	if err := engine.Compile(ctx); err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	result, err := engine.ParseAndReturn(ctx, "main")
	if err != nil {
		t.Fatalf("ParseAndReturn failed: %v", err)
	}

	if result != "caught" {
		t.Errorf("got %q, want %q", result, "caught")
	}
}

func TestTryCatchNoError(t *testing.T) {
	engine := tpl.New()
	engine.Raw.TemplateData["main"] = `{{try}}success{{catch}}caught{{/try}}`

	ctx := context.Background()

	if err := engine.Compile(ctx); err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	result, err := engine.ParseAndReturn(ctx, "main")
	if err != nil {
		t.Fatalf("ParseAndReturn failed: %v", err)
	}

	if result != "success" {
		t.Errorf("got %q, want %q", result, "success")
	}
}

func TestElseIf(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"first_true", `{{if {{_a}}}}A{{elseif {{_b}}}}B{{else}}C{{/if}}`, map[string]any{"_a": true, "_b": false}, "A"},
		{"second_true", `{{if {{_a}}}}A{{elseif {{_b}}}}B{{else}}C{{/if}}`, map[string]any{"_a": false, "_b": true}, "B"},
		{"else", `{{if {{_a}}}}A{{elseif {{_b}}}}B{{else}}C{{/if}}`, map[string]any{"_a": false, "_b": false}, "C"},
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

func TestTemplateInclusion(t *testing.T) {
	engine := tpl.New()
	engine.Raw.TemplateData["main"] = `Hello {{sub}}!`
	engine.Raw.TemplateData["sub"] = `World`

	ctx := context.Background()

	if err := engine.Compile(ctx); err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	result, err := engine.ParseAndReturn(ctx, "main")
	if err != nil {
		t.Fatalf("ParseAndReturn failed: %v", err)
	}

	if result != "Hello World!" {
		t.Errorf("got %q, want %q", result, "Hello World!")
	}
}

