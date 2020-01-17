package tpl

import (
	"errors"
	"fmt"
)

var (
	ErrTplNotFound = errors.New("tpl: Template not found")
)

// Error is a template error, containing details such as where an error occured
// in the template source and details on the actual error.
type Error struct {
	Message    string
	Template   string
	Line, Char int
	Stack      []byte
	Parent     error
}

func (e *Error) Error() string {
	return e.String()
}

func (e *Error) String() string {
	return fmt.Sprintf("At %s on line %d (position %d): %s", e.Template, e.Line, e.Char, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Parent
}
