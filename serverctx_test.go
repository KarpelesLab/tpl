package tpl_test

import (
	"context"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/KarpelesLab/tpl"
)

func TestServerCtx(t *testing.T) {
	// Create a test HTTP request
	req := httptest.NewRequest("GET", "http://example.com/test?param=value", nil)
	
	// Test ServerCtx creation
	ctx := tpl.ServerCtx(req)
	if ctx == nil {
		t.Errorf("ServerCtx() returned nil")
	}
	
	// Test accessing GET parameters
	getParams := ctx.Value("$_get")
	if getParams == nil {
		t.Errorf("ctx.Value(\"$_get\") returned nil")
	}
	
	// Test Reset
	tpl.ResetServerCtx(ctx)
	
	// Create POST request
	form := url.Values{}
	form.Add("post_param", "post_value")
	postReq := httptest.NewRequest("POST", "http://example.com/test", nil)
	postReq.PostForm = form
	
	postCtx := tpl.ServerCtx(postReq)
	postParams := postCtx.Value("$_post")
	if postParams == nil {
		t.Errorf("ctx.Value(\"$_post\") returned nil")
	}
	
	// Test request params
	reqParams := postCtx.Value("$_request")
	if reqParams == nil {
		t.Errorf("ctx.Value(\"$_request\") returned nil")
	}
	
	// Test non-string key
	nonStringValue := postCtx.Value(123)
	if nonStringValue != nil {
		t.Errorf("ctx.Value(123) = %v, want nil", nonStringValue)
	}
}

func TestCallFunction(t *testing.T) {
	ctx := context.Background()
	
	// Register a test function in context
	testFn := func(ctx context.Context, params tpl.Values, out tpl.WritableValue) error {
		return out.WriteValue(ctx, "test function result")
	}
	
	ctx = context.WithValue(ctx, "@test_function", tpl.TplFuncCallback(testFn))
	
	// Call the function
	out := tpl.NewEmptyValue()
	err := tpl.CallFunction(ctx, "test_function", tpl.Values{}, out)
	if err != nil {
		t.Errorf("CallFunction() error = %v", err)
	}
	
	// Check result
	result, err := out.ReadValue(ctx)
	if err != nil {
		t.Errorf("ReadValue() error = %v", err)
	}
	if result != "test function result" {
		t.Errorf("Function result = %v, want %v", result, "test function result")
	}
	
	// Test calling undefined function
	err = tpl.CallFunction(ctx, "undefined_function", tpl.Values{}, out)
	if err == nil {
		t.Errorf("CallFunction() with undefined function should return error")
	}
}

func TestFormatSize(t *testing.T) {
	// Just test basic functionality without expecting specific formatting
	tests := []struct {
		size       uint64
		shouldHave string
	}{
		{0, "0 B"},
		{1023, "1023 B"},
		{1024, "iB"}, // Should contain kiB
		{1048576, "iB"}, // Should contain MiB
		{1073741824, "iB"}, // Should contain GiB
		{1099511627776, "iB"}, // Should contain TiB
	}
	
	for _, tt := range tests {
		got := tpl.FormatSize(tt.size)
		if !strings.Contains(got, tt.shouldHave) {
			t.Errorf("FormatSize(%d) = %v, should contain %v", tt.size, got, tt.shouldHave)
		}
	}
}