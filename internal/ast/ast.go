package ast

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/deniskipeles/pylearn/internal/constants" // Import the constants package
	"github.com/deniskipeles/pylearn/internal/lexer"     // Import the lexer package to use token definitions
)

// --- Interfaces ---

// Node is the base interface for all AST nodes.
type Node interface {
	TokenLiteral() string // Returns the literal value of the token associated with the node
	String() string       // Returns a string representation of the node (for debugging)
}

// Statement represents a statement node in the AST. Statements do not produce values.
// (This is a marker interface, embedding the Node interface)
type Statement interface {
	Node
	statementNode() // Marker method to distinguish statements
}

// Expression represents an expression node in the AST. Expressions produce values.
// (This is a marker interface, embedding the Node interface)
type Expression interface {
	Node
	expressionNode() // Marker method to distinguish expressions
}

// --- Root Node ---

// Program is the root node of every AST our parser produces.
type Program struct {
	Statements []Statement // Sequence of statements in the program
}

// TokenLiteral returns the literal of the first token if statements exist, otherwise "".
func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return constants.AstProgramTokenLiteralDefault
}

// String returns a string representation of the entire program.
func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
		// Optionally add newline or semicolon representation between statements
		// out.WriteString("\n")
	}
	return out.String()
}

// --- Statements ---

// ExpressionStatement represents a statement consisting solely of an expression.
// Example: `x + 5;` or just `myFunction();`
type ExpressionStatement struct {
	Token      lexer.Token // The first token of the expression (e.g., identifier, literal)
	Expression Expression  // The expression being wrapped
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String() // Often just printing the expression is sufficient
	}
	return constants.AstExpressionStatementStringDefault
}

// DelStatement represents `del target`
type DelStatement struct {
	Token  lexer.Token // The 'del' token
	Target Expression  // The expression to delete (e.g., Identifier, IndexExpression)
}

func (ds *DelStatement) statementNode()       {}
func (ds *DelStatement) TokenLiteral() string { return ds.Token.Literal }
func (ds *DelStatement) String() string {
	var out bytes.Buffer
	out.WriteString(ds.TokenLiteral() + " ")
	if ds.Target != nil {
		out.WriteString(ds.Target.String())
	}
	return out.String()
}

// GlobalStatement represents `global name1, name2, ...`
type GlobalStatement struct {
	Token lexer.Token   // The 'global' token
	Names []*Identifier // The list of identifiers declared global
}

func (gs *GlobalStatement) statementNode()       {}
func (gs *GlobalStatement) TokenLiteral() string { return gs.Token.Literal }
func (gs *GlobalStatement) String() string {
	var out bytes.Buffer
	out.WriteString(gs.TokenLiteral() + " ")
	names := []string{}
	for _, n := range gs.Names {
		names = append(names, n.String())
	}
	out.WriteString(strings.Join(names, ", "))
	return out.String()
}

// LetStatement represents a variable assignment (binding a value to a name).
// Example: `x = 5` or `my_var = "hello"`
// Note: Python uses simple assignment, not a 'let' keyword, but this struct represents the assignment structure.
type LetStatement struct {
	Token lexer.Token // The '=' token
	Name  *Identifier // The identifier being assigned to
	Value Expression  // The expression producing the value to assign
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal } // Returns "="
func (ls *LetStatement) String() string {
	var out bytes.Buffer
	out.WriteString(ls.Name.String())
	out.WriteString(constants.AstLetStatementAssignOp)
	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}
	// Optionally add semicolon if language requires it: out.WriteString(";")
	return out.String()
}

// AssignStatement represents a variable assignment (e.g., `x = 5`, `x += 1`).
// For multiple assignment `x, y = 1, 2`, the `Target` will be a `TupleLiteral`.
type AssignStatement struct {
	Token    lexer.Token // The assignment token itself (e.g., '=', '+=')
	Target   Expression  // <<< REVERT: Back to single Target Expression
	Operator string      // The assignment operator (e.g., "=", "+=")
	Value    Expression  // The right-hand side expression
}

func (as *AssignStatement) expressionNode()      {}
func (as *AssignStatement) statementNode()       {}
func (as *AssignStatement) TokenLiteral() string { return as.Token.Literal }

func (as *AssignStatement) String() string {
	var out bytes.Buffer
	out.WriteString(as.Target.String()) // <<< REVERT
	out.WriteString(constants.Space + as.Operator + constants.Space)
	if as.Value != nil {
		out.WriteString(as.Value.String())
	}
	return out.String()
}

// NilLiteral represents the 'None' value.
type NilLiteral struct {
	Token lexer.Token // The lexer.NIL token ("None")
}

func (nl *NilLiteral) expressionNode()      {}
func (nl *NilLiteral) TokenLiteral() string { return nl.Token.Literal } // Returns "None"
func (nl *NilLiteral) String() string       { return nl.Token.Literal } // Returns "None"

// AssertStatement represents `assert condition, [message]`
type AssertStatement struct {
	Token     lexer.Token // The 'assert' token
	Condition Expression
	Message   Expression // Optional, can be nil
}

func (as *AssertStatement) statementNode()       {}
func (as *AssertStatement) TokenLiteral() string { return as.Token.Literal }
func (as *AssertStatement) String() string {
	var out bytes.Buffer
	out.WriteString(as.TokenLiteral() + " ")
	if as.Condition != nil {
		out.WriteString(as.Condition.String())
	}
	if as.Message != nil {
		out.WriteString(", ")
		out.WriteString(as.Message.String())
	}
	return out.String()
}

// YieldExpression represents `yield [expression]`
type YieldExpression struct {
	Token lexer.Token // The 'yield' token
	Value Expression  // The value to yield (can be nil)
}

func (ye *YieldExpression) expressionNode()      {}
func (ye *YieldExpression) TokenLiteral() string { return ye.Token.Literal }
func (ye *YieldExpression) String() string {
	var out bytes.Buffer
	out.WriteString(ye.TokenLiteral())
	if ye.Value != nil {
		out.WriteString(" ")
		out.WriteString(ye.Value.String())
	}
	return out.String()
}

// ReturnStatement represents a return statement in a function.
// Example: `return x * 2`
type ReturnStatement struct {
	Token       lexer.Token // The 'return' token
	ReturnValue Expression  // The expression to be returned
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal } // Returns "return"
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer
	out.WriteString(rs.TokenLiteral() + constants.Space)
	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}
	// Optionally add semicolon: out.WriteString(";")
	return out.String()
}

// BlockStatement represents a sequence of statements, typically enclosed in braces or indented.
type BlockStatement struct {
	Token      lexer.Token // The '{' token or the first token of the block implicitly
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer
	// In Python-like language, blocks don't have explicit braces usually.
	// The string representation might just concatenate the statements.
	for _, s := range bs.Statements {
		out.WriteString(s.String())
		// Add indentation/newlines for better readability if needed
	}
	return out.String()
}

// --- Expressions ---

// Identifier represents a variable or function name.
// Example: `x`, `my_function`
type Identifier struct {
	Token lexer.Token // The lexer.IDENT token
	Value string      // The actual name (e.g., "x")
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

// BooleanLiteral represents boolean values (True or False).
type BooleanLiteral struct {
	Token lexer.Token // The lexer.TRUE or lexer.FALSE token
	Value bool
}

func (b *BooleanLiteral) expressionNode()      {}
func (b *BooleanLiteral) TokenLiteral() string { return b.Token.Literal }
func (b *BooleanLiteral) String() string       { return b.Token.Literal }

// IntegerLiteral represents integer numbers.
// Example: `5`, `100`, `-3`
type IntegerLiteral struct {
	Token lexer.Token // The lexer.INT token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

// FloatLiteral represents floating-point numbers.
// Example: `3.14`, `-0.5`
type FloatLiteral struct {
	Token lexer.Token // The lexer.FLOAT token
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }

// StringLiteral represents string values.
// Example: `"hello"`, `'world'`
type StringLiteral struct {
	Token lexer.Token // The lexer.STRING token
	Value string      // The actual string content (without quotes)
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return fmt.Sprintf(constants.DoubleQuoteFormat, sl.Value) } // Print with quotes for clarity

// PrefixExpression represents expressions with a leading operator.
// Example: `!True`, `-15`
type PrefixExpression struct {
	Token    lexer.Token // The prefix token (e.g., '!', '-')
	Operator string      // The operator itself (e.g., "!", "-")
	Right    Expression  // The expression to the right of the operator
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer
	out.WriteString(constants.OpenParen)
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(constants.CloseParen)
	return out.String()
}

// InfixExpression represents expressions with an operator between two operands.
// Example: `5 + 5`, `x > y`, `a == b`
type InfixExpression struct {
	Token    lexer.Token // The operator token (e.g., '+', '==')
	Left     Expression  // The expression to the left of the operator
	Operator string      // The operator itself (e.g., "+", "==")
	Right    Expression  // The expression to the right of the operator
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer
	out.WriteString(constants.OpenParen)
	out.WriteString(ie.Left.String())
	out.WriteString(constants.Space + ie.Operator + constants.Space)
	out.WriteString(ie.Right.String())
	out.WriteString(constants.CloseParen)
	return out.String()
}

// IfExpression represents an if/elif/else conditional.
// Python's `if` is more of a statement controlling flow, but can sometimes
// be used in expression contexts (`x = a if condition else b`).
// We'll model the primary statement form here.
type IfStatement struct {
	Token       lexer.Token     // The 'if' token
	Condition   Expression      // The condition to evaluate
	Consequence *BlockStatement // Statements to execute if condition is true
	Alternative *BlockStatement // The 'else' block (optional)
	ElifBlocks  []*ElifBlock    // List of 'elif' blocks (optional)
}

// ElifBlock represents a single 'elif' condition and its consequence.
type ElifBlock struct {
	Token       lexer.Token     // The 'elif' token
	Condition   Expression      // The elif condition
	Consequence *BlockStatement // Statements for this elif block
}

func (ifs *IfStatement) statementNode()       {}
func (ifs *IfStatement) TokenLiteral() string { return ifs.Token.Literal } // Returns "if"
func (ifs *IfStatement) String() string {
	var out bytes.Buffer
	out.WriteString(constants.IfKeywordWithSpace)
	out.WriteString(ifs.Condition.String())
	out.WriteString(constants.AstFunctionLiteralBodyColon) // Assuming Python-like structure representation
	out.WriteString(ifs.Consequence.String())              // Add indentation in practice

	for _, eb := range ifs.ElifBlocks {
		out.WriteString(constants.Newline + constants.ElifKeywordWithSpace)
		out.WriteString(eb.Condition.String())
		out.WriteString(constants.AstFunctionLiteralBodyColon)
		out.WriteString(eb.Consequence.String())
	}

	if ifs.Alternative != nil {
		out.WriteString(constants.Newline + constants.ElseKeywordWithColonNewline)
		out.WriteString(ifs.Alternative.String())
	}
	return out.String()
}

// ConditionalExpression represents `value_if_true if condition else value_if_false`
// TernaryExpression represents `value_if_true if condition else value_if_false`
type TernaryExpression struct {
	Token        lexer.Token // The 'if' token
	ValueIfTrue  Expression
	Condition    Expression
	ValueIfFalse Expression
}

func (te *TernaryExpression) expressionNode()      {}
func (te *TernaryExpression) TokenLiteral() string { return te.Token.Literal }
func (te *TernaryExpression) String() string {
	var out bytes.Buffer
	out.WriteString(constants.OpenParen)
	out.WriteString(te.ValueIfTrue.String())
	out.WriteString(constants.Space + constants.IfKeywordWithSpace)
	out.WriteString(te.Condition.String())
	out.WriteString(constants.Space + constants.ElseKeywordWithSpace)
	out.WriteString(te.ValueIfFalse.String())
	out.WriteString(constants.CloseParen)
	return out.String()
}

// Parameter represents a function parameter with an optional default value.
type Parameter struct {
	Token        lexer.Token // The token of the parameter identifier
	Name         *Identifier
	DefaultValue Expression // Optional: nil if no default value
}

// Node interface is not strictly needed for Parameter itself if it's only a struct used by FunctionLiteral
// func (p *Parameter) Node() {}
func (p *Parameter) TokenLiteral() string { return p.Token.Literal }
func (p *Parameter) String() string {
	var out bytes.Buffer
	out.WriteString(p.Name.String())
	if p.DefaultValue != nil {
		out.WriteString(constants.AstParameterDefaultAssignOp)
		out.WriteString(p.DefaultValue.String())
	}
	return out.String()
}

// FunctionLiteral represents the definition of a function.
// Example: `def add(x, y): return x + y`
type FunctionLiteral struct {
	Token       lexer.Token     // The 'def' token
	Name        *Identifier     // Optional function name (for named functions)
	Decorators  []Expression    // List of parameter identifiers
	Parameters  []*Parameter    // <<<< MODIFIED to use new Parameter type
	VarArgParam *Identifier     // Name for *args (e.g., "args"), nil if not present
	KwArgParam  *Identifier     // Name for **kwargs (e.g., "kwargs"), nil if not present
	Body        *BlockStatement // The function body
	IsAsync     bool
	IsGenerator bool
}

func (fl *FunctionLiteral) expressionNode()      {}                          // Can be treated as an expression that yields a function object
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal } // Returns "def"
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer
	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}
	if fl.VarArgParam != nil {
		params = append(params, constants.AstStarredArgumentPositionalPrefix+fl.VarArgParam.String())
	}
	if fl.KwArgParam != nil {
		params = append(params, constants.AstStarredArgumentKeywordPrefix+fl.KwArgParam.String())
	}

	out.WriteString(fl.TokenLiteral()) // "def"
	if fl.Name != nil {
		out.WriteString(constants.Space + fl.Name.String())
	}
	out.WriteString(constants.AstFunctionLiteralCallParens)
	out.WriteString(strings.Join(params, constants.AstFunctionLiteralArgSeparator))
	out.WriteString(constants.AstCallExpressionCloseParens)
	if fl.Body != nil {
		out.WriteString(constants.AstFunctionLiteralBodyColon)
		out.WriteString(fl.Body.String())
	} else {
		out.WriteString(constants.AstFunctionLiteralBodyEllipsis)
	}
	return out.String()
}

// LambdaLiteral represents a lambda function expression.
// Example: `lambda x, y=1: x + y`
type LambdaLiteral struct {
	Token      lexer.Token  // The 'lambda' token
	Parameters []*Parameter // The list of parameters
	Body       Expression   // The single expression that is the body
}

func (ll *LambdaLiteral) expressionNode()      {}
func (ll *LambdaLiteral) TokenLiteral() string { return ll.Token.Literal }
func (ll *LambdaLiteral) String() string {
	var out bytes.Buffer
	params := []string{}
	for _, p := range ll.Parameters {
		params = append(params, p.String())
	}

	out.WriteString(ll.TokenLiteral())
	if len(params) > 0 {
		out.WriteString(" ")
		out.WriteString(strings.Join(params, ", "))
	}
	out.WriteString(": ")
	out.WriteString(ll.Body.String())

	return out.String()
}

type AwaitExpression struct {
	Token      lexer.Token // The 'await' token
	Expression Expression  // The expression being awaited (should evaluate to an AsyncResult)
}

func (ae *AwaitExpression) expressionNode()      {}
func (ae *AwaitExpression) TokenLiteral() string { return ae.Token.Literal }
func (ae *AwaitExpression) String() string {
	return fmt.Sprintf(constants.AstAwaitExpressionFormat, ae.Expression.String())
}

// KeywordArgument represents a `name=value` argument in a function call.
type KeywordArgument struct {
	Token lexer.Token // The '=' token, or the Name's token
	Name  *Identifier
	Value Expression
}

func (ka *KeywordArgument) expressionNode()      {} // Implements Expression so it can be in CallExpression.Arguments
func (ka *KeywordArgument) TokenLiteral() string { return ka.Name.TokenLiteral() }
func (ka *KeywordArgument) String() string {
	return ka.Name.String() + constants.AstKeywordArgumentAssignOp + ka.Value.String()
}

// --- StarredArgument (New AST Node for call-site unpacking) ---
type StarredArgument struct {
	Token      lexer.Token // The '*' or '**' token
	Value      Expression  // The expression being unpacked
	IsKwUnpack bool        // True if '**', false if '*'
}

func (sa *StarredArgument) expressionNode()      {} // Implements Expression
func (sa *StarredArgument) TokenLiteral() string { return sa.Token.Literal }
func (sa *StarredArgument) String() string {
	prefix := constants.AstStarredArgumentPositionalPrefix
	if sa.IsKwUnpack {
		prefix = constants.AstStarredArgumentKeywordPrefix
	}
	return prefix + sa.Value.String()
}

// Arguments can be normal expressions or StarredArgument expressions.
// We keep `Arguments []Expression` and type-assert later, as StarredArgument implements Expression.
type CallExpression struct {
	Token     lexer.Token  // The '(' token
	Function  Expression   // Expression resulting in a function
	Arguments []Expression // List of arguments (can include *ast.StarredArgument)
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer
	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}
	out.WriteString(ce.Function.String())
	out.WriteString(constants.AstFunctionLiteralCallParens)
	out.WriteString(strings.Join(args, constants.AstCallExpressionArgSeparator))
	out.WriteString(constants.AstCallExpressionCloseParens)
	return out.String()
}

// ListLiteral represents the creation of a list.
// Example: `[1, "two", True]`
type ListLiteral struct {
	Token    lexer.Token // The '[' token
	Elements []Expression
}

func (ll *ListLiteral) expressionNode()      {}
func (ll *ListLiteral) TokenLiteral() string { return ll.Token.Literal } // Returns "["
func (ll *ListLiteral) String() string {
	var out bytes.Buffer
	elements := []string{}
	for _, el := range ll.Elements {
		elements = append(elements, el.String())
	}
	out.WriteString(constants.AstListLiteralOpenBracket)
	out.WriteString(strings.Join(elements, constants.AstListLiteralSeparator))
	out.WriteString(constants.AstListLiteralCloseBracket)
	return out.String()
}

// --- NEW: ListComprehension AST Node ---
type ListComprehension struct {
	Token     lexer.Token // The '[' token
	Element   Expression  // The expression to compute for each item (e.g., `i*2`)
	Variable  *Identifier // The loop variable (e.g., `i`)
	Iterable  Expression  // The expression to iterate over (e.g., `range(5)`)
	Condition Expression  // Optional `if` condition (can be nil)
}

func (lc *ListComprehension) expressionNode()      {}
func (lc *ListComprehension) TokenLiteral() string { return lc.Token.Literal }
func (lc *ListComprehension) String() string {
	var out bytes.Buffer
	out.WriteString(constants.AstListComprehensionOpenBracket)
	out.WriteString(lc.Element.String())
	out.WriteString(constants.AstListComprehensionForKeyword)
	out.WriteString(lc.Variable.String())
	out.WriteString(constants.AstListComprehensionInKeyword)
	out.WriteString(lc.Iterable.String())
	if lc.Condition != nil {
		out.WriteString(constants.AstListComprehensionIfKeyword)
		out.WriteString(lc.Condition.String())
	}
	out.WriteString(constants.AstListComprehensionCloseBracket)
	return out.String()
}

// DictLiteral represents the creation of a dictionary (map).
// Example: `{"one": 1, "two": 2}`
type DictLiteral struct {
	Token lexer.Token               // The '{' token
	Pairs map[Expression]Expression // Key-value pairs
}

func (dl *DictLiteral) expressionNode()      {}
func (dl *DictLiteral) TokenLiteral() string { return dl.Token.Literal } // Returns "{"
func (dl *DictLiteral) String() string {
	var out bytes.Buffer
	pairs := []string{}
	// Note: map iteration order is not guaranteed in Go. For consistent string
	// representation in tests, consider sorting keys if they are comparable.
	for key, value := range dl.Pairs {
		pairs = append(pairs, key.String()+constants.AstDictLiteralPairSeparator+value.String())
	}
	out.WriteString(constants.AstDictLiteralOpenBrace)
	out.WriteString(strings.Join(pairs, constants.AstDictLiteralSeparator))
	out.WriteString(constants.AstDictLiteralCloseBrace)
	return out.String()
}

// IndexExpression represents accessing an element by index or key.
// Example: `my_list[0]`, `my_dict["key"]`
type IndexExpression struct {
	Token lexer.Token // The '[' token
	Left  Expression  // The expression being indexed (list, dict, string)
	Index Expression  // The index or key expression
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal } // Returns "["
func (ie *IndexExpression) String() string {
	var out bytes.Buffer
	out.WriteString(constants.AstIndexExpressionOpenParen)
	out.WriteString(ie.Left.String())
	out.WriteString(constants.AstIndexExpressionOpenBracket)
	out.WriteString(ie.Index.String())
	out.WriteString(constants.AstIndexExpressionCloseBracket)
	out.WriteString(constants.AstIndexExpressionCloseParen)
	return out.String()
}

// SliceExpression represents `object[start:stop:step]`
type SliceExpression struct {
	Token lexer.Token // The '[' token
	Left  Expression  // The object being sliced
	Start Expression  // Optional start index
	Stop  Expression  // Optional stop index
	Step  Expression  // Optional step value
}

func (se *SliceExpression) expressionNode()      {}
func (se *SliceExpression) TokenLiteral() string { return se.Token.Literal } // Returns "["
func (se *SliceExpression) String() string {
	var out bytes.Buffer
	out.WriteString(constants.AstSliceExpressionOpenParen)
	out.WriteString(se.Left.String())
	out.WriteString(constants.AstSliceExpressionOpenBracket)
	if se.Start != nil {
		out.WriteString(se.Start.String())
	}
	out.WriteString(constants.AstSliceExpressionColon)
	if se.Stop != nil {
		out.WriteString(se.Stop.String())
	}
	if se.Step != nil {
		out.WriteString(constants.AstSliceExpressionColon)
		out.WriteString(se.Step.String())
	}
	out.WriteString(constants.AstSliceExpressionCloseBracket)
	out.WriteString(constants.AstSliceExpressionCloseParen)
	return out.String()
}

// --- Loop Statements (Added based on requirements) ---

// WhileStatement represents a 'while' loop.
// Example: `while x < 10: x = x + 1`
type WhileStatement struct {
	Token     lexer.Token     // The 'while' token
	Condition Expression      // Loop condition
	Body      *BlockStatement // Statements to execute in the loop
}

func (ws *WhileStatement) statementNode()       {}
func (ws *WhileStatement) TokenLiteral() string { return ws.Token.Literal } // Returns "while"
func (ws *WhileStatement) String() string {
	var out bytes.Buffer
	out.WriteString(constants.AstWhileStatementWhileKeyword)
	out.WriteString(ws.Condition.String())
	out.WriteString(constants.AstWhileStatementColonNewline)
	out.WriteString(ws.Body.String())
	return out.String()
}

// ForStatement represents a 'for...in' loop.
// Example: `for item in my_list: print(item)`
type ForStatement struct {
	Token     lexer.Token   // the 'for' token
	Variable  *Identifier   // for single variable (legacy)
	Variables []*Identifier // for multiple variables (tuple unpacking)
	Iterable  Expression
	Body      *BlockStatement
}

func (fs *ForStatement) statementNode()       {}
func (fs *ForStatement) TokenLiteral() string { return fs.Token.Literal } // Returns "for"
func (fs *ForStatement) String() string {
	var out bytes.Buffer
	out.WriteString(fs.TokenLiteral() + constants.Space)

	if len(fs.Variables) > 0 {
		// Multiple variables
		for i, v := range fs.Variables {
			if i > 0 {
				out.WriteString(constants.AstDictLiteralSeparator)
			}
			out.WriteString(v.String())
		}
	} else if fs.Variable != nil {
		// Single variable (legacy)
		out.WriteString(fs.Variable.String())
	}

	out.WriteString(constants.AstForStatementInKeyword)
	if fs.Iterable != nil {
		out.WriteString(fs.Iterable.String())
	}
	out.WriteString(constants.AstForStatementColon)
	if fs.Body != nil {
		out.WriteString(fs.Body.String())
	}
	return out.String()
}

// --- Loop Control Statements ---

// BreakStatement represents a 'break' statement.
type BreakStatement struct {
	Token lexer.Token // The 'break' token
}

func (bs *BreakStatement) statementNode()       {}
func (bs *BreakStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BreakStatement) String() string       { return bs.Token.Literal }

// ContinueStatement represents a 'continue' statement.
type ContinueStatement struct {
	Token lexer.Token // The 'continue' token
}

func (cs *ContinueStatement) statementNode()       {}
func (cs *ContinueStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *ContinueStatement) String() string       { return cs.Token.Literal }

// WithStatement represents a 'with context_expr [as var]: body' statement
type WithStatement struct {
	Token          lexer.Token     // The 'with' token
	ContextManager Expression      // The expression that results in a context manager
	TargetVariable *Identifier     // Optional: the variable after 'as'
	Body           *BlockStatement // The indented block to execute
}

func (ws *WithStatement) statementNode()       {}
func (ws *WithStatement) TokenLiteral() string { return ws.Token.Literal }
func (ws *WithStatement) String() string {
	var out bytes.Buffer
	out.WriteString(constants.AstWithStatementWithKeyword)
	out.WriteString(ws.ContextManager.String())
	if ws.TargetVariable != nil {
		out.WriteString(constants.AstWithStatementAsKeyword)
		out.WriteString(ws.TargetVariable.String())
	}
	out.WriteString(constants.AstWithStatementColonNewline)
	if ws.Body != nil {
		out.WriteString(ws.Body.String()) // Assuming Body.String() handles visual indent
	} else {
		out.WriteString(constants.AstWithStatementEmptyBodyComment)
	}
	return out.String()
}

// ClassStatement represents a class definition `class Name(Base1, Base2): ... body ...`
type ClassStatement struct {
	Token        lexer.Token     // The 'class' token
	Name         *Identifier     // The name of the class
	Decorators   []Expression    // List of decorator expressions
	Superclasses []*Identifier   // List of superclass identifiers
	Body         *BlockStatement // The block containing method definitions
}

func (cs *ClassStatement) statementNode()       {}
func (cs *ClassStatement) TokenLiteral() string { return cs.Token.Literal }

// And update its String() method to handle the list:
func (cs *ClassStatement) String() string {
	var out bytes.Buffer
	out.WriteString(constants.AstClassStatementClassKeyword)
	out.WriteString(cs.Name.String())
	if len(cs.Superclasses) > 0 { // <<-- CHANGE THIS BLOCK
		out.WriteString(constants.AstClassStatementSuperclassOpenParen)
		superNames := []string{}
		for _, sup := range cs.Superclasses {
			superNames = append(superNames, sup.String())
		}
		out.WriteString(strings.Join(superNames, constants.AstClassStatementSuperclassSeparator))
		out.WriteString(constants.AstClassStatementSuperclassCloseParen)
	}
	out.WriteString(constants.AstClassStatementColonNewline)
	if cs.Body != nil {
		out.WriteString(cs.Body.String())
	} else {
		out.WriteString(constants.AstClassStatementBodyEllipsis)
	}
	return out.String()
}

// DotExpression represents accessing an attribute or method.
// Example: `my_object.attribute`, `my_instance.method`
type DotExpression struct {
	Token      lexer.Token // The '.' token
	Left       Expression  // The object being accessed
	Identifier *Identifier // The attribute or method name being accessed
}

func (de *DotExpression) expressionNode()      {}
func (de *DotExpression) TokenLiteral() string { return de.Token.Literal } // Returns "."
func (de *DotExpression) String() string {
	var out bytes.Buffer
	out.WriteString(constants.AstDotExpressionOpenParen)
	out.WriteString(de.Left.String())
	out.WriteString(constants.AstDotExpressionDotOperator)
	out.WriteString(de.Identifier.String())
	out.WriteString(constants.AstDotExpressionCloseParen)
	return out.String()
}

// PassStatement represents a 'pass' statement.
type PassStatement struct {
	Token lexer.Token // The 'pass' token
}

func (ps *PassStatement) statementNode()       {}
func (ps *PassStatement) TokenLiteral() string { return ps.Token.Literal }
func (ps *PassStatement) String() string       { return constants.AstPassStatementPassKeyword } // Just "pass"

// ImportStatement represents `import name` or `import name as alias`
type ImportStatement struct {
	Token lexer.Token // The 'import' token
	Name  *Identifier // The module identifier being imported
	Alias *Identifier // Optional alias for the module
}

func (is *ImportStatement) statementNode()       {}
func (is *ImportStatement) TokenLiteral() string { return is.Token.Literal }
func (is *ImportStatement) String() string {
	var out bytes.Buffer
	out.WriteString(is.TokenLiteral() + constants.AstImportStatementSpace)
	out.WriteString(is.Name.String())
	if is.Alias != nil {
		out.WriteString(constants.AstExceptHandlerAsKeyword + is.Alias.String())
	}
	return out.String()
}

// ===========================================
// Represents an individual name being imported in a 'from' statement
type ImportNamePair struct {
	Token        lexer.Token // The 'name' token (or '*' token for import *)
	OriginalName *Identifier // The name as it exists in the source module
	Alias        *Identifier // Optional alias for the name in the current scope (e.g., import foo as bar)
}

func (inp *ImportNamePair) String() string {
	if inp.Alias != nil {
		return inp.OriginalName.String() + constants.AstImportNamePairAsKeyword + inp.Alias.String()
	}
	return inp.OriginalName.String()
}
func (inp *ImportNamePair) expressionNode() {}

// FromImportStatement represents 'from module import name1, name2 as alias' or 'from module import *'
type FromImportStatement struct {
	Token      lexer.Token       // The 'from' token
	ModulePath Expression        // Module identifier (or dotted path expression)
	Names      []*ImportNamePair // List of names to import (nil or empty if ImportAll is true)
	ImportAll  bool              // True if 'import *' was used
}

func (fis *FromImportStatement) statementNode()       {}
func (fis *FromImportStatement) TokenLiteral() string { return fis.Token.Literal }
func (fis *FromImportStatement) String() string {
	var out bytes.Buffer
	out.WriteString(fis.TokenLiteral() + constants.AstFromImportStatementSpace)
	out.WriteString(fis.ModulePath.String())
	out.WriteString(constants.AstFromImportStatementImportKeyword)
	if fis.ImportAll {
		out.WriteString(constants.AstFromImportStatementImportAll)
	} else {
		names := []string{}
		for _, np := range fis.Names {
			names = append(names, np.String())
		}
		out.WriteString(strings.Join(names, constants.AstFromImportStatementNameSeparator))
	}
	return out.String()
}

// ===========================================

// TryStatement represents a try...except block.
type TryStatement struct {
	Token    lexer.Token      // The 'try' token
	Body     *BlockStatement  // The block of code to try
	Handlers []*ExceptHandler // List of except handlers
	// Else    *BlockStatement // Optional 'else' block (TODO later)
	Finally *BlockStatement // Optional 'finally' block (TODO later)
}

func (ts *TryStatement) statementNode()       {}
func (ts *TryStatement) TokenLiteral() string { return ts.Token.Literal }
func (ts *TryStatement) String() string {
	var out bytes.Buffer
	out.WriteString(constants.AstTryStatementTryKeyword)
	out.WriteString(ts.Body.String()) // Assuming String() handles indentation notionally
	for _, h := range ts.Handlers {
		out.WriteString(h.String())
	}
	// TODO: Add Else string representation later
	if ts.Finally != nil {
		out.WriteString(constants.AstTryStatementFinallyKeyword) // Or use a constant
		out.WriteString(ts.Finally.String())
	}
	return out.String()
}

// ExceptHandler represents a single 'except' clause.
type ExceptHandler struct {
	Token lexer.Token     // The 'except' token
	Type  Expression      // Optional: The type of exception to catch (e.g., identifier like NameError)
	Var   *Identifier     // Optional: The variable to bind the exception object to (e.g., 'e' in 'as e')
	Body  *BlockStatement // The block of code to execute if exception matches
}

// String representation for ExceptHandler
func (eh *ExceptHandler) String() string {
	var out bytes.Buffer
	out.WriteString(constants.AstExceptHandlerExceptKeyword)
	if eh.Type != nil {
		out.WriteString(constants.Space)
		out.WriteString(eh.Type.String())
	}
	if eh.Var != nil {
		out.WriteString(constants.AstExceptHandlerAsKeyword)
		out.WriteString(eh.Var.String())
	}
	out.WriteString(constants.AstExceptHandlerColonNewline)
	out.WriteString(eh.Body.String())
	return out.String()
}

// RaiseStatement represents `raise [expression]`
type RaiseStatement struct {
	Token     lexer.Token // The 'raise' token
	Exception Expression  // The exception to raise (optional, nil for bare raise)
}

func (rs *RaiseStatement) statementNode()       {}
func (rs *RaiseStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *RaiseStatement) String() string {
	var out bytes.Buffer
	out.WriteString(rs.TokenLiteral())
	if rs.Exception != nil {
		out.WriteString(constants.AstRaiseStatementSpace)
		out.WriteString(rs.Exception.String())
	}
	return out.String()
}

// --- NEW: Tuple Literal ---
// Represents `(1, "two")` or `1, "two"` in some contexts
type TupleLiteral struct {
	Token    lexer.Token // The '(' token or first element's token if implicit
	Elements []Expression
}

func (tl *TupleLiteral) expressionNode()      {}
func (tl *TupleLiteral) TokenLiteral() string { return tl.Token.Literal }
func (tl *TupleLiteral) String() string {
	var out bytes.Buffer
	elements := []string{}
	for _, el := range tl.Elements {
		elements = append(elements, el.String())
	}
	out.WriteString(constants.AstTupleLiteralOpenParen)
	out.WriteString(strings.Join(elements, constants.AstTupleLiteralSeparator))
	// Add trailing comma for single-element tuple representation
	if len(tl.Elements) == 1 {
		out.WriteString(constants.AstTupleLiteralTrailingComma)
	}
	out.WriteString(constants.AstTupleLiteralCloseParen)
	return out.String()
}

// --- NEW: Set Literal ---
// Represents `{1, "two"}` (Parser needs to distinguish from dict)
type SetLiteral struct {
	Token    lexer.Token // The '{' token
	Elements []Expression
}

func (sl *SetLiteral) expressionNode()      {}
func (sl *SetLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *SetLiteral) String() string {
	var out bytes.Buffer
	elements := []string{}
	for _, el := range sl.Elements {
		elements = append(elements, el.String())
	}
	// Note: Set element order isn't guaranteed, but AST string is literal
	out.WriteString(constants.AstSetLiteralOpenBrace)
	out.WriteString(strings.Join(elements, constants.AstSetLiteralSeparator))
	out.WriteString(constants.AstSetLiteralCloseBrace)
	return out.String()
}

// --- NEW: SetComprehension AST Node ---
type SetComprehension struct {
	Token     lexer.Token // The '{' token
	Element   Expression  // The expression to compute for each item
	Variable  *Identifier // The loop variable
	Iterable  Expression  // The expression to iterate over
	Condition Expression  // Optional `if` condition (can be nil)
}

func (sc *SetComprehension) expressionNode()      {}
func (sc *SetComprehension) TokenLiteral() string { return sc.Token.Literal }
func (sc *SetComprehension) String() string {
	var out bytes.Buffer
	out.WriteString(constants.AstSetComprehensionOpenBrace)
	out.WriteString(sc.Element.String())
	out.WriteString(constants.AstSetComprehensionForKeyword)
	out.WriteString(sc.Variable.String())
	out.WriteString(constants.AstSetComprehensionInKeyword)
	out.WriteString(sc.Iterable.String())
	if sc.Condition != nil {
		out.WriteString(constants.AstSetComprehensionIfKeyword)
		out.WriteString(sc.Condition.String())
	}
	out.WriteString(constants.AstSetComprehensionCloseBrace)
	return out.String()
}

type BytesLiteral struct {
	Token lexer.Token // The lexer.BYTES token
	Value []byte
}

func (bl *BytesLiteral) expressionNode()      {}
func (bl *BytesLiteral) TokenLiteral() string { return bl.Token.Literal } // Maybe format like b'...'?
// String returns a string representation of the AST node itself,
// typically resembling the source code literal.
func (bl *BytesLiteral) String() string {
	// Return the original literal string stored in the token.
	// This assumes the lexer correctly stores the full b'...' literal.
	return bl.Token.Literal
}

// GetToken returns the primary token associated with an expression node.
// It uses a type switch to access the Token field from the concrete type.
func GetToken(expr Expression) lexer.Token {
	switch e := expr.(type) {
	case *Identifier:
		return e.Token
	case *IntegerLiteral:
		return e.Token
	case *FloatLiteral:
		return e.Token
	case *StringLiteral:
		return e.Token
	case *BooleanLiteral:
		return e.Token
	case *NilLiteral:
		return e.Token
	case *FunctionLiteral:
		return e.Token
	case *ListLiteral:
		return e.Token
	case *TupleLiteral:
		return e.Token
	case *DictLiteral:
		return e.Token
	case *SetLiteral:
		return e.Token
	case *PrefixExpression:
		return e.Token
	case *InfixExpression:
		return e.Token
	case *IndexExpression:
		return e.Token
	case *SliceExpression:
		return e.Token
	case *CallExpression:
		return e.Token
	case *DotExpression:
		return e.Token
	// Add other expression types as needed
	default:
		// Fallback for any unhandled types
		return lexer.Token{Type: lexer.ILLEGAL, Literal: constants.AstGetTokenIllegalLiteral}
	}
}
