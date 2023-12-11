package tpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/KarpelesLab/pjson"
	"golang.org/x/text/language"
)

type ValueReader interface {
	ReadValue(ctx context.Context) (interface{}, error)
}

type Value interface {
	ReadValue(ctx context.Context) (interface{}, error)
	WithCtx(ctx context.Context) *ValueCtx
}

type WritableValue interface {
	Value
	Printf(fmt string, arg ...interface{}) (int, error)
	Write([]byte) (int, error)
	WriteValue(context.Context, interface{}) error
}

type ValueCtx struct {
	Value
	ctx context.Context
}

type Values []Value

type ArrayAccessGet interface {
	OffsetGet(context.Context, string) (Value, error)
}

type ArrayAccessGetAny interface {
	OffsetGet(context.Context, string) (any, error)
}

type bytableIf interface {
	Bytes() []byte
}

func NewValueCtx(ctx context.Context, v Value) *ValueCtx {
	return &ValueCtx{v, ctx}
}

func (v *ValueCtx) Raw() (interface{}, error) {
	var res interface{}
	var err error

	res = v.Value

	for {
		o, ok := res.(ValueReader)
		if !ok {
			break
		}

		res, err = o.ReadValue(v.ctx)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (v *ValueCtx) MarshalJSON() ([]byte, error) {
	rv, err := v.WithCtx(v.ctx).Raw()
	if err != nil {
		return nil, err
	}
	return json.Marshal(rv)
}

func (v *ValueCtx) ToInt() (int64, bool) {
	return fetchNumberInt(v.ctx, v.Value)
}

func (v *ValueCtx) ToFloat() (float64, bool) {
	return fetchNumberFloat(v.ctx, v.Value)
}

func (v *ValueCtx) ToBool() bool {
	return asBoolIntf(v)
}

func (v *ValueCtx) String() string {
	s, _ := v.StringErr()
	return s
}

// IsString checks if the raw value is either a string, or an interface made to return a string
func (v *ValueCtx) IsString() bool {
	rawValue, err := v.Raw()
	if err != nil {
		return false
	}

	switch rawValue.(type) {
	case string:
		return true
	case []byte:
		return true
	case *bytes.Buffer:
		return true
	case fmt.Stringer:
		return true
	case bytableIf:
		return true
	default:
		return false
	}
}

func (v *ValueCtx) StringErr() (string, error) {
	rawValue, err := v.WithCtx(v.ctx).Raw()
	if err != nil {
		return "", err
	}

	if rawValue == nil {
		return "", nil
	}

	switch rv := rawValue.(type) {
	case string:
		return rv, nil
	case []byte:
		return string(rv), nil
	case fmt.Stringer:
		return rv.String(), nil
	case bytableIf:
		return string(rv.Bytes()), nil
	case func() string:
		return rv(), nil
	case func() []byte:
		return string(rv()), nil
	case func() (Value, error):
		if t, err := rv(); err != nil {
			return "", err
		} else {
			return t.WithCtx(v.ctx).String(), nil
		}
	case Value:
		return rv.WithCtx(v.ctx).StringErr()
	case Values:
		res := &bytes.Buffer{}
		for _, sub := range rv {
			res.WriteString(sub.WithCtx(v.ctx).String())
		}
		return res.String(), nil
	case *ValueCtx:
		return rv.String(), nil
	case bool:
		if rv {
			return "1", nil
		}
		return "", nil
	case json.RawMessage:
		// convert to interface{}, re-run through the process
		var x interface{}
		err := json.Unmarshal(rv, &x)
		if err != nil {
			return "", err
		}
		// run through String() again
		return NewValue(x).WithCtx(v.ctx).StringErr()
	case pjson.RawMessage:
		// convert to interface{}, re-run through the process
		var x interface{}
		err := json.Unmarshal(rv, &x)
		if err != nil {
			return "", err
		}
		// run through String() again
		return NewValue(x).WithCtx(v.ctx).StringErr()
	case interface{ RawJSONBytes() []byte }:
		// convert to any, re-run through the process
		var x any
		err := json.Unmarshal(rv.RawJSONBytes(), &x)
		if err != nil {
			return "", err
		}
		// run through String() again
		return NewValue(x).WithCtx(v.ctx).StringErr()
	case nil:
		return "", nil
	default:
		return fmt.Sprintf("%+v", rv), nil
	}
}

func (v *ValueCtx) Bytes() []byte {
	r, _ := v.BytesErr()
	return r
}

func (v *ValueCtx) BytesErr() ([]byte, error) {
	rawValue, err := v.WithCtx(v.ctx).Raw()
	if err != nil {
		return nil, err
	}

	if rawValue == nil {
		return nil, nil
	}

	switch rv := rawValue.(type) {
	case string:
		return []byte(rv), nil
	case []byte:
		return rv, nil
	case float32:
		return []byte(strconv.FormatFloat(float64(rv), 'g', 14, 32)), nil
	case float64:
		return []byte(strconv.FormatFloat(rv, 'g', 14, 64)), nil
	case bytableIf:
		return rv.Bytes(), nil
	case fmt.Stringer:
		return []byte(rv.String()), nil
	case func() string:
		return []byte(rv()), nil
	case func() []byte:
		return rv(), nil
	case func() (Value, error):
		if t, err := rv(); err != nil {
			return nil, err
		} else {
			return t.WithCtx(v.ctx).Bytes(), nil
		}
	case Value:
		return rv.WithCtx(v.ctx).BytesErr()
	case Values:
		res := &bytes.Buffer{}
		for _, sub := range rv {
			res.Write(sub.WithCtx(v.ctx).Bytes())
		}
		return res.Bytes(), nil
	case *ValueCtx:
		return rv.Bytes(), nil
	case bool:
		if rv {
			return []byte{'1'}, nil
		}
		return []byte{}, nil
	case json.RawMessage:
		// convert to interface{}, re-run through the process
		var x interface{}
		err := json.Unmarshal(rv, &x)
		if err != nil {
			return nil, err
		}
		// run through String() again
		return NewValue(x).WithCtx(v.ctx).BytesErr()
	case pjson.RawMessage:
		// convert to interface{}, re-run through the process
		var x interface{}
		err := json.Unmarshal(rv, &x)
		if err != nil {
			return nil, err
		}
		// run through String() again
		return NewValue(x).WithCtx(v.ctx).BytesErr()
	case interface{ RawJSONBytes() []byte }:
		// convert to any, re-run through the process
		var x any
		err := json.Unmarshal(rv.RawJSONBytes(), &x)
		if err != nil {
			return nil, err
		}
		// run through String() again
		return NewValue(x).WithCtx(v.ctx).BytesErr()
	case nil:
		return nil, nil
	default:
		return []byte(fmt.Sprintf("%+v", rv)), nil
	}
}

func (v *ValueCtx) MatchValueType(t interface{}) (interface{}, error) {
	rv, err := v.WithCtx(v.ctx).Raw()
	if err != nil {
		return nil, err
	}

	switch realT := t.(type) {
	case bool:
		if x, ok := rv.(bool); ok {
			return x, nil
		}
		y, ok := fetchNumberAny(v.ctx, rv)
		if !ok {
			return asBoolIntf(rv), nil
		}
		switch z := y.(type) {
		case int64:
			return z != 0, nil
		case uint64:
			return z != 0, nil
		case float64:
			return z != 0, nil
		case bool: // shouldn't happen I guess...
			return z, nil
		case nil:
			return false, nil
		default:
			return false, nil
		}
	case int8:
		if x, ok := fetchNumberInt(v.ctx, v); ok {
			// TODO check range
			return int8(x), nil
		}
		return nil, fmt.Errorf("%#v not a number", v)
	case int16:
		if x, ok := fetchNumberInt(v.ctx, v); ok {
			// TODO check range
			return int16(x), nil
		}
		return nil, fmt.Errorf("%#v not a number", v)
	case int32:
		if x, ok := fetchNumberInt(v.ctx, v); ok {
			// TODO check range
			return int32(x), nil
		}
		return nil, fmt.Errorf("%#v not a number", v)
	case int64:
		if x, ok := fetchNumberInt(v.ctx, v); ok {
			// TODO check range
			return int64(x), nil
		}
		return nil, fmt.Errorf("%#v not a number", v)
	case int:
		if x, ok := fetchNumberInt(v.ctx, v); ok {
			// TODO check range
			return int(x), nil
		}
		return nil, fmt.Errorf("%#v not a number", v)
	case uint8:
		if x, ok := fetchNumberUint(v.ctx, v); ok {
			// TODO check range
			return uint8(x), nil
		}
		return nil, fmt.Errorf("%#v not a number", v)
	case uint16:
		if x, ok := fetchNumberUint(v.ctx, v); ok {
			// TODO check range
			return uint16(x), nil
		}
		return nil, fmt.Errorf("%#v not a number", v)
	case uint32:
		if x, ok := fetchNumberUint(v.ctx, v); ok {
			// TODO check range
			return uint32(x), nil
		}
		return nil, fmt.Errorf("%#v not a number", v)
	case uint64:
		if x, ok := fetchNumberUint(v.ctx, v); ok {
			// TODO check range
			return uint64(x), nil
		}
		return nil, fmt.Errorf("%#v not a number", v)
	case uint:
		if x, ok := fetchNumberUint(v.ctx, v); ok {
			// TODO check range
			return uint(x), nil
		}
		return nil, fmt.Errorf("%#v not a number", v)
	case float32:
		if x, ok := fetchNumberFloat(v.ctx, v); ok {
			// TODO check range
			return float32(x), nil
		}
		return nil, fmt.Errorf("%#v not a number", v)
	case float64:
		if x, ok := fetchNumberFloat(v.ctx, v); ok {
			// TODO check range
			return x, nil
		}
		return nil, fmt.Errorf("%#v not a number", v)
	case string:
		return v.WithCtx(v.ctx).StringErr()
	case []byte:
		return v.WithCtx(v.ctx).BytesErr()
	case *bytes.Buffer:
		b, err := v.WithCtx(v.ctx).BytesErr()
		buf := &bytes.Buffer{}
		buf.Write(b)
		return buf, err
	case language.Tag:
		return language.Parse(v.WithCtx(v.ctx).String())
	case Value:
		x, err := realT.WithCtx(v.ctx).Raw()
		if err != nil {
			return nil, err
		}
		return v.MatchValueType(x)
	default:
		return nil, fmt.Errorf("unsupported format for conversion: %T", t)
	}
}
