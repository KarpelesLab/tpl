package tpl_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/KarpelesLab/tpl"
)

func TestForeach(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]any
		expected string
	}{
		{
			"slice",
			`{{foreach {{_arr}} as _item}}{{_item}}{{/foreach}}`,
			map[string]any{"_arr": []string{"a", "b", "c"}},
			"abc",
		},
		{
			"slice_with_key",
			`{{foreach {{_arr}} as _item}}{{_item_key}}:{{_item}},{{/foreach}}`,
			map[string]any{"_arr": []string{"a", "b", "c"}},
			"0:a,1:b,2:c,",
		},
		{
			"slice_with_idx",
			`{{foreach {{_arr}} as _item}}{{_item_idx}},{{/foreach}}`,
			map[string]any{"_arr": []string{"a", "b", "c"}},
			"1,2,3,",
		},
		{
			"map",
			`{{foreach {{_map}} as _item}}{{_item_key}}={{_item}};{{/foreach}}`,
			map[string]any{"_map": map[string]any{"x": 1, "y": 2}},
			"",
		},
		{
			"empty_with_else",
			`{{foreach {{_arr}} as _item}}{{_item}}{{else}}empty{{/foreach}}`,
			map[string]any{"_arr": []string{}},
			"empty",
		},
		{
			"interface_slice",
			`{{foreach {{_arr}} as _item}}{{_item}}{{/foreach}}`,
			map[string]any{"_arr": []any{1, 2, 3}},
			"123",
		},
		{
			"json_array",
			`{{foreach {{_arr}} as _item}}{{_item}}{{/foreach}}`,
			map[string]any{"_arr": json.RawMessage(`[1, 2, 3]`)},
			"123",
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

			// For map iteration, order is not guaranteed, so skip strict comparison
			if tt.name == "map" {
				return
			}

			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestForeachPrevValue(t *testing.T) {
	engine := tpl.New()
	engine.Raw.TemplateData["main"] = `{{foreach {{_arr}} as _item}}{{if {{_item_prv}}}}prev={{_item_prv}},{{/if}}cur={{_item}};{{/foreach}}`

	ctx := context.Background()
	ctx = tpl.ValuesCtx(ctx, map[string]any{
		"_arr": []string{"a", "b", "c"},
	})

	if err := engine.Compile(ctx); err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	result, err := engine.ParseAndReturn(ctx, "main")
	if err != nil {
		t.Fatalf("ParseAndReturn failed: %v", err)
	}

	expected := "cur=a;prev=a,cur=b;prev=b,cur=c;"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestForeachMax(t *testing.T) {
	engine := tpl.New()
	engine.Raw.TemplateData["main"] = `{{foreach {{_arr}} as _item}}{{_item_max}}{{/foreach}}`

	ctx := context.Background()
	ctx = tpl.ValuesCtx(ctx, map[string]any{
		"_arr": []string{"a", "b", "c"},
	})

	if err := engine.Compile(ctx); err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	result, err := engine.ParseAndReturn(ctx, "main")
	if err != nil {
		t.Fatalf("ParseAndReturn failed: %v", err)
	}

	if result != "333" {
		t.Errorf("got %q, want %q", result, "333")
	}
}

func TestForeachNil(t *testing.T) {
	engine := tpl.New()
	engine.Raw.TemplateData["main"] = `{{foreach {{_arr}} as _item}}{{_item}}{{else}}empty{{/foreach}}`

	ctx := context.Background()
	ctx = tpl.ValuesCtx(ctx, map[string]any{
		"_arr": nil,
	})

	if err := engine.Compile(ctx); err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	result, err := engine.ParseAndReturn(ctx, "main")
	if err != nil {
		t.Fatalf("ParseAndReturn failed: %v", err)
	}

	if result != "empty" {
		t.Errorf("got %q, want %q", result, "empty")
	}
}

func TestForeachMapStringValue(t *testing.T) {
	engine := tpl.New()
	engine.Raw.TemplateData["main"] = `{{foreach {{_map}} as _item}}{{_item}};{{/foreach}}`

	ctx := context.Background()
	ctx = tpl.ValuesCtx(ctx, map[string]any{
		"_map": map[string]tpl.Value{"a": tpl.NewValue("1")},
	})

	if err := engine.Compile(ctx); err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	result, err := engine.ParseAndReturn(ctx, "main")
	if err != nil {
		t.Fatalf("ParseAndReturn failed: %v", err)
	}

	if result != "1;" {
		t.Errorf("got %q, want %q", result, "1;")
	}
}

func TestForeachMapStringJSON(t *testing.T) {
	engine := tpl.New()
	engine.Raw.TemplateData["main"] = `{{foreach {{_map}} as _item}}{{_item}};{{/foreach}}`

	ctx := context.Background()
	ctx = tpl.ValuesCtx(ctx, map[string]any{
		"_map": map[string]json.RawMessage{"x": json.RawMessage(`"hello"`)},
	})

	if err := engine.Compile(ctx); err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	result, err := engine.ParseAndReturn(ctx, "main")
	if err != nil {
		t.Fatalf("ParseAndReturn failed: %v", err)
	}

	if result != "hello;" {
		t.Errorf("got %q, want %q", result, "hello;")
	}
}

func TestForeachSeq(t *testing.T) {
	engine := tpl.New()
	engine.Raw.TemplateData["main"] = `{{foreach {{@seq(1, 5)}} as _num}}{{_num}}{{/foreach}}`

	ctx := context.Background()

	if err := engine.Compile(ctx); err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	result, err := engine.ParseAndReturn(ctx, "main")
	if err != nil {
		t.Fatalf("ParseAndReturn failed: %v", err)
	}

	if result != "12345" {
		t.Errorf("got %q, want %q", result, "12345")
	}
}

func TestForeachSeqWithStep(t *testing.T) {
	engine := tpl.New()
	engine.Raw.TemplateData["main"] = `{{foreach {{@seq(0, 10, 2)}} as _num}}{{_num}},{{/foreach}}`

	ctx := context.Background()

	if err := engine.Compile(ctx); err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	result, err := engine.ParseAndReturn(ctx, "main")
	if err != nil {
		t.Fatalf("ParseAndReturn failed: %v", err)
	}

	if result != "0,2,4,6,8,10," {
		t.Errorf("got %q, want %q", result, "0,2,4,6,8,10,")
	}
}

func TestForeachWithValues(t *testing.T) {
	engine := tpl.New()
	engine.Raw.TemplateData["main"] = `{{foreach {{_arr}} as _item}}{{_item}},{{/foreach}}`

	ctx := context.Background()
	ctx = tpl.ValuesCtx(ctx, map[string]any{
		"_arr": tpl.Values{tpl.NewValue("x"), tpl.NewValue("y"), tpl.NewValue("z")},
	})

	if err := engine.Compile(ctx); err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	result, err := engine.ParseAndReturn(ctx, "main")
	if err != nil {
		t.Fatalf("ParseAndReturn failed: %v", err)
	}

	if result != "x,y,z," {
		t.Errorf("got %q, want %q", result, "x,y,z,")
	}
}

func TestForeachWithMapAny(t *testing.T) {
	engine := tpl.New()
	engine.Raw.TemplateData["main"] = `{{foreach {{_map}} as _item}}{{_item_key}}={{_item}};{{/foreach}}`

	ctx := context.Background()
	ctx = tpl.ValuesCtx(ctx, map[string]any{
		"_map": map[string]any{"x": "1"},
	})

	if err := engine.Compile(ctx); err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	result, err := engine.ParseAndReturn(ctx, "main")
	if err != nil {
		t.Fatalf("ParseAndReturn failed: %v", err)
	}


	if result != "x=1;" {
		t.Errorf("got %q, want %q", result, "x=1;")
	}
}
