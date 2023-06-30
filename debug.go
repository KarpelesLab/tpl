package tpl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/KarpelesLab/pjson"
)

type debugDumpable interface {
	Dump(o io.Writer, lvl int)
}

func debugDump(o io.Writer, i interface{}, lvl int) {
	pfx := strings.Repeat("\t", lvl)

	switch v := i.(type) {
	case debugDumpable:
		v.Dump(o, lvl)
	case Values:
		fmt.Fprintf(o, "%sValues[", pfx)
		for _, sv := range v {
			debugDump(o, sv, lvl)
		}
		fmt.Fprintf(o, "]")
	case *interfaceValue:
		fmt.Fprintf(o, "%sinterfaceValue[", pfx)
		debugDump(o, v.val, lvl)
		fmt.Fprintf(o, "]")
	case json.RawMessage:
		fmt.Fprintf(o, "%sjson.RawMessage[%s]", pfx, v)
	case pjson.RawMessage:
		fmt.Fprintf(o, "%spjson.RawMessage[%s]", pfx, v)
	case interface{ RawJSONBytes() []byte }:
		fmt.Fprintf(o, "%sjson.RawMessage[%s]", pfx, v.RawJSONBytes())
	case *bytes.Buffer:
		fmt.Fprintf(o, "%s%T(%s)", pfx, v, v.Bytes())
	default:
		fmt.Fprintf(o, "%s%T(%#v)", pfx, i, i)
	}
}

func (f *fragment) Dump(o io.Writer, lvl int) {
	pfx := strings.Repeat("\t", lvl)
	fmt.Fprintf(o, "%sfragment(%s)\n", pfx, f.ftyp)
	if f.text != "" {
		fmt.Fprintf(o, "%sValue: %#v\n", pfx, f.text)
	}
	if len(f.data) > 0 {
		fmt.Fprintf(o, "%sdata:\n", pfx)
		for _, s := range f.data {
			s.Dump(o, lvl+1)
		}
	}
	if len(f.linkextra) > 0 {
		fmt.Fprintf(o, "%slinkextra:\n", pfx)
		for _, s := range f.linkextra {
			s.Dump(o, lvl+1)
		}
	}
}

func (n *internalNode) Dump(o io.Writer, lvl int) {
	pfx := strings.Repeat("\t", lvl)
	fmt.Fprintf(o, "%sNode(%s)\n", pfx, n.typ.String())
	if n.str != "" {
		fmt.Fprintf(o, "%sValue: %#v\n", pfx, n.str)
	}
	if len(n.sub) > 0 {
		for _, s := range n.sub {
			s.Dump(o, lvl+1)
		}
	}
	if len(n.filters) > 0 {
		fmt.Fprintf(o, "%sFilters:\n", pfx)
		n.filters.Dump(o, lvl+1)
	}
	if n.value != nil {
		fmt.Fprintf(o, "%sValue:\n", pfx)
		debugDump(o, n.value, lvl+1)
	}
}

func (a *internalArray) Dump(o io.Writer, lvl int) {
	pfx := strings.Repeat("\t", lvl)
	fmt.Fprintf(o, "%sArray [\n", pfx)
	for i, n := range *a {
		fmt.Fprintf(o, "%s%d =>\n", pfx, i)
		n.Dump(o, lvl+1)
	}
	fmt.Fprintf(o, "%s]\n", pfx)
}

func (e *Page) Dump(o io.Writer, lvl int) {
	pfx := strings.Repeat("\t", lvl)
	for t, a := range e.compiled {
		fmt.Fprintf(o, "%sTpl(%s)\n", pfx, t)
		a.Dump(o, lvl+1)
	}
}
