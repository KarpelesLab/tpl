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
