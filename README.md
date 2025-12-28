[![GoDoc](https://godoc.org/github.com/KarpelesLab/tpl?status.svg)](https://godoc.org/github.com/KarpelesLab/tpl)
[![Coverage Status](https://coveralls.io/repos/github/KarpelesLab/tpl/badge.svg?branch=master)](https://coveralls.io/github/KarpelesLab/tpl?branch=master)

# Template Engine for Go

This is a legacy template engine system, made to be compatible with an older version initially written in PHP.

## Features

- Template syntax with interpolation and control structures
- Support for custom filters and functions
- Parallel template execution
- Structured error handling with location information
- Context-aware template execution

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
	engine.Raw.TemplateData["main"] = "Hello {{_name}}!"

	// Set up a context with variables
	ctx := context.Background()
	ctx = tpl.ValuesCtx(ctx, map[string]any{
		"_name": "World",
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