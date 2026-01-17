package tpl

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
)

type fragment struct {
	ftyp       string
	line, char int
	text       string
	data       fragments
	linkextra  fragments
	ctx        *step1_context
}

type fragments []*fragment

// Error returns a template error suitable for being returned or for panic
func (f *fragment) error(msg string, i ...interface{}) error {
	return &Error{Message: fmt.Sprintf(msg, i...), Template: f.ctx.tpl, Line: f.line, Char: f.char}
}

func (f *fragment) newNode() *internalNode {
	n := new(internalNode)
	n.typ = internalInvalid
	n.e = f.ctx.e
	n.tpl, n.line, n.char = f.ctx.tpl, f.line, f.char
	return n
}

type step1_context struct {
	curLine, curChar     int
	startLine, startChar int
	stack                map[int]*fragment
	level                int
	tmpStr               bytes.Buffer
	cLast, c, cNext      byte
	tpl                  string
	e                    *Page
}

func (ctx *step1_context) newFragment(t string) *fragment {
	f := new(fragment)
	f.ftyp = t
	f.ctx = ctx
	f.data = make(fragments, 0)
	f.line = ctx.startLine
	f.char = ctx.startChar
	ctx.startLine = ctx.curLine
	ctx.startChar = ctx.curChar
	if ctx.level >= 0 && t != "|" {
		ctx.stack[ctx.level].data = append(ctx.stack[ctx.level].data, f)
	} else if ctx.level >= 0 && t == "|" {
		ctx.stack[ctx.level].linkextra = append(ctx.stack[ctx.level].linkextra, f)
	}
	ctx.level++
	ctx.stack[ctx.level] = f
	return f
}

func (ctx *step1_context) newTextFragment(txt string) *fragment {
	f := new(fragment)
	f.ftyp = "text"
	f.ctx = ctx
	f.text = txt
	f.line = ctx.startLine
	f.char = ctx.startChar
	ctx.startLine = ctx.curLine
	ctx.startChar = ctx.curChar
	return f
}

// Compile processes all raw templates and builds the internal representation.
// It returns an error if any template fails to compile.
func (e *Page) Compile(ctx context.Context) error {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return err
	}

	// Reset values
	e.Version = 1
	e.compiled = make(map[string]internalArray)

	// Verify the template data is valid
	if !e.Raw.IsValid() {
		return errors.New("template is not valid (is main template missing?)")
	}

	// For each raw template, compile
	for tpl, data := range e.Raw.TemplateData {
		// Check for context cancellation between compiling templates
		if err := ctx.Err(); err != nil {
			return err
		}

		if err := e.compileTpl_step1(ctx, tpl, data); err != nil {
			// Add context to the error if it's not already our custom error type
			if _, ok := err.(*Error); !ok {
				err = &Error{
					Message:  err.Error(),
					Template: tpl,
					Parent:   err,
				}
			}
			return err
		}
	}

	return nil
}

func (ctx *step1_context) flush() {
	if ctx.tmpStr.Len() == 0 {
		return
	}
	ctx.stack[ctx.level].data = append(ctx.stack[ctx.level].data, ctx.newTextFragment(ctx.tmpStr.String()))
	ctx.tmpStr.Reset()
}

func (e *Page) compileTpl_step1(rctx context.Context, tpl, data string) error {
	// analyze string, detect {{ and }} and append in an array where it is cut
	var ctx step1_context

	// initialize context
	ctx.tpl = tpl
	ctx.e = e
	ctx.level = -1
	ctx.curLine, ctx.startLine = 1, 1
	ctx.stack = make(map[int]*fragment)

	// create root
	root := ctx.newFragment("root")

	for i := 0; i < len(data); i++ {
		// initialize vars
		ctx.curChar++
		ctx.cLast = ctx.c
		ctx.c = data[i]
		if i+1 < len(data) {
			ctx.cNext = data[i+1]
		} else {
			ctx.cNext = 0
		}

		// based on current var...
		switch {
		case ctx.c == '{' && ctx.cNext == '{': // opening expression
			if ctx.cLast == '\\' {
				// this is actually an escaped one, not to be an opening
				ctx.tmpStr.Truncate(ctx.tmpStr.Len() - 1) // remove last char, which should be the \\
				ctx.tmpStr.WriteString("{{")
				i++
				ctx.curChar++
				break
			}
			// is this a literal fragment?
			if strings.HasPrefix(data[i:], "{{literal}}") {
				endPos := strings.Index(data[i:], "{{/literal}}")
				if endPos > 0 {
					ctx.tmpStr.WriteString(data[i+11 : i+endPos])
					i += endPos + 11
					break
				}
			}
			ctx.flush()
			ctx.newFragment("{{")
			i++
			ctx.curChar++
		case ctx.c == '}' && ctx.cNext == '}' && ctx.level > 0 && ctx.stack[ctx.level].ftyp != `"`: // closing expression
			ctx.flush()
			if ctx.stack[ctx.level].ftyp == "|" {
				ctx.level--
			}
			ctx.level--
			i++
			ctx.curChar++
		case ctx.c == '\\' && ctx.cNext == '"' && ctx.stack[ctx.level].ftyp == `"`: // escaped quote
			ctx.tmpStr.WriteByte('"')
			i++
			ctx.curChar++
		case ctx.c == '"' && ctx.level > 0 && ctx.stack[ctx.level].ftyp != `"`: // opening quote
			ctx.flush()
			ctx.newFragment(`"`)
		case ctx.c == '"' && ctx.stack[ctx.level].ftyp == `"`: // closing quote
			ctx.flush()
			ctx.level--
		case ctx.c == '(' && ctx.level > 0 && ctx.stack[ctx.level].ftyp != `"`: // (sub)parenthesis in {{}}
			ctx.flush()
			ctx.newFragment("(")
		case ctx.c == ')' && ctx.stack[ctx.level].ftyp == "(": // closing parenthesis
			ctx.flush()
			ctx.level--
		case ctx.c == '|' && ctx.cNext != '|' && ctx.cLast != '|' && ctx.level > 0 && ctx.stack[ctx.level].ftyp != `"`: // argument separator (filters/etc)
			ctx.flush()
			if ctx.stack[ctx.level].ftyp == "|" {
				ctx.level--
			}
			ctx.newFragment("|")
		case ctx.c == '\n': // new line
			ctx.curLine++
			ctx.curChar = 0
			ctx.tmpStr.WriteByte(ctx.c)
		default:
			ctx.tmpStr.WriteByte(ctx.c)
		}
	}

	ctx.flush()

	if ctx.level != 0 {
		var tags []string
		for i := 0; i <= ctx.level; i++ {
			tags = append(tags, ctx.stack[i].ftyp)
		}
		return ctx.stack[ctx.level].error("Badly constructed page, not all tags are properly closed: %s", tags)
	}

	return e.compileTpl_step2(rctx, tpl, root)
}

func (e *Page) compileTpl_step2(ctx context.Context, tpl string, root *fragment) error {
	newroot, err := e.compileTpl_step2_recurse(ctx, root.data, false)
	if err != nil {
		return err
	}
	e.compiled[tpl] = newroot

	return nil
}

func (e *Page) compileTpl_step2_recurse(ctx context.Context, fl fragments, isExpression bool) (res internalArray, err error) {
	res = internalArray{}
	stack := make(map[int]*internalNode)
	stackArray := make(map[int]*internalArray)
	level := 0
	cur := &res
	stackArray[level] = &res

Loop:
	for i := 0; i < len(fl); i++ {
		cur = stackArray[level]
		f := fl[i]
		n := f.newNode()

		switch f.ftyp {
		case "text":
			if isExpression {
				txt := f.text
				// character by character analysis
				var val string
				for j := 0; j < len(txt); j++ {
					switch txt[j] {
					case ' ', '\r', '\n', '\t':
						// NOOP - skip whitespace
					case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.':
						val = val + string(txt[j])
					default:
						// check if math operators
						var l int
						var ok bool

						if j < len(txt)-1 {
							_, ok = math_operators[txt[j:j+2]]
							l = 2
						}
						if !ok {
							_, ok = math_operators[txt[j:j+1]]
							l = 1
						}

						if !ok {
							return nil, &Error{Message: fmt.Sprintf("unhandled char %c", txt[j]), Template: n.tpl, Line: n.line, Char: n.char}
						}

						// cut text at position
						if len(val) > 0 {
							n.typ = internalText
							n.str = val
							res = append(res, n)
							val = ""
							n = f.newNode()
						}
						n.typ = internalOperator
						n.str = txt[j : j+l]
						res = append(res, n)
						val = ""
						n = f.newNode()
						if l == 2 {
							j++
						}
					}
				}
				if len(val) > 0 {
					n.typ = internalText
					n.str = val
				} else {
					n = nil
				}
			} else {
				// check if text is not empty
				if f.text == "" {
					n = nil
				} else {
					n.typ = internalText
					n.str = f.text
				}
			}
		case `"`:
			n.typ = internalQuote
			n.sub = make([]internalArray, 1)
			n.sub[0], err = e.compileTpl_step2_recurse(ctx, f.data, false)
		case "(":
			n.typ = internalSub
			n.sub = make([]internalArray, 1)
			n.sub[0], err = e.compileTpl_step2_recurse(ctx, f.data, true)
		case "|":
			if len(f.data) != 2 {
				// should be 2 (string func name, then parenthesis args)
				err = f.error("invalid filter call")
				return
			}
			if f.data[0].ftyp != "text" {
				err = f.error("invalid filter call")
				return
			}
			n.typ = internalFilter
			n.str = f.data[0].text
			n.sub = make([]internalArray, 1)
			n.sub[0], err = e.compileTpl_step2_recurse(ctx, f.data[1:2], true)
			if err != nil {
				return
			}
			if n.str[0] == '_' && n.str[len(n.str)-1] == '=' {
				// not a filter but a var assign
				n.typ = internalVar
				n.str = strings.ToLower(n.str[:len(n.str)-1])
			}
			// filter filter? shouldn't happen but who knows...
			n.filters, err = e.compileTpl_step2_recurse(ctx, f.linkextra, false)
		case "{{":
			// check for text
			if f.data[0].ftyp != "text" {
				// Check if it's a standalone quote (for direct string output)
				// e.g., {{"hello"}} or {{"hello"|filter()}}
				if len(f.data) == 1 && f.data[0].ftyp == `"` {
					// Direct string output
					n.typ = internalQuote
					n.sub = make([]internalArray, 1)
					n.sub[0], err = e.compileTpl_step2_recurse(ctx, f.data[0].data, false)
					if err != nil {
						return
					}
					n.filters, err = e.compileTpl_step2_recurse(ctx, f.linkextra, false)
					break
				}
				// Check if it starts with parenthesis - treat as expression
				// e.g., {{(1+2)}} or {{(1 + 2) * 3}}
				if f.data[0].ftyp == "(" {
					// Parse as expression
					n.typ = internalSub
					n.sub = make([]internalArray, 1)
					n.sub[0], err = e.compileTpl_step2_recurse(ctx, f.data, true) // true = isExpression
					if err != nil {
						return
					}
					n.filters, err = e.compileTpl_step2_recurse(ctx, f.linkextra, false)
					break
				}
				// consider it a link
				n.typ = internalLink
				n.sub = make([]internalArray, 1)
				n.sub[0], err = e.compileTpl_step2_recurse(ctx, f.data, false)
				if err != nil {
					return
				}
				n.filters, err = e.compileTpl_step2_recurse(ctx, f.linkextra, false)
				break
			}

			// extract first word in text, analyze it (fallback to link if none)
			txt := f.data[0].text
			cmd := txt
			if cmd[0] == '@' {
				if len(f.data) != 2 {
					// should be 2 (string func name, then parenthesis args)
					err = f.error("invalid method call")
					return
				}
				n.typ = internalFunc
				n.str = cmd[1:]
				n.sub = make([]internalArray, 1)
				n.sub[0], err = e.compileTpl_step2_recurse(ctx, f.data[1:2], true)
				if n.sub[0].isStatic() {
					t := &interfaceValue{}
					fnc, ok := tplFunctions[n.str]
					if ok && fnc.CanCompile {
						if params, err2 := n.sub[0].ToValues(ctx); err2 == nil {
							err = fnc.Method(ctx, params, t)
							if err != nil {
								err = f.error("error running method %s: %s", n.str, err)
								return
							}
							if err == nil {
								// update things
								n.typ = internalValue
								n.str = ""
								n.sub = nil
								n.value = t
							}
						}
					}
				}
				if err != nil {
					return
				}
				n.filters, err = e.compileTpl_step2_recurse(ctx, f.linkextra, false)
				break
			}
			pos := strings.IndexByte(txt, ' ')
			if pos != -1 {
				cmd = txt[:pos]
			}
			switch strings.ToLower(cmd) {
			case "foreach":
				n.typ = internalForeach
				n.sub = make([]internalArray, 2) // will add one more if has else
				if strings.TrimSpace(txt) != "foreach" {
					err = f.error("foreach invalid syntax")
					return
				}
				if len(f.data) != 3 {
					// wat??? we expect "foreach " [var] "as (varname)"
					err = f.error("foreach invalid syntax")
					return
				}
				if f.data[2].ftyp != "text" {
					// we expect text to contain "as _varname"
					err = f.error("foreach invalid syntax")
					return
				}
				// parse foreach variable
				n.sub[0], err = e.compileTpl_step2_recurse(ctx, fragments{f.data[1]}, false) // condition for {{if}}
				if err != nil {
					return
				}
				n.sub[1] = internalArray{}
				foreachVar := strings.TrimSpace(f.data[2].text)
				if !strings.HasPrefix(foreachVar, "as ") {
					err = f.error("foreach invalid syntax")
					return
				}
				foreachVar = strings.TrimSpace(foreachVar[2:])
				if len(foreachVar) < 2 || !strings.HasPrefix(foreachVar, "_") {
					err = f.error("foreach invalid syntax")
					return
				}
				n.str = strings.ToLower(foreachVar)
				level++
				stack[level] = n
				stackArray[level] = &n.sub[1]
			case "/foreach":
				if level < 1 || stack[level].typ != internalForeach {
					err = f.error("/foreach at invalid position")
					return
				}
				level--
				cur = stackArray[level]
				n = nil
			case "set":
				n.typ = internalSet
				n.sub = make([]internalArray, 1) // will add one more if has else
				if len(txt) <= 4 {
					if len(f.data) > 1 {
						f.data = f.data[1:]
					} else {
						//hum? let's just avoid crashing
						f.data = make(fragments, 0)
					}
				} else {
					f.data[0].text = f.data[0].text[4:]
				}

				// let's just parse all we now have in f.data, should be one text node ( _key=) and a value (quoted, link, etc)
				for len(f.data) > 1 {
					if f.data[0].ftyp != "text" {
						err = f.data[0].error("invalid set instruction")
						return
					}

					tmp := strings.TrimSpace(f.data[0].text)
					if !strings.HasSuffix(tmp, "=") {
						err = f.data[0].error("invalid set instruction")
						return
					}
					tmp = strings.ToLower(strings.TrimSpace(strings.TrimSuffix(tmp, "=")))

					if tmp[0] != '_' {
						err = f.data[0].error("invalid set instruction")
						return
					}

					nVar := f.data[0].newNode()
					nVar.typ = internalVar
					nVar.str = tmp
					nVar.sub = make([]internalArray, 1)
					nVar.sub[0], err = e.compileTpl_step2_recurse(ctx, f.data[1:2], true)
					if err != nil {
						return
					}

					n.filters = append(n.filters, nVar)

					f.data = f.data[2:]
				}

				n.sub[0] = internalArray{}
				level++
				stack[level] = n
				stackArray[level] = &n.sub[0]
			case "/set":
				if level < 1 || stack[level].typ != internalSet {
					err = f.error("/set at invalid position")
					return
				}
				level--
				cur = stackArray[level]
				n = nil
			case "if":
				n.typ = internalIf
				n.sub = make([]internalArray, 2) // will add one more if has else
				if len(txt) <= 3 {
					if len(f.data) > 1 {
						f.data = f.data[1:]
					} else {
						//hum? let's just avoid crashing
						f.data = make(fragments, 0)
					}
				} else {
					f.data[0].text = f.data[0].text[3:]
				}
				n.sub[0], err = e.compileTpl_step2_recurse(ctx, f.data, true) // condition for {{if}}
				if err != nil {
					return
				}
				n.sub[1] = internalArray{}
				level++
				stack[level] = n
				stackArray[level] = &n.sub[1]
			case "/if":
				if level < 1 || stack[level].typ != internalIf {
					err = f.error("/if at invalid position")
					return
				}
				level--
				cur = stackArray[level]
				n = nil
			case "else":
				if level < 1 || (stack[level].typ != internalIf && stack[level].typ != internalForeach) {
					err = f.error("else at invalid position")
					return
				}
				if len(stack[level].sub) > 2 {
					err = f.error("else present more than once")
					return
				}
				stack[level].sub = append(stack[level].sub, internalArray{})
				stackArray[level] = &stack[level].sub[len(stack[level].sub)-1]
				n = nil
			case "elseif":
				// create a new if in the current if's sub[2], and record it at the same level in the stack
				if level < 1 || (stack[level].typ != internalIf) {
					err = f.error("elseif at invalid position")
					return
				}
				if len(stack[level].sub) > 2 {
					err = f.error("elseif after else")
					return
				}

				n.typ = internalIf
				n.sub = make([]internalArray, 2) // will add one more if has else
				if len(txt) <= 7 {
					if len(f.data) > 1 {
						f.data = f.data[1:]
					} else {
						//hum? let's just avoid crashing
						f.data = make(fragments, 0)
					}
				} else {
					f.data[0].text = f.data[0].text[7:]
				}
				n.sub[0], err = e.compileTpl_step2_recurse(ctx, f.data, true) // condition for {{if}}
				if err != nil {
					return
				}
				n.sub[1] = internalArray{}

				// add this if as else of previous if
				stack[level].sub = append(stack[level].sub, internalArray{n})
				// then record it at the same level
				stack[level] = n
				stackArray[level] = &n.sub[1]
			case "try":
				n.typ = internalTry
				n.sub = make([]internalArray, 1) // will add one more if has else
				n.sub[0] = internalArray{}
				level++
				stack[level] = n
				stackArray[level] = &n.sub[0]
			case "catch":
				// create a new if in the current if's sub[2], and record it at the same level in the stack
				if level < 1 || (stack[level].typ != internalTry) {
					err = f.error("catch at invalid position")
					return
				}
				if len(stack[level].sub) > 1 {
					err = f.error("catch after catch?")
					return
				}

				if len(txt) <= 6 {
					if len(f.data) > 1 {
						err = f.error("invalid parameters in catch")
						return
					}
				} else {
					stack[level].str = strings.ToLower(strings.TrimSpace(txt[6:]))
				}

				stack[level].sub = append(stack[level].sub, internalArray{})
				stackArray[level] = &stack[level].sub[len(stack[level].sub)-1]
				n = nil
			case "/try":
				if level < 1 || stack[level].typ != internalTry {
					err = f.error("/try at invalid position")
					return
				}
				level--
				cur = stackArray[level]
				n = nil
			}
			if n == nil {
				break
			}
			if n.typ == internalInvalid {
				// Check if the text looks like a numeric expression (starts with digit or decimal point)
				// e.g., {{1+1}}, {{3.14}}, {{1 + 2 * 3}}
				trimmedTxt := strings.TrimSpace(txt)
				if len(trimmedTxt) > 0 && (isDigit(trimmedTxt[0]) || (trimmedTxt[0] == '.' && len(trimmedTxt) > 1 && isDigit(trimmedTxt[1]))) {
					// This looks like a numeric expression - parse as expression and output
					n.typ = internalSub
					n.sub = make([]internalArray, 1)
					n.sub[0], err = e.compileTpl_step2_recurse(ctx, f.data, true) // true = isExpression
					if err != nil {
						return
					}
					n.filters, err = e.compileTpl_step2_recurse(ctx, f.linkextra, false)
					break
				}

				// consider it a link, but start by making text lowercase
				pos := strings.IndexByte(txt, '/')
				if pos > 0 {
					f.data[0].text = strings.ToLower(txt[:pos]) + txt[pos:]
				} else {
					f.data[0].text = strings.ToLower(txt)
				}
				n.typ = internalLink
				n.sub = make([]internalArray, 1)
				n.sub[0], err = e.compileTpl_step2_recurse(ctx, f.data, false)
				if err != nil {
					return
				}
				n.filters, err = e.compileTpl_step2_recurse(ctx, f.linkextra, false)
			}
		}

		if err != nil {
			break Loop
		}

		if n != nil {
			if n.typ == internalInvalid {
				f.Dump(os.Stdout, 0)
			}
			*cur = append(*cur, n)
		}
	}

	if level != 0 {
		// something isn't closed on the stack
		err = stack[level].error("element isn't closed")
		return
	}

	// if not an expression, processing ends here
	if !isExpression {
		return
	}

	// range over res to check things that needs to be made recursive (operators, etc)
	// we should always have an operator between values
	var last_comma *internalNode

	for {
		has_op := false
		op_pos := 0
		op_weight := 999

		// look for operator with lowest weight
		for i, n := range res {
			if n.typ == internalOperator && len(n.sub) == 0 {
				has_op = true
				op_w := math_operators[n.str]
				if op_w == 0 {
					// syntax error
					return nil, n.error("invalid operator %s", n.str)
				}
				if op_w < op_weight {
					op_weight = op_w
					op_pos = i
				}
			}
		}
		if !has_op {
			break
		}

		n := res[op_pos]

		if op_pos == len(res)-1 {
			return nil, n.error("invalid operator %s at end of expression", n.str)
		}

		next_n := res[op_pos+1]

		if op_pos == 0 {
			if n.str == "!" || n.str == "~" {
				// special case!
				n.sub = []internalArray{internalArray{next_n}}
				res = res[1:]
				res[0] = n
				continue
			}
			return nil, n.error("invalid operator %s at start of expression", n.str)
		}

		if n.str == "!" || n.str == "~" {
			// special case!
			n.sub = []internalArray{internalArray{next_n}}
			copy(res[op_pos:], res[op_pos+1:])
			res[len(res)-1] = nil
			res[op_pos] = n
			res = res[:len(res)-1]
			continue
		}

		prev_n := res[op_pos-1]

		if n.str == "," {
			if last_comma != nil {
				// we already got a comma, group what we have here with it
				if prev_n != last_comma {
					return nil, n.error("invalid comma position")
				}
				// we need to remove cur and next, not prev
				prev_n.sub = append(prev_n.sub, internalArray{next_n})
				copy(res[op_pos:], res[op_pos+2:])
				res[len(res)-1] = nil
				res[len(res)-2] = nil
				res = res[:len(res)-2]
				continue
			} else {
				// special case
				n.typ = internalList
				n.str = ""
				last_comma = n
			}
		}

		// remove prev/next from res
		copy(res[op_pos-1:], res[op_pos+1:])
		res[len(res)-1] = nil
		res[len(res)-2] = nil
		res[op_pos-1] = n
		res = res[:len(res)-2]

		n.sub = []internalArray{internalArray{prev_n}, internalArray{next_n}}

	}
	return
}

// isDigit returns true if the byte is a digit (0-9)
func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}