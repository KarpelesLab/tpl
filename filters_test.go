package tpl_test

import (
	"bytes"
	"context"
	"strings"
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

func TestFilterLengthArray(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"slice_len", `{{_val|length()}}`, map[string]any{"_val": []string{"a", "b", "c"}}, "3"},
		{"empty_slice", `{{_val|length()}}`, map[string]any{"_val": []string{}}, "0"},
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

func TestFilterReverseArray(t *testing.T) {
	engine := tpl.New()
	engine.Raw.TemplateData["main"] = `{{_val|reverse()|implode(",")}}`

	ctx := context.Background()
	ctx = tpl.ValuesCtx(ctx, map[string]any{
		"_val": []string{"a", "b", "c"},
	})

	if err := engine.Compile(ctx); err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	result, err := engine.ParseAndReturn(ctx, "main")
	if err != nil {
		t.Fatalf("ParseAndReturn failed: %v", err)
	}

	if result != "c,b,a" {
		t.Errorf("got %q, want %q", result, "c,b,a")
	}
}

func TestFilterSizeNegative(t *testing.T) {
	engine := tpl.New()
	engine.Raw.TemplateData["main"] = `{{_val|size()}}`

	ctx := context.Background()
	ctx = tpl.ValuesCtx(ctx, map[string]any{
		"_val": -1024,
	})

	if err := engine.Compile(ctx); err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	result, err := engine.ParseAndReturn(ctx, "main")
	if err != nil {
		t.Fatalf("ParseAndReturn failed: %v", err)
	}

	if result != "-1.00 kiB" {
		t.Errorf("got %q, want %q", result, "-1.00 kiB")
	}
}

func TestFilterJsonWithMap(t *testing.T) {
	engine := tpl.New()
	engine.Raw.TemplateData["main"] = `{{_val|json()}}`

	ctx := context.Background()
	ctx = tpl.ValuesCtx(ctx, map[string]any{
		"_val": map[string]int{"a": 1},
	})

	if err := engine.Compile(ctx); err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	result, err := engine.ParseAndReturn(ctx, "main")
	if err != nil {
		t.Fatalf("ParseAndReturn failed: %v", err)
	}

	if result != `{"a":1}` {
		t.Errorf("got %q, want %q", result, `{"a":1}`)
	}
}

func TestFilterRoundModes(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"round_2_decimals", `{{_val|round(2)}}`, map[string]any{"_val": 3.14159}, "3.14"},
		{"round_0_decimals", `{{_val|round(0)}}`, map[string]any{"_val": 3.5}, "4"},
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

func TestFilterToInt(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"from_string", `{{_val|toint()}}`, map[string]any{"_val": "42"}, "42"},
		{"from_float", `{{_val|toint()}}`, map[string]any{"_val": 3.7}, "4"},
		{"from_bool_true", `{{_val|toint()}}`, map[string]any{"_val": true}, "1"},
		{"from_bool_false", `{{_val|toint()}}`, map[string]any{"_val": false}, "0"},
		{"from_int8", `{{_val|toint()}}`, map[string]any{"_val": int8(42)}, "42"},
		{"from_int16", `{{_val|toint()}}`, map[string]any{"_val": int16(42)}, "42"},
		{"from_int32", `{{_val|toint()}}`, map[string]any{"_val": int32(42)}, "42"},
		{"from_int64", `{{_val|toint()}}`, map[string]any{"_val": int64(42)}, "42"},
		{"from_uint8", `{{_val|toint()}}`, map[string]any{"_val": uint8(42)}, "42"},
		{"from_uint16", `{{_val|toint()}}`, map[string]any{"_val": uint16(42)}, "42"},
		{"from_uint32", `{{_val|toint()}}`, map[string]any{"_val": uint32(42)}, "42"},
		{"from_nil", `{{_val|toint()}}`, map[string]any{"_val": nil}, "0"},
		{"from_bytes", `{{_val|toint()}}`, map[string]any{"_val": []byte("123")}, "123"},
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

func TestLengthFilterVariousTypes(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"string", `{{_val|length()}}`, map[string]any{"_val": "hello"}, "5"},
		{"slice_interface", `{{_val|length()}}`, map[string]any{"_val": []any{"a", "b", "c"}}, "3"},
		{"slice_string", `{{_val|length()}}`, map[string]any{"_val": []string{"a", "b"}}, "2"},
		{"map_string_interface", `{{_val|length()}}`, map[string]any{"_val": map[string]any{"a": 1, "b": 2}}, "2"},
		{"empty_string", `{{_val|length()}}`, map[string]any{"_val": ""}, "0"},
		{"empty_slice", `{{_val|length()}}`, map[string]any{"_val": []any{}}, "0"},
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

func TestArraySliceFilter(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"string_slice", `{{_val|arrayslice(1, 3)}}`, map[string]any{"_val": "hello"}, "ell"},
		{"string_one_param", `{{_val|arrayslice(2)}}`, map[string]any{"_val": "hello"}, "he"},
		{"string_from_exceeds", `{{_val|arrayslice(10, 2)}}`, map[string]any{"_val": "hello"}, ""},
		{"string_to_exceeds", `{{_val|arrayslice(3, 10)}}`, map[string]any{"_val": "hello"}, "lo"},
		{"slice_string", `{{_val|arrayslice(0, 2)|implode(",")}}`, map[string]any{"_val": []string{"x", "y", "z"}}, "x,y"},
		{"slice_string_from_exceeds", `{{_val|arrayslice(10, 2)}}`, map[string]any{"_val": []string{"a", "b"}}, ""},
		{"slice_string_to_exceeds", `{{_val|arrayslice(1, 10)|implode(",")}}`, map[string]any{"_val": []string{"a", "b", "c"}}, "b,c"},
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

func TestImplodeFilter(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"slice_string", `{{_val|implode(",")}}`, map[string]any{"_val": []string{"a", "b", "c"}}, "a,b,c"},
		{"empty_slice", `{{_val|implode(",")}}`, map[string]any{"_val": []string{}}, ""},
		{"single_element", `{{_val|implode(",")}}`, map[string]any{"_val": []string{"only"}}, "only"},
		{"with_space", `{{_val|implode(" - ")}}`, map[string]any{"_val": []string{"one", "two"}}, "one - two"},
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

func TestReverseFilter(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"string", `{{_val|reverse()}}`, map[string]any{"_val": "hello"}, "olleh"},
		{"unicode", `{{_val|reverse()}}`, map[string]any{"_val": "日本語"}, "語本日"},
		{"slice_string", `{{_val|reverse()|implode(",")}}`, map[string]any{"_val": []string{"a", "b", "c"}}, "c,b,a"},
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

func TestExplodeFilter(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"simple", `{{_val|explode(",")}}`, map[string]any{"_val": "a,b,c"}, "[a b c]"},
		{"single", `{{_val|explode(",")}}`, map[string]any{"_val": "single"}, "[single]"},
		{"empty", `{{_val|explode(",")}}`, map[string]any{"_val": ""}, "[]"},
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

func TestSubstrFilter(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"basic", `{{_val|substr(0, 3)}}`, map[string]any{"_val": "hello"}, "hel"},
		{"middle", `{{_val|substr(1, 3)}}`, map[string]any{"_val": "hello"}, "ell"},
		{"from_end", `{{_val|substr({{_start}}, 2)}}`, map[string]any{"_val": "hello", "_start": int64(-2)}, "lo"},
		{"exceeds", `{{_val|substr(0, 100)}}`, map[string]any{"_val": "hi"}, "hi"},
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

func TestJsonParseFilter(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"string", `{{_val|jsonparse()}}`, map[string]any{"_val": `"hello"`}, "hello"},
		{"number", `{{_val|jsonparse()}}`, map[string]any{"_val": `42`}, "42"},
		{"bool", `{{_val|jsonparse()}}`, map[string]any{"_val": `true`}, "1"},
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

func TestBoolConversion(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"empty_map", `{{if {{_val}}}}yes{{else}}no{{/if}}`, map[string]any{"_val": map[string]any{}}, "no"},
		{"nonempty_map", `{{if {{_val}}}}yes{{else}}no{{/if}}`, map[string]any{"_val": map[string]any{"a": 1}}, "yes"},
		{"empty_slice", `{{if {{_val}}}}yes{{else}}no{{/if}}`, map[string]any{"_val": []any{}}, "no"},
		{"nonempty_slice", `{{if {{_val}}}}yes{{else}}no{{/if}}`, map[string]any{"_val": []any{1}}, "yes"},
		{"string_zero", `{{if {{_val}}}}yes{{else}}no{{/if}}`, map[string]any{"_val": "0"}, "no"},
		{"string_one", `{{if {{_val}}}}yes{{else}}no{{/if}}`, map[string]any{"_val": "1"}, "yes"},
		{"bytes_zero", `{{if {{_val}}}}yes{{else}}no{{/if}}`, map[string]any{"_val": []byte("0")}, "no"},
		{"bytes_one", `{{if {{_val}}}}yes{{else}}no{{/if}}`, map[string]any{"_val": []byte("1")}, "yes"},
		{"int64_zero", `{{if {{_val}}}}yes{{else}}no{{/if}}`, map[string]any{"_val": int64(0)}, "no"},
		{"int64_nonzero", `{{if {{_val}}}}yes{{else}}no{{/if}}`, map[string]any{"_val": int64(42)}, "yes"},
		{"float64_zero", `{{if {{_val}}}}yes{{else}}no{{/if}}`, map[string]any{"_val": float64(0)}, "no"},
		{"float64_nonzero", `{{if {{_val}}}}yes{{else}}no{{/if}}`, map[string]any{"_val": float64(3.14)}, "yes"},
		{"uint64_nonzero", `{{if {{_val}}}}yes{{else}}no{{/if}}`, map[string]any{"_val": uint64(1)}, "yes"},
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

func TestTypeFilter(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"string", `{{_val|type()}}`, map[string]any{"_val": "hello"}, "string"},
		{"int", `{{_val|type()}}`, map[string]any{"_val": int64(42)}, "int64"},
		{"bool", `{{_val|type()}}`, map[string]any{"_val": true}, "bool"},
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

func TestNl2brFilter(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"basic", `{{_val|nl2br()}}`, map[string]any{"_val": "a\nb\nc"}, "a<br/>b<br/>c"},
		{"empty", `{{_val|nl2br()}}`, map[string]any{"_val": ""}, ""},
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

func TestTrimFilter(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"spaces", `{{_val|trim()}}`, map[string]any{"_val": "  hello  "}, "hello"},
		{"tabs", `{{_val|trim()}}`, map[string]any{"_val": "\thello\t"}, "hello"},
		{"newlines", `{{_val|trim()}}`, map[string]any{"_val": "\nhello\n"}, "hello"},
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

func TestToIntFilter(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"from_string", `{{_val|toint()}}`, map[string]any{"_val": "42"}, "42"},
		{"from_float", `{{_val|toint()}}`, map[string]any{"_val": 3.9}, "4"},
		{"from_bool_true", `{{_val|toint()}}`, map[string]any{"_val": true}, "1"},
		{"from_bool_false", `{{_val|toint()}}`, map[string]any{"_val": false}, "0"},
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

func TestRoundFilter(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"default_precision", `{{_val|round()}}`, map[string]any{"_val": 3.14159}, "3.14"},
		{"precision_0", `{{_val|round(0)}}`, map[string]any{"_val": 3.7}, "4"},
		{"precision_1", `{{_val|round(1)}}`, map[string]any{"_val": 3.14159}, "3.1"},
		{"precision_3", `{{_val|round(3)}}`, map[string]any{"_val": 3.14159}, "3.142"},
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

func TestStripcrlfFilter(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"crlf", `{{_val|stripcrlf()}}`, map[string]any{"_val": "a\r\nb\r\nc"}, "abc"},
		{"lf_only", `{{_val|stripcrlf()}}`, map[string]any{"_val": "a\nb\nc"}, "abc"},
		{"cr_only", `{{_val|stripcrlf()}}`, map[string]any{"_val": "a\rb\rc"}, "abc"},
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

func TestB64EncDecFilters(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"encode", `{{_val|b64enc()}}`, map[string]any{"_val": "hello"}, "aGVsbG8="},
		{"decode", `{{_val|b64dec()}}`, map[string]any{"_val": "aGVsbG8="}, "hello"},
		{"roundtrip", `{{_val|b64enc()|b64dec()}}`, map[string]any{"_val": "test data"}, "test data"},
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

func TestStriptagsFilter(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"basic", `{{_val|striptags()}}`, map[string]any{"_val": "<p>Hello</p>"}, "Hello"},
		{"multiple", `{{_val|striptags()}}`, map[string]any{"_val": "<div><p>Test</p></div>"}, "Test"},
		{"with_attrs", `{{_val|striptags()}}`, map[string]any{"_val": `<a href="test">Link</a>`}, "Link"},
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

func TestIsNullFilter(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"null_true", `{{if {{_val|isnull()}}}}yes{{else}}no{{/if}}`, map[string]any{"_val": nil}, "yes"},
		{"null_false", `{{if {{_val|isnull()}}}}yes{{else}}no{{/if}}`, map[string]any{"_val": "hello"}, "no"},
		{"null_empty_string", `{{if {{_val|isnull()}}}}yes{{else}}no{{/if}}`, map[string]any{"_val": ""}, "no"},
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

func TestDumpFilter(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
	}{
		{"string", `{{_val|dump()}}`, map[string]any{"_val": "hello"}},
		{"int", `{{_val|dump()}}`, map[string]any{"_val": 42}},
		{"map", `{{_val|dump()}}`, map[string]any{"_val": map[string]any{"a": 1}}},
		{"slice", `{{_val|dump()}}`, map[string]any{"_val": []string{"a", "b"}}},
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

			_, err := engine.ParseAndReturn(ctx, "main")
			if err != nil {
				t.Fatalf("ParseAndReturn failed: %v", err)
			}
			// dump() outputs debug info, we just check it doesn't error
		})
	}
}

func TestPriceFilter(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"with_display", `{{_val|price()}}`, map[string]any{"_val": map[string]any{"display": "$12.34"}}, "$12.34"},
		{"no_display", `{{_val|price()}}`, map[string]any{"_val": map[string]any{"value": 1234}}, "N/A"},
		{"non_map", `{{_val|price()}}`, map[string]any{"_val": int64(1234)}, "N/A"},
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

func TestJsonFilter(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"string", `{{_val|json()}}`, map[string]any{"_val": "hello"}, `"hello"`},
		{"int", `{{_val|json()}}`, map[string]any{"_val": 42}, "42"},
		{"map", `{{_val|json()}}`, map[string]any{"_val": map[string]any{"a": 1}}, `{"a":1}`},
		{"slice", `{{_val|json()}}`, map[string]any{"_val": []string{"a", "b"}}, `["a","b"]`},
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

func TestReplaceFilter(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"basic", `{{_val|replace("world", "earth")}}`, map[string]any{"_val": "hello world"}, "hello earth"},
		{"multiple", `{{_val|replace("a", "x")}}`, map[string]any{"_val": "abracadabra"}, "xbrxcxdxbrx"},
		{"no_match", `{{_val|replace("z", "x")}}`, map[string]any{"_val": "hello"}, "hello"},
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

func TestSizeFilter(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"bytes", `{{_val|size()}}`, map[string]any{"_val": int64(500)}, "500 B"},
		{"kilobytes", `{{_val|size()}}`, map[string]any{"_val": int64(1536)}, "1.50 kiB"},
		{"megabytes", `{{_val|size()}}`, map[string]any{"_val": int64(1572864)}, "1.50 MiB"},
		{"gigabytes", `{{_val|size()}}`, map[string]any{"_val": int64(1610612736)}, "1.50 GiB"},
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

func TestEntitiesFilter(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"basic", `{{_val|entities()}}`, map[string]any{"_val": "<div>"}, "&lt;div&gt;"},
		{"ampersand", `{{_val|entities()}}`, map[string]any{"_val": "a&b"}, "a&amp;b"},
		{"quotes", `{{_val|entities()}}`, map[string]any{"_val": `"test"`}, "&#34;test&#34;"},
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

func TestImplodeMoreCases(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"values", `{{_arr|implode("-")}}`, map[string]any{"_arr": tpl.Values{tpl.NewValue("a"), tpl.NewValue("b")}}, "ab"},
		{"single_elem", `{{_arr|implode(",")}}`, map[string]any{"_arr": []string{"x"}}, "x"},
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

func TestLengthMoreCases(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"bytes", `{{_val|length()}}`, map[string]any{"_val": []byte("hello")}, "5"},
		{"buffer", `{{_val|length()}}`, map[string]any{"_val": bytes.NewBufferString("test")}, "4"},
		{"values", `{{_val|length()}}`, map[string]any{"_val": tpl.Values{tpl.NewValue("a"), tpl.NewValue("b")}}, "2"},
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

func TestBoolOperatorMoreCases(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"bool_and_bool", `{{if ({{_a}} && {{_b}})}}yes{{else}}no{{/if}}`, map[string]any{"_a": true, "_b": true}, "yes"},
		{"bool_and_false", `{{if ({{_a}} && {{_b}})}}yes{{else}}no{{/if}}`, map[string]any{"_a": true, "_b": false}, "no"},
		{"bool_or_bool", `{{if ({{_a}} || {{_b}})}}yes{{else}}no{{/if}}`, map[string]any{"_a": false, "_b": true}, "yes"},
		{"bool_or_false", `{{if ({{_a}} || {{_b}})}}yes{{else}}no{{/if}}`, map[string]any{"_a": false, "_b": false}, "no"},
		{"bool_not", `{{set _x=(~{{_a}})}}{{if {{_x}}}}yes{{else}}no{{/if}}{{/set}}`, map[string]any{"_a": true}, "no"},
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

func TestJsonDumpFilter(t *testing.T) {
	engine := tpl.New()
	engine.Raw.TemplateData["main"] = `{{_val|jsondump()}}`

	ctx := context.Background()
	ctx = tpl.ValuesCtx(ctx, map[string]any{
		"_val": map[string]any{"key": "value"},
	})

	if err := engine.Compile(ctx); err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	result, err := engine.ParseAndReturn(ctx, "main")
	if err != nil {
		t.Fatalf("ParseAndReturn failed: %v", err)
	}

	// jsondump returns pretty-printed JSON
	if !strings.Contains(result, "key") || !strings.Contains(result, "value") {
		t.Errorf("unexpected result: %q", result)
	}
}

func TestStripBbCodeFilter(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"bold", `{{_val|stripbbcode()}}`, map[string]any{"_val": "[b]hello[/b]"}, "hello"},
		{"italic", `{{_val|stripbbcode()}}`, map[string]any{"_val": "[i]world[/i]"}, "world"},
		{"nested", `{{_val|stripbbcode()}}`, map[string]any{"_val": "[b][i]text[/i][/b]"}, "text"},
		{"no_bbcode", `{{_val|stripbbcode()}}`, map[string]any{"_val": "plain text"}, "plain text"},
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

func TestToStringFilter(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{"from_int", `{{_val|tostring()}}`, map[string]any{"_val": 42}, "42"},
		{"from_float", `{{_val|tostring()}}`, map[string]any{"_val": 3.14}, "3.14"},
		{"from_bool", `{{_val|tostring()}}`, map[string]any{"_val": true}, "1"},
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
