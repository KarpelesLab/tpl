//go:build darwin
// +build darwin

package tpl

import (
	"context"
	"errors"
	"os"
	"runtime"
)

func init() {
	RegisterFunction("uname", &TplFunction{Method: fncUname})
}

func fncUname(ctx context.Context, params Values, out WritableValue) error {
	if len(params) < 1 {
		return errors.New("uname() function requires 1 argument")
	}

	switch params[0].WithCtx(ctx).String() {
	case "s":
		out.WriteValue(ctx, "Darwin")
	case "n":
		n, err := os.Hostname()
		if err != nil {
			return err
		}
		out.WriteValue(ctx, n)
	case "r":
		out.WriteValue(ctx, "?")
	case "m":
		out.WriteValue(ctx, runtime.GOARCH)
	default:
		fallthrough
	case "a":
		n, err := os.Hostname()
		if err != nil {
			return err
		}
		// return full uname, ie "s n r v m"
		out.WriteValue(ctx, "Darwin "+n+" "+runtime.GOARCH)
	}
	return nil
}
