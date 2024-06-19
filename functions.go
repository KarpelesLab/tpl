package tpl

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"runtime"

	"github.com/KarpelesLab/webutil"
)

type TplFuncCallback func(ctx context.Context, params Values, out WritableValue) error

type TplFunction struct {
	Method     TplFuncCallback
	CanCompile bool
}

func init() {
	RegisterFunction("error", &TplFunction{Method: fncError})
	RegisterFunction("redirect", &TplFunction{Method: fncRedirect})
	RegisterFunction("string", &TplFunction{Method: fncString, CanCompile: true})
	RegisterFunction("rand", &TplFunction{Method: fncRand})
	RegisterFunction("printf", &TplFunction{Method: fncPrintf, CanCompile: true})
	// exists
	// urlstamp
	// import
	// match
	// assert
	// urlget
	// price
	RegisterFunction("phpversion", &TplFunction{Method: fncPhpversion, CanCompile: true})
	RegisterFunction("seq", &TplFunction{Method: fncSeq, CanCompile: true})
	// locale
	// switchlanguage
	// request
	// audio
}

func CallFunction(ctx context.Context, funcName string, params Values, target WritableValue) error {
	if f, ok := ctx.Value("@" + funcName).(TplFuncCallback); ok {
		// custom in-context function
		return f(ctx, params, target)
	} else if f, ok := tplFunctions[funcName]; ok {
		// only call if ok
		return f.Method(ctx, params, target)
	} else {
		return fmt.Errorf("tpl: call to undefined function %s", funcName)
	}
}

func fncError(ctx context.Context, params Values, out WritableValue) error {
	if len(params) == 0 {
		return errors.New("error() requires at least a format")
	}
	var pformat string
	var arg []interface{}
	for i, p := range params {
		if i == 0 {
			pformat = p.WithCtx(ctx).String()
			continue
		}
		if rv, err := p.ReadValue(ctx); err != nil {
			return err
		} else {
			arg = append(arg, rv)
		}
	}

	return fmt.Errorf(pformat, arg...)
}

func fncRedirect(ctx context.Context, params Values, out WritableValue) error {
	if len(params) != 1 {
		return errors.New("redirect() requires one parameter")
	}

	u := params[0].WithCtx(ctx).String()
	uP, err := url.Parse(u)
	if err != nil {
		return err
	}

	return webutil.RedirectError(uP)
}

func fncString(ctx context.Context, params Values, out WritableValue) error {
	for _, p := range params {
		out.Write(p.WithCtx(ctx).Bytes())
	}
	return nil
}

func fncRand(ctx context.Context, params Values, out WritableValue) error {
	if len(params) < 2 {
		return errors.New("rand() requires 2 arguments")
	}

	min, ok := params[0].WithCtx(ctx).ToInt()
	if !ok {
		return errors.New("rand() failed to parse minimum value")
	}
	max, ok := params[1].WithCtx(ctx).ToInt()
	if !ok {
		return errors.New("rand() failed to parse maximum value")
	}

	if min == max {
		return out.WriteValue(ctx, min)
	}

	if min >= max {
		return errors.New("rand() invalid values, min needs to be lower than max")
	}

	return out.WriteValue(ctx, rand.Int63n(max-min)+min)
}

func fncPhpversion(ctx context.Context, params Values, out WritableValue) error {
	out.WriteValue(ctx, runtime.Version())
	return nil
}

func fncUnameHelperToString[T int8 | uint8](v [65]T) string {
	out := make([]byte, len(v))
	for i := 0; i < len(v); i++ {
		if v[i] == 0 {
			return string(out[:i])
		}
		out[i] = byte(v[i])
	}
	return string(out)
}

func fncSeq(ctx context.Context, params Values, out WritableValue) error {
	if len(params) < 2 {
		return errors.New("seq() function requires 2 arguments")
	}

	start := AsOutValue(ctx, params[0]).AsInt(ctx)
	end := AsOutValue(ctx, params[1]).AsInt(ctx)
	step := int64(1)
	if len(params) >= 3 {
		step = AsOutValue(ctx, params[2]).AsInt(ctx)
	}

	if end < start {
		out.WriteValue(ctx, Values{})
		return nil
	}

	res := make(Values, (1+end-start)/step)

	for i := start; i <= end; i += step {
		res[i-start] = NewValue(i)
	}
	out.WriteValue(ctx, res)
	return nil
}

func fncPrintf(ctx context.Context, params Values, out WritableValue) error {
	if len(params) == 0 {
		return errors.New("printf() requires at least a format")
	}
	var pformat string
	var arg []interface{}
	for i, p := range params {
		if i == 0 {
			pformat = p.WithCtx(ctx).String()
			continue
		}
		if rv, err := p.ReadValue(ctx); err != nil {
			return err
		} else {
			arg = append(arg, rv)
		}
	}

	_, err := out.Printf(pformat, arg...)
	return err
}
