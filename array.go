package tpl

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
)

func ResolveValueIndex(ctx context.Context, v interface{}, s string) (interface{}, error) {
	switch o := v.(type) {
	case ArrayAccessGet:
		return o.OffsetGet(ctx, s)
	case map[string]interface{}:
		return o[s], nil
	case map[string]Value:
		return o[s], nil
	case url.Values:
		return o[s], nil
	case Values:
		n, err := strconv.ParseInt(s, 0, 64)
		if err != nil {
			log.Printf("[tpl] failed to access array element #%s", s)
			return nil, nil
		}
		if n < 0 || int(n) >= len(o) {
			return nil, nil
		}
		return o[n], nil
	case []interface{}:
		n, err := strconv.ParseInt(s, 0, 64)
		if err != nil {
			log.Printf("[tpl] failed to access array element #%s", s)
			return nil, nil
		}
		if n < 0 || int(n) >= len(o) {
			return nil, nil
		}
		return o[n], nil
	case []string:
		n, err := strconv.ParseInt(s, 0, 64)
		if err != nil {
			log.Printf("[tpl] failed to access array element #%s", s)
			return nil, nil
		}
		if n < 0 || int(n) >= len(o) {
			return nil, nil
		}
		return o[n], nil
	case json.RawMessage:
		// parse at json object
		var sub interface{}
		err := json.Unmarshal(o, &sub)
		if err != nil {
			return nil, fmt.Errorf("failed to parse json: %s", err)
		}
		return ResolveValueIndex(ctx, sub, s)
	case Value:
		val, err := o.ReadValue(ctx)
		if err != nil {
			return nil, err
		}
		return ResolveValueIndex(ctx, val, s)
	case nil:
		return nil, nil
	default:
		//log.Printf("unhandled type: %T", val)
		return nil, fmt.Errorf("unhandled type: %T", v)
	}
}
