package tpl

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"math"
	"net/url"
	"strings"

	"github.com/frustra/bbcode"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
)

type TplFiltCallback func(ctx context.Context, params Values, in Value, out WritableValue) error

var BBCodeCompiler = bbcode.NewCompiler(true, true)

func init() {
	// select
	// selecton
	// price
	RegisterFilter("price", fltPrice)
	// unicode
	RegisterFilter("unicode", fltUnicode)
	RegisterFilter("size", fltSize)
	RegisterFilter("replace", fltReplace)
	RegisterFilter("isnull", fltIsNull)
	RegisterFilter("json", fltJson)
	RegisterFilter("jsondump", fltJsonDump)
	RegisterFilter("jsonparse", fltJsonParse)
	RegisterFilter("nl2br", fltNl2br)
	RegisterFilter("entities", fltEntities)
	RegisterFilter("stripcrlf", fltStripcrlf)
	// ascii
	RegisterFilter("nbsp", fltNbsp)
	RegisterFilter("uppercase", fltUpper)
	RegisterFilter("lowercase", fltLower)
	// ucfirst
	RegisterFilter("urlencode", fltUrlencode)
	RegisterFilter("rawurlencode", fltRawurlencode)
	// urldecode
	RegisterFilter("striptags", fltStriptags)
	RegisterFilter("substr", fltSubstr)
	// striphtmlheaders
	RegisterFilter("null", fltNull)
	RegisterFilter("export", fltDump)
	RegisterFilter("dump", fltDump)
	RegisterFilter("length", fltLength)
	RegisterFilter("count", fltLength)
	RegisterFilter("truncate", fltTruncate)
	RegisterFilter("trim", fltTrim)
	// pad
	RegisterFilter("type", fltType)
	// last
	RegisterFilter("toint", fltToInt)
	RegisterFilter("tostring", fltToString)
	RegisterFilter("round", fltRound)
	RegisterFilter("b64enc", fltB64Enc)
	RegisterFilter("b64dec", fltB64Dec)
	RegisterFilter("explode", fltExplode)
	RegisterFilter("implode", fltImplode)
	RegisterFilter("reverse", fltReverse)
	RegisterFilter("keyval", fltKeyVal)
	// arrayvalues
	RegisterFilter("arrayslice", fltArraySlice)
	RegisterFilter("arrayfilter", fltArrayFilter)
	// arrayvalue
	RegisterFilter("columns", fltColumns)
	RegisterFilter("lines", fltLines)

	RegisterFilter("bbcode", fltBbCode)
	RegisterFilter("stripbbcode", fltStripBbCode)
	RegisterFilter("markdown", fltMarkdown)
	// barcode
	// qrcode
	// isarray

	// For bbcode
	BBCodeCompiler.SetTag("sup", func(node *bbcode.BBCodeNode) (*bbcode.HTMLTag, bool) {
		// create sup element
		out := bbcode.NewHTMLTag("")
		out.Name = "sup"

		// Returning true here means continue to parse child nodes.
		return out, true
	})

	BBCodeCompiler.SetTag("sub", func(node *bbcode.BBCodeNode) (*bbcode.HTMLTag, bool) {
		// create sup element
		out := bbcode.NewHTMLTag("")
		out.Name = "sub"

		// Returning true here means continue to parse child nodes.
		return out, true
	})

	BBCodeCompiler.SetTag("mail", func(node *bbcode.BBCodeNode) (*bbcode.HTMLTag, bool) {
		// create a href
		out := bbcode.NewHTMLTag("")
		out.Name = "a"
		value := node.GetOpeningTag().Value
		if value == "" {
			text := bbcode.CompileText(node)
			if len(text) > 0 {
				out.Attrs["href"] = bbcode.ValidURL("mailto:" + text)
			}
		} else {
			out.Attrs["href"] = bbcode.ValidURL("mailto:" + value)
		}
		return out, true
	})
}

func fltPrice(ctx context.Context, params Values, in Value, out WritableValue) error {
	// grab in
	v, err := in.WithCtx(ctx).Raw()
	if err != nil {
		return err
	}

	switch s := v.(type) {
	case map[string]interface{}:
		if p, ok := s["display"]; ok {
			return out.WriteValue(ctx, p)
		}
		return out.WriteValue(ctx, "N/A")
	default:
		return out.WriteValue(ctx, "N/A")
	}
}

func fltUnicode(ctx context.Context, params Values, in Value, out WritableValue) error {
	// split input into unicode chars as objects
	var r Values

	s, err := in.WithCtx(ctx).StringErr()
	if err != nil {
		return err
	}

	// s is a "string" so c is a rune
	for _, c := range s {
		r = append(r, unicodeInfoObj(c))
	}

	return out.WriteValue(ctx, r)
}

func fltSize(ctx context.Context, params Values, in Value, out WritableValue) error {
	v, ok := in.WithCtx(ctx).ToInt()
	if !ok {
		return errors.New("size(): input needs to be numeric")
	}
	if v < 0 {
		out.WriteValue(ctx, "-")
		v = 0 - v
	}
	return out.WriteValue(ctx, FormatSize(uint64(v)))
}

func fltReplace(ctx context.Context, params Values, in Value, out WritableValue) error {
	if len(params) < 2 {
		return errors.New("replace() filters requires at least 2 arguments")
	}

	replace_from := params[0].WithCtx(ctx).String()
	replace_to := params[1].WithCtx(ctx).String()

	return out.WriteValue(ctx, strings.Replace(in.WithCtx(ctx).String(), replace_from, replace_to, -1))
}

func fltIsNull(ctx context.Context, params Values, in Value, out WritableValue) error {
	val, err := in.ReadValue(ctx)
	if err != nil {
		return err
	}
	return out.WriteValue(ctx, val == nil)
}

func fltJson(ctx context.Context, params Values, in Value, out WritableValue) error {
	inB, err := in.WithCtx(ctx).Raw()
	if err != nil {
		return err
	}

	switch b := inB.(type) {
	case *bytes.Buffer:
		inB = b.String()
	case []byte:
		inB = string(b)
	}

	r, err := json.Marshal(inB)
	if err != nil {
		return err
	}

	_, err = out.Write(r)
	return err
}

func fltJsonDump(ctx context.Context, params Values, in Value, out WritableValue) error {
	inB, err := in.WithCtx(ctx).Raw()
	if err != nil {
		return err
	}

	if b, ok := inB.(*bytes.Buffer); ok {
		inB = b.String()
	}

	r, err := json.Marshal(inB)
	if err != nil {
		return err
	}

	b := &bytes.Buffer{}
	json.Indent(b, r, "", "\t")
	out.WriteValue(ctx, b)
	return nil
}

func fltJsonParse(ctx context.Context, params Values, in Value, out WritableValue) error {
	var jsonVal json.RawMessage
	err := json.Unmarshal(in.WithCtx(ctx).Bytes(), &jsonVal)
	if err != nil {
		return err
	}
	out.WriteValue(ctx, jsonVal)
	return nil
}

func fltEntities(ctx context.Context, params Values, in Value, out WritableValue) error {
	return out.WriteValue(ctx, html.EscapeString(in.WithCtx(ctx).String()))
}

func fltStripcrlf(ctx context.Context, params Values, in Value, out WritableValue) error {
	return out.WriteValue(ctx, strings.Replace(strings.Replace(in.WithCtx(ctx).String(), "\r", "", -1), "\n", "", -1))
}

func fltUrlencode(ctx context.Context, params Values, in Value, out WritableValue) error {
	return out.WriteValue(ctx, QueryEscapeAny(ctx, in))
}

func fltRawurlencode(ctx context.Context, params Values, in Value, out WritableValue) error {
	return out.WriteValue(ctx, url.PathEscape(in.WithCtx(ctx).String()))
}

func fltToInt(ctx context.Context, params Values, in Value, out WritableValue) error {
	i, ok := in.WithCtx(ctx).ToInt()
	if !ok {
		return errors.New("toint(): failed to parse int")
	}
	return out.WriteValue(ctx, i)
}

func fltToString(ctx context.Context, params Values, in Value, out WritableValue) error {
	_, err := out.Write(in.WithCtx(ctx).Bytes())
	return err
}

func fltRound(ctx context.Context, params Values, in Value, out WritableValue) error {
	var precision int64 = 2
	var ok bool

	if len(params) >= 1 {
		precision, ok = params[0].WithCtx(ctx).ToInt()
		if !ok {
			return errors.New("round() filter first argument should be an integer")
		}
	}

	v, ok := in.WithCtx(ctx).ToFloat()
	if !ok {
		return errors.New("round() filter can only be applied on numbers")
	}

	shift := math.Pow(10, float64(precision))
	return out.WriteValue(ctx, math.Round(v*shift)/shift)
}

func fltB64Enc(ctx context.Context, params Values, in Value, out WritableValue) error {
	dat := in.WithCtx(ctx).Bytes()
	res := make([]byte, base64.StdEncoding.EncodedLen(len(dat)))
	base64.StdEncoding.Encode(res, dat)
	return out.WriteValue(ctx, res)
}

func fltB64Dec(ctx context.Context, params Values, in Value, out WritableValue) error {
	dat := in.WithCtx(ctx).Bytes()
	res := make([]byte, base64.StdEncoding.DecodedLen(len(dat)))
	n, err := base64.StdEncoding.Decode(res, dat)
	if err != nil {
		return err
	}
	return out.WriteValue(ctx, res[:n])
}

func fltExplode(ctx context.Context, params Values, in Value, out WritableValue) error {
	if len(params) < 1 {
		return errors.New("explode() filter requires one argument")
	}
	return out.WriteValue(ctx, strings.Split(in.WithCtx(ctx).String(), params[0].WithCtx(ctx).String()))
}

func fltImplode(ctx context.Context, params Values, in Value, out WritableValue) error {
	if len(params) < 1 {
		return errors.New("implode() filter requires one argument")
	}
	val, err := in.ReadValue(ctx)
	if err != nil {
		return err
	}
	switch val.(type) {
	case []string:
		return out.WriteValue(ctx, strings.Join(val.([]string), params[0].WithCtx(ctx).String()))
	case [][]byte:
		return out.WriteValue(ctx, bytes.Join(val.([][]byte), params[0].WithCtx(ctx).Bytes()))
	default:
		return out.WriteValue(ctx, val)
	}
}

func fltStriptags(ctx context.Context, params Values, in Value, out WritableValue) error {
	str, err := in.WithCtx(ctx).StringErr()
	if err != nil {
		return err
	}

	for {
		p := strings.IndexByte(str, '<')
		if p == -1 {
			break
		}
		p2 := strings.IndexByte(str[p:], '>')
		if p2 == -1 {
			break
		}

		str = str[:p] + str[p+p2+1:]
	}

	return out.WriteValue(ctx, str)
}

func fltSubstr(ctx context.Context, params Values, in Value, out WritableValue) error {
	if len(params) < 2 {
		return errors.New("substr() filter requires 2 arguments")
	}

	start64, ok := params[0].WithCtx(ctx).ToInt()
	if !ok {
		return errors.New("substr() parameter should be an integer")
	}
	end64, ok := params[1].WithCtx(ctx).ToInt()
	if !ok {
		return errors.New("substr() parameter should be an integer")
	}
	start, end := int(start64), int(end64)

	// check values
	str := in.WithCtx(ctx).String()

	if start < 0 {
		start = len(str) + start
	}
	if start >= len(str) {
		return nil
	}
	if end < 0 {
		end = (len(str) - start) + end
	}
	if start+end > len(str) {
		end = len(str) - start
	}

	return out.WriteValue(ctx, str[start:start+end])
}

func fltTruncate(ctx context.Context, params Values, in Value, out WritableValue) error {
	vIn, err := in.WithCtx(ctx).StringErr()
	if err != nil {
		return err
	}

	l := 100
	if len(params) >= 1 {
		l64, ok := params[0].WithCtx(ctx).ToInt()
		if !ok {
			return errors.New("truncate() filter first parameter must be an int")
		}
		l = int(l64)
	}

	if len(vIn) < l {
		return out.WriteValue(ctx, vIn)
	}

	pad := "…"
	if len(params) >= 2 {
		pad = params[1].WithCtx(ctx).String()
	}
	wordCut := false
	if len(params) >= 3 {
		wordCut = params[2].WithCtx(ctx).ToBool()
	}

	if wordCut {
		return out.WriteValue(ctx, vIn[:l]+pad)
	}

	pos := strings.LastIndexByte(vIn[:l], ' ')
	if pos == -1 {
		return out.WriteValue(ctx, vIn[:l]+pad)
	}

	return out.WriteValue(ctx, vIn[:pos]+pad)
}

func fltTrim(ctx context.Context, params Values, in Value, out WritableValue) error {
	return out.WriteValue(ctx, strings.TrimSpace(in.WithCtx(ctx).String()))
}

func fltType(ctx context.Context, params Values, in Value, out WritableValue) error {
	a, err := in.ReadValue(ctx)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "%T", a)
	return err
}

func fltNl2br(ctx context.Context, params Values, in Value, out WritableValue) error {
	return out.WriteValue(ctx, strings.Replace(in.WithCtx(ctx).String(), "\n", "<br/>", -1))
}

func fltNbsp(ctx context.Context, params Values, in Value, out WritableValue) error {
	return out.WriteValue(ctx, strings.Replace(in.WithCtx(ctx).String(), " ", "\xc2\xa0", -1))
}

func fltUpper(ctx context.Context, params Values, in Value, out WritableValue) error {
	return out.WriteValue(ctx, strings.ToUpper(in.WithCtx(ctx).String()))
}

func fltLower(ctx context.Context, params Values, in Value, out WritableValue) error {
	return out.WriteValue(ctx, strings.ToLower(in.WithCtx(ctx).String()))
}

func fltNull(ctx context.Context, params Values, in Value, out WritableValue) error {
	return nil
}

func fltDump(ctx context.Context, params Values, in Value, out WritableValue) error {
	debugDump(out, in, 0)
	return nil
}

func fltLength(ctx context.Context, params Values, in Value, out WritableValue) error {
	inObj, err := in.ReadValue(ctx)
	if err != nil {
		return err
	}

	switch i := inObj.(type) {
	case map[string]WritableValue:
		return out.WriteValue(ctx, len(i))
	case string:
		return out.WriteValue(ctx, len(i))
	case *bytes.Buffer:
		return out.WriteValue(ctx, i.Len())
	case []interface{}:
		return out.WriteValue(ctx, len(i))
	case Values:
		return out.WriteValue(ctx, len(i))
	case []string:
		return out.WriteValue(ctx, len(i))
	case map[string]interface{}:
		return out.WriteValue(ctx, len(i))
	case map[string]Value:
		return out.WriteValue(ctx, len(i))
	case map[string]json.RawMessage:
		return out.WriteValue(ctx, len(i))
	default:
		return out.WriteValue(ctx, len(in.WithCtx(ctx).String()))
	}
}

func fltReverse(ctx context.Context, params Values, in Value, out WritableValue) error {
	inObj, err := in.WithCtx(ctx).Raw()
	if err != nil {
		return err
	}

	switch i := inObj.(type) {
	case string:
		r := []rune(i)

		for k, j := 0, len(r)-1; k < j; k, j = k+1, j-1 {
			r[k], r[j] = r[j], r[k]
		}

		return out.WriteValue(ctx, string(r))
	case []interface{}:
		r := make([]interface{}, len(i))

		l := len(i) - 1

		for j := l; j >= 0; j-- {
			r[l-j] = i[j]
		}

		return out.WriteValue(ctx, r)
	case Values:
		r := make(Values, len(i))

		l := len(i) - 1

		for j := l; j >= 0; j-- {
			r[l-j] = i[j]
		}

		return out.WriteValue(ctx, r)
	case []string:
		r := make([]string, len(i))

		l := len(i) - 1

		for j := l; j >= 0; j-- {
			r[l-j] = i[j]
		}

		return out.WriteValue(ctx, r)
	default:
		return fmt.Errorf("reverse() filter argument should be an array or a string, type %T not supported", inObj)
	}
}

func fltColumns(ctx context.Context, params Values, in Value, out WritableValue) error {
	l := 3 // columns count
	if len(params) >= 1 {
		v, ok := params[0].WithCtx(ctx).ToInt()
		if ok && v > 0 {
			l = int(v)
		}
	}

	var res Values
	sub := make([]Values, l)

	_, err := foreachAny(ctx, in, func(k, v interface{}, idx, max int64) error {
		sub[int(idx-1)%l] = append(sub[int(idx-1)%l], NewValue(v))

		return nil
	})

	if err != nil {
		return err
	}

	for _, x := range sub {
		res = append(res, NewValue(x))
	}

	return out.WriteValue(ctx, NewValue(res))
}

func fltLines(ctx context.Context, params Values, in Value, out WritableValue) error {
	l := 3 // line length
	if len(params) >= 1 {
		v, ok := params[0].WithCtx(ctx).ToInt()
		if ok && v > 0 {
			l = int(v)
		}
	}

	var res Values
	var sub Values

	_, err := foreachAny(ctx, in, func(k, v interface{}, idx, max int64) error {
		sub = append(sub, NewValue(v))
		if len(sub) >= l {
			res = append(res, NewValue(sub))
			sub = nil
		}
		return nil
	})
	if err != nil {
		return err
	}
	if sub != nil {
		res = append(res, NewValue(sub))
	}

	return out.WriteValue(ctx, NewValue(res))
}

func fltKeyVal(ctx context.Context, params Values, in Value, out WritableValue) error {
	// convert in value (some kind of array) to a keyval using foreach
	if len(params) < 1 {
		return errors.New("keyval() filter requires at least one parameter")
	}

	refK := params[0].WithCtx(ctx).String()
	refV := refK
	if len(params) >= 2 {
		refV = params[1].WithCtx(ctx).String()
	}

	final := make(map[string]Value)

	// foreach value in input
	_, err := foreachAny(ctx, in, func(k, v interface{}, idx, max int64) error {
		// resolve localK and localV
		localK, err := ResolveValueIndex(ctx, v, refK)
		if err != nil {
			return err
		}
		// localK is a key (so a string)
		localKstr, err := NewValue(localK).WithCtx(ctx).StringErr()
		if err != nil {
			return err
		}
		localV, err := ResolveValueIndex(ctx, v, refV)
		if err != nil {
			return err
		}

		// store in final
		final[localKstr] = NewValue(localV)
		return nil
	})

	if err != nil {
		return err
	}

	return out.WriteValue(ctx, final)
}

func fltArraySlice(ctx context.Context, params Values, in Value, out WritableValue) error {
	inObj, err := in.WithCtx(ctx).Raw()
	if err != nil {
		return err
	}

	if len(params) < 1 {
		return errors.New("arraySlice() filter requires at least one parameter")
	}

	from, ok := params[0].WithCtx(ctx).ToInt()
	if !ok {
		return errors.New("arraySlice() filter first parameter must be an int")
	}

	to := from

	if len(params) > 1 {
		to, ok = params[1].WithCtx(ctx).ToInt()

		if !ok {
			return errors.New("arraySlice() filter second parameter must be an int")
		}
	} else {
		from = 0
	}

	if from < 0 || to < 0 {
		// illegal
		return fmt.Errorf("arraySlice(%d,%d) filter argument cannot be negative", from, to)
	}

	switch i := inObj.(type) {
	case string:
		r := []rune(i)
		if from >= int64(len(r)) {
			return nil
		}
		if from+to > int64(len(r)) {
			return out.WriteValue(ctx, string(r[from:]))
		}
		return out.WriteValue(ctx, string(r[from:from+to]))
	case []interface{}:
		if from >= int64(len(i)) {
			return nil
		}
		if from+to > int64(len(i)) {
			return out.WriteValue(ctx, i[from:])
		}
		return out.WriteValue(ctx, i[from:from+to])
	case Values:
		if from >= int64(len(i)) {
			return nil
		}
		if from+to > int64(len(i)) {
			return out.WriteValue(ctx, i[from:])
		}
		return out.WriteValue(ctx, i[from:from+to])
	case []string:
		if from >= int64(len(i)) {
			return nil
		}
		if from+to > int64(len(i)) {
			return out.WriteValue(ctx, i[from:])
		}
		return out.WriteValue(ctx, i[from:from+to])
	default:
		return fmt.Errorf("arraySlice() filter argument should be an array or a string, type %T not supported", inObj)
	}
}

func fltArrayFilter(ctx context.Context, params Values, in Value, out WritableValue) error {
	if len(params) < 2 {
		return errors.New("arrayfilter() requires at least 2 parameters")
	}

	path, err := params[0].WithCtx(ctx).StringErr()
	if err != nil {
		return err
	}
	pathA := strings.Split(path, "/")

	checkV, err := params[1].WithCtx(ctx).Raw()
	if err != nil {
		return err
	}

	if params[1].WithCtx(ctx).IsString() {
		switch params[1].WithCtx(ctx).String() {
		case "true":
			checkV = true
		case "false":
			checkV = false
		}
	}

	var res []interface{}

	_, err = foreachAny(ctx, in, func(k, v interface{}, idx, max int64) error {
		sv := v
		var err error
		for _, subP := range pathA {
			sv, err = ResolveValueIndex(ctx, sv, subP)
			if err != nil {
				return err
			}
		}

		// make sure v is a root object and not a Value
		sv, err = NewValue(sv).WithCtx(ctx).Raw()
		if err != nil {
			return err
		}
		r, err := CompareValues(ctx, checkV, sv)
		if err != nil {
			return err
		}

		if r {
			res = append(res, v)
		}
		return nil
	})
	if err != nil {
		return err
	}

	return out.WriteValue(ctx, res)
}

func fltBbCode(ctx context.Context, params Values, in Value, out WritableValue) error {
	return out.WriteValue(ctx, BBCodeCompiler.Compile(in.WithCtx(ctx).String()))
}

func fltStripBbCode(ctx context.Context, params Values, in Value, out WritableValue) error {
	str, err := in.WithCtx(ctx).StringErr()
	if err != nil {
		return err
	}

	for {
		p := strings.IndexByte(str, '[')
		if p == -1 {
			break
		}
		p2 := strings.IndexByte(str[p:], ']')
		if p2 == -1 {
			break
		}

		str = str[:p] + str[p+p2+1:]
	}

	return out.WriteValue(ctx, str)
}

func fltMarkdown(ctx context.Context, params Values, in Value, out WritableValue) error {
	res := blackfriday.Run([]byte(in.WithCtx(ctx).String()))
	final := bluemonday.UGCPolicy().SanitizeBytes(res)
	return out.WriteValue(ctx, final)
}
