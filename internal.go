package tpl

import (
	"fmt"
	"runtime/debug"
)

//go:generate stringer -output stringer.go -type=internalType

const (
	// internalInvalid nodes are invalid (problem during compilation?)
	internalInvalid internalType = iota
	// internalText is a simple string to be passed as is
	internalText
	internalLink     // Link to another template to be included
	internalQuote    // String "in quotes"
	internalValue    // Value passed in value
	internalIf       // If (Sub[0]) Sub[1] else Sub[2]
	internalTry      // Try Sub[0] Catch(Str) Sub[1]
	internalForeach  // foreach(Sub[0] as Str) Sub[1] else(if empty) Sub[2]
	internalJs       // Str
	internalFunc     // Str(Sub[0])
	internalFilter   // Str(Sub[0]) <- (input)
	internalVar      // var "Str" = Sub[0]
	internalOperator // Str=oneOf("+-*/ || && ! etc...") Sub contains 1 entry (!~) or two (other operators)
	internalSub      // Sub[0] - used in expressions
	internalList     // Sub[*] (for example when values are separated by commas), parsed as Values
	internalSet      // Sub[0] + filters (to set variables)
	internalIndex    // Sub[0][Sub[1]] - bracket index access, Sub[0]=base, Sub[1]=index expression
)

// internalNode contains a sub-element in a given page
type internalNode struct {
	typ        internalType
	str        string          // if any text data, or name of var for foreach, catch
	sub        []internalArray // eg. Sub[0]=Expr Sub[1]=Sub Sub[2]=Else
	filters    internalArray   // an array of TPL_FILTER
	value      Value
	line, char int
	tpl        string
	e          *Page
}

type internalArray []*internalNode

// Error returns a template error suitable for being returned or for panic
func (n *internalNode) error(msg string, arg ...interface{}) error {
	return &Error{Message: fmt.Sprintf(msg, arg...), Template: n.tpl, Line: n.line, Char: n.char, Stack: debug.Stack()}
}

// Error returns a template error suitable for being returned or for panic
func (n *internalNode) subError(sub error, msg string, arg ...interface{}) error {
	return &Error{Message: fmt.Sprintf(msg, arg...), Template: n.tpl, Line: n.line, Char: n.char, Stack: debug.Stack(), Parent: sub}
}

// internalType represents the type of a given internal template entry
type internalType int

var tplFunctions = map[string]*TplFunction{}
var tplfilters = map[string]TplFiltCallback{}

func RegisterFunction(name string, f *TplFunction) {
	tplFunctions[name] = f
}

func RegisterFilter(name string, f TplFiltCallback) {
	tplfilters[name] = f
}
