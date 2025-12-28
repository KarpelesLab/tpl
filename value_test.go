package tpl_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/KarpelesLab/tpl"
)

func TestValueCtxStringErr(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		value    any
		expected string
	}{
		{"string", "hello", "hello"},
		{"bytes", []byte("hello"), "hello"},
		{"buffer", bytes.NewBufferString("hello"), "hello"},
		{"bool_true", true, "1"},
		{"bool_false", false, ""},
		{"nil", nil, ""},
		{"int", 42, "42"},
		{"float", 3.14, "3.14"},
		{"json_raw", json.RawMessage(`"hello"`), "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tpl.NewValue(tt.value)
			result := v.WithCtx(ctx).String()
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestValueCtxBytesErr(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		value    any
		expected []byte
	}{
		{"string", "hello", []byte("hello")},
		{"bytes", []byte("hello"), []byte("hello")},
		{"buffer", bytes.NewBufferString("hello"), []byte("hello")},
		{"bool_true", true, []byte{'1'}},
		{"bool_false", false, []byte{}},
		{"nil", nil, nil},
		{"float32", float32(1.5), []byte("1.5")}, // Use value with exact binary representation
		{"float64", float64(3.14), []byte("3.14")},
		{"json_raw", json.RawMessage(`"hello"`), []byte("hello")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tpl.NewValue(tt.value)
			result := v.WithCtx(ctx).Bytes()
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestValueCtxIsString(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		value    any
		expected bool
	}{
		{"string", "hello", true},
		{"bytes", []byte("hello"), true},
		{"buffer", bytes.NewBufferString("hello"), true},
		{"int", 42, false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tpl.NewValue(tt.value)
			result := v.WithCtx(ctx).IsString()
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValueCtxToInt(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		value    any
		expected int64
		ok       bool
	}{
		{"int", 42, 42, true},
		{"string", "42", 42, true},
		{"float", 3.7, 4, true},
		{"invalid", "abc", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tpl.NewValue(tt.value)
			result, ok := v.WithCtx(ctx).ToInt()
			if ok != tt.ok {
				t.Errorf("ok: got %v, want %v", ok, tt.ok)
			}
			if ok && result != tt.expected {
				t.Errorf("got %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestValueCtxToFloat(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		value    any
		expected float64
		ok       bool
	}{
		{"int", 42, 42.0, true},
		{"float", 3.14, 3.14, true},
		{"string", "3.14", 3.14, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tpl.NewValue(tt.value)
			result, ok := v.WithCtx(ctx).ToFloat()
			if ok != tt.ok {
				t.Errorf("ok: got %v, want %v", ok, tt.ok)
			}
			if ok && result != tt.expected {
				t.Errorf("got %f, want %f", result, tt.expected)
			}
		})
	}
}

func TestValueCtxToBool(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		value    any
		expected bool
	}{
		{"true", true, true},
		{"false", false, false},
		{"int_1", 1, true},
		{"int_0", 0, false},
		{"string_1", "1", true},
		{"string_0", "0", false},
		{"string_empty", "", false},
		{"string_nonempty", "hello", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tpl.NewValue(tt.value)
			result := v.WithCtx(ctx).ToBool()
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValueCtxMarshalJSON(t *testing.T) {
	ctx := context.Background()

	v := tpl.NewValue("hello")
	vc := v.WithCtx(ctx)

	data, err := json.Marshal(vc)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	if string(data) != `"hello"` {
		t.Errorf("got %s, want %s", data, `"hello"`)
	}
}

func TestValueCtxMatchValueType(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		value    any
		target   any
		expected any
	}{
		{"to_bool_true", "1", true, true},
		{"to_bool_false", "0", false, false},
		{"to_bool_from_int", 42, true, true},
		{"to_bool_from_zero", 0, false, false},
		{"to_int", "42", int(0), int(42)},
		{"to_int8", "42", int8(0), int8(42)},
		{"to_int16", "42", int16(0), int16(42)},
		{"to_int32", "42", int32(0), int32(42)},
		{"to_int64", "42", int64(0), int64(42)},
		{"to_uint8", "42", uint8(0), uint8(42)},
		{"to_uint16", "42", uint16(0), uint16(42)},
		{"to_uint32", "42", uint32(0), uint32(42)},
		{"to_uint64", "42", uint64(0), uint64(42)},
		{"to_uint", "42", uint(0), uint(42)},
		{"to_float32", "3.14", float32(0), float32(3.14)},
		{"to_float64", "3.14", float64(0), float64(3.14)},
		{"to_string", 42, "", "42"},
		{"to_bytes", "hello", []byte{}, []byte("hello")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tpl.NewValue(tt.value)
			result, err := v.WithCtx(ctx).MatchValueType(tt.target)
			if err != nil {
				t.Fatalf("MatchValueType failed: %v", err)
			}

			switch exp := tt.expected.(type) {
			case []byte:
				if !bytes.Equal(result.([]byte), exp) {
					t.Errorf("got %v, want %v", result, tt.expected)
				}
			default:
				if result != tt.expected {
					t.Errorf("got %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestValueCtxMatchValueTypeBoolFromNumber(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		value    any
		expected bool
	}{
		{"from_int64_nonzero", int64(42), true},
		{"from_int64_zero", int64(0), false},
		{"from_uint64_nonzero", uint64(42), true},
		{"from_uint64_zero", uint64(0), false},
		{"from_float64_nonzero", float64(3.14), true},
		{"from_float64_zero", float64(0), false},
		{"from_nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tpl.NewValue(tt.value)
			result, err := v.WithCtx(ctx).MatchValueType(true)
			if err != nil {
				t.Fatalf("MatchValueType failed: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValueCtxMatchValueTypeBuffer(t *testing.T) {
	ctx := context.Background()

	v := tpl.NewValue("hello")
	result, err := v.WithCtx(ctx).MatchValueType(&bytes.Buffer{})
	if err != nil {
		t.Fatalf("MatchValueType failed: %v", err)
	}
	buf, ok := result.(*bytes.Buffer)
	if !ok {
		t.Fatalf("expected *bytes.Buffer, got %T", result)
	}
	if buf.String() != "hello" {
		t.Errorf("got %q, want %q", buf.String(), "hello")
	}
}

