package interpreter

import (
	"github.com/deniskipeles/pylearn/internal/ast"
	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)

func evalSliceExpression(node *ast.SliceExpression, ctx *InterpreterContext) object.Object {
	left := Eval(node.Left, ctx)
	if object.IsError(left) {
		return left
	}

	// The object must be sliceable (for now, List, Tuple, String, Bytes).
	switch obj := left.(type) {
	case *object.List:
		return sliceList(obj, node, ctx)
	case *object.Tuple:
		return sliceTuple(obj, node, ctx)
	case *object.String:
		return sliceString(obj, node, ctx)
	case *object.Bytes:
		return sliceBytes(obj, node, ctx)
	default:
		return object.NewError(constants.TypeError, constants.InterpreterEvalSlicesObjectNotSliceable, left.Type())
	}
}

// --- NEW SLICING HELPER FUNCTIONS ---

// calculateSliceBounds is a generic helper to compute integer bounds for slicing.
func calculateSliceBounds(node *ast.SliceExpression, seqLen int64, ctx *InterpreterContext) (start, stop, step int64, err object.Object) {
	// Default values
	start, stop, step = 0, seqLen, 1

	// Evaluate Start expression
	if node.Start != nil {
		startObj := Eval(node.Start, ctx)
		if object.IsError(startObj) {
			return 0, 0, 0, startObj
		}
		if startInt, ok := startObj.(*object.Integer); ok {
			start = startInt.Value
		} else if startObj != object.NULL {
			return 0, 0, 0, object.NewError(constants.TypeError, constants.InterpreterEvalSlicesIndexTypeError, startObj.Type())
		}
	}

	// Evaluate Stop expression
	if node.Stop != nil {
		stopObj := Eval(node.Stop, ctx)
		if object.IsError(stopObj) {
			return 0, 0, 0, stopObj
		}
		if stopInt, ok := stopObj.(*object.Integer); ok {
			stop = stopInt.Value
		} else if stopObj != object.NULL {
			return 0, 0, 0, object.NewError(constants.TypeError, constants.InterpreterEvalSlicesIndexTypeError, stopObj.Type())
		}
	}

	// Evaluate Step expression
	if node.Step != nil {
		stepObj := Eval(node.Step, ctx)
		if object.IsError(stepObj) {
			return 0, 0, 0, stepObj
		}
		if stepInt, ok := stepObj.(*object.Integer); ok {
			step = stepInt.Value
			if step == 0 {
				return 0, 0, 0, object.NewError(constants.ValueError, constants.InterpreterEvalSlicesStepCannotBeZeroError)
			}
		} else if stepObj != object.NULL {
			return 0, 0, 0, object.NewError(constants.TypeError, constants.InterpreterEvalSlicesIndexTypeError, stepObj.Type())
		}
	}

	// Apply Python's slice semantics to the integer values
	if step > 0 {
		if start < 0 {
			start += seqLen
		}
		if start < 0 {
			start = 0
		}
		if start > seqLen {
			start = seqLen
		}

		if stop < 0 {
			stop += seqLen
		}
		if stop < 0 {
			stop = 0
		}
		if stop > seqLen {
			stop = seqLen
		}
	} else { // step < 0
		if start < 0 {
			start += seqLen
		}
		if start < -1 {
			start = -1
		} // Clamps to one before the beginning
		if start >= seqLen {
			start = seqLen - 1
		}

		if stop < 0 {
			stop += seqLen
		}
		if stop < -1 {
			stop = -1
		}
		if stop > seqLen {
			stop = seqLen - 1
		}
	}

	return start, stop, step, nil
}

func sliceList(list *object.List, node *ast.SliceExpression, ctx *InterpreterContext) object.Object {
	start, stop, step, err := calculateSliceBounds(node, int64(len(list.Elements)), ctx)
	if err != nil {
		return err
	}

	newElements := []object.Object{}
	if step > 0 {
		for i := start; i < stop; i += step {
			newElements = append(newElements, list.Elements[i])
		}
	} else { // step < 0
		for i := start; i > stop; i += step {
			newElements = append(newElements, list.Elements[i])
		}
	}
	return &object.List{Elements: newElements}
}

func sliceTuple(tuple *object.Tuple, node *ast.SliceExpression, ctx *InterpreterContext) object.Object {
	start, stop, step, err := calculateSliceBounds(node, int64(len(tuple.Elements)), ctx)
	if err != nil {
		return err
	}

	newElements := []object.Object{}
	if step > 0 {
		for i := start; i < stop; i += step {
			newElements = append(newElements, tuple.Elements[i])
		}
	} else { // step < 0
		for i := start; i > stop; i += step {
			newElements = append(newElements, tuple.Elements[i])
		}
	}
	return &object.Tuple{Elements: newElements}
}

func sliceString(str *object.String, node *ast.SliceExpression, ctx *InterpreterContext) object.Object {
	runes := []rune(str.Value)
	start, stop, step, err := calculateSliceBounds(node, int64(len(runes)), ctx)
	if err != nil {
		return err
	}

	var result []rune
	if step > 0 {
		for i := start; i < stop; i += step {
			result = append(result, runes[i])
		}
	} else {
		for i := start; i > stop; i += step {
			result = append(result, runes[i])
		}
	}
	return &object.String{Value: string(result)}
}

func sliceBytes(bytes *object.Bytes, node *ast.SliceExpression, ctx *InterpreterContext) object.Object {
	start, stop, step, err := calculateSliceBounds(node, int64(len(bytes.Value)), ctx)
	if err != nil {
		return err
	}

	var result []byte
	if step > 0 {
		for i := start; i < stop; i += step {
			result = append(result, bytes.Value[i])
		}
	} else {
		for i := start; i > stop; i += step {
			result = append(result, bytes.Value[i])
		}
	}
	return &object.Bytes{Value: result}
}
