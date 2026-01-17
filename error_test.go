package tpl_test

import (
	"errors"
	"testing"

	"github.com/KarpelesLab/tpl"
)

func TestErrorString(t *testing.T) {
	// Create a test error
	err := &tpl.Error{
		Message:  "test error",
		Template: "test.tpl",
		Line:     10,
		Char:     20,
	}

	// Test Error() method
	expected := "At test.tpl on line 10 (position 20): test error"
	if err.Error() != expected {
		t.Errorf("Error.Error() = %v, want %v", err.Error(), expected)
	}

	// Test String() method
	if err.String() != expected {
		t.Errorf("Error.String() = %v, want %v", err.String(), expected)
	}
}

func TestErrorUnwrap(t *testing.T) {
	// Create a parent error
	parentErr := errors.New("parent error")

	// Create a test error with a parent
	err := &tpl.Error{
		Message:  "test error",
		Template: "test.tpl",
		Line:     10,
		Char:     20,
		Parent:   parentErr,
	}

	// Test Unwrap method
	unwrapped := err.Unwrap()
	if unwrapped != parentErr {
		t.Errorf("Error.Unwrap() = %v, want %v", unwrapped, parentErr)
	}

	// Test with errors.Is
	if !errors.Is(err, parentErr) {
		t.Errorf("errors.Is() failed to match parent error")
	}
}

func TestErrorIs(t *testing.T) {
	// Create two errors with same fields
	err1 := &tpl.Error{
		Message:  "test error",
		Template: "test.tpl",
		Line:     10,
		Char:     20,
	}

	err2 := &tpl.Error{
		Message:  "test error",
		Template: "test.tpl",
		Line:     10,
		Char:     20,
	}

	// Create an error with different fields
	err3 := &tpl.Error{
		Message:  "different error",
		Template: "test.tpl",
		Line:     10,
		Char:     20,
	}

	// Test Is method
	if !err1.Is(err2) {
		t.Errorf("Error.Is() should have matched identical errors")
	}

	if err1.Is(err3) {
		t.Errorf("Error.Is() should not have matched different errors")
	}

	// Test with errors.Is
	if !errors.Is(err1, err2) {
		t.Errorf("errors.Is() failed to match identical errors")
	}

	if errors.Is(err1, err3) {
		t.Errorf("errors.Is() should not have matched different errors")
	}
}

func TestErrTplNotFound(t *testing.T) {
	if tpl.ErrTplNotFound.Error() != "tpl: Template not found" {
		t.Errorf("ErrTplNotFound has incorrect message: %v", tpl.ErrTplNotFound.Error())
	}
}
