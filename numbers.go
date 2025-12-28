package tpl

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"math"
	"net/url"
	"strconv"
)

type valueRawExtractor interface {
	Raw() (interface{}, error)
}

// some helper functions related to numbers
func asBoolIntf(v interface{}) bool {
	switch r := v.(type) {
	case bool:
		return r
	case int:
		return r != 0
	case int64:
		return r != 0
	case uint64:
		return r != 0
	case float64:
		return r != 0
	case *bytes.Buffer:
		if r.Len() > 1 {
			return true
		}
		if r.Len() == 0 || r.String() == "0" {
			return false
		}
		return true
	case string:
		if len(r) > 1 {
			return true
		}
		if len(r) == 0 || r == "0" {
			return false
		}
		return true
	case []byte:
		if len(r) > 1 {
			return true
		}
		if len(r) == 0 || r[0] == '0' {
			return false
		}
		return true
	case map[string]interface{}:
		if len(r) > 0 {
			return true
		}
		return false
	case []interface{}:
		if len(r) > 0 {
			return true
		}
		return false
	case Values:
		if len(r) > 0 {
			return true
		}
		return false
	case json.RawMessage:
		// convert to interface{}, re-run through the process
		var x interface{}
		err := json.Unmarshal(r, &x)
		if err != nil {
			return false
		}
		return asBoolIntf(x)
	case url.Values:
		return len(r) > 0
	case valueRawExtractor:
		rV, err := r.Raw()
		if err != nil {
			return false
		}
		return asBoolIntf(rV)
	case *interfaceValue:
		return asBoolIntf(r.val)
	default:
		return false
	}
}

func fetchNumberInt(ctx context.Context, v interface{}) (int64, bool) {
	switch n := v.(type) {
	case int8:
		return int64(n), true
	case int16:
		return int64(n), true
	case int32:
		return int64(n), true
	case int64:
		return n, true
	case int:
		return int64(n), true
	case uint8:
		return int64(n), true
	case uint16:
		return int64(n), true
	case uint32:
		return int64(n), true
	case uint64:
		if n&(1<<63) != 0 {
			return int64(n), false
		}
		return int64(n), true
	case uint:
		return int64(n), true
	case bool:
		if n {
			return 1, true
		} else {
			return 0, true
		}
	case float32:
		x := math.Round(float64(n))
		y := int64(x)
		return y, float64(y) == x
	case float64:
		x := math.Round(n)
		y := int64(x)
		return y, float64(y) == x
	case string:
		res, err := strconv.ParseInt(n, 0, 64)
		return res, err == nil
	case []byte:
		res, err := strconv.ParseInt(string(n), 0, 64)
		return res, err == nil
	case *bytes.Buffer:
		return fetchNumberInt(ctx, n.String())
	case nil:
		return 0, true
	case ValueReader:
		if nVal, err := n.ReadValue(ctx); err != nil {
			return 0, false
		} else {
			return fetchNumberInt(ctx, nVal)
		}
	default:
		log.Printf("[number] failed to parse type %T", n)
	}

	return 0, false
}

func fetchNumberUint(ctx context.Context, v interface{}) (uint64, bool) {
	switch n := v.(type) {
	case int8:
		return uint64(n), n > 0
	case int16:
		return uint64(n), n > 0
	case int32:
		return uint64(n), n > 0
	case int64:
		return uint64(n), n > 0
	case int:
		return uint64(n), n > 0
	case uint8:
		return uint64(n), true
	case uint16:
		return uint64(n), true
	case uint32:
		return uint64(n), true
	case uint64:
		return n, true
	case uint:
		return uint64(n), true
	case float32:
		if n < 0 {
			return 0, false
		}
		x := math.Round(float64(n))
		y := uint64(x)
		return y, float64(y) == x
	case float64:
		if n < 0 {
			return 0, false
		}
		x := math.Round(n)
		y := uint64(x)
		return y, float64(y) == x
	case bool:
		if n {
			return 1, true
		} else {
			return 0, true
		}
	case string:
		res, err := strconv.ParseUint(n, 0, 64)
		return res, err == nil
	case nil:
		return 0, true
	case ValueReader:
		if nVal, err := n.ReadValue(ctx); err != nil {
			return 0, false
		} else {
			return fetchNumberUint(ctx, nVal)
		}
	}

	return 0, false
}

func fetchNumberFloat(ctx context.Context, v interface{}) (float64, bool) {
	switch n := v.(type) {
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case int:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uintptr:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	case string:
		res, err := strconv.ParseFloat(n, 64)
		return res, err == nil
	case nil:
		return 0, true
	case ValueReader:
		if nVal, err := n.ReadValue(ctx); err != nil {
			return 0, false
		} else {
			return fetchNumberFloat(ctx, nVal)
		}
	}

	res, ok := fetchNumberInt(ctx, v)
	return float64(res), ok
}

func fetchNumberAny(ctx context.Context, v interface{}) (interface{}, bool) {
	switch n := v.(type) {
	case int8:
		return int64(n), true
	case int16:
		return int64(n), true
	case int32:
		return int64(n), true
	case int64:
		return n, true
	case int:
		return int64(n), true
	case uint8:
		return int64(n), true
	case uint16:
		return int64(n), true
	case uint32:
		return int64(n), true
	case uint64:
		return uint64(n), true
	case uintptr:
		return uint64(n), true
	case uint:
		return int64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	case nil:
		return 0, true
	case bool:
		return n, true
	case string:
		if res, err := strconv.ParseInt(n, 0, 64); err == nil {
			return res, true
		}
		if res, err := strconv.ParseFloat(n, 64); err == nil {
			return res, true
		}
		if res, err := strconv.ParseUint(n, 0, 64); err == nil {
			return res, true
		}
		return asBoolIntf(n), false
	case *bytes.Buffer:
		if n.Len() > 100 {
			return nil, false
		}
		return fetchNumberAny(ctx, n.String())
	case *ValueCtx:
		if nVal, err := n.Raw(); err != nil {
			return nil, false
		} else {
			return fetchNumberAny(ctx, nVal)
		}
	case ValueReader:
		if nVal, err := n.ReadValue(ctx); err != nil {
			return nil, false
		} else {
			return fetchNumberAny(ctx, nVal)
		}
	}

	return nil, false
}
