package tpl_test

import (
	"context"
	"testing"

	"github.com/KarpelesLab/tpl"
)

func TestValuesCtx(t *testing.T) {
	// Base context
	baseCtx := context.Background()

	// Test with empty values map
	emptyValues := map[string]any{}
	ctx1 := tpl.ValuesCtx(baseCtx, emptyValues)

	// Should return the original context if the map is empty
	if ctx1 != baseCtx {
		t.Errorf("ValuesCtx with empty map should return original context")
	}

	// Test with non-empty values map
	testValues := map[string]any{
		"key1": "value1",
		"key2": 123,
	}
	ctx2 := tpl.ValuesCtx(baseCtx, testValues)

	// Should return a new context if the map is not empty
	if ctx2 == baseCtx {
		t.Errorf("ValuesCtx with non-empty map should return a new context")
	}

	// Test value retrieval
	if val := ctx2.Value("key1"); val != "value1" {
		t.Errorf("ctx.Value(\"key1\") = %v, want %v", val, "value1")
	}

	if val := ctx2.Value("key2"); val != 123 {
		t.Errorf("ctx.Value(\"key2\") = %v, want %v", val, 123)
	}

	// Test non-existent key should return nil
	if val := ctx2.Value("nonexistent"); val != nil {
		t.Errorf("ctx.Value(\"nonexistent\") = %v, want nil", val)
	}

	// Test with ValuesCtxAlways which should always return a new context
	ctx3 := tpl.ValuesCtxAlways(baseCtx, emptyValues)
	if ctx3 == baseCtx {
		t.Errorf("ValuesCtxAlways should always return a new context")
	}

	// Test parent context value retrieval
	parentCtx := context.WithValue(baseCtx, "parent_key", "parent_value") //lint:ignore SA1029 template variables use string keys by design
	ctx4 := tpl.ValuesCtx(parentCtx, testValues)

	if val := ctx4.Value("parent_key"); val != "parent_value" {
		t.Errorf("ctx.Value(\"parent_key\") = %v, want %v", val, "parent_value")
	}

	// Test with non-string key (should return parent value)
	if val := ctx4.Value(123); val != nil {
		t.Errorf("ctx.Value(123) = %v, want nil", val)
	}
}

func TestValuesCtxString(t *testing.T) {
	baseCtx := context.Background()
	testValues := map[string]any{
		"key1": "value1",
	}

	ctx := tpl.ValuesCtxAlways(baseCtx, testValues)

	// Cast to valuesCtx to access String method
	if vCtx, ok := ctx.(interface{ String() string }); ok {
		str := vCtx.String()
		if str == "" {
			t.Errorf("valuesCtx.String() returned empty string")
		}
	} else {
		t.Errorf("Failed to cast context to access String method")
	}
}
