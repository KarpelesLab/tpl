package tpl

import (
	"context"
	"fmt"
)

var math_operators = map[string]int{
	// from https://en.wikipedia.org/wiki/Order_of_operations
	"!":  2,
	"~":  2,
	"*":  3,
	"/":  3,
	"%":  3,
	"+":  4,
	"-":  4,
	"<<": 5,
	">>": 5,
	"<":  6,
	"<=": 6,
	">":  6,
	">=": 6,
	"==": 7,
	"!=": 7,
	"&":  8,
	"^":  9,
	"|":  10,
	"&&": 11,
	"||": 12,
	",":  900,
}

func mathSingleValueOperator(ctx context.Context, op string, val1 Value) (*interfaceValue, error) {
	switch op {
	case "!":
		return &interfaceValue{val: !AsOutValue(ctx, val1).AsBool(ctx)}, nil
	case "~":
		val1o, err := val1.ReadValue(ctx)
		if err != nil {
			return nil, err
		}
		switch o := val1o.(type) {
		case bool:
			return &interfaceValue{!o}, nil
		case int64:
			return &interfaceValue{int64(0x7fffffffffffffff ^ o)}, nil
		case uint64:
			return &interfaceValue{uint64(0xffffffffffffffff ^ o)}, nil
		default:
			return &interfaceValue{}, nil // ???
		}
	}
	return &interfaceValue{}, fmt.Errorf("unrecognized operator %s", op)
}

func mathValueOperator(ctx context.Context, op string, val1, val2 Value) (Value, error) {
	if op == "==" || op == "!=" {
		// handle comparisons with go reflection
		o1, err := val1.WithCtx(ctx).Raw()
		if err != nil {
			return nil, err
		}
		o2, err := val2.WithCtx(ctx).Raw()
		if err != nil {
			return nil, err
		}

		r, err := CompareValues(ctx, o1, o2)
		if op == "!=" {
			r = !r
		}

		return NewValue(r), err
	}

	if op == "&&" || op == "||" {
		// need to be bool
		b1, err := val1.WithCtx(ctx).Raw()
		if err != nil {
			return nil, err
		}
		b2, err := val2.WithCtx(ctx).Raw()
		if err != nil {
			return nil, err
		}
		return &interfaceValue{mathValueOperatorBool(op, asBoolIntf(b1), asBoolIntf(b2))}, nil
	}

	o1, err := AsOutValue(ctx, val1).AsNumeric(ctx).ReadValue(ctx)
	if err != nil {
		return nil, err
	}
	o2 := AsOutValue(ctx, val2).AsNumeric(ctx)

	switch v1 := o1.(type) {
	case int64:
		o2v, err := o2.ReadValue(ctx)
		if err != nil {
			return nil, err
		}

		switch v2 := o2v.(type) {
		case int64:
			return &interfaceValue{mathValueOperatorInt64(op, v1, v2)}, nil
		case float64:
			return &interfaceValue{mathValueOperatorFloat64(op, float64(v1), v2)}, nil
		default:
			return &interfaceValue{mathValueOperatorInt64(op, v1, o2.AsInt(ctx))}, nil
		}
	case float64:
		return &interfaceValue{mathValueOperatorFloat64(op, v1, o2.AsFloat(ctx))}, nil
	case bool:
		return &interfaceValue{mathValueOperatorBool(op, v1, o2.AsBool(ctx))}, nil
	}

	return &interfaceValue{}, nil
}

func mathValueOperatorInt64(op string, v1, v2 int64) interface{} {
	switch op {
	case "*":
		return v1 * v2
	case "/":
		return v1 / v2
	case "%":
		return v1 % v2
	case "+":
		return v1 + v2
	case "-":
		return v1 - v2
	case "<<":
		return v1 << uint(v2)
	case ">>":
		return v1 >> uint(v2)
	case "<":
		return v1 < v2
	case "<=":
		return v1 <= v2
	case ">":
		return v1 > v2
	case ">=":
		return v1 >= v2
	case "==":
		return v1 == v2
	case "!=":
		return v1 != v2
	case "&":
		return v1 & v2
	case "^":
		return v1 ^ v2
	case "|":
		return v1 | v2
	default:
		return nil
	}
}

func mathValueOperatorFloat64(op string, v1, v2 float64) interface{} {
	switch op {
	case "*":
		return v1 * v2
	case "/":
		return v1 / v2
	case "+":
		return v1 + v2
	case "-":
		return v1 - v2
	case "<":
		return v1 < v2
	case "<=":
		return v1 <= v2
	case ">":
		return v1 > v2
	case ">=":
		return v1 >= v2
	case "==":
		return v1 == v2
	case "!=":
		return v1 != v2
	default:
		return nil
	}
}

func mathValueOperatorBool(op string, v1, v2 bool) interface{} {
	switch op {
	case "==":
		return v1 == v2
	case "!=":
		return v1 != v2
	case "&&":
		return v1 && v2
	case "||":
		return v1 || v2
	default:
		return nil
	}
}
