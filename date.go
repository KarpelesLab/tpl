package tpl

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/MagicalTux/strftime"
	"golang.org/x/text/language"
)

func init() {
	RegisterFilter("date", fltDate)
	RegisterFilter("duration", fltDuration)
}

func fltDate(ctx context.Context, params Values, in Value, out WritableValue) error {
	f := "%c"
	if len(params) > 0 {
		f = params[0].WithCtx(ctx).String()
	}
	if f == "" {
		f = "%c"
	}

	// we need in
	in_obj, err := in.WithCtx(ctx).Raw()
	if err != nil {
		return out.WriteValue(ctx, "N/A")
	}

	t, err := parseDate(ctx, in_obj)
	if err != nil {
		return out.WriteValue(ctx, "N/A")
	}

	lng := ctx.Value("_language").(language.Tag)
	var loc *time.Location
	ctx.Value(&loc)
	if loc != nil {
		t = t.In(loc)
	}

	return out.WriteValue(ctx, strftime.Format(lng, f, t))
}

func padLeft(str, pad string, length int) string {
	for {
		str = pad + str
		if len(str) >= length {
			return str[len(str)-length:]
		}
	}
}

func fltDuration(ctx context.Context, params Values, in Value, out WritableValue) error {
	// input is assumed to be seconds, and will be formatted days:hh:mm:ss
	v, _ := fetchNumberInt(ctx, in)

	var res string
	if v < 0 {
		res = "-"
		v = -v
	}

	switch {
	case v >= 86400: // days
		d := v / 86400
		res += strconv.FormatInt(d, 10) + ":"
		v -= d * 86400
		fallthrough
	case v >= 3600: // hours
		h := v / 3600
		res += padLeft(strconv.FormatInt(h, 10), "0", 2) + ":"
		v -= h * 3600
		fallthrough
	default: // minutes
		m := v / 60
		res += padLeft(strconv.FormatInt(m, 10), "0", 2) + ":"
		v -= m * 60
		// seconds
		res += padLeft(strconv.FormatInt(v, 10), "0", 2)
	}

	return out.WriteValue(ctx, res)
}

func parseDate(ctx context.Context, in interface{}) (time.Time, error) {
	switch r := in.(type) {
	case string:
		r = strings.TrimSpace(r)
		if r == "" {
			return time.Time{}, errors.New("unable to parse empty string as time")
		}
		if r == "now" {
			return time.Now(), nil
		}
		if r[0] == '@' {
			// followed by timestamp in unix time, possibly with decimals
			r = r[1:]
			pos := strings.IndexByte(r, '.')
			if pos == -1 {
				// pure integer
				v, err := strconv.ParseInt(r, 64, 10)
				if err != nil {
					return time.Time{}, err
				}
				return time.Unix(v, 0), nil
			}
			unix_t, err := strconv.ParseInt(r[:pos], 64, 10)
			if err != nil {
				return time.Time{}, err
			}
			r = r[pos+1:]
			if r == "" {
				// no decimal part?
				return time.Unix(unix_t, 0), nil
			}
			// add zeroes
			if len(r) < 9 {
				r = r + strings.Repeat("0", 9-len(r))
			} else if len(r) > 9 {
				r = r[:9]
			}
			unix_us, err := strconv.ParseInt(r, 64, 10)
			if err != nil {
				return time.Time{}, err
			}
			return time.Unix(unix_t, unix_us), nil

		}
		return time.Time{}, fmt.Errorf("failed to parse time: %s", r)
	case json.RawMessage:
		var x interface{}
		err := json.Unmarshal(r, &x)
		if err != nil {
			return time.Time{}, err
		}
		return parseDate(ctx, x)
	case map[string]interface{}:
		if us, ok := r["full"].(string); ok {
			// microtime!
			l := len(us)
			if l < 7 {
				return time.Time{}, fmt.Errorf("failed to parse microtime")
			}
			us_p := us[l-6:]
			us = us[:l-6]
			us_n, err := strconv.ParseInt(us, 10, 64)
			if err != nil {
				return time.Time{}, err
			}
			us_p_n, err := strconv.ParseInt(us_p, 10, 64)
			if err != nil {
				return time.Time{}, err
			}
			return time.Unix(us_n, us_p_n*1000), nil
		}
		return time.Time{}, fmt.Errorf("failed to parse array as time")
	case []byte:
		return parseDate(ctx, string(r))
	case *bytes.Buffer:
		return parseDate(ctx, string(r.Bytes()))
	case Value:
		v, err := r.ReadValue(ctx)
		if err != nil {
			return time.Time{}, err
		}
		return parseDate(ctx, v)
	default:
		return time.Time{}, fmt.Errorf("failed to parse time of type %T", in)
	}

}
