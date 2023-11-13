package tpl

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/KarpelesLab/webutil"
)

func QueryEscapeAny(ctx context.Context, val interface{}) string {
	switch v := val.(type) {
	case string:
		return url.QueryEscape(v)
	case []byte:
		return url.QueryEscape(string(v))
	case fmt.Stringer:
		return url.QueryEscape(v.String())
	case bytableIf:
		return url.QueryEscape(string(v.Bytes()))
	case func() string:
		return url.QueryEscape(v())
	case func() []byte:
		return url.QueryEscape(string(v()))
	case func() (Value, error):
		rv, err := v()
		if err != nil {
			LogWarn(ctx, "unable to read value in QueryEscapeAny: %s", err)
			return ""
		}
		return QueryEscapeAny(ctx, rv)
	case bytes.Buffer:
		return url.QueryEscape(v.String())
	case strings.Builder:
		return url.QueryEscape(v.String())
	case ValueReader:
		rv, err := v.ReadValue(ctx)
		if err != nil {
			LogWarn(ctx, "unable to read value in QueryEscapeAny: %s", err)
			return ""
		}
		return QueryEscapeAny(ctx, rv)
	case map[string]interface{}:
		return webutil.EncodePhpQuery(v)
	case int, int64, uint64:
		return fmt.Sprintf("%d", v)
	case nil:
		return ""
	default:
		LogWarn(ctx, "unable to handle type %T in QueryEscapeAny", val)
		return ""
	}
}
