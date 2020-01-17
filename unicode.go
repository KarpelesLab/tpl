package tpl

import (
	"context"
	"unicode"

	"golang.org/x/text/unicode/runenames"
	"golang.org/x/text/width"
)

type unicodeInfoObj rune
type unicodeInfoProp rune

func (r unicodeInfoObj) WithCtx(ctx context.Context) *ValueCtx {
	return &ValueCtx{r, ctx}
}

func (r unicodeInfoObj) ReadValue(ctx context.Context) (interface{}, error) {
	return string([]rune{rune(r)}), nil
}

func (r unicodeInfoObj) OffsetGet(ctx context.Context, offset string) (Value, error) {
	// Age NumericValue Property
	switch offset {
	case "Name":
		return NewValue(runenames.Name(rune(r))), nil
	case "Property":
		return unicodeInfoProp(r), nil
	case "Value":
		return NewValue(int(r)), nil
	case "ToLower":
		return unicodeInfoObj(unicode.ToLower(rune(r))), nil
	case "ToUpper":
		return unicodeInfoObj(unicode.ToUpper(rune(r))), nil
	case "IsDigit":
		return NewValue(unicode.IsDigit(rune(r))), nil
	default:
		return NewValue(nil), nil
	}
}

func (r unicodeInfoProp) WithCtx(ctx context.Context) *ValueCtx {
	return &ValueCtx{r, ctx}
}

func (r unicodeInfoProp) ReadValue(ctx context.Context) (interface{}, error) {
	return string([]rune{rune(r)}), nil
}

func (r unicodeInfoProp) OffsetGet(ctx context.Context, offset string) (Value, error) {
	// see https://en.wikipedia.org/wiki/Template:General_Category_(Unicode) for categories
	switch offset {
	case "ascii_hex_digit":
		return NewValue(unicode.Is(unicode.ASCII_Hex_Digit, rune(r))), nil
	case "bidi_control":
		return NewValue(unicode.Is(unicode.Bidi_Control, rune(r))), nil
	case "dash":
		return NewValue(unicode.Is(unicode.Dash, rune(r))), nil
	case "diacritic":
		return NewValue(unicode.Is(unicode.Diacritic, rune(r))), nil
	case "extender":
		return NewValue(unicode.Is(unicode.Extender, rune(r))), nil
	case "hex_digit":
		return NewValue(unicode.Is(unicode.Hex_Digit, rune(r))), nil
	case "hyphen":
		return NewValue(unicode.Is(unicode.Hyphen, rune(r))), nil
	case "ideographic":
		return NewValue(unicode.Is(unicode.Ideographic, rune(r))), nil

	case "alphabetic": // is this an alphabetic character?
		return NewValue(unicode.Is(unicode.Other_Alphabetic, rune(r))), nil // TODO this might not be exact
	case "lowercase":
		return NewValue(unicode.Is(unicode.Lower, rune(r))), nil
	case "uppercase":
		return NewValue(unicode.Is(unicode.Upper, rune(r))), nil
	case "white_space":
		return NewValue(unicode.Is(unicode.White_Space, rune(r))), nil
	case "math":
		return NewValue(unicode.Is(unicode.Sm, rune(r))), nil
	case "quotation_mark":
		return NewValue(unicode.Is(unicode.Quotation_Mark, rune(r))), nil
	case "radical":
		return NewValue(unicode.Is(unicode.Radical, rune(r))), nil
	case "unified_ideograph":
		return NewValue(unicode.Is(unicode.Unified_Ideograph, rune(r))), nil
	case "grapheme_extend":
		return NewValue(unicode.Is(unicode.Other_Grapheme_Extend, rune(r))), nil
	case "east_asian_width":
		switch width.LookupRune(rune(r)).Kind() {
		case width.Neutral:
			return NewValue("NEUTRAL"), nil
		case width.EastAsianAmbiguous:
			return NewValue("AMBIGUOUS"), nil
		case width.EastAsianWide:
			return NewValue("WIDE"), nil
		case width.EastAsianNarrow:
			return NewValue("NARROW"), nil
		case width.EastAsianFullwidth:
			return NewValue("FULLWIDTH"), nil
		case width.EastAsianHalfwidth:
			return NewValue("HALFWIDTH"), nil
		default:
			return NewValue("Unsupported"), nil
		}

		/*
		   case 'grapheme_base':
		   case 'grapheme_link':
		   case 'case_sensitive':
		   case 'line_break':
		           return \IntlChar::getIntPropertyValue($this->char, constant('IntlChar::PROPERTY_'.strtoupper($prop)));
		   case 'numeric_type':
		           $res = \IntlChar::getIntPropertyValue($this->char, constant('IntlChar::PROPERTY_'.strtoupper($prop)));
		           switch($res) {
		                   case \IntlChar::NT_NONE: return 'NONE';
		                   case \IntlChar::NT_DECIMAL: return 'DECIMAL';
		                   case \IntlChar::NT_DIGIT: return 'DIGIT';
		                   case \IntlChar::NT_NUMERIC: return 'NUMERIC';
		                   case \IntlChar::NT_COUNT: return 'COUNT';
		                   default: return $res;
		           }
		*/
	default:
		return NewValue(nil), nil
	}
}
