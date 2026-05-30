package builtins

import (
	"fmt"
	"sort"

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/lexer"
	"github.com/deniskipeles/pylearn/internal/object"
)

func pyIterFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	token := lexer.Token{Line: -1, Column: 0, Type: lexer.ILLEGAL, Literal: constants.BuiltinsIterTokenLiteral}
	var obj object.Object

	if len(args) < 1 || len(args) > 2 {
		return object.NewError(constants.TypeError, constants.BuiltinsIterArgCountError, len(args))
	}
	obj = args[0]

	if len(args) == 2 {
		sentinel := args[1]
		callableBuiltin, ok := Builtins[constants.BuiltinsCallableFuncName]
		if !ok {
			return object.NewError(constants.InternalError, constants.BuiltinsIterCallableBuiltinNotFound)
		}
		callableResult := ctx.Execute(callableBuiltin, obj)
		if object.IsError(callableResult) {
			return callableResult
		}

		if callableResult != object.TRUE {
			return object.NewErrorWithLocation(token, constants.TypeError, constants.BuiltinsIterVMustBeCallable)
		}

		return &object.GenericIterator{
			Source: constants.BuiltinsIterCallableSource,
			NextFn: func() (object.Object, bool) {
				result := ctx.Execute(obj)
				if object.IsError(result) {
					fmt.Printf(constants.BuiltinsIterWarningErrorIgnored, result.Inspect())
				}

				isEqual := (result.Type() == sentinel.Type() && result.Inspect() == sentinel.Inspect())
				if isEqual {
					return nil, true
				}
				return result, false
			},
		}
	}

	iterator, errObj := object.GetObjectIterator(ctx, obj, token)
	if errObj != nil {
		return errObj
	}
	return iterator
}

func pyNextFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	token := lexer.Token{Line: -1, Column: 0, Type: lexer.ILLEGAL, Literal: constants.BuiltinsNextTokenLiteral}
	hasDefault := false
	var defaultVal object.Object

	if len(args) < 1 || len(args) > 2 {
		return object.NewErrorWithLocation(token, constants.TypeError, constants.BuiltinsNextArgCountError, len(args))
	}

	iteratorArg := args[0]
	if len(args) == 2 {
		hasDefault = true
		defaultVal = args[1]
	}

	iterator, ok := iteratorArg.(object.Iterator)
	if !ok {
		return object.NewErrorWithLocation(token, constants.TypeError, constants.BuiltinsNextObjectNotIterator, iteratorArg.Type())
	}

	nextItem, stop := iterator.Next()
	if stop {
		if hasDefault {
			return defaultVal
		}
		return object.STOP_ITERATION
	}

	if object.IsError(nextItem) {
		return nextItem
	}
	return nextItem
}

func pyEnumerateFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	var kwargs *object.Dict
	var positionalArgs []object.Object

	if len(args) > 0 {
		lastArg := args[len(args)-1]
		if dictObj, isDict := lastArg.(*object.Dict); isDict {
			kwargs = dictObj
			positionalArgs = args[:len(args)-1]
		} else {
			positionalArgs = args
		}
	}

	if len(positionalArgs) < 1 || len(positionalArgs) > 2 {
		return object.NewError(constants.TypeError, constants.BuiltinsEnumerateArgCountError, len(positionalArgs))
	}

	iterableArg := positionalArgs[0]
	startValue := int64(0)

	if len(positionalArgs) == 2 {
		if positionalArgs[1] != object.NULL {
			startObj, ok := positionalArgs[1].(*object.Integer)
			if !ok {
				return object.NewError(constants.TypeError, constants.BuiltinsEnumerateStartArgTypeError, positionalArgs[1].Type())
			}
			startValue = startObj.Value
		}
	}

	if kwargs != nil {
		startKeyObj := &object.String{Value: constants.BuiltinsStartParam}
		startHash, _ := startKeyObj.HashKey()
		if pair, ok := kwargs.Pairs[startHash]; ok {
			if len(positionalArgs) == 2 {
				return object.NewError(constants.TypeError, constants.BuiltinsEnumerateMultipleValuesError)
			}
			startObj, ok := pair.Value.(*object.Integer)
			if !ok {
				return object.NewError(constants.TypeError, constants.BuiltinsEnumerateStartArgTypeError, pair.Value.Type())
			}
			startValue = startObj.Value
		}
	}

	sourceIter, errObj := object.GetObjectIterator(ctx, iterableArg, object.NoToken)
	if errObj != nil {
		return errObj
	}

	return &object.EnumerateIterator{
		SourceIterator: sourceIter,
		CurrentIndex:   startValue,
	}
}

func pyZipFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) == 0 {
		return &object.GenericIterator{
			Source: constants.BuiltinsZipSourceName,
			NextFn: func() (object.Object, bool) { return nil, true },
		}
	}

	iters := make([]object.Iterator, len(args))
	for i, arg := range args {
		iter, errObj := object.GetObjectIterator(ctx, arg, object.NoToken)
		if errObj != nil {
			return errObj
		}
		iters[i] = iter
	}

	return &object.GenericIterator{
		Source: constants.BuiltinsZipSourceName,
		NextFn: func() (object.Object, bool) {
			elements := make([]object.Object, len(iters))
			for i, iter := range iters {
				item, stop := iter.Next()
				if stop {
					return nil, true
				}
				if object.IsError(item) {
					return item, true
				}
				elements[i] = item
			}
			return &object.Tuple{Elements: elements}, false
		},
	}
}

func pyMapFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) < 2 {
		return object.NewError(constants.TypeError, constants.BuiltinsMapAtLeastTwoArgsError)
	}

	fn := args[0]
	if !object.IsCallable(fn) {
		return object.NewError(constants.TypeError, constants.ErrNotCallable, fn.Type())
	}

	iters := make([]object.Iterator, len(args)-1)
	for i, arg := range args[1:] {
		iter, errObj := object.GetObjectIterator(ctx, arg, object.NoToken)
		if errObj != nil {
			return errObj
		}
		iters[i] = iter
	}

	return &object.GenericIterator{
		Source: constants.BuiltinsMapSourceName,
		NextFn: func() (object.Object, bool) {
			callArgs := make([]object.Object, len(iters))
			for i, iter := range iters {
				item, stop := iter.Next()
				if stop {
					return nil, true
				}
				if object.IsError(item) {
					return item, true
				}
				callArgs[i] = item
			}

			res := ctx.Execute(fn, callArgs...)
			if object.IsError(res) {
				return res, true
			}
			return res, false
		},
	}
}

func pyFilterFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 2 {
		return object.NewError(constants.TypeError, constants.BuiltinsFilterArgCountError, len(args))
	}

	fn := args[0]
	iter, errObj := object.GetObjectIterator(ctx, args[1], object.NoToken)
	if errObj != nil {
		return errObj
	}

	return &object.GenericIterator{
		Source: constants.BuiltinsFilterSourceName,
		NextFn: func() (object.Object, bool) {
			for {
				item, stop := iter.Next()
				if stop {
					return nil, true
				}
				if object.IsError(item) {
					return item, true
				}

				var isTrue bool
				if fn == object.NULL {
					isTrue, _ = object.IsTruthy(ctx, item)
				} else {
					res := ctx.Execute(fn, item)
					if object.IsError(res) {
						return res, true
					}
					isTrue, _ = object.IsTruthy(ctx, res)
				}

				if isTrue {
					return item, false
				}
			}
		},
	}
}

func pyReversedFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, constants.BuiltinsReversedArgCountError, len(args))
	}

	iter, errObj := object.GetObjectIterator(ctx, args[0], object.NoToken)
	if errObj != nil {
		return errObj
	}

	items, unpackErr := object.UnpackIterator(iter)
	if unpackErr != nil {
		return object.NewError(constants.RuntimeError, unpackErr.Error())
	}

	idx := len(items) - 1
	return &object.GenericIterator{
		Source: constants.BuiltinsReversedSourceName,
		NextFn: func() (object.Object, bool) {
			if idx < 0 {
				return nil, true
			}
			item := items[idx]
			idx--
			return item, false
		},
	}
}

func pySortedFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	var kwargs *object.Dict
	var positionalArgs []object.Object

	if len(args) > 0 {
		lastArg := args[len(args)-1]
		if dictObj, isDict := lastArg.(*object.Dict); isDict {
			kwargs = dictObj
			positionalArgs = args[:len(args)-1]
		} else {
			positionalArgs = args
		}
	}

	if len(positionalArgs) != 1 {
		return object.NewError(constants.TypeError, constants.BuiltinsSortedArgCountError, len(positionalArgs))
	}

	reverse := false
	var keyFn object.Object = object.NULL

	if kwargs != nil {
		revKey, _ := (&object.String{Value: constants.BuiltinsReverseParam}).HashKey()
		if pair, ok := kwargs.Pairs[revKey]; ok {
			r, _ := object.IsTruthy(ctx, pair.Value)
			reverse = r
		}

		keyKey, _ := (&object.String{Value: constants.BuiltinsKeyParam}).HashKey()
		if pair, ok := kwargs.Pairs[keyKey]; ok {
			keyFn = pair.Value
		}
	}

	iter, errObj := object.GetObjectIterator(ctx, positionalArgs[0], object.NoToken)
	if errObj != nil {
		return errObj
	}

	items, unpackErr := object.UnpackIterator(iter)
	if unpackErr != nil {
		return object.NewError(constants.RuntimeError, unpackErr.Error())
	}

	var sortErr object.Object

	sort.SliceStable(items, func(i, j int) bool {
		if sortErr != nil {
			return false
		}

		a := items[i]
		b := items[j]

		if keyFn != object.NULL {
			a = ctx.Execute(keyFn, a)
			if object.IsError(a) {
				sortErr = a
				return false
			}
			b = ctx.Execute(keyFn, b)
			if object.IsError(b) {
				sortErr = b
				return false
			}
		}

		comp := object.CompareObjects(constants.LessThanOp, a, b, ctx)
		if object.IsError(comp) {
			sortErr = comp
			return false
		}

		isLt := comp == object.TRUE
		if reverse {
			compRev := object.CompareObjects(constants.GreaterThanOp, a, b, ctx)
			if object.IsError(compRev) {
				sortErr = compRev
				return false
			}
			return compRev == object.TRUE
		}
		return isLt
	})

	if sortErr != nil {
		return sortErr
	}

	return &object.List{Elements: items}
}

func pyAllFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, constants.BuiltinsAllArgCountError, len(args))
	}
	iterator, errObj := object.GetObjectIterator(ctx, args[0], object.NoToken)
	if errObj != nil {
		return errObj
	}

	for {
		item, stop := iterator.Next()
		if stop {
			break
		}
		if object.IsError(item) {
			return item
		}

		isTrue, truthErr := object.IsTruthy(ctx, item)
		if truthErr != nil {
			if pyErr, ok := truthErr.(object.Object); ok && object.IsError(pyErr) {
				return pyErr
			}
			return object.NewError(constants.RuntimeError, constants.BuiltinsAllIsTruthyPropagatedError, truthErr)
		}
		if !isTrue {
			return object.FALSE
		}
	}
	return object.TRUE
}

func pyAnyFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, constants.BuiltinsAnyArgCountError, len(args))
	}
	iterator, errObj := object.GetObjectIterator(ctx, args[0], object.NoToken)
	if errObj != nil {
		return errObj
	}

	for {
		item, stop := iterator.Next()
		if stop {
			break
		}
		if object.IsError(item) {
			return item
		}

		isTrue, truthErr := object.IsTruthy(ctx, item)
		if truthErr != nil {
			if pyErr, ok := truthErr.(object.Object); ok && object.IsError(pyErr) {
				return pyErr
			}
			return object.NewError(constants.RuntimeError, constants.BuiltinsAnyIsTruthyPropagatedError, truthErr)
		}
		if isTrue {
			return object.TRUE
		}
	}
	return object.FALSE
}

// --- Registration ---
func init() {
	registerBuiltin(constants.BuiltinsIterFuncName, &object.Builtin{Name: constants.BuiltinsIterFuncName, Fn: pyIterFn})
	registerBuiltin(constants.BuiltinsNextFuncName, &object.Builtin{Name: constants.BuiltinsNextFuncName, Fn: pyNextFn})

	// Enumerate with keyword support
	registerBuiltin(constants.BuiltinsEnumerateFuncName, &object.Builtin{
		Name:            constants.BuiltinsEnumerateFuncName,
		Fn:              pyEnumerateFn,
		AcceptsKeywords: map[string]bool{constants.BuiltinsStartParam: true},
	})

	registerBuiltin(constants.BuiltinsZipFuncName, &object.Builtin{Name: constants.BuiltinsZipFuncName, Fn: pyZipFn})
	registerBuiltin(constants.BuiltinsMapFuncName, &object.Builtin{Name: constants.BuiltinsMapFuncName, Fn: pyMapFn})
	registerBuiltin(constants.BuiltinsFilterFuncName, &object.Builtin{Name: constants.BuiltinsFilterFuncName, Fn: pyFilterFn})
	registerBuiltin(constants.BuiltinsReversedFuncName, &object.Builtin{Name: constants.BuiltinsReversedFuncName, Fn: pyReversedFn})

	// Sorted with keyword support
	registerBuiltin(constants.BuiltinsSortedFuncName, &object.Builtin{
		Name:            constants.BuiltinsSortedFuncName,
		Fn:              pySortedFn,
		AcceptsKeywords: map[string]bool{constants.BuiltinsKeyParam: true, constants.BuiltinsReverseParam: true},
	})

	registerBuiltin(constants.BuiltinsAllFuncName, &object.Builtin{Name: constants.BuiltinsAllFuncName, Fn: pyAllFn})
	registerBuiltin(constants.BuiltinsAnyFuncName, &object.Builtin{Name: constants.BuiltinsAnyFuncName, Fn: pyAnyFn})
}