package rule

import (
	"testing"
)

func TestLexer(t *testing.T) {
	tests := []struct {
		input    string
		expected []Token
	}{
		{
			input: "country == 'US'",
			expected: []Token{
				{Type: TokenIdentifier, Value: "country"},
				{Type: TokenEqual, Value: "=="},
				{Type: TokenString, Value: "US"},
				{Type: TokenEOF},
			},
		},
		{
			input: "age > 18",
			expected: []Token{
				{Type: TokenIdentifier, Value: "age"},
				{Type: TokenGreater, Value: ">"},
				{Type: TokenNumber, Value: "18"},
				{Type: TokenEOF},
			},
		},
		{
			input: "active == true",
			expected: []Token{
				{Type: TokenIdentifier, Value: "active"},
				{Type: TokenEqual, Value: "=="},
				{Type: TokenBoolean, Value: "TRUE"},
				{Type: TokenEOF},
			},
		},
		{
			input: "country == 'US' AND age > 18",
			expected: []Token{
				{Type: TokenIdentifier, Value: "country"},
				{Type: TokenEqual, Value: "=="},
				{Type: TokenString, Value: "US"},
				{Type: TokenAnd, Value: "AND"},
				{Type: TokenIdentifier, Value: "age"},
				{Type: TokenGreater, Value: ">"},
				{Type: TokenNumber, Value: "18"},
				{Type: TokenEOF},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			for i, expected := range tt.expected {
				token := lexer.NextToken()
				if token.Type != expected.Type {
					t.Errorf("token %d: expected type %s, got %s", i, expected.Type, token.Type)
				}
				if token.Value != expected.Value {
					t.Errorf("token %d: expected value %s, got %s", i, expected.Value, token.Value)
				}
			}
		})
	}
}

func TestParseRule(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "simple equality",
			input:   "country == 'US'",
			wantErr: false,
		},
		{
			name:    "simple inequality",
			input:   "country != 'US'",
			wantErr: false,
		},
		{
			name:    "numeric comparison",
			input:   "age > 18",
			wantErr: false,
		},
		{
			name:    "numeric greater or equal",
			input:   "age >= 21",
			wantErr: false,
		},
		{
			name:    "numeric less",
			input:   "age < 65",
			wantErr: false,
		},
		{
			name:    "numeric less or equal",
			input:   "age <= 65",
			wantErr: false,
		},
		{
			name:    "boolean comparison",
			input:   "active == true",
			wantErr: false,
		},
		{
			name:    "regex match",
			input:   `email ~= '^[a-z]+@example\.com$'`,
			wantErr: false,
		},
		{
			name:    "AND expression",
			input:   "country == 'US' AND age > 18",
			wantErr: false,
		},
		{
			name:    "OR expression",
			input:   "tier == 'premium' OR age > 65",
			wantErr: false,
		},
		{
			name:    "NOT expression",
			input:   "NOT country == 'US'",
			wantErr: false,
		},
		{
			name:    "complex expression with parentheses",
			input:   "(country == 'US' AND age > 18) OR tier == 'premium'",
			wantErr: false,
		},
		{
			name:    "nested parentheses",
			input:   "((country == 'US') AND (age > 18))",
			wantErr: false,
		},
		{
			name:    "multiple ANDs",
			input:   "a == 1 AND b == 2 AND c == 3",
			wantErr: false,
		},
		{
			name:    "multiple ORs",
			input:   "a == 1 OR b == 2 OR c == 3",
			wantErr: false,
		},
		{
			name:    "mixed AND and OR",
			input:   "a == 1 AND b == 2 OR c == 3",
			wantErr: false,
		},
		{
			name:    "NOT with AND",
			input:   "NOT a == 1 AND b == 2",
			wantErr: false,
		},
		{
			name:    "double NOT",
			input:   "NOT NOT a == 1",
			wantErr: false,
		},
		{
			name:    "date comparison",
			input:   "signupDate > '2020-01-01'",
			wantErr: false,
		},
		{
			name:    "negative number",
			input:   "balance >= -100",
			wantErr: false,
		},
		{
			name:    "decimal number",
			input:   "rate > 0.5",
			wantErr: false,
		},
		{
			name:    "scientific notation",
			input:   "value > 1e10",
			wantErr: false,
		},
		{
			name:    "double quoted string",
			input:   `name == "John"`,
			wantErr: false,
		},
		{
			name:      "empty string",
			input:     "",
			wantErr:   true,
			errSubstr: "empty condition",
		},
		{
			name:      "whitespace only",
			input:     "   ",
			wantErr:   true,
			errSubstr: "empty condition",
		},
		{
			name:      "missing operator",
			input:     "country 'US'",
			wantErr:   true,
			errSubstr: "comparison operator",
		},
		{
			name:      "missing value",
			input:     "country ==",
			wantErr:   true,
			errSubstr: "literal value",
		},
		{
			name:      "unterminated string",
			input:     "country == 'US",
			wantErr:   true,
			errSubstr: "Error",
		},
		{
			name:      "missing closing paren",
			input:     "(country == 'US'",
			wantErr:   true,
			errSubstr: "')'",
		},
		{
			name:      "missing opening paren",
			input:     "country == 'US')",
			wantErr:   true,
			errSubstr: "unexpected",
		},
		{
			name:      "invalid operator",
			input:     "country = 'US'",
			wantErr:   true,
			errSubstr: "Error",
		},
		{
			name:      "identifier starting with number",
			input:     "1country == 'US'",
			wantErr:   true,
			errSubstr: "unexpected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule, err := ParseRule(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errSubstr)
				} else if tt.errSubstr != "" && !containsString(err.Error(), tt.errSubstr) {
					t.Errorf("expected error containing %q, got %q", tt.errSubstr, err.Error())
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if rule == nil {
				t.Error("expected rule, got nil")
				return
			}
			if rule.AST == nil {
				t.Error("expected AST, got nil")
				return
			}
			// Validate the parsed rule
			if err := rule.Validate(); err != nil {
				t.Errorf("validation failed: %v", err)
			}
		})
	}
}

func TestRuleEvaluation(t *testing.T) {
	tests := []struct {
		name       string
		condition  string
		attributes map[string]interface{}
		expected   bool
		wantErr    bool
	}{
		{
			name:       "simple string equality - true",
			condition:  "country == 'US'",
			attributes: map[string]interface{}{"country": "US"},
			expected:   true,
		},
		{
			name:       "simple string equality - false",
			condition:  "country == 'US'",
			attributes: map[string]interface{}{"country": "UK"},
			expected:   false,
		},
		{
			name:       "string inequality",
			condition:  "country != 'US'",
			attributes: map[string]interface{}{"country": "UK"},
			expected:   true,
		},
		{
			name:       "numeric greater than - true",
			condition:  "age > 18",
			attributes: map[string]interface{}{"age": 25},
			expected:   true,
		},
		{
			name:       "numeric greater than - false",
			condition:  "age > 18",
			attributes: map[string]interface{}{"age": 15},
			expected:   false,
		},
		{
			name:       "numeric less than or equal",
			condition:  "age <= 65",
			attributes: map[string]interface{}{"age": 65},
			expected:   true,
		},
		{
			name:       "numeric greater than or equal",
			condition:  "age >= 21",
			attributes: map[string]interface{}{"age": 21},
			expected:   true,
		},
		{
			name:       "boolean equality - true",
			condition:  "active == true",
			attributes: map[string]interface{}{"active": true},
			expected:   true,
		},
		{
			name:       "boolean equality - false",
			condition:  "active == true",
			attributes: map[string]interface{}{"active": false},
			expected:   false,
		},
		{
			name:       "AND expression - both true",
			condition:  "country == 'US' AND age > 18",
			attributes: map[string]interface{}{"country": "US", "age": 25},
			expected:   true,
		},
		{
			name:       "AND expression - one false",
			condition:  "country == 'US' AND age > 18",
			attributes: map[string]interface{}{"country": "US", "age": 15},
			expected:   false,
		},
		{
			name:       "OR expression - both false",
			condition:  "tier == 'premium' OR age > 65",
			attributes: map[string]interface{}{"tier": "basic", "age": 30},
			expected:   false,
		},
		{
			name:       "OR expression - one true",
			condition:  "tier == 'premium' OR age > 65",
			attributes: map[string]interface{}{"tier": "premium", "age": 30},
			expected:   true,
		},
		{
			name:       "NOT expression",
			condition:  "NOT country == 'US'",
			attributes: map[string]interface{}{"country": "UK"},
			expected:   true,
		},
		{
			name:       "complex expression with parentheses",
			condition:  "(country == 'US' AND age > 18) OR tier == 'premium'",
			attributes: map[string]interface{}{"country": "US", "age": 25, "tier": "basic"},
			expected:   true,
		},
		{
			name:       "complex expression - false",
			condition:  "(country == 'US' AND age > 18) OR tier == 'premium'",
			attributes: map[string]interface{}{"country": "UK", "age": 30, "tier": "basic"},
			expected:   false,
		},
		{
			name:       "date comparison - after",
			condition:  "signupDate > '2020-01-01'",
			attributes: map[string]interface{}{"signupDate": "2023-01-01"},
			expected:   true,
		},
		{
			name:       "date comparison - before",
			condition:  "signupDate < '2020-01-01'",
			attributes: map[string]interface{}{"signupDate": "2019-01-01"},
			expected:   true,
		},
		{
			name:       "regex match - true",
			condition:  `email ~= '^[a-z]+@example\.com$'`,
			attributes: map[string]interface{}{"email": "test@example.com"},
			expected:   true,
		},
		{
			name:       "regex match - false",
			condition:  `email ~= '^[a-z]+@example\.com$'`,
			attributes: map[string]interface{}{"email": "test@other.com"},
			expected:   false,
		},
		{
			name:       "float comparison",
			condition:  "rate > 0.5",
			attributes: map[string]interface{}{"rate": 0.75},
			expected:   true,
		},
		{
			name:       "negative number comparison",
			condition:  "balance >= -100",
			attributes: map[string]interface{}{"balance": -50},
			expected:   true,
		},
		{
			name:       "missing attribute",
			condition:  "country == 'US'",
			attributes: map[string]interface{}{"age": 25},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule, err := ParseRule(tt.condition)
			if err != nil {
				t.Fatalf("failed to parse rule: %v", err)
			}

			result, err := rule.Evaluate(tt.attributes)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("evaluation error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestASTString(t *testing.T) {
	tests := []struct {
		condition string
		expected  string
	}{
		{
			condition: "country == 'US'",
			expected:  "(country == US)",
		},
		{
			condition: "country == 'US' AND age > 18",
			expected:  "((country == US) AND (age > 18))",
		},
		{
			condition: "NOT country == 'US'",
			expected:  "(NOT (country == US))",
		},
		{
			condition: "(a == 1 OR b == 2) AND c == 3",
			expected:  "(((a == 1) OR (b == 2)) AND (c == 3))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.condition, func(t *testing.T) {
			rule, err := ParseRule(tt.condition)
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}
			if rule.AST.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, rule.AST.String())
			}
		})
	}
}

func TestOperatorPrecedence(t *testing.T) {
	// AND should have higher precedence than OR
	// a == 1 OR b == 2 AND c == 3 should be parsed as:
	// a == 1 OR (b == 2 AND c == 3)
	condition := "a == 1 OR b == 2 AND c == 3"
	rule, err := ParseRule(condition)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Test with attributes where:
	// a = 0 (false), b = 2 (true for b==2), c = 3 (true for c==3)
	// Result should be: false OR (true AND true) = true
	attrs := map[string]interface{}{
		"a": 0,
		"b": 2,
		"c": 3,
	}
	result, err := rule.Evaluate(attrs)
	if err != nil {
		t.Fatalf("evaluation error: %v", err)
	}
	if !result {
		t.Errorf("expected true, got false")
	}

	// Test with attributes where:
	// a = 0 (false), b = 2 (true for b==2), c = 4 (false for c==3)
	// Result should be: false OR (true AND false) = false
	attrs = map[string]interface{}{
		"a": 0,
		"b": 2,
		"c": 4,
	}
	result, err = rule.Evaluate(attrs)
	if err != nil {
		t.Fatalf("evaluation error: %v", err)
	}
	if result {
		t.Errorf("expected false, got true")
	}
}

func TestNOTPrecedence(t *testing.T) {
	// NOT should have higher precedence than AND
	// NOT a == 1 AND b == 2 should be parsed as:
	// (NOT (a == 1)) AND (b == 2)
	condition := "NOT a == 1 AND b == 2"
	rule, err := ParseRule(condition)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Test with a = 0 (NOT false = true), b = 2 (true)
	// Result should be: true AND true = true
	attrs := map[string]interface{}{
		"a": 0,
		"b": 2,
	}
	result, err := rule.Evaluate(attrs)
	if err != nil {
		t.Fatalf("evaluation error: %v", err)
	}
	if !result {
		t.Errorf("expected true, got false")
	}
}

func TestTypeCoercion(t *testing.T) {
	tests := []struct {
		name       string
		condition  string
		attributes map[string]interface{}
		expected   bool
	}{
		{
			name:       "int to float comparison",
			condition:  "value > 10.5",
			attributes: map[string]interface{}{"value": 15}, // int
			expected:   true,
		},
		{
			name:       "string number comparison",
			condition:  "value == 42",
			attributes: map[string]interface{}{"value": "42"}, // string
			expected:   true,
		},
		{
			name:       "float to int comparison",
			condition:  "value == 10",
			attributes: map[string]interface{}{"value": 10.0}, // float64
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule, err := ParseRule(tt.condition)
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}
			result, err := rule.Evaluate(tt.attributes)
			if err != nil {
				t.Fatalf("evaluation error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestNestedJSONEvaluation(t *testing.T) {
	tests := []struct {
		name       string
		condition  string
		attributes map[string]interface{}
		expected   bool
		wantErr    bool
	}{
		{
			name:      "nested field - simple greater than",
			condition: "savedTagsCounts.fitness > 3",
			attributes: map[string]interface{}{
				"savedTagsCounts": map[string]interface{}{
					"entertainment": 12,
					"fitness":       5,
					"relationships": 8,
				},
			},
			expected: true,
		},
		{
			name:      "nested field - simple less than",
			condition: "savedTagsCounts.fitness < 10",
			attributes: map[string]interface{}{
				"savedTagsCounts": map[string]interface{}{
					"entertainment": 12,
					"fitness":       5,
					"relationships": 8,
				},
			},
			expected: true,
		},
		{
			name:      "nested field - equality",
			condition: "savedTagsCounts.fitness == 5",
			attributes: map[string]interface{}{
				"savedTagsCounts": map[string]interface{}{
					"entertainment": 12,
					"fitness":       5,
					"relationships": 8,
				},
			},
			expected: true,
		},
		{
			name:      "nested field combined with top-level field",
			condition: "savedTagsCounts.fitness > 3 AND country == 'US'",
			attributes: map[string]interface{}{
				"country": "US",
				"savedTagsCounts": map[string]interface{}{
					"entertainment": 12,
					"fitness":       5,
					"relationships": 8,
				},
			},
			expected: true,
		},
		{
			name:      "nested field combined with top-level field - false case",
			condition: "savedTagsCounts.fitness > 10 AND country == 'US'",
			attributes: map[string]interface{}{
				"country": "US",
				"savedTagsCounts": map[string]interface{}{
					"entertainment": 12,
					"fitness":       5,
					"relationships": 8,
				},
			},
			expected: false,
		},
		{
			name:      "nested field with OR condition",
			condition: "savedTagsCounts.fitness > 10 OR savedTagsCounts.entertainment > 10",
			attributes: map[string]interface{}{
				"savedTagsCounts": map[string]interface{}{
					"entertainment": 12,
					"fitness":       5,
					"relationships": 8,
				},
			},
			expected: true,
		},
		{
			name:      "deeply nested field - two levels",
			condition: "user.profile.age > 18",
			attributes: map[string]interface{}{
				"user": map[string]interface{}{
					"profile": map[string]interface{}{
						"age":    25,
						"gender": "male",
					},
					"id": 123,
				},
			},
			expected: true,
		},
		{
			name:      "deeply nested field - three levels",
			condition: "user.profile.settings.notifications == true",
			attributes: map[string]interface{}{
				"user": map[string]interface{}{
					"profile": map[string]interface{}{
						"settings": map[string]interface{}{
							"notifications": true,
							"darkMode":      false,
						},
						"age": 25,
					},
					"id": 123,
				},
			},
			expected: true,
		},
		{
			name:      "deeply nested string comparison",
			condition: "user.address.country == 'US'",
			attributes: map[string]interface{}{
				"user": map[string]interface{}{
					"address": map[string]interface{}{
						"country": "US",
						"city":    "New York",
					},
				},
			},
			expected: true,
		},
		{
			name:      "multiple nested fields in same expression",
			condition: "savedTagsCounts.fitness > 3 AND savedTagsCounts.entertainment > 10",
			attributes: map[string]interface{}{
				"savedTagsCounts": map[string]interface{}{
					"entertainment": 12,
					"fitness":       5,
					"relationships": 8,
				},
			},
			expected: true,
		},
		{
			name:      "nested field not found",
			condition: "savedTagsCounts.nonexistent > 3",
			attributes: map[string]interface{}{
				"savedTagsCounts": map[string]interface{}{
					"fitness": 5,
				},
			},
			wantErr: true,
		},
		{
			name:      "parent path not found",
			condition: "nonexistent.field > 3",
			attributes: map[string]interface{}{
				"savedTagsCounts": map[string]interface{}{
					"fitness": 5,
				},
			},
			wantErr: true,
		},
		{
			name:      "nested field with NOT",
			condition: "NOT savedTagsCounts.fitness > 10",
			attributes: map[string]interface{}{
				"savedTagsCounts": map[string]interface{}{
					"fitness": 5,
				},
			},
			expected: true,
		},
		{
			name:      "complex nested expression with parentheses",
			condition: "(savedTagsCounts.fitness > 3 AND country == 'US') OR tier == 'premium'",
			attributes: map[string]interface{}{
				"country": "US",
				"tier":    "basic",
				"savedTagsCounts": map[string]interface{}{
					"fitness": 5,
				},
			},
			expected: true,
		},
		{
			name:      "nested field with regex match",
			condition: `user.email ~= '^[a-z]+@example\.com$'`,
			attributes: map[string]interface{}{
				"user": map[string]interface{}{
					"email": "test@example.com",
					"id":    123,
				},
			},
			expected: true,
		},
		{
			name:      "nested field with date comparison",
			condition: "user.subscription.expiryDate > '2024-01-01'",
			attributes: map[string]interface{}{
				"user": map[string]interface{}{
					"subscription": map[string]interface{}{
						"expiryDate": "2025-06-15",
						"plan":       "premium",
					},
				},
			},
			expected: true,
		},
		{
			name:      "nested boolean field",
			condition: "user.settings.emailVerified == true",
			attributes: map[string]interface{}{
				"user": map[string]interface{}{
					"settings": map[string]interface{}{
						"emailVerified": true,
						"twoFactorAuth": false,
					},
				},
			},
			expected: true,
		},
		{
			name:      "nested field inequality",
			condition: "user.status != 'banned'",
			attributes: map[string]interface{}{
				"user": map[string]interface{}{
					"status": "active",
					"id":     123,
				},
			},
			expected: true,
		},
		{
			name:      "nested field with negative number",
			condition: "account.balance.minimum >= -100",
			attributes: map[string]interface{}{
				"account": map[string]interface{}{
					"balance": map[string]interface{}{
						"minimum": -50,
						"current": 1000,
					},
				},
			},
			expected: true,
		},
		{
			name:      "mixed nested and top-level in complex expression",
			condition: "country == 'US' AND (user.tier == 'premium' OR user.age > 65)",
			attributes: map[string]interface{}{
				"country": "US",
				"user": map[string]interface{}{
					"tier": "basic",
					"age":  70,
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule, err := ParseRule(tt.condition)
			if err != nil {
				t.Fatalf("failed to parse rule: %v", err)
			}

			result, err := rule.Evaluate(tt.attributes)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("evaluation error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestNestedJSONLexer(t *testing.T) {
	tests := []struct {
		input    string
		expected []Token
	}{
		{
			input: "savedTagsCounts.fitness > 3",
			expected: []Token{
				{Type: TokenIdentifier, Value: "savedTagsCounts.fitness"},
				{Type: TokenGreater, Value: ">"},
				{Type: TokenNumber, Value: "3"},
				{Type: TokenEOF},
			},
		},
		{
			input: "user.profile.age > 18",
			expected: []Token{
				{Type: TokenIdentifier, Value: "user.profile.age"},
				{Type: TokenGreater, Value: ">"},
				{Type: TokenNumber, Value: "18"},
				{Type: TokenEOF},
			},
		},
		{
			input: "a.b.c.d == 'test'",
			expected: []Token{
				{Type: TokenIdentifier, Value: "a.b.c.d"},
				{Type: TokenEqual, Value: "=="},
				{Type: TokenString, Value: "test"},
				{Type: TokenEOF},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			for i, expected := range tt.expected {
				token := lexer.NextToken()
				if token.Type != expected.Type {
					t.Errorf("token %d: expected type %s, got %s", i, expected.Type, token.Type)
				}
				if token.Value != expected.Value {
					t.Errorf("token %d: expected value %s, got %s", i, expected.Value, token.Value)
				}
			}
		})
	}
}

func TestGetNestedValue(t *testing.T) {
	tests := []struct {
		name       string
		attributes map[string]interface{}
		path       string
		expected   interface{}
		exists     bool
	}{
		{
			name: "simple field",
			attributes: map[string]interface{}{
				"country": "US",
			},
			path:     "country",
			expected: "US",
			exists:   true,
		},
		{
			name: "nested field one level",
			attributes: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "John",
				},
			},
			path:     "user.name",
			expected: "John",
			exists:   true,
		},
		{
			name: "nested field two levels",
			attributes: map[string]interface{}{
				"user": map[string]interface{}{
					"profile": map[string]interface{}{
						"age": 25,
					},
				},
			},
			path:     "user.profile.age",
			expected: 25,
			exists:   true,
		},
		{
			name: "field not found",
			attributes: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "John",
				},
			},
			path:   "user.nonexistent",
			exists: false,
		},
		{
			name: "parent not found",
			attributes: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "John",
				},
			},
			path:   "nonexistent.field",
			exists: false,
		},
		{
			name: "nested numeric value",
			attributes: map[string]interface{}{
				"stats": map[string]interface{}{
					"count": 42,
				},
			},
			path:     "stats.count",
			expected: 42,
			exists:   true,
		},
		{
			name: "nested boolean value",
			attributes: map[string]interface{}{
				"settings": map[string]interface{}{
					"enabled": true,
				},
			},
			path:     "settings.enabled",
			expected: true,
			exists:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, exists := getNestedValue(tt.attributes, tt.path)
			if exists != tt.exists {
				t.Errorf("expected exists=%v, got exists=%v", tt.exists, exists)
				return
			}
			if tt.exists && value != tt.expected {
				t.Errorf("expected value=%v, got value=%v", tt.expected, value)
			}
		})
	}
}

// Helper function
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
