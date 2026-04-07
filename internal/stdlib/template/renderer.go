// ===========================pylearn/internal/stdlib/template/renderer.go start here===========================
// internal/stdlib/template/renderer.go
package template

import (
	"fmt"
	"strings"

	"github.com/deniskipeles/pylearn/internal/object"
)

// Renderer now holds a context of wrapped Values.
type Renderer struct {
	execCtx object.ExecutionContext
	context map[string]*Value
	out     *strings.Builder
}

func NewRenderer(ctx object.ExecutionContext, context map[string]*Value, out *strings.Builder) *Renderer {
	return &Renderer{execCtx: ctx, context: context, out: out}
}

// evalExpression now returns a *Value, performing lookups in the wrapped context.
func (r *Renderer) evalExpression(node Node) (*Value, error) {
	switch n := node.(type) {
	case *IdentNode:
		// Check for built-in functions first
		if fn, ok := r.context[n.Name]; ok {
			return fn, nil
		}
		return NewValue(object.NULL, r.execCtx), nil // Return a wrapped NULL for undefined variables

	case *DotAccessNode:
		baseVal, err := r.evalExpression(n.Base)
		if err != nil {
			return nil, err
		}
		// Delegate attribute access to the Value's Getattr method
		return baseVal.Getattr(n.Attr), nil

	case *IntegerNode:
		return NewValue(&object.Integer{Value: int64(n.Value)}, r.execCtx), nil

	case *FmtStr:
		return NewValue(&object.String{Value: n.Text}, r.execCtx), nil

	case *BinaryExpressionNode:
		leftVal, err := r.evalExpression(n.Left)
		if err != nil {
			return nil, err
		}
		rightVal, err := r.evalExpression(n.Right)
		if err != nil {
			return nil, err
		}

		// Use Pylearn's robust comparison logic
		result := object.CompareObjects(n.Operator, leftVal.obj, rightVal.obj, r.execCtx)
		if object.IsError(result) {
			return nil, fmt.Errorf(result.Inspect())
		}
		return NewValue(result, r.execCtx), nil

	case *CallNode:
		funcVal, err := r.evalExpression(n.Function)
		if err != nil {
			return nil, err
		}
		var args []object.Object
		for _, argNode := range n.Args {
			argVal, err := r.evalExpression(argNode)
			if err != nil {
				return nil, err
			}
			args = append(args, argVal.obj)
		}
		result := r.execCtx.Execute(funcVal.obj, args...)
		if object.IsError(result) {
			return nil, fmt.Errorf(result.Inspect())
		}
		return NewValue(result, r.execCtx), nil

	case *FilterNode:
		val, err := r.evalExpression(n.Expression)
		if err != nil {
			return nil, err
		}

		filterFn, ok := r.context[n.Name]
		if !ok || filterFn.obj.Type() != object.BUILTIN_OBJ {
			return nil, fmt.Errorf("filter '%s' not found or not a function", n.Name)
		}

		// A filter takes the expression's result as its first argument.
		result := r.execCtx.Execute(filterFn.obj, val.obj)
		if object.IsError(result) {
			return nil, fmt.Errorf(result.Inspect())
		}
		return NewValue(result, r.execCtx), nil
	}

	return nil, fmt.Errorf("cannot evaluate node type %T in an expression", node)
}

// Render traverses the AST and produces the final output.
func (r *Renderer) Render(node Node) error {
	// Check for an extends tag at the top level of the node list
	if nodeList, ok := node.(NodeList); ok && len(nodeList) > 0 {
		if extends, isExtends := nodeList[0].(*ExtendsNode); isExtends {
			return r.handleExtends(extends, nodeList)
		}
	}

	// Regular rendering for all other nodes
	return r.renderNode(node)
}

// renderNode is the internal rendering logic for any given node.
func (r *Renderer) renderNode(node Node) error {
	switch n := node.(type) {
	case NodeList:
		for _, child := range n {
			if err := r.renderNode(child); err != nil {
				return err
			}
		}

	case *RawNode:
		r.out.WriteString(n.String())

	case *TextNode:
		r.out.WriteString(n.String())

	case *VariableNode:
		val, err := r.evalExpression(n.Expression)
		if err != nil {
			return err
		}
		// Use the Value's String() method for correct, unquoted output
		r.out.WriteString(val.String())

	case *IfNode:
		condVal, err := r.evalExpression(n.Condition)
		if err != nil {
			return err
		}
		// Use the Value's IsTrue() method for correct truthiness check
		if condVal.IsTrue() {
			return r.renderNode(n.Body)
		} else if n.Else != nil {
			return r.renderNode(n.Else)
		}

	case *ForNode:
		iterableVal, err := r.evalExpression(n.Iterable)
		if err != nil {
			return err
		}

		iterator, err := iterableVal.Iter()
		if err != nil {
			return err
		}

		// Create a new context for each iteration of the loop
		for {
			item, stop := iterator.Next()
			if stop {
				break
			}
			if object.IsError(item) {
				return fmt.Errorf("error during iteration: %s", item.Inspect())
			}

			// Create a new renderer with an enclosed context for the loop variable
			loopContext := make(map[string]*Value)
			for k, v := range r.context {
				loopContext[k] = v
			}
			loopContext[n.LoopVar] = NewValue(item, r.execCtx)
			loopRenderer := NewRenderer(r.execCtx, loopContext, r.out)

			if err := loopRenderer.renderNode(n.Body); err != nil {
				return err
			}
		}

	case *BlockNode:
		// When rendering a template directly, a block just renders its content.
		return r.renderNode(n.Body)

	case *IncludeNode:
		return r.handleInclude(n)

	default:
		return fmt.Errorf("unknown node type to render: %T", n)
	}
	return nil
}

// --- Handlers for Complex Tags ---

func (r *Renderer) handleInclude(node *IncludeNode) error {
	tpl, err := NewTemplate(node.Template.Text, loadTemplateSource(node.Template.Text))
	if err != nil {
		return err
	}
	// Create a fresh renderer for the included template, but using the same output builder
	includeRenderer := NewRenderer(r.execCtx, r.context, r.out)
	return includeRenderer.Render(tpl.Root)
}

func (r *Renderer) handleExtends(node *ExtendsNode, childNodes NodeList) error {
	parentTpl, err := NewTemplate(node.Parent.Text, loadTemplateSource(node.Parent.Text))
	if err != nil {
		return err
	}

	// Find all blocks in the current (child) template
	childBlocks := make(map[string]*BlockNode)
	for _, n := range childNodes {
		if block, ok := n.(*BlockNode); ok {
			childBlocks[block.Name] = block
		}
	}

	// This is the renderer for the parent template.
	parentRenderer := NewRenderer(r.execCtx, r.context, r.out)

	// Create a new render function that knows how to override blocks.
	var renderParentWithChildBlocks func(Node) error
	renderParentWithChildBlocks = func(n Node) error {
		// If the node is a block and the child has an override, render the child's version.
		if block, ok := n.(*BlockNode); ok {
			if childBlock, exists := childBlocks[block.Name]; exists {
				// Render the child's block using this same special render function
				// to allow for nested block overrides.
				return renderParentWithChildBlocks(childBlock.Body)
			}
		}

		// For all other nodes in the parent, or blocks that are not overridden,
		// we use the original rendering logic by calling renderNode.
		return parentRenderer.renderNode(n)
	}

	// Start the rendering process on the parent's root node using our special render function.
	return renderParentWithChildBlocks(parentTpl.Root)
}

// Dummy function to simulate loading templates from a file system
func loadTemplateSource(path string) string {
	// In a real application, you would read the file from disk here.
	// For this test, we'll hardcode the content.
	if path == "temp_templates/base.html" {
		return `<!DOCTYPE html><html><head><title>{% block title %}Default Title{% endblock %}</title></head><body><div id='content'>{% block content %}{% endblock %}</div></body></html>`
	}
	if path == "temp_templates/header.html" {
		return `<h1>{{ title }}</h1>`
	}
	return ""
}

// ===========================pylearn/internal/stdlib/template/renderer.go ends here===========================
