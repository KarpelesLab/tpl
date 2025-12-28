package tpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
)

// interfaceValue is a simple container for interface{} type that will automatically
// spawn a bytes.Buffer when attempting to write on an empty value, and will
// handle string conversion of various kinds of values.
type interfaceValue struct {
	val interface{}
}

func NewValue(v interface{}) Value {
	if valueV, ok := v.(Value); ok {
		return valueV
	}
	return &interfaceValue{v}
}

func NewEmptyValue() WritableValue {
	return &interfaceValue{nil}
}

func makeValue(v interface{}) *interfaceValue {
	if valueV, ok := v.(*interfaceValue); ok {
		return valueV
	}
	return &interfaceValue{v}
}

func AsOutValue(ctx context.Context, v interface{}) *interfaceValue {
	if valueV, ok := v.(*interfaceValue); ok {
		return valueV
	}
	if valueV, ok := v.(Value); ok {
		if final, err := valueV.WithCtx(ctx).Raw(); err != nil {
			panic(err)
		} else {
			return &interfaceValue{final}
		}
	}
	return &interfaceValue{v}
}

func (v *interfaceValue) WithCtx(ctx context.Context) *ValueCtx {
	return &ValueCtx{v, ctx}
}

func (v *interfaceValue) AsBool(ctx context.Context) bool {
	return asBoolIntf(v.val)
}

func (v *interfaceValue) AsFloat(ctx context.Context) float64 {
	nv := v.AsNumeric(ctx)
	switch t := nv.val.(type) {
	case float64:
		return t
	case int64:
		return float64(t)
	case uint64:
		return float64(t)
	case bool:
		if t {
			return 1
		} else {
			return 0
		}
	default:
		return 0
	}
}

func (v *interfaceValue) AsInt(ctx context.Context) int64 {
	nv := v.AsNumeric(ctx)
	switch t := nv.val.(type) {
	case int64:
		return t
	case uint64:
		return int64(t)
	case float64:
		return int64(math.Round(t))
	case bool:
		if t {
			return 1
		} else {
			return 0
		}
	default:
		return 0
	}
}

func (v *interfaceValue) AsNumeric(ctx context.Context) *interfaceValue {
	// try to parse value
	if r, ok := fetchNumberAny(ctx, v.val); ok {
		return &interfaceValue{r}
	}

	// bad
	return &interfaceValue{}
}

func (v *interfaceValue) WriteValue(ctx context.Context, val interface{}) error {
	if v.val == nil {
		switch rv := val.(type) {
		case *interfaceValue:
			return v.WriteValue(ctx, rv.val)
		case *bytes.Buffer:
			v.val = rv.Bytes()
			return nil
		default:
			v.val = val
			return nil
		}
	}

	if err := v.forceBuffer(ctx); err != nil {
		return err
	}

	wv, err := NewValue(val).WithCtx(ctx).BytesErr()
	if err != nil {
		return err
	}
	_, err = v.Write(wv)
	return err
}

func (v *interfaceValue) ReadValue(ctx context.Context) (interface{}, error) {
	switch rv := v.val.(type) {
	case *ValueCtx:
		return rv.Raw()
	default:
		return v.val, nil
	}
}

func (v *interfaceValue) forceBuffer(ctx context.Context) error {
	// force v to be a buffer
	if _, ok := v.val.(io.Writer); ok {
		return nil
	}

	dat, err := v.WithCtx(ctx).BytesErr()
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	buf.Write(dat)

	v.val = buf
	return nil
}

func (v *interfaceValue) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if v.val == nil {
		buf := &bytes.Buffer{}
		v.val = buf
		return buf.Write(p)
	}

	switch rv := v.val.(type) {
	case io.Writer: // will match bytes.Buffer
		return rv.Write(p)
	case string:
		v.val = rv + string(p)
		return len(p), nil
	case []byte:
		v.val = append(rv, p...)
		return len(p), nil
	case fmt.Stringer:
		b := &bytes.Buffer{}
		b.Write([]byte(rv.String()))
		v.val = b
		return b.Write(p)
	default:
		b := &bytes.Buffer{}
		fmt.Fprintf(b, "%+v", v.val)
		v.val = b
		return b.Write(p)
	}
}

func (v *interfaceValue) WriteString(s string) (int, error) {
	// for convenience
	return v.Write([]byte(s))
}

func (v *interfaceValue) Printf(format string, arg ...interface{}) (int, error) {
	return fmt.Fprintf(v, format, arg...)
}

func (v *interfaceValue) MarshalJSON() ([]byte, error) {
	switch r := v.val.(type) {
	case []byte:
		return json.Marshal(string(r))
	case *bytes.Buffer:
		return json.Marshal(r.String())
	default:
		return json.Marshal(v.val)
	}
}
