package tpl_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/KarpelesLab/tpl"
)

func TestResolveValueIndex(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		value    any
		index    string
		expected any
		hasError bool
	}{
		{
			"map_string",
			map[string]any{"key": "value"},
			"key",
			"value",
			false,
		},
		{
			"map_int",
			map[string]any{"num": 42},
			"num",
			42,
			false,
		},
		{
			"slice_index",
			[]any{"a", "b", "c"},
			"1",
			"b",
			false,
		},
		{
			"string_slice_index",
			[]string{"a", "b", "c"},
			"0",
			"a",
			false,
		},
		{
			"json_object",
			json.RawMessage(`{"key": "value"}`),
			"key",
			"value",
			false,
		},
		{
			"json_array",
			json.RawMessage(`["a", "b", "c"]`),
			"1",
			"b",
			false,
		},
		{
			"nested_map",
			map[string]any{"outer": map[string]any{"inner": "value"}},
			"outer",
			map[string]any{"inner": "value"},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tpl.ResolveValueIndex(ctx, tt.value, tt.index)

			if tt.hasError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("ResolveValueIndex failed: %v", err)
			}

			// For simple types, compare directly
			switch exp := tt.expected.(type) {
			case string:
				if result != exp {
					t.Errorf("got %v, want %v", result, tt.expected)
				}
			case int:
				// JSON unmarshals numbers as float64
				switch r := result.(type) {
				case float64:
					if int(r) != exp {
						t.Errorf("got %v, want %v", result, tt.expected)
					}
				case int:
					if r != exp {
						t.Errorf("got %v, want %v", result, tt.expected)
					}
				default:
					t.Errorf("got type %T, want int", result)
				}
			}
		})
	}
}

func TestResolveValueIndexEdgeCases(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		value    any
		index    string
		expected any
		hasError bool
	}{
		{
			"nil_value",
			nil,
			"key",
			nil,
			false,
		},
		{
			"slice_out_of_bounds_high",
			[]any{"a", "b", "c"},
			"10",
			nil,
			false,
		},
		{
			"slice_out_of_bounds_negative",
			[]any{"a", "b", "c"},
			"-1",
			nil,
			false,
		},
		{
			"slice_invalid_index",
			[]any{"a", "b", "c"},
			"notanumber",
			nil,
			false,
		},
		{
			"string_slice_out_of_bounds",
			[]string{"a", "b", "c"},
			"10",
			nil,
			false,
		},
		{
			"string_slice_invalid_index",
			[]string{"a", "b", "c"},
			"invalid",
			nil,
			false,
		},
		{
			"json_invalid",
			json.RawMessage(`{invalid json}`),
			"key",
			nil,
			true,
		},
		{
			"unhandled_type",
			42, // int is not a valid type for indexing
			"0",
			nil,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tpl.ResolveValueIndex(ctx, tt.value, tt.index)

			if tt.hasError {
				if err == nil {
					t.Errorf("expected error, got nil (result: %v)", result)
				}
				return
			}

			if err != nil {
				t.Fatalf("ResolveValueIndex failed: %v", err)
			}

			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestResolveValueIndexMapJSON(t *testing.T) {
	ctx := context.Background()
	m := map[string]json.RawMessage{"key": json.RawMessage(`"value"`)}
	result, err := tpl.ResolveValueIndex(ctx, m, "key")
	if err != nil {
		t.Fatalf("ResolveValueIndex failed: %v", err)
	}
	resJSON, ok := result.(json.RawMessage)
	if !ok {
		t.Fatalf("expected json.RawMessage, got %T", result)
	}
	if string(resJSON) != `"value"` {
		t.Errorf("got %s, want %s", string(resJSON), `"value"`)
	}
}

func TestArrayAccess(t *testing.T) {
	engine := tpl.New()
	engine.Raw.TemplateData["main"] = `{{_obj/key}}`

	ctx := context.Background()
	ctx = tpl.ValuesCtx(ctx, map[string]any{
		"_obj": map[string]any{"key": "value"},
	})

	if err := engine.Compile(ctx); err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	result, err := engine.ParseAndReturn(ctx, "main")
	if err != nil {
		t.Fatalf("ParseAndReturn failed: %v", err)
	}

	if result != "value" {
		t.Errorf("got %q, want %q", result, "value")
	}
}

func TestNestedArrayAccess(t *testing.T) {
	engine := tpl.New()
	engine.Raw.TemplateData["main"] = `{{_obj/level1/level2}}`

	ctx := context.Background()
	ctx = tpl.ValuesCtx(ctx, map[string]any{
		"_obj": map[string]any{
			"level1": map[string]any{
				"level2": "deep value",
			},
		},
	})

	if err := engine.Compile(ctx); err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	result, err := engine.ParseAndReturn(ctx, "main")
	if err != nil {
		t.Fatalf("ParseAndReturn failed: %v", err)
	}

	if result != "deep value" {
		t.Errorf("got %q, want %q", result, "deep value")
	}
}
