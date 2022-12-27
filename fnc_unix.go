//go:build linux

package tpl

import (
	"context"
	"errors"
	"syscall"
)

func init() {
	RegisterFunction("uname", &TplFunction{Method: fncUname, CanCompile: true})
}

func fncUname(ctx context.Context, params Values, out WritableValue) error {
	if len(params) < 1 {
		return errors.New("uname() function requires 1 argument")
	}

	var name syscall.Utsname
	if err := syscall.Uname(&name); err != nil {
		return err
	}

	switch params[0].WithCtx(ctx).String() {
	case "s":
		out.WriteValue(ctx, fncUnameHelperToString(name.Sysname))
	case "n":
		out.WriteValue(ctx, fncUnameHelperToString(name.Nodename)+"."+fncUnameHelperToString(name.Domainname))
	case "r":
		out.WriteValue(ctx, fncUnameHelperToString(name.Release))
	case "v":
		out.WriteValue(ctx, fncUnameHelperToString(name.Version))
	case "m":
		out.WriteValue(ctx, fncUnameHelperToString(name.Machine))
	default:
		fallthrough
	case "a":
		// return full uname, ie "s n r v m"
		out.WriteValue(ctx, fncUnameHelperToString(name.Sysname)+" "+fncUnameHelperToString(name.Nodename)+"."+fncUnameHelperToString(name.Domainname)+" "+fncUnameHelperToString(name.Release)+" "+fncUnameHelperToString(name.Version)+" "+fncUnameHelperToString(name.Machine))
	}
	return nil
}
