package tpl_test

import (
	"context"
	"testing"

	"github.com/KarpelesLab/tpl"
	"golang.org/x/text/language"
)

func TestCompare(t *testing.T) {
	comparePts := []struct {
		A, B   interface{}
		Expect bool
	}{
		// the following results are based on PHP
		{"1", 1, true},
		{false, nil, true},
		{true, nil, false},
		{"1", true, true},
		{true, "hello", true},
		{"hello", true, true},
		{3.14, "3.140", true},
		{"3.140", 3.14, true},
		{language.MustParse("en-US"), "en-US", true},
		{language.MustParse("ja-JP"), "ja-JP", true},
	}

	ctx := context.Background()
	for _, x := range comparePts {
		v, err := tpl.CompareValues(ctx, x.A, x.B)
		if err != nil {
			t.Errorf("Test failed, %#v == %#v resulted in error %s", x.A, x.B, err)
			continue
		}
		if v != x.Expect {
			t.Errorf("Test failed, %#v == %#v returned %#v, expected %#v", x.A, x.B, v, x.Expect)
		}
	}
}
