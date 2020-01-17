package tpl

import (
	"context"
	"log"
)

type TplCtxValue int

const (
	TplCtxLog TplCtxValue = iota + 1
)

type CtxLog interface {
	LogWarn(msg string, arg ...interface{})
}

func LogWarn(ctx context.Context, msg string, arg ...interface{}) {
	if c, ok := ctx.Value(TplCtxLog).(CtxLog); !ok {
		log.Printf("[tpl] WARN: "+msg, arg...)
	} else {
		c.LogWarn(msg, arg...)
	}
}
