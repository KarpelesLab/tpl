package tpl_test

import (
	"context"
	"testing"

	"github.com/KarpelesLab/tpl"
)

func TestQueryEscapeAny(t *testing.T) {
	ctx := context.Background()
	
	tests := []struct {
		input any
		want  string
	}{
		{"test string", "test+string"},
		{123, "123"},
		{[]byte("test bytes"), "test+bytes"},
		{nil, ""},
	}
	
	for _, tt := range tests {
		// Direct input without conversion to Value to test the type switch
		got := tpl.QueryEscapeAny(ctx, tt.input)
		if got != tt.want {
			t.Errorf("QueryEscapeAny(%v) = %v, want %v", tt.input, got, tt.want)
		}
	}
}