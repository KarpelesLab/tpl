[![GoDoc](https://godoc.org/github.com/KarpelesLab/tpl?status.svg)](https://godoc.org/github.com/KarpelesLab/tpl)

# Template Engine for Go

This is a legacy template engine system, made to be compatible with an older version initially written in PHP.

## Features

- Template syntax with interpolation and control structures
- Support for custom filters and functions
- Parallel template execution
- Structured error handling with location information
- Context-aware template execution

## Recent Modernization

This codebase has been updated to use modern Go practices:

- Added comprehensive documentation with GoDoc-compatible comments
- Replaced `interface{}` with `any` type alias
- Enhanced error handling with `errors.Is()` and `errors.As()` support
- Added proper context cancellation handling
- Implemented structured logging with log/slog
- Improved concurrency patterns
- Applied generics for type-safe operations

## Usage

See the [SYNTAX.md](SYNTAX.md) file for template syntax documentation.

```go
package main

import (
	"context"
	"fmt"
	
	"github.com/KarpelesLab/tpl"
)

func main() {
	// Create a new template engine
	engine := tpl.New()
	
	// Add a template
	engine.Raw.TemplateData["main"] = "Hello {{name}}!"
	
	// Set up a context with variables
	ctx := context.Background()
	ctx = tpl.ValuesCtx(ctx, map[string]any{
		"name": "World",
	})
	
	// Compile templates
	if err := engine.Compile(ctx); err != nil {
		panic(err)
	}
	
	// Execute template
	result, err := engine.ParseAndReturn(ctx, "main")
	if err != nil {
		panic(err)
	}
	
	fmt.Println(result) // Output: Hello World!
}
```

## License

This project is released under the MIT license.