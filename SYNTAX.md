# Template Engine Syntax Documentation

This document describes the syntax and capabilities of the TPL-GO template engine.

## Basic Template Syntax

### Delimiters
- Templates use `{{` and `}}` as delimiters to mark expressions and control structures
- Static text outside delimiters is rendered as-is
- Escape delimiters with backslash: `\{{` to render literal `{{`
- Use `{{literal}}...{{/literal}}` to render enclosed content as literal text without processing

### Variables
- Variables are accessed within delimiters: `{{VARIABLE_NAME}}`
- Variable names are not case-sensitive
- Prefixing variables with underscore is conventional (e.g., `{{_VARIABLE}}`)
- Access nested properties with forward slash: `{{_OBJECT/property}}` (properties are case-sensitive)
- Array elements accessed by index: `{{_ARRAY/0}}`

### Expressions
- Mathematical expressions: `{{1 + 2}}`, `{{1 + 2 * 3}}`, `{{(1 + 2) * 3}}`
- Comparison expressions: `{{_X == "value"}}`, `{{_X != "value"}}`
- Logical operators: `&&` (AND), `||` (OR), `!` (NOT)
- Parentheses for grouping: `{{(_X == "a") || (_Y == "b")}}`

## Functions

Functions are called with the `@` prefix:

```
{{@function_name("param1", "param2")}}
```

### Core Functions
- `@string(value)` - Convert value to string
- `@printf(format, args...)` - Format string using specified format
- `@error(format, args...)` - Generate an error with formatted message
- `@redirect(url)` - Redirect to the specified URL
- `@rand(min, max)` - Generate random number between min and max
- `@seq(start, end, [step])` - Generate sequence of numbers

## Filters

Filters transform values and are applied with the pipe symbol:

```
{{variable|filter(params)}}
```

### String Filters
- `|uppercase()` - Convert string to uppercase
- `|lowercase()` - Convert string to lowercase
- `|trim()` - Remove whitespace from start and end
- `|substr(start, length)` - Extract substring
- `|truncate(length, [ellipsis], [word_cut])` - Truncate string with ellipsis
- `|entities()` - Convert HTML characters to entities
- `|striptags()` - Remove HTML tags
- `|replace(from, to)` - Replace text
- `|nl2br()` - Convert newlines to `<br/>` tags

### Array Filters
- `|json()` - Convert to JSON
- `|jsonparse()` - Parse JSON string to object
- `|explode(delimiter)` - Split string into array
- `|implode(glue)` - Join array elements with glue
- `|reverse()` - Reverse array or string
- `|arrayslice(offset, [length])` - Extract portion of array
- `|arrayfilter(path, value)` - Filter array by property value
- `|length()` or `|count()` - Get array length or string length
- `|lines(count)` - Group array items into lines
- `|columns(count)` - Group array items into columns

### Type Conversion Filters
- `|toint()` - Convert to integer
- `|tostring()` - Convert to string
- `|round([precision])` - Round number to specified precision

### Encoding Filters
- `|b64enc()` - Base64 encode
- `|b64dec()` - Base64 decode
- `|urlencode()` - URL encode
- `|rawurlencode()` - Raw URL encode

### Formatting Filters
- `|size()` - Format byte sizes
- `|duration()` - Format duration (seconds to time format)

### Special Filters
- `|dump()` or `|export()` - Debug output of variable
- `|type()` - Get variable type

## Control Structures

### Conditional Statements
```
{{if condition}}
  content when true
{{elseif other_condition}}
  content when other condition is true
{{else}}
  content when all conditions are false
{{/if}}
```

### Loops
```
{{foreach {{array}} as _item}}
  Access: {{_item}}
  Index: {{_item_idx}}
  Key: {{_item_key}}
  Max: {{_item_max}}
  Previous: {{_item_prv}}
{{else}}
  Content displayed when array is empty
{{/foreach}}
```

### Variable Assignment
```
{{set _X="value"}}
  Content with {{_X}} set
{{/set}}

{{set _X="value" _Y="another value"}}
  Multiple variables: {{_X}} and {{_Y}}
{{/set}}

{{set _X=(1 + 2)}}
  Expression result: {{_X}}
{{/set}}
```

### Error Handling
```
{{try}}
  {{@error("Something went wrong")}}
{{catch _E}}
  Error occurred: {{_E}}
{{/try}}
```

## Variables and Scope

- Variables defined with `{{set}}` are available within the block
- Special variables:
  - `{{_TPL_PAGE}}` - Current template page
  - `{{_EXCEPTION}}` - Error in try/catch blocks

## Special Features

### Literal Text
```
{{literal}}
  Content with {{tags}} that won't be processed
{{/literal}}
```

### Format Conversion
- Markdown to HTML: `{{content|markdown()}}`
- BBCode to HTML: `{{content|bbcode()}}`

### Function Result Chaining
```
{{@string("some text")|uppercase()|json()}}
```

### Compact Syntax for Simple Foreach
```
{{foreach (1,2,3) as _X}}{{_X}}{{/foreach}}
```

### Foreaching Complex Objects
Access nested properties in loops:
```
{{foreach {{_COMPLEX_OBJECT}} as _X}}
  {{if {{_X/property}}=="value"}}
    {{_X/property}}
  {{/if}}
{{/foreach}}
```
