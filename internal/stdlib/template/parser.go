// ===========================pylearn/internal/stdlib/template/parser.go start here===========================
// internal/stdlib/template/parser.go
package template

import (
	"fmt"
	"strconv"
)

// Precedence levels for Pratt parser
const (
	_ int = iota
	LOWEST
	EQUALS  // ==, !=, >, <, >=, <=
	SUM     // +, -
	PRODUCT // *, /
	PREFIX  // -X
	CALL    // myfunction(X)
	INDEX   // array[X]
	DOT     // object.attribute
	PIPE    // |
)

// Maps token types to their precedence
var precedences = map[TokenType]int{
	TokenIdent:  EQUALS, // Placeholder for operators lexed as Ident
	TokenPipe:   PIPE,
	TokenDot:    DOT,
	TokenLparen: CALL,
}

type (
	prefixParseFn func() (Node, error)
	infixParseFn  func(Node) (Node, error)
)

// Parser for template tokens.
type Parser struct {
	tokens []Token
	pos    int

	prefixParseFns map[TokenType]prefixParseFn
	infixParseFns  map[TokenType]infixParseFn
}

func NewParser(tokens []Token) *Parser {
	p := &Parser{tokens: tokens}

	p.prefixParseFns = make(map[TokenType]prefixParseFn)
	p.registerPrefix(TokenIdent, p.parseIdentifier)
	p.registerPrefix(TokenInteger, p.parseInteger)
	p.registerPrefix(TokenText, p.parseString) // String literals
	p.registerPrefix(TokenLparen, p.parseGroupedExpression)

	p.infixParseFns = make(map[TokenType]infixParseFn)
	p.registerInfix(TokenIdent, p.parseInfixExpression) // For operators like >
	p.registerInfix(TokenDot, p.parseDotAccess)
	p.registerInfix(TokenPipe, p.parseFilter)
	p.registerInfix(TokenLparen, p.parseCall)

	return p
}

func (p *Parser) registerPrefix(tokenType TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}
func (p *Parser) registerInfix(tokenType TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) current() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.pos]
}
func (p *Parser) peek() Token {
	if p.pos+1 >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.pos+1]
}
func (p *Parser) next() { p.pos++ }
func (p *Parser) expect(typ TokenType) (Token, error) {
	if p.current().Type == typ {
		tok := p.current()
		p.next()
		return tok, nil
	}
	return Token{}, fmt.Errorf("expected token %v, got %v at line %d", typ, p.current().Type, p.current().Line)
}
func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peek().Type]; ok {
		return p
	}
	// For binary operators lexed as identifiers
	if p.peek().Type == TokenIdent {
		switch p.peek().Val {
		case ">", "<", "==", "!=", ">=", "<=":
			return EQUALS
		}
	}
	return LOWEST
}

func (p *Parser) Parse() (Node, error) { return p.parseNodeList(nil) }

// --- Expression Parsing (Pratt Parser Implementation) ---
func (p *Parser) parseExpression(precedence int) (Node, error) {
	prefix := p.prefixParseFns[p.current().Type]
	if prefix == nil {
		return nil, fmt.Errorf("no prefix parse function for '%s' found", p.current().Val)
	}
	leftExp, err := prefix()
	if err != nil {
		return nil, err
	}
	for p.current().Type != TokenEOF && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peek().Type]
		if infix == nil {
			return leftExp, nil
		}
		p.next()
		leftExp, err = infix(leftExp)
		if err != nil {
			return nil, err
		}
	}
	return leftExp, nil
}
func (p *Parser) parseIdentifier() (Node, error) {
	defer p.next()
	return &IdentNode{Token: p.current(), Name: p.current().Val}, nil
}
func (p *Parser) parseInteger() (Node, error) {
	defer p.next()
	val, _ := strconv.Atoi(p.current().Val)
	return &IntegerNode{Token: p.current(), Value: val}, nil
}
func (p *Parser) parseString() (Node, error) {
	defer p.next()
	return &FmtStr{Token: p.current(), Text: p.current().Val}, nil
}
func (p *Parser) parseGroupedExpression() (Node, error) {
	p.next() // consume '('
	exp, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(TokenRparen); err != nil {
		return nil, err
	}
	return exp, nil
}
func (p *Parser) parseInfixExpression(left Node) (Node, error) {
	tok := p.current()
	precedence := precedences[tok.Type]
	if tok.Type == TokenIdent {
		precedence = EQUALS // Operator precedence
	}
	p.next()
	right, err := p.parseExpression(precedence)
	if err != nil {
		return nil, err
	}
	return &BinaryExpressionNode{Token: tok, Left: left, Operator: tok.Val, Right: right}, nil
}
func (p *Parser) parseDotAccess(left Node) (Node, error) {
	tok := p.current()
	p.next() // consume '.'
	attr, err := p.expect(TokenIdent)
	if err != nil {
		return nil, err
	}
	return &DotAccessNode{Token: tok, Base: left, Attr: attr.Val}, nil
}
func (p *Parser) parseFilter(left Node) (Node, error) {
	p.next() // consume '|'
	filterName, err := p.expect(TokenIdent)
	if err != nil {
		return nil, err
	}
	return &FilterNode{Token: filterName, Expression: left, Name: filterName.Val}, nil
}
func (p *Parser) parseCall(function Node) (Node, error) {
	call := &CallNode{Token: p.current(), Function: function}
	p.next() // consume '('
	var args []Node
	if p.current().Type != TokenRparen {
		for {
			arg, err := p.parseExpression(LOWEST)
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
			if p.current().Type != TokenComma {
				break
			}
			p.next() // consume ','
		}
	}
	if _, err := p.expect(TokenRparen); err != nil {
		return nil, err
	}
	call.Args = args
	return call, nil
}

// --- Statement Parsing ---
func (p *Parser) parseNodeList(endTag *string) (NodeList, error) {
	var nodes NodeList
	for p.current().Type != TokenEOF {
		switch p.current().Type {
		case TokenText:
			nodes = append(nodes, &TextNode{Token: p.current()})
			p.next()
		case TokenVarStart:
			node, err := p.parseVariable()
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, node)
		case TokenBlockStart:
			tag := p.peek()
			if endTag != nil && (*endTag == tag.Val || (tag.Val == "elif" || tag.Val == "else") && *endTag == "endif") {
				return nodes, nil
			}
			var node Node
			var err error
			switch tag.Val {
			case "if":
				node, err = p.parseIf()
			case "for":
				node, err = p.parseFor()
			case "extends":
				node, err = p.parseExtends()
			case "block":
				node, err = p.parseBlock()
			case "include":
				node, err = p.parseInclude()
			case "raw":
				node, err = p.parseRaw()
			default:
				return nil, fmt.Errorf("unknown tag '%s'", tag.Val)
			}
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, node)
		default:
			return nil, fmt.Errorf("unexpected token %v", p.current())
		}
	}
	if endTag != nil {
		return nil, fmt.Errorf("unexpected EOF, expected 'end%s'", *endTag)
	}
	return nodes, nil
}
func (p *Parser) parseSubParser(start, end TokenType) (*Parser, error) {
	if _, err := p.expect(start); err != nil {
		return nil, err
	}
	var subTokens []Token
	for p.current().Type != end && p.current().Type != TokenEOF {
		subTokens = append(subTokens, p.current())
		p.next()
	}
	if _, err := p.expect(end); err != nil {
		return nil, err
	}
	return NewParser(subTokens), nil
}
func (p *Parser) parseVariable() (Node, error) {
	startTok := p.current()
	sub, err := p.parseSubParser(TokenVarStart, TokenVarEnd)
	if err != nil {
		return nil, err
	}
	expr, err := sub.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}
	return &VariableNode{Token: startTok, Expression: expr}, nil
}
func (p *Parser) parseIf() (*IfNode, error) {
	p.next()
	start, _ := p.expect(TokenIdent)
	sub, err := p.parseSubParser(TokenIdent, TokenBlockEnd)
	if err != nil {
		return nil, err
	}
	cond, err := sub.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}

	endIf := "endif"
	body, err := p.parseNodeList(&endIf)
	if err != nil {
		return nil, err
	}

	node := &IfNode{Token: start, Condition: cond, Body: body}
	current := node

	for p.peek().Val == "elif" {
		p.next()
		p.next()
		sub, err := p.parseSubParser(TokenIdent, TokenBlockEnd)
		if err != nil {
			return nil, err
		}
		elifCond, err := sub.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}
		elifBody, err := p.parseNodeList(&endIf)
		if err != nil {
			return nil, err
		}
		newIf := &IfNode{Token: start, Condition: elifCond, Body: elifBody}
		current.Else = newIf
		current = newIf
	}
	if p.peek().Val == "else" {
		p.next()
		p.next()
		if _, err := p.expect(TokenBlockEnd); err != nil {
			return nil, err
		}
		elseBody, err := p.parseNodeList(&endIf)
		if err != nil {
			return nil, err
		}
		current.Else = elseBody
	}
	if _, err := p.expect(TokenBlockStart); err != nil {
		return nil, err
	}
	if _, err := p.expect(TokenIdent); err != nil {
		return nil, err
	}
	if _, err := p.expect(TokenBlockEnd); err != nil {
		return nil, err
	}
	return node, nil
}
func (p *Parser) parseFor() (*ForNode, error) {
	p.next()
	start, _ := p.expect(TokenIdent)
	sub, err := p.parseSubParser(TokenIdent, TokenBlockEnd)
	if err != nil {
		return nil, err
	}
	loopVar, err := sub.expect(TokenIdent)
	if err != nil {
		return nil, err
	}
	if _, err := sub.expect(TokenIdent); err != nil {
		return nil, err
	} // 'in'
	iterable, err := sub.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}
	endFor := "endfor"
	body, err := p.parseNodeList(&endFor)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(TokenBlockStart); err != nil {
		return nil, err
	}
	if _, err := p.expect(TokenIdent); err != nil {
		return nil, err
	}
	if _, err := p.expect(TokenBlockEnd); err != nil {
		return nil, err
	}
	return &ForNode{Token: start, LoopVar: loopVar.Val, Iterable: iterable, Body: body}, nil
}
func (p *Parser) parseExtends() (*ExtendsNode, error) {
	p.next()
	start, _ := p.expect(TokenIdent)
	sub, err := p.parseSubParser(TokenIdent, TokenBlockEnd)
	if err != nil {
		return nil, err
	}
	parentExpr, err := sub.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}
	parent, ok := parentExpr.(*FmtStr)
	if !ok {
		return nil, fmt.Errorf("extends tag requires a string literal")
	}
	return &ExtendsNode{Token: start, Parent: *parent}, nil
}
func (p *Parser) parseBlock() (*BlockNode, error) {
	p.next()
	start, _ := p.expect(TokenIdent)
	sub, err := p.parseSubParser(TokenIdent, TokenBlockEnd)
	if err != nil {
		return nil, err
	}
	name, err := sub.expect(TokenIdent)
	if err != nil {
		return nil, err
	}
	endBlock := "endblock"
	body, err := p.parseNodeList(&endBlock)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(TokenBlockStart); err != nil {
		return nil, err
	}
	if _, err := p.expect(TokenIdent); err != nil {
		return nil, err
	}
	if _, err := p.expect(TokenBlockEnd); err != nil {
		return nil, err
	}
	return &BlockNode{Token: start, Name: name.Val, Body: body}, nil
}
func (p *Parser) parseInclude() (*IncludeNode, error) {
	p.next()
	start, _ := p.expect(TokenIdent)
	sub, err := p.parseSubParser(TokenIdent, TokenBlockEnd)
	if err != nil {
		return nil, err
	}
	templateExpr, err := sub.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}
	template, ok := templateExpr.(*FmtStr)
	if !ok {
		return nil, fmt.Errorf("include tag requires a string literal")
	}
	return &IncludeNode{Token: start, Template: *template}, nil
}
func (p *Parser) parseRaw() (*RawNode, error) {
	p.next()
	_, err := p.expect(TokenIdent)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(TokenBlockEnd); err != nil {
		return nil, err
	}
	text, err := p.expect(TokenText)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(TokenBlockStart); err != nil {
		return nil, err
	}
	if _, err := p.expect(TokenIdent); err != nil {
		return nil, err
	}
	if _, err := p.expect(TokenBlockEnd); err != nil {
		return nil, err
	}
	return &RawNode{TextNode{Token: text}}, nil
}

// ===========================pylearn/internal/stdlib/template/parser.go ends here===========================
