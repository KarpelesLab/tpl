package tpl

import (
	"context"
	"encoding/json"
	"fmt"
)

func foreachAny(ctx context.Context, val interface{}, elementF func(k, v interface{}, idx, max int64) error) (int64, error) {
	idx := int64(0)

	switch valT := val.(type) {
	case Values:
		max := int64(len(valT))
		for kT, vT := range valT {
			idx += 1
			if err := elementF(kT, vT, idx, max); err != nil {
				return idx, err
			}
		}
		return idx, nil
	case []interface{}:
		max := int64(len(valT))
		for kT, vT := range valT {
			idx += 1
			if err := elementF(kT, vT, idx, max); err != nil {
				return idx, err
			}
		}
		return idx, nil
	case []string:
		max := int64(len(valT))
		for kT, vT := range valT {
			idx += 1
			if err := elementF(kT, vT, idx, max); err != nil {
				return idx, err
			}
		}
		return idx, nil
	case map[string]Value:
		max := int64(len(valT))
		for kT, vT := range valT {
			idx += 1
			if err := elementF(kT, vT, idx, max); err != nil {
				return idx, err
			}
		}
		return idx, nil
	case map[string]interface{}:
		max := int64(len(valT))
		for kT, vT := range valT {
			idx += 1
			if err := elementF(kT, vT, idx, max); err != nil {
				return idx, err
			}
		}
		return idx, nil
	case json.RawMessage:
		var d interface{}
		if err := json.Unmarshal(valT, &d); err != nil {
			return 0, err
		}
		return foreachAny(ctx, d, elementF)
	case Value:
		v, err := valT.ReadValue(ctx)
		if err != nil {
			return 0, err
		}
		return foreachAny(ctx, v, elementF)
	// various cases from https://godoc.org/github.com/emirpasic/gods/containers
	case interface {
		All(func(key interface{}, value interface{}) bool) bool
		Size() int
	}:
		var err error
		max := int64(valT.Size())

		valT.All(func(key, value interface{}) bool {
			idx += 1
			if e := elementF(key, value, idx, max); e != nil {
				err = e
				return false
			}
			return true
		})
		return idx, err
	case interface {
		All(func(index int, value interface{}) bool) bool
		Size() int
	}:
		var err error
		max := int64(valT.Size())

		valT.All(func(key int, value interface{}) bool {
			idx += 1
			if e := elementF(key, value, idx, max); e != nil {
				err = e
				return false
			}
			return true
		})
		return idx, err
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unsupported type for foreach: %T", val)
	}
}
