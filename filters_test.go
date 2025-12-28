package tpl_test

import (
	"context"
	"testing"

	"github.com/KarpelesLab/tpl"
)

func TestFilters(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		// String filters (using variables since literals are treated as template names)
		{"uppercase", `{{_val|uppercase()}}`, map[string]any{"_val": "hello"}, "HELLO"},
		{"lowercase", `{{_val|lowercase()}}`, map[string]any{"_val": "HELLO"}, "hello"},
		{"trim", `{{_val|trim()}}`, map[string]any{"_val": "  hello  "}, "hello"},
		{"nl2br", `{{_val|nl2br()}}`, map[string]any{"_val": "\n"}, "<br/>"},
		{"nbsp", `{{_val|nbsp()}}`, map[string]any{"_val": "a b"}, "a\xc2\xa0b"},
		{"entities", `{{_val|entities()}}`, map[string]any{"_val": "<>&"}, "&lt;&gt;&amp;"},
		{"stripcrlf", `{{_val|stripcrlf()}}`, map[string]any{"_val": "\r\n"}, ""},

		// URL encoding
		{"urlencode", `{{_val|urlencode()}}`, map[string]any{"_val": "hello world"}, "hello+world"},
		{"rawurlencode", `{{_val|rawurlencode()}}`, map[string]any{"_val": "hello world"}, "hello%20world"},

		// Type conversion
		{"toint", `{{_val|toint()}}`, map[string]any{"_val": "42"}, "42"},
		{"tostring", `{{_val|tostring()}}`, map[string]any{"_val": 42}, "42"},

		// Base64
		{"b64enc", `{{_val|b64enc()}}`, map[string]any{"_val": "hello"}, "aGVsbG8="},
		{"b64dec", `{{_val|b64dec()}}`, map[string]any{"_val": "aGVsbG8="}, "hello"},

		// String manipulation
		{"replace", `{{_val|replace("l", "x")}}`, map[string]any{"_val": "hello"}, "hexxo"},
		{"substr", `{{_val|substr(1, 3)}}`, map[string]any{"_val": "hello"}, "ell"},
		{"truncate", `{{_val|truncate(5)}}`, map[string]any{"_val": "hello world"}, "hello…"},
		{"truncate_wordcut", `{{_val|truncate(8, "...", 1)}}`, map[string]any{"_val": "hello world"}, "hello wo..."},
		{"striptags", `{{_val|striptags()}}`, map[string]any{"_val": "<b>hello</b>"}, "hello"},
		{"explode", `{{_val|explode(",")}}`, map[string]any{"_val": "a,b,c"}, "[a b c]"},

		// Length/count
		{"length_string", `{{_val|length()}}`, map[string]any{"_val": "hello"}, "5"},
		{"count_string", `{{_val|count()}}`, map[string]any{"_val": "hello"}, "5"},

		// Null filter
		{"null", `{{_val|null()}}`, map[string]any{"_val": "hello"}, ""},

		// isnull filter (bool true renders as "1")
		{"isnull_true", `{{_val|isnull()}}`, map[string]any{"_val": nil}, "1"},

		// Type filter
		{"type", `{{_val|type()}}`, map[string]any{"_val": "hello"}, "string"},

		// JSON
		{"json", `{{_val|json()}}`, map[string]any{"_val": "hello"}, `"hello"`},

		// Round
		{"round", `{{_val|round(2)}}`, map[string]any{"_val": 3.14159}, "3.14"},

		// Size filter
		{"size_bytes", `{{_val|size()}}`, map[string]any{"_val": 1023}, "1023 B"},
		{"size_kib", `{{_val|size()}}`, map[string]any{"_val": 1024}, "1.00 kiB"},
		{"size_mib", `{{_val|size()}}`, map[string]any{"_val": 1048576}, "1.00 MiB"},

		// Reverse
		{"reverse_string", `{{_val|reverse()}}`, map[string]any{"_val": "hello"}, "olleh"},

		// Duration
		{"duration_seconds", `{{_val|duration()}}`, map[string]any{"_val": 65}, "01:05"},
		{"duration_hours", `{{_val|duration()}}`, map[string]any{"_val": 3665}, "01:01:05"},
		{"duration_days", `{{_val|duration()}}`, map[string]any{"_val": 90065}, "1:01:01:05"},
		{"duration_negative", `{{_val|duration()}}`, map[string]any{"_val": -65}, "-01:05"},
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

func TestFilterImplode(t *testing.T) {
	engine := tpl.New()
	engine.Raw.TemplateData["main"] = `{{_arr|explode(",")|implode("-")}}`

	ctx := context.Background()
	ctx = tpl.ValuesCtx(ctx, map[string]any{
		"_arr": "a,b,c",
	})

	if err := engine.Compile(ctx); err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	result, err := engine.ParseAndReturn(ctx, "main")
	if err != nil {
		t.Fatalf("ParseAndReturn failed: %v", err)
	}

	if result != "a-b-c" {
		t.Errorf("got %q, want %q", result, "a-b-c")
	}
}

func TestFilterPrice(t *testing.T) {
	engine := tpl.New()
	engine.Raw.TemplateData["main"] = `{{_price|price()}}`

	ctx := context.Background()
	ctx = tpl.ValuesCtx(ctx, map[string]any{
		"_price": map[string]any{"display": "$10.00"},
	})

	if err := engine.Compile(ctx); err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	result, err := engine.ParseAndReturn(ctx, "main")
	if err != nil {
		t.Fatalf("ParseAndReturn failed: %v", err)
	}

	if result != "$10.00" {
		t.Errorf("got %q, want %q", result, "$10.00")
	}
}

func TestFilterArraySlice(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{
			"string_slice",
			`{{_val|arrayslice(1, 3)}}`,
			map[string]any{"_val": "hello"},
			"ell",
		},
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

func TestFilterUnicode(t *testing.T) {
	engine := tpl.New()
	engine.Raw.TemplateData["main"] = `{{@string("a")|unicode()|length()}}`

	ctx := context.Background()

	if err := engine.Compile(ctx); err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	result, err := engine.ParseAndReturn(ctx, "main")
	if err != nil {
		t.Fatalf("ParseAndReturn failed: %v", err)
	}

	if result != "1" {
		t.Errorf("got %q, want %q", result, "1")
	}
}

func TestFilterBbCode(t *testing.T) {
	engine := tpl.New()
	engine.Raw.TemplateData["main"] = `{{_val|bbcode()}}`

	ctx := context.Background()
	ctx = tpl.ValuesCtx(ctx, map[string]any{
		"_val": "[b]hello[/b]",
	})

	if err := engine.Compile(ctx); err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	result, err := engine.ParseAndReturn(ctx, "main")
	if err != nil {
		t.Fatalf("ParseAndReturn failed: %v", err)
	}

	if result != "<b>hello</b>" {
		t.Errorf("got %q, want %q", result, "<b>hello</b>")
	}
}

func TestFilterStripBbCode(t *testing.T) {
	engine := tpl.New()
	engine.Raw.TemplateData["main"] = `{{_val|stripbbcode()}}`

	ctx := context.Background()
	ctx = tpl.ValuesCtx(ctx, map[string]any{
		"_val": "[b]hello[/b]",
	})

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

func TestFilterMarkdown(t *testing.T) {
	engine := tpl.New()
	engine.Raw.TemplateData["main"] = `{{_val|markdown()}}`

	ctx := context.Background()
	ctx = tpl.ValuesCtx(ctx, map[string]any{
		"_val": "**hello**",
	})

	if err := engine.Compile(ctx); err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	result, err := engine.ParseAndReturn(ctx, "main")
	if err != nil {
		t.Fatalf("ParseAndReturn failed: %v", err)
	}

	if result != "<p><strong>hello</strong></p>\n" {
		t.Errorf("got %q, want %q", result, "<p><strong>hello</strong></p>\n")
	}
}
