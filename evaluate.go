package rule

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// EvaluateResult contains the result of rule evaluation
type EvaluateResult struct {
	Result          bool   // The boolean result of the evaluation
	FailedCondition string // The first condition that caused the rule to fail (empty if result is true)
	FailurePosition int    // The starting position of the failing condition in the original string (-1 if no failure)
}

// Evaluate evaluates the rule against the given attributes and returns the result.
func (r *Rule) Evaluate(attributes map[string]interface{}) (bool, error) {
	result, err := r.EvaluateWithFailure(attributes)
	if err != nil {
		return false, err
	}
	return result.Result, nil
}

// EvaluateWithFailure evaluates the rule against the given attributes and returns
// the result along with the first failing condition (if any).
func (r *Rule) EvaluateWithFailure(attributes map[string]interface{}) (EvaluateResult, error) {
	if r.AST == nil {
		return EvaluateResult{}, errors.New("rule has no AST")
	}

	result, failedCondition, pos, err := evaluateNodeWithFailure(r.AST, attributes, r.Condition)
	if err != nil {
		return EvaluateResult{}, err
	}

	return EvaluateResult{
		Result:          result,
		FailedCondition: failedCondition,
		FailurePosition: pos,
	}, nil
}

// evaluateNodeWithFailure recursively evaluates an AST node and tracks the first failing condition
func evaluateNodeWithFailure(node *Node, attributes map[string]interface{}, condition string) (bool, string, int, error) {
	switch node.Type {
	case NodeLiteral:
		// This shouldn't be evaluated directly
		return false, "", -1, errors.New("cannot evaluate literal directly")

	case NodeIdentifier:
		// Get the value from attributes (supports nested paths with dot notation)
		value, exists := getNestedValue(attributes, node.Field)
		if !exists {
			return false, "", -1, fmt.Errorf("attribute '%s' not found", node.Field)
		}
		// For standalone identifiers, treat as boolean
		switch v := value.(type) {
		case bool:
			if !v {
				return false, node.Field, node.Pos, nil
			}
			return true, "", -1, nil
		default:
			return false, "", -1, fmt.Errorf("cannot evaluate non-boolean attribute '%s' as standalone expression", node.Field)
		}

	case NodeComparison:
		return evaluateComparisonWithFailure(node, attributes, condition)

	case NodeLogical:
		return evaluateLogicalWithFailure(node, attributes, condition)

	case NodeNot:
		operand, _, _, err := evaluateNodeWithFailure(node.Operand, attributes, condition)
		if err != nil {
			return false, "", -1, err
		}
		// For NOT, we invert the result but the failed condition changes
		// If the inner condition was true, NOT makes it false, so that's the failing condition
		// If the inner condition was false, NOT makes it true, so no failing condition
		if operand {
			// NOT true = false, so the NOT expression itself fails
			// Build a descriptive failure message showing the actual condition that was negated
			failedStr := extractConditionString(condition, node.Operand.Pos, node.Operand.EndPos)
			return false, fmt.Sprintf("NOT (%s)", failedStr), node.Pos, nil
		}
		return true, "", -1, nil

	default:
		return false, "", -1, fmt.Errorf("unknown node type: %d", node.Type)
	}
}

// extractConditionString extracts a substring from the condition based on position
func extractConditionString(condition string, start, end int) string {
	if start < 0 || end > len(condition) || start >= end {
		return ""
	}
	return strings.TrimSpace(condition[start:end])
}

// evaluateLogicalWithFailure evaluates logical (AND/OR) expressions
func evaluateLogicalWithFailure(node *Node, attributes map[string]interface{}, condition string) (bool, string, int, error) {
	switch node.Operator {
	case "AND":
		// For AND, we need to find the first false condition
		leftResult, leftFailed, leftPos, err := evaluateNodeWithFailure(node.Left, attributes, condition)
		if err != nil {
			return false, "", -1, err
		}
		if !leftResult {
			// Left side failed, return its failing condition
			return false, leftFailed, leftPos, nil
		}
		// Left side passed, check right side
		rightResult, rightFailed, rightPos, err := evaluateNodeWithFailure(node.Right, attributes, condition)
		if err != nil {
			return false, "", -1, err
		}
		if !rightResult {
			// Right side failed
			return false, rightFailed, rightPos, nil
		}
		// Both passed
		return true, "", -1, nil

	case "OR":
		// For OR, we need both sides to be false to report a failure
		leftResult, leftFailed, leftPos, err := evaluateNodeWithFailure(node.Left, attributes, condition)
		if err != nil {
			return false, "", -1, err
		}
		if leftResult {
			// Left side passed, OR is satisfied
			return true, "", -1, nil
		}
		// Left side failed, try right side
		rightResult, _, _, err := evaluateNodeWithFailure(node.Right, attributes, condition)
		if err != nil {
			return false, "", -1, err
		}
		if rightResult {
			// Right side passed, OR is satisfied
			return true, "", -1, nil
		}
		// Both sides failed, return the first failing condition (left)
		return false, leftFailed, leftPos, nil

	default:
		return false, "", -1, fmt.Errorf("unknown logical operator: %s", node.Operator)
	}
}

// evaluateComparisonWithFailure evaluates a comparison node and returns the failing condition
func evaluateComparisonWithFailure(node *Node, attributes map[string]interface{}, condition string) (bool, string, int, error) {
	// Get the field value
	if node.Left.Type != NodeIdentifier {
		return false, "", -1, errors.New("left side of comparison must be an identifier")
	}

	fieldName := node.Left.Field
	fieldValue, exists := getNestedValue(attributes, fieldName)
	if !exists {
		return false, "", -1, fmt.Errorf("attribute '%s' not found", fieldName)
	}

	// Get the comparison value
	compareValue := node.Right.Value

	// Build the condition string for failure reporting
	conditionStr := fmt.Sprintf("%s %s %v", fieldName, node.Operator, formatValue(compareValue))

	// Perform the comparison based on types
	var result bool
	var err error

	switch node.Operator {
	case "==":
		result, err = compareEqual(fieldValue, compareValue)
	case "!=":
		result, err = compareEqual(fieldValue, compareValue)
		result = !result
	case ">", ">=", "<", "<=":
		result, err = compareNumericOrDate(fieldValue, compareValue, node.Operator)
	case "~=":
		result, err = matchRegex(fieldValue, compareValue)
	default:
		return false, "", -1, fmt.Errorf("unknown comparison operator: %s", node.Operator)
	}

	if err != nil {
		return false, "", -1, err
	}

	if !result {
		return false, conditionStr, node.Pos, nil
	}

	return true, "", -1, nil
}

// formatValue formats a value for display in the failed condition string
func formatValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("'%s'", v)
	case float64:
		// Check if it's a whole number
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%v", v)
	case bool:
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// evaluateNode recursively evaluates an AST node (kept for backward compatibility)
func evaluateNode(node *Node, attributes map[string]interface{}) (bool, error) {
	result, _, _, err := evaluateNodeWithFailure(node, attributes, "")
	return result, err
}

// evaluateComparison evaluates a comparison node (kept for backward compatibility)
func evaluateComparison(node *Node, attributes map[string]interface{}) (bool, error) {
	result, _, _, err := evaluateComparisonWithFailure(node, attributes, "")
	return result, err
}

// compareEqual compares two values for equality
func compareEqual(a, b interface{}) (bool, error) {
	// Convert both to comparable types
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)

	// Try numeric comparison first
	aFloat, aErr := toFloat64(a)
	bFloat, bErr := toFloat64(b)
	if aErr == nil && bErr == nil {
		return aFloat == bFloat, nil
	}

	// Try boolean comparison
	aBool, aErr := toBool(a)
	bBool, bErr := toBool(b)
	if aErr == nil && bErr == nil {
		return aBool == bBool, nil
	}

	// Fall back to string comparison
	return aStr == bStr, nil
}

// compareNumericOrDate compares numeric or date values
func compareNumericOrDate(a, b interface{}, op string) (bool, error) {
	// Try date comparison first
	aDate, aErr := toDate(a)
	bDate, bErr := toDate(b)
	if aErr == nil && bErr == nil {
		switch op {
		case ">":
			return aDate.After(bDate), nil
		case ">=":
			return aDate.After(bDate) || aDate.Equal(bDate), nil
		case "<":
			return aDate.Before(bDate), nil
		case "<=":
			return aDate.Before(bDate) || aDate.Equal(bDate), nil
		}
	}

	// Try numeric comparison
	aFloat, aErr := toFloat64(a)
	bFloat, bErr := toFloat64(b)
	if aErr == nil && bErr == nil {
		switch op {
		case ">":
			return aFloat > bFloat, nil
		case ">=":
			return aFloat >= bFloat, nil
		case "<":
			return aFloat < bFloat, nil
		case "<=":
			return aFloat <= bFloat, nil
		}
	}

	return false, fmt.Errorf("cannot compare values: %v and %v", a, b)
}

// matchRegex matches a value against a regex pattern
func matchRegex(value, pattern interface{}) (bool, error) {
	patternStr := fmt.Sprintf("%v", pattern)
	valueStr := fmt.Sprintf("%v", value)

	return regexp.MatchString(patternStr, valueStr)
}

// toFloat64 attempts to convert a value to float64
func toFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}

// toBool attempts to convert a value to bool
func toBool(value interface{}) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		return strconv.ParseBool(v)
	default:
		return false, fmt.Errorf("cannot convert %T to bool", value)
	}
}

// toDate attempts to parse a value as a date
func toDate(value interface{}) (time.Time, error) {
	str := fmt.Sprintf("%v", value)
	return time.Parse("2006-01-02", str)
}

// getNestedValue retrieves a value from a nested map using dot notation
// e.g., "savedTagsCounts.fitness" will traverse the nested structure
func getNestedValue(attributes map[string]interface{}, path string) (interface{}, bool) {
	parts := strings.Split(path, ".")
	var current interface{} = attributes

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			var exists bool
			current, exists = v[part]
			if !exists {
				return nil, false
			}
		case map[string]string:
			val, exists := v[part]
			if !exists {
				return nil, false
			}
			current = val
		case map[string]int:
			val, exists := v[part]
			if !exists {
				return nil, false
			}
			current = val
		case map[string]float64:
			val, exists := v[part]
			if !exists {
				return nil, false
			}
			current = val
		case map[string]bool:
			val, exists := v[part]
			if !exists {
				return nil, false
			}
			current = val
		default:
			return nil, false
		}
	}

	return current, true
}
