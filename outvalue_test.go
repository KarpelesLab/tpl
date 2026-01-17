package tpl_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/KarpelesLab/tpl"
)

func TestOutValueMethods(t *testing.T) {
	ctx := context.Background()

	// Test NewEmptyValue
	emptyVal := tpl.NewEmptyValue()
	if emptyVal == nil {
		t.Errorf("NewEmptyValue() returned nil")
	}

	// Write some value to it
	emptyVal.Write([]byte("test string"))
	result, err := emptyVal.ReadValue(ctx)
	if err != nil {
		t.Errorf("ReadValue() error = %v", err)
	}

	if s, ok := result.(*bytes.Buffer); !ok || s.String() != "test string" {
		t.Errorf("ReadValue() = %v, want buffer with 'test string'", result)
	}

	// Test AsFloat
	numVal := tpl.NewValue(3.14)
	floatVal := tpl.AsOutValue(ctx, numVal).AsFloat(ctx)
	if floatVal != 3.14 {
		t.Errorf("AsFloat() = %v, want %v", floatVal, 3.14)
	}

	// Test AsInt
	intVal := tpl.AsOutValue(ctx, numVal).AsInt(ctx)
	if intVal != 3 {
		t.Errorf("AsInt() = %v, want %v", intVal, 3)
	}

	// Test WriteString
	buf := &bytes.Buffer{}
	bufVal := tpl.NewValue(buf)
	wVal := tpl.AsOutValue(ctx, bufVal)
	n, err := wVal.WriteString("test string")
	if err != nil {
		t.Errorf("WriteString() error = %v", err)
	}
	if n != 11 {
		t.Errorf("WriteString() = %v, want %v", n, 11)
	}
	if buf.String() != "test string" {
		t.Errorf("Buffer content = %v, want %v", buf.String(), "test string")
	}

	// Test MarshalJSON
	jsonVal := tpl.AsOutValue(ctx, tpl.NewValue(map[string]string{"key": "value"}))
	jsonData, err := jsonVal.MarshalJSON()
	if err != nil {
		t.Errorf("MarshalJSON() error = %v", err)
	}
	if string(jsonData) != `{"key":"value"}` {
		t.Errorf("MarshalJSON() = %v, want %v", string(jsonData), `{"key":"value"}`)
	}
}

func TestNewValueCtx(t *testing.T) {
	ctx := context.Background()
	val := tpl.NewValue("test")

	// Create ValueCtx with NewValueCtx
	valueCtx := tpl.NewValueCtx(ctx, val)
	if valueCtx == nil {
		t.Errorf("NewValueCtx() returned nil")
	}

	// Test Raw method
	raw, err := valueCtx.Raw()
	if err != nil {
		t.Errorf("Raw() error = %v", err)
	}
	if raw != "test" {
		t.Errorf("Raw() = %v, want %v", raw, "test")
	}
}

func TestAsFloatVariousTypes(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		value    any
		expected float64
	}{
		{"float64", float64(3.14), 3.14},
		{"int64", int64(42), 42.0},
		{"uint64", uint64(42), 42.0},
		{"bool_true", true, 1.0},
		{"bool_false", false, 0.0},
		{"string_number", "3.14", 3.14},
		{"nil", nil, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tpl.AsOutValue(ctx, tpl.NewValue(tt.value)).AsFloat(ctx)
			if result != tt.expected {
				t.Errorf("AsFloat() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAsIntVariousTypes(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		value    any
		expected int64
	}{
		{"int64", int64(42), 42},
		{"uint64", uint64(42), 42},
		{"float64", float64(3.7), 4},
		{"bool_true", true, 1},
		{"bool_false", false, 0},
		{"string_number", "42", 42},
		{"nil", nil, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tpl.AsOutValue(ctx, tpl.NewValue(tt.value)).AsInt(ctx)
			if result != tt.expected {
				t.Errorf("AsInt() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAsNumericEdgeCases(t *testing.T) {
	ctx := context.Background()

	// Test with a value that can't be converted to a number
	nonNumVal := tpl.AsOutValue(ctx, tpl.NewValue(map[string]any{}))
	floatResult := nonNumVal.AsFloat(ctx)
	if floatResult != 0 {
		t.Errorf("AsFloat() on non-numeric = %v, want 0", floatResult)
	}
	intResult := nonNumVal.AsInt(ctx)
	if intResult != 0 {
		t.Errorf("AsInt() on non-numeric = %v, want 0", intResult)
	}
}
