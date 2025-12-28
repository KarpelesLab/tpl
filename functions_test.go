package tpl_test

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/KarpelesLab/tpl"
)

func TestFunctions(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
		check    func(string) bool
	}{
		{"string", `{{@string("hello", " ", "world")}}`, nil, "hello world", nil},
		{"printf_string", `{{@printf("%s", "test")}}`, nil, "test", nil},
		{"phpversion", `{{@phpversion()}}`, nil, "", func(s string) bool { return strings.HasPrefix(s, "go") }},
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

			if tt.check != nil {
				if !tt.check(result) {
					t.Errorf("check failed for result %q", result)
				}
			} else if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFunctionRand(t *testing.T) {
	engine := tpl.New()
	engine.Raw.TemplateData["main"] = `{{@rand(1, 10)}}`

	ctx := context.Background()

	if err := engine.Compile(ctx); err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	result, err := engine.ParseAndReturn(ctx, "main")
	if err != nil {
		t.Fatalf("ParseAndReturn failed: %v", err)
	}

	n, err := strconv.Atoi(result)
	if err != nil {
		t.Fatalf("Failed to parse result as int: %v", err)
	}

	if n < 1 || n >= 10 {
		t.Errorf("rand result %d out of range [1, 10)", n)
	}
}

func TestFunctionError(t *testing.T) {
	engine := tpl.New()
	engine.Raw.TemplateData["main"] = `{{@error("test error %s", "message")}}`

	ctx := context.Background()

	if err := engine.Compile(ctx); err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	_, err := engine.ParseAndReturn(ctx, "main")
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	if !strings.Contains(err.Error(), "test error message") {
		t.Errorf("Error message %q should contain 'test error message'", err.Error())
	}
}

func TestCustomFunction(t *testing.T) {
	engine := tpl.New()
	engine.Raw.TemplateData["main"] = `{{@custom("arg1", "arg2")}}`

	ctx := context.Background()

	customFn := func(ctx context.Context, params tpl.Values, out tpl.WritableValue) error {
		out.Write([]byte("custom:"))
		for i, p := range params {
			if i > 0 {
				out.Write([]byte(","))
			}
			out.Write(p.WithCtx(ctx).Bytes())
		}
		return nil
	}

	ctx = context.WithValue(ctx, "@custom", tpl.TplFuncCallback(customFn)) //lint:ignore SA1029 template functions use string keys by design

	if err := engine.Compile(ctx); err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	result, err := engine.ParseAndReturn(ctx, "main")
	if err != nil {
		t.Fatalf("ParseAndReturn failed: %v", err)
	}

	if result != "custom:arg1,arg2" {
		t.Errorf("got %q, want %q", result, "custom:arg1,arg2")
	}
}
