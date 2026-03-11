package rule

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Evaluate evaluates the rule against the given attributes
func (r *Rule) Evaluate(attributes map[string]interface{}) (bool, error) {
	if r.AST == nil {
		return false, errors.New("rule has no AST")
	}
	return evaluateNode(r.AST, attributes)
}

// evaluateNode recursively evaluates an AST node
func evaluateNode(node *Node, attributes map[string]interface{}) (bool, error) {
	switch node.Type {
	case NodeLiteral:
		// This shouldn't be evaluated directly
		return false, errors.New("cannot evaluate literal directly")

	case NodeIdentifier:
		// Get the value from attributes (supports nested paths with dot notation)
		value, exists := getNestedValue(attributes, node.Field)
		if !exists {
			return false, fmt.Errorf("attribute '%s' not found", node.Field)
		}
		// For standalone identifiers, treat as boolean
		switch v := value.(type) {
		case bool:
			return v, nil
		default:
			return false, fmt.Errorf("cannot evaluate non-boolean attribute '%s' as standalone expression", node.Field)
		}

	case NodeComparison:
		return evaluateComparison(node, attributes)

	case NodeLogical:
		left, err := evaluateNode(node.Left, attributes)
		if err != nil {
			return false, err
		}
		right, err := evaluateNode(node.Right, attributes)
		if err != nil {
			return false, err
		}
		switch node.Operator {
		case "AND":
			return left && right, nil
		case "OR":
			return left || right, nil
		default:
			return false, fmt.Errorf("unknown logical operator: %s", node.Operator)
		}

	case NodeNot:
		operand, err := evaluateNode(node.Operand, attributes)
		if err != nil {
			return false, err
		}
		return !operand, nil

	default:
		return false, fmt.Errorf("unknown node type: %d", node.Type)
	}
}

// evaluateComparison evaluates a comparison node
func evaluateComparison(node *Node, attributes map[string]interface{}) (bool, error) {
	// Get the field value
	if node.Left.Type != NodeIdentifier {
		return false, errors.New("left side of comparison must be an identifier")
	}

	fieldName := node.Left.Field
	fieldValue, exists := getNestedValue(attributes, fieldName)
	if !exists {
		return false, fmt.Errorf("attribute '%s' not found", fieldName)
	}

	// Get the comparison value
	compareValue := node.Right.Value

	// Perform the comparison based on types
	switch node.Operator {
	case "==":
		return compareEqual(fieldValue, compareValue)
	case "!=":
		result, err := compareEqual(fieldValue, compareValue)
		return !result, err
	case ">", ">=", "<", "<=":
		return compareNumericOrDate(fieldValue, compareValue, node.Operator)
	case "~=":
		return matchRegex(fieldValue, compareValue)
	default:
		return false, fmt.Errorf("unknown comparison operator: %s", node.Operator)
	}
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
