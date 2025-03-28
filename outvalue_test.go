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