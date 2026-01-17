# Template Engine Syntax Documentation

This document describes the syntax and capabilities of the TPL-GO template engine.

## Basic Template Syntax

### Delimiters
- Templates use `{{` and `}}` as delimiters to mark expressions and control structures
- Static text outside delimiters is rendered as-is
- Escape delimiters with backslash: `\{{` to render literal `{{`
- Use `{{literal}}...{{/literal}}` to render enclosed content as literal text without processing

### Template Inclusion
- `{{NAME}}` without underscore prefix includes another template file (e.g., `{{HEADER}}` includes header.tpl)
- Templates cannot be included with direct parameters
- Variables for included templates must be set using the set block:
  ```
  {{set _X="abc"}}
    {{HEADER}}
  {{/set}}
  ```
- Variables set before inclusion become available in the included template

### Variables
- Variables are accessed within delimiters: `{{_VARIABLE_NAME}}`
- Variable names are case-sensitive
- Variables **must** be prefixed with underscore (e.g., `{{_VARIABLE}}`) to distinguish from template inclusion

### Array and Object Access

Access array elements and object properties using the forward slash (`/`) notation:

```
{{_ARRAY/0}}              Access first element of array (0-indexed)
{{_ARRAY/1}}              Access second element
{{_OBJECT/property}}      Access object property by name
{{_DATA/users/0/name}}    Chain multiple levels of access
```

**Supported types for indexing:**
- Arrays/slices: Access by numeric index (0-based)
- Maps/objects: Access by string key
- JSON data: Automatically parsed and accessed
- url.Values: Access query parameter values

**Examples:**
```
{{set _USER={"name": "John", "address": {"city": "Paris"}}}}
  Name: {{_USER/name}}
  City: {{_USER/address/city}}
{{/set}}

{{set _ITEMS=("apple", "banana", "cherry")}}
  First: {{_ITEMS/0}}
  Last: {{_ITEMS/2}}
{{/set}}
```

**Note:** Out of bounds array access returns null/empty rather than an error.

### Expressions

Expressions can be output directly or used in control structures:

**Direct output** (numbers, strings, and expressions starting with digit, quote, or parenthesis):
```
{{42}}                     outputs "42"
{{1+1}}                    outputs "2"
{{1 + 2 * 3}}              outputs "7" (precedence respected)
{{3.14}}                   outputs "3.14"
{{"hello"}}                outputs "hello"
{{"hello"|uppercase()}}    outputs "HELLO"
{{(1 + 2) * 3}}            outputs "9"
```

**In `set` blocks** (assign to variable):
```
{{set _X=(1 + 2)}}{{_X}}{{/set}}    outputs "3"
```

**In conditions**:
```
{{if (1 < 2)}}yes{{/if}}
{{if 1 < 2}}yes{{/if}}              also valid
```

**With variables** (use double braces inside expressions):
```
{{set _X=({{_A}} + {{_B}})}}{{_X}}{{/set}}
{{if ({{_A}} < {{_B}})}}A is less{{/if}}
```

- Comparison operators: `==`, `!=`, `<`, `<=`, `>`, `>=`
- Logical operators: `&&` (AND), `||` (OR), `!` (NOT)
- Parentheses for grouping: `{{(_X == "a") || (_Y == "b")}}`

**Note:** Template names cannot start with digits, so `{{42}}` is unambiguously a number expression.

### Operators

The following operators are supported, listed by precedence (lowest number = highest precedence):

| Precedence | Operator | Description |
|------------|----------|-------------|
| 2 | `!` | Logical NOT |
| 2 | `~` | Bitwise NOT |
| 3 | `*` | Multiplication |
| 3 | `/` | Division |
| 3 | `%` | Modulo |
| 4 | `+` | Addition |
| 4 | `-` | Subtraction |
| 5 | `<<` | Left bit shift |
| 5 | `>>` | Right bit shift |
| 6 | `<` | Less than |
| 6 | `<=` | Less than or equal |
| 6 | `>` | Greater than |
| 6 | `>=` | Greater than or equal |
| 7 | `==` | Equal |
| 7 | `!=` | Not equal |
| 8 | `&` | Bitwise AND |
| 9 | `^` | Bitwise XOR |
| 10 | `\|` | Bitwise OR |
| 11 | `&&` | Logical AND |
| 12 | `\|\|` | Logical OR |

**Examples:**
```
{{5 % 3}}           outputs 2
{{8 >> 2}}          outputs 2
{{3 << 2}}          outputs 12
{{5 & 3}}           outputs 1
{{5 | 3}}           outputs 7
{{5 ^ 3}}           outputs 6
{{~0}}              outputs -1
```

## Functions

Functions are called with the `@` prefix:

```
{{@function_name("param1", "param2")}}
```

### Core Functions

#### @string(value)
Converts a value of any type to its string representation.

**Example:**
```
{{@string(123)}} outputs "123"
{{@string(true)}} outputs "true"
```

#### @printf(format, args...)
Formats a string using the specified format string and arguments, following Go's fmt.Sprintf conventions.

**Example:**
```
{{@printf("Hello, %s! You are %d years old.", "John", 30)}}
outputs "Hello, John! You are 30 years old."
```

#### @error(format, args...)
Generates an error with the formatted message. This will halt template processing unless caught in a try/catch block.

**Example:**
```
{{try}}
  {{@error("Invalid value: %s", _VALUE)}}
{{catch _E}}
  Error occurred: {{_E}}
{{/try}}
```

#### @redirect(url)
Redirects to the specified URL. This is typically used in web applications.

**Example:**
```
{{if _USER_LOGGED_OUT}}
  {{@redirect("/login")}}
{{/if}}
```

#### @rand(min, max)
Generates a random integer between min and max (inclusive).

**Example:**
```
Random number between 1 and 10: {{@rand(1, 10)}}
```

#### @seq(start, end, [step])
Generates a sequence of numbers from start to end with optional step increment (default step is 1).

**Note:** Returns an empty array if `end < start`.

**Example:**
```
{{foreach {{@seq(1, 5)}} as _NUM}}
  {{_NUM}}
{{/foreach}}
outputs "1 2 3 4 5"

{{foreach {{@seq(0, 10, 2)}} as _NUM}}
  {{_NUM}}
{{/foreach}}
outputs "0 2 4 6 8 10"
```

#### @phpversion()
Returns the Go runtime version (for compatibility with PHP-style templates).

**Example:**
```
{{@phpversion()}} outputs something like "go1.23.0"
```

## Filters

Filters transform values and are applied with the pipe symbol:

```
{{variable|filter(params)}}
```

### String Filters

#### |uppercase()
Converts a string to uppercase.

**Example:**
```
{{"hello world"|uppercase()}} outputs "HELLO WORLD"
```

#### |lowercase()
Converts a string to lowercase.

**Example:**
```
{{"HELLO WORLD"|lowercase()}} outputs "hello world"
```

#### |trim()
Removes whitespace from the beginning and end of a string.

**Example:**
```
{{" hello world "|trim()}} outputs "hello world"
```

#### |substr(start, length)
Extracts a substring starting at the specified position with the specified length.

**Supports negative indices:**
- Negative `start`: counts from end of string
- Negative `length`: reduces the extracted length from the remaining string

**Example:**
```
{{"hello world"|substr(0, 5)}} outputs "hello"
{{"hello world"|substr(6, 5)}} outputs "world"
{{"hello world"|substr(-5, 5)}} outputs "world"
{{"hello world"|substr(0, -6)}} outputs "hello"
```

#### |truncate(length, [ellipsis], [word_cut])
Truncates a string to the specified length, adding an ellipsis if truncated.

**Parameters:**
- `length`: Maximum length (default: 100)
- `ellipsis`: String to append when truncated (default: `…`)
- `word_cut`: If `true`, cut exactly at length; if `false` (default), cut at last word boundary

**Example:**
```
{{"This is a very long sentence"|truncate(10)}} outputs "This is a…"
{{"This is a very long sentence"|truncate(10, "...")}} outputs "This is a..."
{{"This is a very long sentence"|truncate(10, "...", true)}} outputs "This is a ..."
```

#### |entities()
Converts HTML special characters to their entity equivalents.

**Example:**
```
{{"<div>Hello & World</div>"|entities()}} outputs "&lt;div&gt;Hello &amp; World&lt;/div&gt;"
```

#### |striptags()
Removes HTML tags from a string.

**Example:**
```
{{"<div>Hello <strong>World</strong></div>"|striptags()}} outputs "Hello World"
```

#### |replace(from, to)
Replaces all occurrences of a substring with another substring.

**Example:**
```
{{"hello world"|replace("world", "universe")}} outputs "hello universe"
```

#### |nl2br()
Converts newlines (\n) to HTML line breaks (<br/>).

**Example:**
```
{{"Line 1\nLine 2"|nl2br()}} outputs "Line 1<br/>Line 2"
```

#### |nbsp()
Converts spaces to non-breaking spaces (`&nbsp;` / `\xc2\xa0`).

**Example:**
```
{{"hello world"|nbsp()}} outputs "hello\xc2\xa0world"
```

#### |stripcrlf()
Removes carriage returns (`\r`) and line feeds (`\n`) from a string.

**Example:**
```
{{"Line 1\r\nLine 2"|stripcrlf()}} outputs "Line 1Line 2"
```

### Array Filters

#### |json()
Converts a value to its JSON representation.

**Example:**
```
{{set _ARRAY=(1,2,3,4,5)}}
{{_ARRAY|json()}} outputs "[1,2,3,4,5]"

{{set _OBJ={"name": "John", "age": 30}}}
{{_OBJ|json()}} outputs "{"name":"John","age":30}"
```

#### |jsondump()
Converts a value to pretty-printed JSON with indentation.

**Example:**
```
{{set _OBJ={"name": "John", "age": 30}}}
{{_OBJ|jsondump()}}
outputs:
{
	"age": 30,
	"name": "John"
}
```

#### |jsonparse()
Parses a JSON string into an object or array.

**Example:**
```
{{set _JSON_STR='{"name":"John","age":30}'}}
{{set _OBJ=_JSON_STR|jsonparse()}}
{{_OBJ/name}} outputs "John"
```

#### |explode(delimiter)
Splits a string into an array using the specified delimiter.

**Example:**
```
{{set _PARTS="apple,orange,banana"|explode(",")}}
{{_PARTS/0}} outputs "apple"
{{_PARTS/1}} outputs "orange"
```

#### |implode(glue)
Joins array elements with the specified glue string.

**Example:**
```
{{set _ARRAY=("apple", "orange", "banana")}}
{{_ARRAY|implode(", ")}} outputs "apple, orange, banana"
```

#### |reverse()
Reverses an array or string.

**Example:**
```
{{set _ARRAY=("apple", "orange", "banana")}}
{{_ARRAY|reverse()|implode(", ")}} outputs "banana, orange, apple"

{{"hello"|reverse()}} outputs "olleh"
```

#### |arrayslice(offset, [length])
Extracts a portion of an array or string starting at the specified offset with optional length.

**Parameters:**
- With one parameter: `arrayslice(length)` - extracts from start to length
- With two parameters: `arrayslice(offset, length)` - extracts from offset for length items

**Example:**
```
{{set _ARRAY=("apple", "orange", "banana", "grape", "kiwi")}}
{{_ARRAY|arrayslice(1, 2)|implode(", ")}} outputs "orange, banana"
{{_ARRAY|arrayslice(2)|implode(", ")}} outputs "apple, orange"
{{"hello world"|arrayslice(0, 5)}} outputs "hello"
```

#### |arrayfilter(path, value)
Filters an array by keeping only elements where the specified path equals the specified value.

**Example:**
```
{{set _USERS=(
  {"name": "John", "age": 30, "active": true},
  {"name": "Jane", "age": 25, "active": false},
  {"name": "Bob", "age": 40, "active": true}
)}}
{{_USERS|arrayfilter("active", true)|json()}}
outputs '[{"name":"John","age":30,"active":true},{"name":"Bob","age":40,"active":true}]'
```

#### |length() or |count()
Gets the length of an array or string.

**Example:**
```
{{set _ARRAY=("apple", "orange", "banana")}}
{{_ARRAY|length()}} outputs "3"

{{"hello"|length()}} outputs "5"
```

#### |lines(count)
Groups array items into lines with the specified count of items per line.

**Example:**
```
{{set _ITEMS=("item1", "item2", "item3", "item4", "item5", "item6")}}
{{foreach {{_ITEMS|lines(2)}} as _LINE}}
  <div class="row">
  {{foreach {{_LINE}} as _ITEM}}
    <div class="col">{{_ITEM}}</div>
  {{/foreach}}
  </div>
{{/foreach}}
```

#### |columns(count)
Groups array items into columns with the specified count of items per column.

**Example:**
```
{{set _ITEMS=("item1", "item2", "item3", "item4", "item5", "item6")}}
{{foreach {{_ITEMS|columns(2)}} as _COL}}
  <div class="column">
  {{foreach {{_COL}} as _ITEM}}
    <div class="item">{{_ITEM}}</div>
  {{/foreach}}
  </div>
{{/foreach}}
```

#### |keyval(keyPath, [valuePath])
Converts an array of objects into a key-value map, using the specified paths to extract keys and values.

**Example:**
```
{{set _USERS=(
  {"id": "u1", "name": "John"},
  {"id": "u2", "name": "Jane"},
  {"id": "u3", "name": "Bob"}
)}}
{{set _MAP=_USERS|keyval("id", "name")}}
{{_MAP/u1}} outputs "John"
{{_MAP/u2}} outputs "Jane"
```

If only one parameter is provided, the same path is used for both key and value:
```
{{set _MAP=_USERS|keyval("id")}}
{{_MAP/u1}} outputs "u1"
```

### Type Conversion Filters

#### |toint()
Converts a value to an integer.

**Example:**
```
{{"123"|toint()}} outputs 123
{{"123.45"|toint()}} outputs 123
```

#### |tostring()
Converts a value to a string. Similar to the @string function.

**Example:**
```
{{123|tostring()}} outputs "123"
{{true|tostring()}} outputs "true"
```

#### |round([precision])
Rounds a number to the specified decimal precision (default is 2).

**Example:**
```
{{3.14159|round()}} outputs 3.14
{{3.14159|round(0)}} outputs 3
{{3.14159|round(4)}} outputs 3.1416
```

### Encoding Filters

#### |b64enc()
Encodes a string to Base64.

**Example:**
```
{{"hello world"|b64enc()}} outputs "aGVsbG8gd29ybGQ="
```

#### |b64dec()
Decodes a Base64 encoded string.

**Example:**
```
{{"aGVsbG8gd29ybGQ="|b64dec()}} outputs "hello world"
```

#### |urlencode()
URL encodes a string, replacing spaces with plus signs.

**Example:**
```
{{"hello world"|urlencode()}} outputs "hello+world"
{{"a=b&c=d"|urlencode()}} outputs "a%3Db%26c%3Dd"
```

#### |rawurlencode()
Raw URL encodes a string, maintaining spaces as %20.

**Example:**
```
{{"hello world"|rawurlencode()}} outputs "hello%20world"
```

### Formatting Filters

#### |size()
Formats byte sizes into human-readable formats (KB, MB, GB, etc.).

**Example:**
```
{{1024|size()}} outputs "1 KB"
{{1048576|size()}} outputs "1 MB"
{{1073741824|size()}} outputs "1 GB"
```

#### |duration()
Formats duration in seconds to a human-readable time format (days:hh:mm:ss).

**Example:**
```
{{3600|duration()}} outputs "01:00:00" (1 hour)
{{90|duration()}} outputs "01:30" (1 minute, 30 seconds)
{{90061|duration()}} outputs "1:01:01:01" (1 day, 1 hour, 1 minute, 1 second)
{{-3600|duration()}} outputs "-01:00:00" (negative durations supported)
```

#### |date([format])
Formats a date/time value using strftime format specifiers. Default format is `%c` (locale-appropriate date and time).

**Input formats:**
- `"now"` - Current time
- `"@1234567890"` - Unix timestamp (with optional decimal for microseconds)
- Objects with `full` property containing microtime string

**Common format specifiers:**
- `%Y` - 4-digit year
- `%m` - Month (01-12)
- `%d` - Day (01-31)
- `%H` - Hour (00-23)
- `%M` - Minute (00-59)
- `%S` - Second (00-59)
- `%c` - Locale-appropriate date and time

**Example:**
```
{{"@1609459200"|date("%Y-%m-%d")}} outputs "2021-01-01"
{{"now"|date("%H:%M:%S")}} outputs current time
```

### Special Filters

#### |dump() or |export()
Outputs debug information about a variable, showing its type and value.

**Example:**
```
{{set _OBJ={"name": "John", "age": 30}}}
{{_OBJ|dump()}}
outputs something like:
[object map[string]interface {}: {"age":30,"name":"John"}]
```

#### |type()
Returns the type of a variable as a string.

**Example:**
```
{{123|type()}} outputs "int"
{{"hello"|type()}} outputs "string"
{{(1,2,3)|type()}} outputs "array"
```

#### |null()
Suppresses output. Useful when you want to execute an expression without displaying its result.

**Example:**
```
{{set _X="value"|null()}}  No output, but _X is set
{{@redirect("/login")|null()}}  Redirect without output
```

#### |isnull()
Checks if a value is null/nil and returns a boolean.

**Example:**
```
{{if {{_VALUE|isnull()}}}}
  Value is not set
{{/if}}
```

#### |unicode()
Splits a string into an array of unicode character objects, each containing character information.

**Example:**
```
{{foreach {{"héllo"|unicode()}} as _CHAR}}
  {{_CHAR/char}} - {{_CHAR/codepoint}}
{{/foreach}}
```

#### |price()
Formats a price object. Expects an object with a `display` property.

**Example:**
```
{{_PRODUCT/price|price()}}
```
Returns the `display` property value, or "N/A" if not available.

#### |stripbbcode()
Removes BBCode tags from a string.

**Example:**
```
{{"[b]Bold[/b] text"|stripbbcode()}} outputs "Bold text"
```

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