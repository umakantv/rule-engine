package rule

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// TokenType represents the type of a token
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenError
	// Literals
	TokenIdentifier // field names
	TokenString     // quoted strings
	TokenNumber     // numeric literals
	TokenBoolean    // true/false
	// Comparison operators
	TokenEqual    // ==
	TokenNotEqual // !=
	TokenGreater  // >
	TokenLess     // <
	TokenGreaterEqual // >=
	TokenLessEqual    // <=
	TokenRegexMatch   // ~=
	// Logical operators
	TokenAnd // AND
	TokenOr  // OR
	TokenNot // NOT
	// Grouping
	TokenLeftParen  // (
	TokenRightParen // )
)

// Token represents a lexical token
type Token struct {
	Type  TokenType
	Value string
	Pos   int // position in input string
}

// String returns a string representation of the token
func (t Token) String() string {
	switch t.Type {
	case TokenEOF:
		return "EOF"
	case TokenError:
		return fmt.Sprintf("Error(%s)", t.Value)
	default:
		return fmt.Sprintf("%s(%s)", t.Type.String(), t.Value)
	}
}

// String returns a string representation of TokenType
func (tt TokenType) String() string {
	switch tt {
	case TokenEOF:
		return "EOF"
	case TokenError:
		return "Error"
	case TokenIdentifier:
		return "Identifier"
	case TokenString:
		return "String"
	case TokenNumber:
		return "Number"
	case TokenBoolean:
		return "Boolean"
	case TokenEqual:
		return "=="
	case TokenNotEqual:
		return "!="
	case TokenGreater:
		return ">"
	case TokenLess:
		return "<"
	case TokenGreaterEqual:
		return ">="
	case TokenLessEqual:
		return "<="
	case TokenRegexMatch:
		return "~="
	case TokenAnd:
		return "AND"
	case TokenOr:
		return "OR"
	case TokenNot:
		return "NOT"
	case TokenLeftParen:
		return "("
	case TokenRightParen:
		return ")"
	default:
		return "Unknown"
	}
}

// Lexer tokenizes the input string
type Lexer struct {
	input string
	pos   int
}

// NewLexer creates a new lexer for the given input
func NewLexer(input string) *Lexer {
	return &Lexer{input: input, pos: 0}
}

// peek returns the current character without advancing
func (l *Lexer) peek() byte {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

// next returns the current character and advances position
func (l *Lexer) next() byte {
	ch := l.peek()
	l.pos++
	return ch
}

// skipWhitespace skips any whitespace characters
func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) && isWhitespace(l.peek()) {
		l.pos++
	}
}

// NextToken returns the next token from the input
func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	if l.pos >= len(l.input) {
		return Token{Type: TokenEOF, Pos: l.pos}
	}

	pos := l.pos
	ch := l.peek()

	switch {
	case isLetter(ch) || ch == '_':
		return l.readIdentifier()
	case ch == '"' || ch == '\'':
		return l.readString()
	case isDigit(ch) || (ch == '-' && l.pos+1 < len(l.input) && isDigit(l.input[l.pos+1])):
		return l.readNumber()
	case ch == '=':
		l.next()
		if l.peek() == '=' {
			l.next()
			return Token{Type: TokenEqual, Value: "==", Pos: pos}
		}
		return Token{Type: TokenError, Value: "expected '=' after '='", Pos: pos}
	case ch == '!':
		l.next()
		if l.peek() == '=' {
			l.next()
			return Token{Type: TokenNotEqual, Value: "!=", Pos: pos}
		}
		return Token{Type: TokenError, Value: "expected '=' after '!'", Pos: pos}
	case ch == '>':
		l.next()
		if l.peek() == '=' {
			l.next()
			return Token{Type: TokenGreaterEqual, Value: ">=", Pos: pos}
		}
		return Token{Type: TokenGreater, Value: ">", Pos: pos}
	case ch == '<':
		l.next()
		if l.peek() == '=' {
			l.next()
			return Token{Type: TokenLessEqual, Value: "<=", Pos: pos}
		}
		return Token{Type: TokenLess, Value: "<", Pos: pos}
	case ch == '~':
		l.next()
		if l.peek() == '=' {
			l.next()
			return Token{Type: TokenRegexMatch, Value: "~=", Pos: pos}
		}
		return Token{Type: TokenError, Value: "expected '=' after '~'", Pos: pos}
	case ch == '(':
		l.next()
		return Token{Type: TokenLeftParen, Value: "(", Pos: pos}
	case ch == ')':
		l.next()
		return Token{Type: TokenRightParen, Value: ")", Pos: pos}
	default:
		l.next()
		return Token{Type: TokenError, Value: fmt.Sprintf("unexpected character: %c", ch), Pos: pos}
	}
}

// readIdentifier reads an identifier or keyword
func (l *Lexer) readIdentifier() Token {
	pos := l.pos
	for l.pos < len(l.input) && (isLetter(l.peek()) || isDigit(l.peek()) || l.peek() == '_' || l.peek() == '.') {
		l.next()
	}

	value := l.input[pos:l.pos]
	upper := strings.ToUpper(value)

	// Check for keywords (only for simple identifiers, not dotted paths)
	if !strings.Contains(value, ".") {
		switch upper {
		case "AND":
			return Token{Type: TokenAnd, Value: value, Pos: pos}
		case "OR":
			return Token{Type: TokenOr, Value: value, Pos: pos}
		case "NOT":
			return Token{Type: TokenNot, Value: value, Pos: pos}
		case "TRUE", "FALSE":
			return Token{Type: TokenBoolean, Value: upper, Pos: pos}
		}
	}

	return Token{Type: TokenIdentifier, Value: value, Pos: pos}
}

// readString reads a quoted string
func (l *Lexer) readString() Token {
	pos := l.pos
	quote := l.next() // consume opening quote

	var sb strings.Builder
	for l.pos < len(l.input) {
		ch := l.next()
		if ch == quote {
			return Token{Type: TokenString, Value: sb.String(), Pos: pos}
		}
		if ch == '\\' {
			if l.pos >= len(l.input) {
				return Token{Type: TokenError, Value: "unterminated string escape", Pos: pos}
			}
			escaped := l.next()
			switch escaped {
			case 'n':
				sb.WriteByte('\n')
			case 't':
				sb.WriteByte('\t')
			case 'r':
				sb.WriteByte('\r')
			case '\\':
				sb.WriteByte('\\')
			case '"', '\'':
				sb.WriteByte(escaped)
			default:
				// For regex patterns and other escape sequences, preserve the backslash
				sb.WriteByte('\\')
				sb.WriteByte(escaped)
			}
		} else {
			sb.WriteByte(ch)
		}
	}
	return Token{Type: TokenError, Value: "unterminated string", Pos: pos}
}

// readNumber reads a numeric literal
func (l *Lexer) readNumber() Token {
	pos := l.pos
	var sb strings.Builder

	// Handle negative sign
	if l.peek() == '-' {
		sb.WriteByte(l.next())
	}

	// Read integer part
	for l.pos < len(l.input) && isDigit(l.peek()) {
		sb.WriteByte(l.next())
	}

	// Read decimal part
	if l.peek() == '.' {
		sb.WriteByte(l.next())
		for l.pos < len(l.input) && isDigit(l.peek()) {
			sb.WriteByte(l.next())
		}
	}

	// Read exponent part
	if l.peek() == 'e' || l.peek() == 'E' {
		sb.WriteByte(l.next())
		if l.peek() == '+' || l.peek() == '-' {
			sb.WriteByte(l.next())
		}
		for l.pos < len(l.input) && isDigit(l.peek()) {
			sb.WriteByte(l.next())
		}
	}

	return Token{Type: TokenNumber, Value: sb.String(), Pos: pos}
}

// Helper functions
func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

// AST Node types

// NodeType represents the type of AST node
type NodeType int

const (
	NodeComparison NodeType = iota
	NodeLogical
	NodeNot
	NodeLiteral
	NodeIdentifier
)

// Node represents a node in the AST
type Node struct {
	Type     NodeType
	Value    interface{} // for literals
	Field    string      // for identifiers
	Operator string      // for comparison and logical operators
	Left     *Node       // left operand
	Right    *Node       // right operand
	Operand  *Node       // for NOT operator
}

// String returns a string representation of the node
func (n *Node) String() string {
	switch n.Type {
	case NodeLiteral:
		return fmt.Sprintf("%v", n.Value)
	case NodeIdentifier:
		return n.Field
	case NodeComparison:
		return fmt.Sprintf("(%s %s %s)", n.Left.String(), n.Operator, n.Right.String())
	case NodeLogical:
		return fmt.Sprintf("(%s %s %s)", n.Left.String(), n.Operator, n.Right.String())
	case NodeNot:
		return fmt.Sprintf("(NOT %s)", n.Operand.String())
	default:
		return "Unknown"
	}
}

// Parser parses tokens into an AST
type Parser struct {
	lexer *Lexer
	curr  Token
	peek  Token
}

// NewParser creates a new parser for the given input
func NewParser(input string) *Parser {
	lexer := NewLexer(input)
	p := &Parser{lexer: lexer}
	p.nextToken() // initialize curr
	p.nextToken() // initialize peek
	return p
}

// nextToken advances to the next token
func (p *Parser) nextToken() {
	p.curr = p.peek
	p.peek = p.lexer.NextToken()
}

// currentToken returns the current token
func (p *Parser) currentToken() Token {
	return p.curr
}

// expect checks if the current token is of the expected type
func (p *Parser) expect(t TokenType) error {
	if p.curr.Type != t {
		return fmt.Errorf("expected %s, got %s at position %d", t, p.curr.Type, p.curr.Pos)
	}
	return nil
}

// Parse parses the input and returns an AST
func (p *Parser) Parse() (*Node, error) {
	node, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if p.curr.Type != TokenEOF {
		return nil, fmt.Errorf("unexpected token %s after expression at position %d", p.curr.Type, p.curr.Pos)
	}

	return node, nil
}

// parseExpression parses a logical expression (OR has lowest precedence)
func (p *Parser) parseExpression() (*Node, error) {
	return p.parseOr()
}

// parseOr parses OR expressions
func (p *Parser) parseOr() (*Node, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}

	for p.curr.Type == TokenOr {
		p.nextToken()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &Node{
			Type:     NodeLogical,
			Operator: "OR",
			Left:     left,
			Right:    right,
		}
	}

	return left, nil
}

// parseAnd parses AND expressions
func (p *Parser) parseAnd() (*Node, error) {
	left, err := p.parseNot()
	if err != nil {
		return nil, err
	}

	for p.curr.Type == TokenAnd {
		p.nextToken()
		right, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		left = &Node{
			Type:     NodeLogical,
			Operator: "AND",
			Left:     left,
			Right:    right,
		}
	}

	return left, nil
}

// parseNot parses NOT expressions
func (p *Parser) parseNot() (*Node, error) {
	if p.curr.Type == TokenNot {
		p.nextToken()
		operand, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		return &Node{
			Type:     NodeNot,
			Operator: "NOT",
			Operand:  operand,
		}, nil
	}
	return p.parsePrimary()
}

// parsePrimary parses primary expressions (identifiers, literals, parenthesized expressions)
func (p *Parser) parsePrimary() (*Node, error) {
	switch p.curr.Type {
	case TokenLeftParen:
		p.nextToken()
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if p.curr.Type != TokenRightParen {
			return nil, fmt.Errorf("expected ')' at position %d", p.curr.Pos)
		}
		p.nextToken()
		return expr, nil

	case TokenIdentifier:
		return p.parseComparison()

	case TokenNot:
		return p.parseNot()

	default:
		return nil, fmt.Errorf("unexpected token %s at position %d", p.curr.Type, p.curr.Pos)
	}
}

// parseComparison parses a comparison expression
func (p *Parser) parseComparison() (*Node, error) {
	if p.curr.Type != TokenIdentifier {
		return nil, fmt.Errorf("expected identifier, got %s at position %d", p.curr.Type, p.curr.Pos)
	}

	left := &Node{
		Type:  NodeIdentifier,
		Field: p.curr.Value,
	}
	p.nextToken()

	// Check for comparison operator
	operator := ""
	switch p.curr.Type {
	case TokenEqual, TokenNotEqual, TokenGreater, TokenLess,
		TokenGreaterEqual, TokenLessEqual, TokenRegexMatch:
		operator = p.curr.Value
		p.nextToken()
	default:
		return nil, fmt.Errorf("expected comparison operator, got %s at position %d", p.curr.Type, p.curr.Pos)
	}

	// Parse right operand (literal value)
	right, err := p.parseLiteral()
	if err != nil {
		return nil, err
	}

	return &Node{
		Type:     NodeComparison,
		Operator: operator,
		Left:     left,
		Right:    right,
	}, nil
}

// parseLiteral parses a literal value (string, number, boolean)
func (p *Parser) parseLiteral() (*Node, error) {
	switch p.curr.Type {
	case TokenString:
		node := &Node{
			Type:  NodeLiteral,
			Value: p.curr.Value,
		}
		p.nextToken()
		return node, nil

	case TokenNumber:
		value, err := strconv.ParseFloat(p.curr.Value, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number format: %s at position %d", p.curr.Value, p.curr.Pos)
		}
		node := &Node{
			Type:  NodeLiteral,
			Value: value,
		}
		p.nextToken()
		return node, nil

	case TokenBoolean:
		value := strings.ToUpper(p.curr.Value) == "TRUE"
		node := &Node{
			Type:  NodeLiteral,
			Value: value,
		}
		p.nextToken()
		return node, nil

	case TokenIdentifier:
		// Allow identifiers as values (for referencing other fields or constants)
		node := &Node{
			Type:  NodeLiteral,
			Value: p.curr.Value,
		}
		p.nextToken()
		return node, nil

	default:
		return nil, fmt.Errorf("expected literal value, got %s at position %d", p.curr.Type, p.curr.Pos)
	}
}

// Rule represents a parsed rule with its AST
type Rule struct {
	Condition string
	AST       *Node
}

// ParseRule parses a rule string and returns a Rule object
func ParseRule(condition string) (*Rule, error) {
	if strings.TrimSpace(condition) == "" {
		return nil, errors.New("empty condition string")
	}

	parser := NewParser(condition)
	ast, err := parser.Parse()
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	return &Rule{
		Condition: condition,
		AST:       ast,
	}, nil
}

// Validate checks if the rule is valid
func (r *Rule) Validate() error {
	if r.AST == nil {
		return errors.New("rule has no AST")
	}
	return validateNode(r.AST)
}

// validateNode recursively validates an AST node
func validateNode(node *Node) error {
	if node == nil {
		return errors.New("nil node in AST")
	}

	switch node.Type {
	case NodeLiteral:
		// Literals are always valid
		return nil

	case NodeIdentifier:
		if node.Field == "" {
			return errors.New("identifier has empty field name")
		}
		return nil

	case NodeComparison:
		if node.Left == nil || node.Right == nil {
			return errors.New("comparison node missing operands")
		}
		if !isValidComparisonOperator(node.Operator) {
			return fmt.Errorf("invalid comparison operator: %s", node.Operator)
		}
		if err := validateNode(node.Left); err != nil {
			return err
		}
		return validateNode(node.Right)

	case NodeLogical:
		if node.Left == nil || node.Right == nil {
			return errors.New("logical node missing operands")
		}
		if !isValidLogicalOperator(node.Operator) {
			return fmt.Errorf("invalid logical operator: %s", node.Operator)
		}
		if err := validateNode(node.Left); err != nil {
			return err
		}
		return validateNode(node.Right)

	case NodeNot:
		if node.Operand == nil {
			return errors.New("NOT node missing operand")
		}
		return validateNode(node.Operand)

	default:
		return fmt.Errorf("unknown node type: %d", node.Type)
	}
}

// isValidComparisonOperator checks if the operator is valid for comparisons
func isValidComparisonOperator(op string) bool {
	switch op {
	case "==", "!=", ">", "<", ">=", "<=", "~=":
		return true
	default:
		return false
	}
}

// isValidLogicalOperator checks if the operator is valid for logical operations
func isValidLogicalOperator(op string) bool {
	switch op {
	case "AND", "OR":
		return true
	default:
		return false
	}
}

// String returns the string representation of the rule
func (r *Rule) String() string {
	if r.AST == nil {
		return "Rule{empty}"
	}
	return fmt.Sprintf("Rule{%s}", r.AST.String())
}
