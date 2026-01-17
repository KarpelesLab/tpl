// Package tpl provides a template engine with string interpolation and control structures.
package tpl

import (
	"bytes"
	"context"
	"io"
	"strings"
	"sync"
)

// Page represents a template engine instance containing compiled templates.
type Page struct {
	// Version will be populated on compile
	Version int
	// Raw contains the template source data
	Raw RawData
	// compiled holds the processed templates ready for execution
	compiled map[string]internalArray

	// MaxProcess controls parallel execution of templates
	// 0 means unlimited concurrency, 1 means serial execution
	MaxProcess int
}

// New creates a new template engine instance.
// The returned Page is ready to have templates added to it
// and then compiled.
func New() *Page {
	e := &Page{}
	e.Raw.init()
	return e
}

// GetMime returns the MIME type for the page.
// The default value is "text/html" if not explicitly set.
// If a charset is specified, it will be appended to the MIME type.
func (e *Page) GetMime() string {
	res, ok := e.Raw.PageProperties["Content-Type"]
	if !ok {
		res = "text/html"
	}
	charset, ok := e.Raw.PageProperties["Charset"]
	if ok {
		res = res + "; charset=" + charset
	}
	return res
}

// GetProperty returns the value of a page property.
// Returns an empty string if the property doesn't exist.
func (e *Page) GetProperty(p string) string {
	return e.Raw.PageProperties[p]
}

func (n *internalNode) internalParseLink(ctx context.Context, out *interfaceValue) error {
	buf := &interfaceValue{}
	err := n.sub[0].run(ctx, buf)
	if err != nil {
		return err
	}
	key := buf.WithCtx(ctx).String()

	if len(key) == 0 {
		return nil // nothing to output
	}

	keyA := strings.Split(key, "/")
	key = keyA[0]

	if len(key) == 0 {
		return n.subError(err, "invalid attempt to read empty variable, invalid keyword?")
	}

	var val any

	if key[0] == '_' || key[0] == '$' {
		// check for context
		val = ctx.Value(key)
	}
	if val == nil {
		key = strings.ToLower(key)
		if tpl, ok := n.e.compiled[key]; ok {
			val = tpl.WithCtx(ctx) // keep context in value so when we resolve it we have vars
		} else {
			// calling a non-existent link is not an error
			LogDebug(ctx, "Accessing non-existing key returns null", "key", key)
			return nil
		}
	}

	// we always have val at this point
	if len(keyA) > 1 {
		for _, s := range keyA[1:] {
			val, err = ResolveValueIndex(ctx, val, s)
			if err != nil {
				return n.subError(err, "failed to access array: %s", err)
			}
		}
	}

	return out.WriteValue(ctx, val)
}

// run executes the template array in the given context, writing output to the provided interfaceValue.
// For arrays with multiple nodes, it can run them concurrently if MaxProcess is not 1.
func (tpl internalArray) run(ctx context.Context, out *interfaceValue) (err error) {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return err
	}

	// Handle empty array or single node cases
	if len(tpl) == 0 {
		return nil
	}
	if len(tpl) == 1 {
		return tpl[0].run(ctx, out)
	}

	// Run serially if MaxProcess is 1
	if tpl[0].e.MaxProcess == 1 {
		for _, n := range tpl {
			// Check for context cancellation between nodes
			if err := ctx.Err(); err != nil {
				return err
			}

			if err = n.run(ctx, out); err != nil {
				return err
			}
		}
		return nil
	}

	// Run concurrently
	wg := &sync.WaitGroup{}
	wg.Add(len(tpl))
	tOut := make([]*interfaceValue, len(tpl))
	errorC := make(chan error, len(tpl))

	// Create a separate context that we can cancel if needed
	execCtx, cancel := context.WithCancel(ctx)
	defer cancel() // Ensure all goroutines are canceled when we exit

	for i, n := range tpl {
		tOut[i] = &interfaceValue{}

		go func(o *interfaceValue, n *internalNode) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					select {
					case errorC <- n.error("panic: %s", r):
						LogError(ctx, n.error("panic: %s", r), "Panic in template execution")
					default:
						// Channel might be closed if we're already handling errors
					}
				}
			}()

			// Run with the cancellable context
			e := n.run(execCtx, o)
			if e != nil {
				select {
				case errorC <- e:
					// Signal other goroutines to stop
					cancel()
				default:
					// Channel might be closed if we're already handling errors
				}
			}
		}(tOut[i], n)
	}
	wg.Wait()

	// Check if original context was canceled during execution
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Read all errors from channel
	if len(errorC) > 0 {
		err = <-errorC
		close(errorC)
		return err
	}

	close(errorC)

	// Write output from all nodes
	for _, lOut := range tOut {
		if err = out.WriteValue(ctx, lOut); err != nil {
			LogError(ctx, err, "Error writing template value")
			return err
		}
	}

	return nil
}

// ReadValue executes the template array and returns its value.
func (n internalArray) ReadValue(ctx context.Context) (any, error) {
	buf := &interfaceValue{}
	err := n.run(ctx, buf)
	if err != nil {
		return nil, err
	}
	return buf.ReadValue(ctx)
}

// WithCtx returns a ValueCtx that wraps this internalArray with the given context.
func (n internalArray) WithCtx(ctx context.Context) *ValueCtx {
	return &ValueCtx{n, ctx}
}

// ReadValue executes the node and returns its value.
func (n *internalNode) ReadValue(ctx context.Context) (any, error) {
	buf := &interfaceValue{}
	err := n.run(ctx, buf)
	if err != nil {
		return nil, err
	}
	return buf.WithCtx(ctx).Raw()
}

// WithCtx returns a ValueCtx that wraps this node with the given context.
func (n *internalNode) WithCtx(ctx context.Context) *ValueCtx {
	return &ValueCtx{n, ctx}
}

func (a internalArray) isStatic() bool {
	for _, n := range a {
		if !n.isStatic() {
			return false
		}
	}
	return true
}

func (n *internalNode) isStatic() bool {
	// check if this node is static
	if len(n.filters) > 0 {
		return false
	}

	switch n.typ {
	case internalText:
		return true
	case internalLink:
		return false
	case internalQuote:
		return n.sub[0].isStatic()
	case internalList:
		for _, x := range n.sub {
			if !x.isStatic() {
				return false
			}
		}
		return true
	case internalSub:
		return n.sub[0].isStatic()
	case internalValue:
		return false // n.value → can it be dynamic? yes.
	default:
		return false
	}
}

// ToValues converts the array's value to Values type.
// If the value is already a Values type, it's returned as is.
// Otherwise, it wraps the value in a single-element Values slice.
func (a internalArray) ToValues(ctx context.Context) (Values, error) {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Get raw value
	params, err := a.WithCtx(ctx).Raw()
	if err != nil {
		return nil, err
	}

	// Convert to Values
	vparams, ok := params.(Values)
	if !ok {
		vparams = Values{AsOutValue(ctx, params)}
	}
	return vparams, nil
}

func (n *internalNode) run(ctx context.Context, out *interfaceValue) error {
	target := out
	if len(n.filters) > 0 {
		// variable setting filters
		v := make(map[string]interface{})
		ctx2 := ValuesCtxAlways(ctx, v)
		var err error
		for _, f := range n.filters {
			if f.typ == internalFilter {
				if target == out {
					target = new(interfaceValue)
				}
				continue
			}
			if f.typ != internalVar {
				continue
			}
			v[f.str], err = f.sub[0].ReadValue(ctx2)
			if err != nil {
				return err
			}
		}
		if len(v) > 0 {
			ctx = ctx2
		}
	}

	switch n.typ {
	case internalText:
		target.Write([]byte(n.str))
	case internalLink:
		err := n.internalParseLink(ctx, target)
		if err != nil {
			return err
		}
	case internalQuote:
		if err := n.sub[0].run(ctx, target); err != nil {
			return err
		}
	case internalList:
		res := make(Values, len(n.sub))
		// we store Ctx() value in order to guarantee context will not be modified for the parsing
		for i, x := range n.sub {
			res[i] = x.WithCtx(ctx)
		}
		out.WriteValue(ctx, res)
	case internalOperator:
		switch n.str {
		case "!", "~":
			// special operator (only one argument)
			if res, err := mathSingleValueOperator(ctx, n.str, n.sub[0]); err != nil {
				return n.subError(err, "operator failed: %s", err)
			} else {
				err = out.WriteValue(ctx, res)
				if err != nil {
					return err
				}
			}
		default:
			if res, err := mathValueOperator(ctx, n.str, n.sub[0], n.sub[1]); err != nil {
				return n.subError(err, "operator failed: %s", err)
			} else {
				err = out.WriteValue(ctx, res)
				if err != nil {
					return err
				}
			}
		}
	case internalForeach:
		// handle value in sub[0] and foreach on it
		varName := n.str

		var prevVal interface{}
		cnt, err := foreachAny(ctx, n.sub[0], func(k, v interface{}, idx, max int64) error {
			nv := map[string]interface{}{
				varName + "_max": NewValue(max),
				varName:          NewValue(v),
				varName + "_key": NewValue(k),
				varName + "_idx": NewValue(idx),
				varName + "_prv": NewValue(prevVal),
			}
			if err := n.sub[1].run(ValuesCtx(ctx, nv), target); err != nil {
				return err
			}

			prevVal = v
			return nil
		})
		if err != nil {
			return n.subError(err, "error in foreach: %s", err)
		}

		if cnt == 0 && len(n.sub) > 2 {
			// else
			if err := n.sub[2].run(ctx, target); err != nil {
				return err
			}
		}
	case internalIf:
		// handle condition in sub[0]
		cond := &interfaceValue{}
		if err := n.sub[0].run(ctx, cond); err != nil {
			return err
		}
		if cond.AsBool(ctx) {
			if err := n.sub[1].run(ctx, target); err != nil {
				return err
			}
		} else {
			if len(n.sub) > 2 {
				if err := n.sub[2].run(ctx, target); err != nil {
					return err
				}
			}
		}
	case internalFunc:
		// process call to function in str
		if f, ok := ctx.Value("@" + n.str).(TplFuncCallback); ok {
			// custom in-context function
			params, err := n.sub[0].ToValues(ctx)
			if err != nil {
				return err
			}
			if err := f(ctx, params, target); err != nil {
				return n.subError(err, "function call failed: %s", err)
			}
		} else if f, ok := tplFunctions[n.str]; ok {
			// only call if ok
			params, err := n.sub[0].ToValues(ctx)
			if err != nil {
				return n.subError(err, "failed to prepare method arguments: %s", err)
			}
			if err := f.Method(ctx, params, target); err != nil {
				return n.subError(err, "function call failed: %s", err)
			}
		} else {
			return n.error("tpl: call to undefined function %s", n.str)
		}
	case internalSet:
		if err := n.sub[0].run(ctx, target); err != nil {
			return err
		}
	case internalSub:
		if err := n.sub[0].run(ctx, target); err != nil {
			return err
		}
	case internalTry:
		t := new(interfaceValue)
		if err := n.sub[0].run(ctx, t); err != nil {
			// catch the error
			if len(n.sub) > 1 {
				ctx2 := ctx
				if n.str != "" {
					ctx2 = context.WithValue(ctx2, n.str, err) //lint:ignore SA1029 template variables use string keys by design
				} else {
					ctx2 = context.WithValue(ctx2, "_exception", err) //lint:ignore SA1029 template variables use string keys by design
				}
				err2 := n.sub[1].run(ctx2, target)
				if err2 != nil {
					return err2
				}
			}
		} else {
			target.WriteValue(ctx, t)
		}
	case internalValue:
		target.WriteValue(ctx, n.value)
	case internalIndex:
		// Bracket index access: sub[0] is the base, sub[1] is the index expression
		// str may contain additional path like "/a" to resolve after indexing
		// First evaluate the base
		baseVal := &interfaceValue{}
		if err := n.sub[0].run(ctx, baseVal); err != nil {
			return err
		}
		base, err := baseVal.ReadValue(ctx)
		if err != nil {
			return n.subError(err, "failed to read base value: %s", err)
		}

		// Then evaluate the index expression
		indexVal := &interfaceValue{}
		if err := n.sub[1].run(ctx, indexVal); err != nil {
			return err
		}
		indexKey := indexVal.WithCtx(ctx).String()

		// Resolve the indexed value
		result, err := ResolveValueIndex(ctx, base, indexKey)
		if err != nil {
			return n.subError(err, "failed to access index [%s]: %s", indexKey, err)
		}

		// If there's additional path in str (like "/a"), resolve it
		if n.str != "" {
			pathParts := strings.Split(n.str, "/")
			for _, part := range pathParts {
				if part == "" {
					continue
				}
				result, err = ResolveValueIndex(ctx, result, part)
				if err != nil {
					return n.subError(err, "failed to access path [%s]: %s", part, err)
				}
			}
		}

		target.WriteValue(ctx, result)
	default:
		return n.error("unable to process node of type %s", n.typ.String())
		//n.Dump(out, 1)
	}
	//if n.filters != nil && len(*n.filters) > 0 && target != out {
	if target != out {
		// apply filters
		for _, f := range n.filters {
			if f.typ != internalFilter {
				continue
			}
			if flt, ok := tplfilters[f.str]; ok {
				params, err := f.sub[0].WithCtx(ctx).Raw()
				if err != nil {
					return err
				}
				var vparams Values
				if params != nil {
					vparams, ok = params.(Values)
					if !ok {
						vparams = Values{AsOutValue(ctx, params)}
					}
				}
				newtarget := &interfaceValue{}
				if err := flt(ctx, vparams, target, newtarget); err != nil {
					return n.subError(err, "failed to run filter %s: %s", f.str, err)
				}
				target = newtarget
			} else {
				return n.error("tpl: call to undefined filter %s", f.str)
			}
		}
		return out.WriteValue(ctx, target)
	}

	return nil
}

// HasTpl returns true if the template exists in the compiled templates.
func (e *Page) HasTpl(tpl string) bool {
	_, ok := e.compiled[tpl]
	return ok
}

// Parse executes the named template in the given context, writing output to the provided interfaceValue.
// Returns ErrTplNotFound if the template doesn't exist.
func (e *Page) Parse(ctx context.Context, tpl string, out *interfaceValue) error {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return err
	}

	tplData, ok := e.compiled[tpl]
	if !ok {
		return ErrTplNotFound
	}
	return tplData.run(ctx, out)
}

// ParseAndWrite executes the named template in the given context, writing output to the provided io.Writer.
// Returns ErrTplNotFound if the template doesn't exist.
func (e *Page) ParseAndWrite(ctx context.Context, tpl string, out io.Writer) error {
	return e.Parse(ctx, tpl, makeValue(out))
}

// ParseAndReturn executes the named template in the given context and returns the result as a string.
// Returns an empty string and ErrTplNotFound if the template doesn't exist.
func (e *Page) ParseAndReturn(ctx context.Context, tpl string) (string, error) {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return "", err
	}

	buf := &bytes.Buffer{}
	err := e.ParseAndWrite(ctx, tpl, buf)
	return buf.String(), err
}
