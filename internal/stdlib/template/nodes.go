// ===========================pylearn/internal/stdlib/template/nodes.go start here===========================
// internal/stdlib/template/nodes.go
package template

import (
	"fmt"
	"strings"
)

// Node is the interface for all AST nodes in a template.
type Node interface {
	String() string
	Position() Token // For error reporting
}

// NodeList holds a sequence of nodes.
type NodeList []Node

func (nl NodeList) String() string {
	var sb strings.Builder
	for _, n := range nl {
		sb.WriteString(n.String())
	}
	return sb.String()
}
func (nl NodeList) Position() Token {
	if len(nl) > 0 {
		return nl[0].Position()
	}
	return Token{} // Return zero-value token for empty list
}

// TextNode represents plain text in the template.
type TextNode struct {
	Token Token
}

func (t *TextNode) String() string      { return t.Token.Val }
func (t *TextNode) Position() Token { return t.Token }

// RawNode represents a {% raw %} block.
type RawNode struct {
	TextNode
}

// VariableNode represents a {{ ... }} block for printing an expression.
type VariableNode struct {
	Token      Token
	Expression Node
}

func (v *VariableNode) String() string      { return "{{ " + v.Expression.String() + " }}" }
func (v *VariableNode) Position() Token { return v.Token }

// IfNode represents an {% if ... %} ... {% else %} ... {% endif %} block.
type IfNode struct {
	Token     Token
	Condition Node
	Body      Node
	Else      Node // Can be another IfNode for elif, or a NodeList for else
}

func (i *IfNode) String() string      { return "{% if " + i.Condition.String() + " %}" }
func (i *IfNode) Position() Token { return i.Token }

// ForNode represents an {% for ... in ... %} ... {% endfor %} loop.
type ForNode struct {
	Token    Token
	LoopVar  string
	Iterable Node
	Body     Node
}

func (f *ForNode) String() string      { return "{% for " + f.LoopVar + " in " + f.Iterable.String() + " %}" }
func (f *ForNode) Position() Token { return f.Token }

// ExtendsNode represents an {% extends "..." %} tag.
type ExtendsNode struct {
	Token    Token
	Parent FmtStr
}

func (e *ExtendsNode) String() string      { return "{% extends " + e.Parent.String() + " %}" }
func (e *ExtendsNode) Position() Token { return e.Token }

// BlockNode represents a {% block name %} ... {% endblock %} tag.
type BlockNode struct {
	Token Token
	Name  string
	Body  Node
}

func (b *BlockNode) String() string      { return "{% block " + b.Name + " %}" }
func (b *BlockNode) Position() Token { return b.Token }

// IncludeNode represents an {% include "..." %} tag.
type IncludeNode struct {
	Token    Token
	Template FmtStr
}

func (i *IncludeNode) String() string      { return "{% include " + i.Template.String() + " %}" }
func (i *IncludeNode) Position() Token { return i.Token }

// === Expression Nodes ===

// IdentNode represents an identifier (variable name).
type IdentNode struct {
	Token Token
	Name  string
}

func (i *IdentNode) String() string      { return i.Name }
func (i *IdentNode) Position() Token { return i.Token }

// FmtStr represents a string literal.
type FmtStr struct {
	Token Token
	Text  string
}

func (s *FmtStr) String() string      { return fmt.Sprintf("%q", s.Text) }
func (s *FmtStr) Position() Token { return s.Token }

// IntegerNode represents an integer literal.
type IntegerNode struct {
	Token Token
	Value int
}

func (i *IntegerNode) String() string      { return i.Token.Val }
func (i *IntegerNode) Position() Token { return i.Token }

// DotAccessNode represents attribute access (e.g., user.name).
type DotAccessNode struct {
	Token Token
	Base  Node
	Attr  string
}

func (d *DotAccessNode) String() string      { return d.Base.String() + "." + d.Attr }
func (d *DotAccessNode) Position() Token { return d.Token }

// FilterNode represents a piped filter operation (e.g., name | upper).
type FilterNode struct {
	Token      Token // The '|' token
	Expression Node
	Name       string
	Args       []Node
}

func (f *FilterNode) String() string      { return f.Expression.String() + " | " + f.Name }
func (f *FilterNode) Position() Token { return f.Token }

// CallNode represents a function call (e.g., range(5)).
type CallNode struct {
	Token    Token // The '(' token
	Function Node
	Args     []Node
}

func (c *CallNode) String() string      { return c.Function.String() + "()" }
func (c *CallNode) Position() Token { return c.Token }

// BinaryExpressionNode represents an operation between two expressions.
type BinaryExpressionNode struct {
	Token    Token
	Left     Node
	Operator string
	Right    Node
}

func (b *BinaryExpressionNode) String() string {
	return fmt.Sprintf("(%s %s %s)", b.Left.String(), b.Operator, b.Right.String())
}
func (b *BinaryExpressionNode) Position() Token { return b.Token }

