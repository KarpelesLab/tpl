package tpl

import (
	"context"
	"fmt"
	"sync"
)

type valuesCtx struct {
	context.Context
	values map[string]interface{}
	mutex  sync.RWMutex
}

func ValuesCtx(parent context.Context, values map[string]interface{}) context.Context {
	if len(values) == 0 {
		return parent
	}
	return &valuesCtx{Context: parent, values: values}
}

func ValuesCtxAlways(parent context.Context, values map[string]interface{}) context.Context {
	return &valuesCtx{Context: parent, values: values}
}

func (c *valuesCtx) String() string {
	return fmt.Sprintf("%v.WithValues(%#v)", c.Context, c.values)
}

func (c *valuesCtx) Value(key interface{}) interface{} {
	if v, ok := key.(string); ok {
		c.mutex.RLock()
		defer c.mutex.RUnlock()
		final, ok := c.values[v]

		if ok {
			return final
		}
	}
	return c.Context.Value(key)
}
