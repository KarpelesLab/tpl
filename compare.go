package tpl

import (
	"bytes"
	"context"
)

// TODO use reflection and implement better comparision?

func CompareValues(ctx context.Context, o1, o2 interface{}) (bool, error) {
	// make sure o1 is never null
	if o1 == nil {
		if o2 == nil {
			// both nil
			return true, nil
		}
		o1 = o2
		o2 = nil
	}

	// avoid comparing pointers
	if o1b, ok := o1.(*bytes.Buffer); ok {
		o1 = o1b.Bytes()
	}

	if typePriority(o1) < typePriority(o2) {
		// should use o2's type rather than o1. swap values
		o2, o1 = o1, o2
	}

	var err error
	o2, err = NewValue(o2).WithCtx(ctx).MatchValueType(o1)
	if err != nil {
		return false, err
	}

	switch o1v := o1.(type) {
	case []byte:
		return bytes.Equal(o1v, o2.([]byte)), nil
	default:
		return o1 == o2, nil
	}
}

func typePriority(v interface{}) int {
	switch v.(type) {
	case bool:
		return 3
	case int, int64, uint64, float64:
		return 3
	case string, []byte, *bytes.Buffer:
		return 2
	case nil:
		return -1
	default:
		return 1 // ??
	}
}
