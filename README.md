# Rule Engine

A powerful and flexible rule engine for Go that evaluates conditions against entity attributes using a simple expression language.

## Features

- **Simple Expression Language**: Intuitive syntax for defining rules
- **Multiple Comparison Operators**: Support for `==`, `!=`, `>`, `<`, `>=`, `<=`, `~=`
- **Logical Operators**: `AND`, `OR`, `NOT` with proper precedence
- **Parentheses Grouping**: Control evaluation order with parentheses
- **Nested JSON Support**: Access nested fields using dot notation (e.g., `user.profile.age`)
- **Type Coercion**: Automatic conversion between numeric types and string-to-number
- **Date Comparison**: Parse and compare dates in `YYYY-MM-DD` format
- **Regex Matching**: Pattern matching using the `~=` operator
- **Failure Tracking**: Identify the first condition that causes a rule to fail

## Installation

```bash
go get github.com/umakantv/rule-engine
```

## Quick Start

```go
package main

import (
    "fmt"
    "rule-engine"
)

func main() {
    // Parse a rule
    rule, err := rule.ParseRule("country == 'US' AND age > 18")
    if err != nil {
        panic(err)
    }

    // Define attributes
    attributes := map[string]interface{}{
        "country": "US",
        "age":     25,
    }

    // Evaluate the rule
    result, err := rule.Evaluate(attributes)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Result: %v\n", result) // Output: Result: true
}
```

## Expression Language

### Comparison Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `==` | Equality | `country == 'US'` |
| `!=` | Inequality | `country != 'US'` |
| `>` | Greater than | `age > 18` |
| `<` | Less than | `age < 65` |
| `>=` | Greater than or equal | `age >= 21` |
| `<=` | Less than or equal | `age <= 65` |
| `~=` | Regex match | `email ~= '^[a-z]+@example\.com$'` |

### Logical Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `AND` | Logical AND | `a == 1 AND b == 2` |
| `OR` | Logical OR | `a == 1 OR b == 2` |
| `NOT` | Logical NOT | `NOT a == 1` |

### Operator Precedence

From highest to lowest:
1. `NOT` (highest)
2. `AND`
3. `OR` (lowest)
4. Parentheses `()` can override precedence

### Literals

- **Strings**: Single or double quoted: `'US'` or `"US"`
- **Numbers**: Integers, decimals, negative numbers, scientific notation: `42`, `3.14`, `-100`, `1e10`
- **Booleans**: `true` or `false` (case-insensitive)

## Usage Examples

### Basic Comparisons

```go
// String comparison
rule.ParseRule("country == 'US'")

// Numeric comparison
rule.ParseRule("age > 18")

// Boolean comparison
rule.ParseRule("active == true")

// Date comparison (YYYY-MM-DD format)
rule.ParseRule("signupDate > '2020-01-01'")

// Regex matching
rule.ParseRule("email ~= '^[a-z]+@example\\.com$'")
```

### Complex Expressions

```go
// AND expression
rule.ParseRule("country == 'US' AND age > 18")

// OR expression
rule.ParseRule("tier == 'premium' OR age > 65")

// NOT expression
rule.ParseRule("NOT country == 'US'")

// Parentheses for grouping
rule.ParseRule("(country == 'US' AND age > 18) OR tier == 'premium'")

// Multiple conditions
rule.ParseRule("a == 1 AND b == 2 AND c == 3")
```

### Nested JSON Attributes

Access nested fields using dot notation:

```go
attributes := map[string]interface{}{
    "country": "US",
    "savedTagsCounts": map[string]interface{}{
        "entertainment": 12,
        "fitness":       5,
        "relationships": 8,
    },
    "user": map[string]interface{}{
        "profile": map[string]interface{}{
            "age": 25,
            "settings": map[string]interface{}{
                "notifications": true,
            },
        },
    },
}

// Single level nesting
rule.ParseRule("savedTagsCounts.fitness > 3")

// Deep nesting
rule.ParseRule("user.profile.age > 18")
rule.ParseRule("user.profile.settings.notifications == true")

// Combined with top-level fields
rule.ParseRule("savedTagsCounts.fitness > 3 AND country == 'US'")
```

### Type Coercion

The engine automatically handles type conversion:

```go
// String number to number comparison
attributes := map[string]interface{}{"value": "42"}
rule.ParseRule("value == 42") // true

// Int to float comparison
attributes := map[string]interface{}{"value": 15} // int
rule.ParseRule("value > 10.5") // true

// Float to int comparison
attributes := map[string]interface{}{"value": 10.0} // float64
rule.ParseRule("value == 10") // true
```

## Failure Tracking

Get detailed information about why a rule fails:

```go
rule, _ := rule.ParseRule("country == 'US' AND age > 18")

attributes := map[string]interface{}{
    "country": "UK",
    "age":     25,
}

result, err := rule.EvaluateWithFailure(attributes)
if err != nil {
    panic(err)
}

fmt.Printf("Result: %v\n", result.Result)                    // false
fmt.Printf("Failed Condition: %s\n", result.FailedCondition) // "country == 'US'"
fmt.Printf("Failure Position: %d\n", result.FailurePosition) // 0 (starting position in the condition string)
```

### Failure Tracking Behavior

- **AND expressions**: Returns the first condition that fails (evaluated left-to-right)
- **OR expressions**: Returns the first failing condition only when all conditions fail
- **NOT expressions**: Returns `NOT (condition)` showing the actual condition that was negated
- **Passing rules**: Returns an empty string for `FailedCondition` and -1 for `FailurePosition`
- **FailurePosition**: The 0-based starting position of the failing condition in the original rule string

```go
// AND - first condition fails
rule.ParseRule("a == 1 AND b == 2")
// attrs: {a: 0, b: 2}
// FailedCondition: "a == 1", FailurePosition: 0

// AND - second condition fails
rule.ParseRule("a == 1 AND b == 2")
// attrs: {a: 1, b: 0}
// FailedCondition: "b == 2", FailurePosition: 11

// OR - both fail, returns first
rule.ParseRule("a == 1 OR b == 2")
// attrs: {a: 0, b: 0}
// FailedCondition: "a == 1", FailurePosition: 0

// OR - one passes, no failure reported
rule.ParseRule("a == 1 OR b == 2")
// attrs: {a: 0, b: 2}
// Result: true, FailedCondition: "", FailurePosition: -1

// NOT - inner condition passes, NOT fails
rule.ParseRule("NOT country == 'US'")
// attrs: {country: "US"}
// FailedCondition: "NOT (country == 'US')", FailurePosition: 0
```

## API Reference

### Types

```go
// Rule represents a parsed rule with its AST
type Rule struct {
    Condition string
    AST       *Node
}

// EvaluateResult contains the result of rule evaluation
type EvaluateResult struct {
    Result          bool   // The boolean result of the evaluation
    FailedCondition string // The first condition that caused the rule to fail
    FailurePosition int    // The starting position of the failing condition (-1 if no failure)
}
```

### Functions

```go
// ParseRule parses a rule string and returns a Rule object
func ParseRule(condition string) (*Rule, error)

// Validate checks if the rule is valid
func (r *Rule) Validate() error

// Evaluate evaluates the rule against the given attributes
func (r *Rule) Evaluate(attributes map[string]interface{}) (bool, error)

// EvaluateWithFailure evaluates the rule and returns the first failing condition
func (r *Rule) EvaluateWithFailure(attributes map[string]interface{}) (EvaluateResult, error)

// String returns the string representation of the rule
func (r *Rule) String() string
```

## Error Handling

The engine returns errors for various scenarios:

```go
// Empty condition
_, err := rule.ParseRule("")
// err: "empty condition string"

// Invalid syntax
_, err := rule.ParseRule("country = 'US'")  // single =
// err: parse error...

// Missing attribute
rule, _ := rule.ParseRule("country == 'US'")
_, err := rule.Evaluate(map[string]interface{}{"age": 25})
// err: "attribute 'country' not found"

// Unterminated string
_, err := rule.ParseRule("country == 'US")
// err: parse error...
```

## Supported Value Types

The engine supports the following value types in attributes:

- `string`
- `int`, `int8`, `int16`, `int32`, `int64`
- `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- `float32`, `float64`
- `bool`
- `map[string]interface{}` (for nested structures)
- `map[string]string`, `map[string]int`, `map[string]float64`, `map[string]bool`

## License

MIT License