package tpl_test

import (
	"context"
	"strings"
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

func TestMathOperatorsInt(t *testing.T) {
	tests := []struct {
		name     string
		template string
		expected string
	}{
		{"add", `{{set _x=(1 + 2)}}{{_x}}{{/set}}`, "3"},
		{"subtract", `{{set _x=(5 - 3)}}{{_x}}{{/set}}`, "2"},
		{"multiply", `{{set _x=(3 * 4)}}{{_x}}{{/set}}`, "12"},
		{"divide", `{{set _x=(10 / 2)}}{{_x}}{{/set}}`, "5"},
		{"modulo", `{{set _x=(10 % 3)}}{{_x}}{{/set}}`, "1"},
		{"precedence", `{{set _x=(2 + 3 * 4)}}{{_x}}{{/set}}`, "14"},
		{"parentheses", `{{set _x=((2 + 3) * 4)}}{{_x}}{{/set}}`, "20"},
		{"negative", `{{set _x=(0 - 5)}}{{_x}}{{/set}}`, "-5"},
		{"shift_left", `{{set _x=(1 << 3)}}{{_x}}{{/set}}`, "8"},
		{"shift_right", `{{set _x=(8 >> 2)}}{{_x}}{{/set}}`, "2"},
		{"bitwise_and", `{{set _x=(7 & 3)}}{{_x}}{{/set}}`, "3"},
		{"bitwise_xor", `{{set _x=(5 ^ 3)}}{{_x}}{{/set}}`, "6"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := tpl.New()
			engine.Raw.TemplateData["main"] = tt.template

			ctx := context.Background()

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

func TestMathOperatorsFloat(t *testing.T) {
	tests := []struct {
		name     string
		template string
		expected string
	}{
		{"add", `{{set _x=(1.5 + 2.5)}}{{_x}}{{/set}}`, "4"},
		{"subtract", `{{set _x=(5.5 - 3.0)}}{{_x}}{{/set}}`, "2.5"},
		{"multiply", `{{set _x=(2.5 * 4)}}{{_x}}{{/set}}`, "10"},
		{"divide", `{{set _x=(10.0 / 4)}}{{_x}}{{/set}}`, "2.5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := tpl.New()
			engine.Raw.TemplateData["main"] = tt.template

			ctx := context.Background()

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

func TestComparisonOperatorsInt(t *testing.T) {
	tests := []struct {
		name     string
		template string
		expected string
	}{
		{"less_than_true", `{{if (1 < 2)}}yes{{else}}no{{/if}}`, "yes"},
		{"less_than_false", `{{if (2 < 1)}}yes{{else}}no{{/if}}`, "no"},
		{"less_equal_true", `{{if (2 <= 2)}}yes{{else}}no{{/if}}`, "yes"},
		{"less_equal_false", `{{if (3 <= 2)}}yes{{else}}no{{/if}}`, "no"},
		{"greater_than_true", `{{if (3 > 2)}}yes{{else}}no{{/if}}`, "yes"},
		{"greater_than_false", `{{if (1 > 2)}}yes{{else}}no{{/if}}`, "no"},
		{"greater_equal_true", `{{if (2 >= 2)}}yes{{else}}no{{/if}}`, "yes"},
		{"greater_equal_false", `{{if (1 >= 2)}}yes{{else}}no{{/if}}`, "no"},
		{"equal_true", `{{if (5 == 5)}}yes{{else}}no{{/if}}`, "yes"},
		{"equal_false", `{{if (5 == 6)}}yes{{else}}no{{/if}}`, "no"},
		{"not_equal_true", `{{if (5 != 6)}}yes{{else}}no{{/if}}`, "yes"},
		{"not_equal_false", `{{if (5 != 5)}}yes{{else}}no{{/if}}`, "no"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := tpl.New()
			engine.Raw.TemplateData["main"] = tt.template

			ctx := context.Background()

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

func TestLogicalOperators(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"and_true_true", `{{if ({{_a}} && {{_b}})}}yes{{else}}no{{/if}}`, map[string]any{"_a": true, "_b": true}, "yes"},
		{"and_true_false", `{{if ({{_a}} && {{_b}})}}yes{{else}}no{{/if}}`, map[string]any{"_a": true, "_b": false}, "no"},
		{"and_false_true", `{{if ({{_a}} && {{_b}})}}yes{{else}}no{{/if}}`, map[string]any{"_a": false, "_b": true}, "no"},
		{"and_false_false", `{{if ({{_a}} && {{_b}})}}yes{{else}}no{{/if}}`, map[string]any{"_a": false, "_b": false}, "no"},
		{"or_true_true", `{{if ({{_a}} || {{_b}})}}yes{{else}}no{{/if}}`, map[string]any{"_a": true, "_b": true}, "yes"},
		{"or_true_false", `{{if ({{_a}} || {{_b}})}}yes{{else}}no{{/if}}`, map[string]any{"_a": true, "_b": false}, "yes"},
		{"or_false_true", `{{if ({{_a}} || {{_b}})}}yes{{else}}no{{/if}}`, map[string]any{"_a": false, "_b": true}, "yes"},
		{"or_false_false", `{{if ({{_a}} || {{_b}})}}yes{{else}}no{{/if}}`, map[string]any{"_a": false, "_b": false}, "no"},
		{"not_true", `{{if (!{{_a}})}}yes{{else}}no{{/if}}`, map[string]any{"_a": true}, "no"},
		{"not_false", `{{if (!{{_a}})}}yes{{else}}no{{/if}}`, map[string]any{"_a": false}, "yes"},
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

func TestMathWithVariables(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"add_vars", `{{set _x=({{_a}} + {{_b}})}}{{_x}}{{/set}}`, map[string]any{"_a": 3, "_b": 4}, "7"},
		{"multiply_vars", `{{set _x=({{_a}} * {{_b}})}}{{_x}}{{/set}}`, map[string]any{"_a": 3, "_b": 4}, "12"},
		{"compare_vars", `{{if ({{_a}} < {{_b}})}}yes{{else}}no{{/if}}`, map[string]any{"_a": 3, "_b": 5}, "yes"},
		{"float_vars", `{{set _x=({{_a}} + {{_b}})}}{{_x}}{{/set}}`, map[string]any{"_a": 1.5, "_b": 2.5}, "4"},
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

func TestBitwiseNotOperator(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"not_bool_true", `{{set _x=(~{{_a}})}}{{_x}}{{/set}}`, map[string]any{"_a": true}, ""},
		{"not_bool_false", `{{set _x=(~{{_a}})}}{{_x}}{{/set}}`, map[string]any{"_a": false}, "1"},
		{"not_int", `{{set _x=(~{{_a}})}}{{_x}}{{/set}}`, map[string]any{"_a": int64(0)}, "9223372036854775807"},
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

func TestFloatComparisonOperators(t *testing.T) {
	tests := []struct {
		name     string
		template string
		expected string
	}{
		{"float_less_than", `{{if (1.5 < 2.5)}}yes{{else}}no{{/if}}`, "yes"},
		{"float_less_equal", `{{if (2.5 <= 2.5)}}yes{{else}}no{{/if}}`, "yes"},
		{"float_greater_than", `{{if (3.5 > 2.5)}}yes{{else}}no{{/if}}`, "yes"},
		{"float_greater_equal", `{{if (2.5 >= 2.5)}}yes{{else}}no{{/if}}`, "yes"},
		{"float_equal", `{{if (2.5 == 2.5)}}yes{{else}}no{{/if}}`, "yes"},
		{"float_not_equal", `{{if (2.5 != 3.5)}}yes{{else}}no{{/if}}`, "yes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := tpl.New()
			engine.Raw.TemplateData["main"] = tt.template

			ctx := context.Background()

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

func TestBoolOperatorEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"bool_equal_true", `{{if ({{_a}} == {{_b}})}}yes{{else}}no{{/if}}`, map[string]any{"_a": true, "_b": true}, "yes"},
		{"bool_equal_false", `{{if ({{_a}} == {{_b}})}}yes{{else}}no{{/if}}`, map[string]any{"_a": true, "_b": false}, "no"},
		{"bool_not_equal", `{{if ({{_a}} != {{_b}})}}yes{{else}}no{{/if}}`, map[string]any{"_a": true, "_b": false}, "yes"},
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

func TestMathWithUintTypes(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"uint8_add", `{{set _x=({{_a}} + {{_b}})}}{{_x}}{{/set}}`, map[string]any{"_a": uint8(10), "_b": uint8(5)}, "15"},
		{"uint16_add", `{{set _x=({{_a}} + {{_b}})}}{{_x}}{{/set}}`, map[string]any{"_a": uint16(100), "_b": uint16(50)}, "150"},
		{"uint32_add", `{{set _x=({{_a}} + {{_b}})}}{{_x}}{{/set}}`, map[string]any{"_a": uint32(1000), "_b": uint32(500)}, "1500"},
		{"uint_add", `{{set _x=({{_a}} + {{_b}})}}{{_x}}{{/set}}`, map[string]any{"_a": uint(42), "_b": uint(8)}, "50"},
		{"uint8_multiply", `{{set _x=({{_a}} * {{_b}})}}{{_x}}{{/set}}`, map[string]any{"_a": uint8(7), "_b": uint8(6)}, "42"},
		{"uint16_comparison_gt", `{{if ({{_a}} > {{_b}})}}yes{{else}}no{{/if}}`, map[string]any{"_a": uint16(10), "_b": uint16(5)}, "yes"},
		{"uint32_comparison_lt", `{{if ({{_a}} < {{_b}})}}yes{{else}}no{{/if}}`, map[string]any{"_a": uint32(3), "_b": uint32(8)}, "yes"},
		{"uint8_comparison_eq", `{{if ({{_a}} == {{_b}})}}yes{{else}}no{{/if}}`, map[string]any{"_a": uint8(42), "_b": uint8(42)}, "yes"},
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

func TestMathWithFloatTypes(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"float32_add", `{{set _x=({{_a}} + {{_b}})}}{{_x}}{{/set}}`, map[string]any{"_a": float32(1.5), "_b": float32(2.5)}, "4"},
		{"float64_add", `{{set _x=({{_a}} + {{_b}})}}{{_x}}{{/set}}`, map[string]any{"_a": float64(10.25), "_b": float64(5.75)}, "16"},
		{"float_multiply", `{{set _x=({{_a}} * {{_b}})}}{{_x}}{{/set}}`, map[string]any{"_a": float64(2.5), "_b": float64(4.0)}, "10"},
		{"float_divide", `{{set _x=({{_a}} / {{_b}})}}{{_x}}{{/set}}`, map[string]any{"_a": float64(10.0), "_b": float64(4.0)}, "2.5"},
		{"float_comparison_gt", `{{if ({{_a}} > {{_b}})}}yes{{else}}no{{/if}}`, map[string]any{"_a": float64(3.14), "_b": float64(2.71)}, "yes"},
		{"float_comparison_lt", `{{if ({{_a}} < {{_b}})}}yes{{else}}no{{/if}}`, map[string]any{"_a": float64(1.5), "_b": float64(2.5)}, "yes"},
		{"float_int_mixed", `{{set _x=({{_a}} + {{_b}})}}{{_x}}{{/set}}`, map[string]any{"_a": float64(1.5), "_b": int64(2)}, "3.5"},
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

func TestMathWithMixedIntTypes(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"int8_add", `{{set _x=({{_a}} + {{_b}})}}{{_x}}{{/set}}`, map[string]any{"_a": int8(10), "_b": int8(5)}, "15"},
		{"int16_add", `{{set _x=({{_a}} + {{_b}})}}{{_x}}{{/set}}`, map[string]any{"_a": int16(100), "_b": int16(50)}, "150"},
		{"int32_add", `{{set _x=({{_a}} + {{_b}})}}{{_x}}{{/set}}`, map[string]any{"_a": int32(1000), "_b": int32(500)}, "1500"},
		{"int_add", `{{set _x=({{_a}} + {{_b}})}}{{_x}}{{/set}}`, map[string]any{"_a": int(42), "_b": int(8)}, "50"},
		{"int8_int64_mixed", `{{set _x=({{_a}} + {{_b}})}}{{_x}}{{/set}}`, map[string]any{"_a": int8(10), "_b": int64(20)}, "30"},
		{"int_uint_mixed", `{{set _x=({{_a}} + {{_b}})}}{{_x}}{{/set}}`, map[string]any{"_a": int64(10), "_b": uint64(20)}, "30"},
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


func TestRandFunction(t *testing.T) {
	engine := tpl.New()
	engine.Raw.TemplateData["main"] = `{{@rand(1, 100)}}`

	ctx := context.Background()

	if err := engine.Compile(ctx); err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	result, err := engine.ParseAndReturn(ctx, "main")
	if err != nil {
		t.Fatalf("ParseAndReturn failed: %v", err)
	}

	// rand() should return some value
	if result == "" {
		t.Error("rand() returned empty string")
	}
}

func TestRandFunctionMultipleCalls(t *testing.T) {
	engine := tpl.New()
	engine.Raw.TemplateData["main"] = `{{@rand(0, 10)}}-{{@rand(0, 10)}}`

	ctx := context.Background()

	if err := engine.Compile(ctx); err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	result, err := engine.ParseAndReturn(ctx, "main")
	if err != nil {
		t.Fatalf("ParseAndReturn failed: %v", err)
	}

	// Should contain a hyphen separating two numbers
	if result == "" || !strings.Contains(result, "-") {
		t.Errorf("rand() returned unexpected result: %q", result)
	}
}

func TestPrintfFunction(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"string", `{{@printf("%s", {{_val}})}}`, map[string]any{"_val": "hello"}, "hello"},
		{"int", `{{@printf("%d", {{_val}})}}`, map[string]any{"_val": 42}, "42"},
		{"float", `{{@printf("%.2f", {{_val}})}}`, map[string]any{"_val": 3.14159}, "3.14"},
		{"multiple", `{{@printf("%s=%d", {{_a}}, {{_b}})}}`, map[string]any{"_a": "x", "_b": 5}, "x=5"},
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

func TestStringFunction(t *testing.T) {
	tests := []struct {
		name     string
		template string
		expected string
	}{
		{"simple", `{{@string("hello")}}`, "hello"},
		{"with_special", `{{@string("a=b&c=d")}}`, "a=b&c=d"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := tpl.New()
			engine.Raw.TemplateData["main"] = tt.template

			ctx := context.Background()

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

func TestErrorFunction(t *testing.T) {
	engine := tpl.New()
	engine.Raw.TemplateData["main"] = `{{@error("test error message")}}`

	ctx := context.Background()

	if err := engine.Compile(ctx); err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	_, err := engine.ParseAndReturn(ctx, "main")
	if err == nil {
		t.Error("Expected error from @error function")
	}
}

func TestComparisonOperatorsWithStrings(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"string_eq", `{{if ({{_a}} == {{_b}})}}yes{{else}}no{{/if}}`, map[string]any{"_a": "hello", "_b": "hello"}, "yes"},
		{"string_ne", `{{if ({{_a}} != {{_b}})}}yes{{else}}no{{/if}}`, map[string]any{"_a": "hello", "_b": "world"}, "yes"},
		{"string_eq_false", `{{if ({{_a}} == {{_b}})}}yes{{else}}no{{/if}}`, map[string]any{"_a": "hello", "_b": "world"}, "no"},
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

func TestNestedMathExpressions(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"nested_add", `{{set _x=(({{_a}} + {{_b}}) + {{_c}})}}{{_x}}{{/set}}`, map[string]any{"_a": int64(1), "_b": int64(2), "_c": int64(3)}, "6"},
		{"nested_mult_add", `{{set _x=(({{_a}} * {{_b}}) + {{_c}})}}{{_x}}{{/set}}`, map[string]any{"_a": int64(2), "_b": int64(3), "_c": int64(4)}, "10"},
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
