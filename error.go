// Package tpl provides a template engine with string interpolation and control structures.
package tpl

import (
	"errors"
	"fmt"
)

// Common template errors.
var (
	// ErrTplNotFound is returned when a requested template is not found.
	ErrTplNotFound = errors.New("tpl: Template not found")
)

// Error is a template error, containing details such as where an error occurred
// in the template source and details on the actual error.
// It implements the error interface and supports error wrapping with Unwrap().
type Error struct {
	// Message describes the error
	Message string
	// Template is the name of the template where the error occurred
	Template string
	// Line is the line number where the error occurred
	Line int
	// Char is the character position where the error occurred
	Char int
	// Stack contains a stack trace, if available
	Stack []byte
	// Parent is the wrapped error, if any
	Parent error
}

// Error returns a string representation of the error.
// This implements the error interface.
func (e *Error) Error() string {
	return e.String()
}

// String returns a formatted error message with location details.
func (e *Error) String() string {
	return fmt.Sprintf("At %s on line %d (position %d): %s", e.Template, e.Line, e.Char, e.Message)
}

// Unwrap returns the wrapped error, enabling compatibility with errors.Is and errors.As.
func (e *Error) Unwrap() error {
	return e.Parent
}

// Is allows for error comparison using errors.Is.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}

	// Compare the relevant fields
	return e.Template == t.Template &&
		e.Line == t.Line &&
		e.Char == t.Char &&
		e.Message == t.Message
}
