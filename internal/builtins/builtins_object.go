package builtins

import (
	"reflect" // For id()
	"strings" // For getattr error message

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)

// --- isinstance() ---
// --- isinstance() ---
// Accepts ExecutionContext (may be needed for future __instancecheck__)
func pyIsInstanceFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 2 {
		return object.NewError(constants.TypeError, constants.BuiltinsObjectIsInstanceArgCountError, len(args))
	}
	obj := args[0]
	classinfo := args[1]

	// --- THIS IS THE FIX ---
	// Check if classinfo is a tuple. If it is, iterate over it and check
	// against each type in the tuple.
	if classTuple, isTuple := classinfo.(*object.Tuple); isTuple {
		for _, typeInTuple := range classTuple.Elements {
			// Recursively call our own logic for each type in the tuple.
			// We can use a helper to avoid duplicating the main logic.
			isMatch, errObj := isInstanceCheckHelper(ctx, obj, typeInTuple)
			if errObj != nil {
				// If any check inside the tuple raises an error (e.g., it contains a non-type), propagate it.
				return errObj
			}
			if isMatch {
				// If we find a match with any type in the tuple, return True immediately.
				return object.TRUE
			}
		}
		// If the loop completes without finding a match, return False.
		return object.FALSE
	}
	// --- END OF FIX ---

	// If classinfo is not a tuple, proceed with the original logic.
	isMatch, errObj := isInstanceCheckHelper(ctx, obj, classinfo)
	if errObj != nil {
		return errObj
	}
	return object.NativeBoolToBooleanObject(isMatch)
}

// isInstanceCheckHelper contains the logic for checking a single type.
func isInstanceCheckHelper(ctx object.ExecutionContext, obj, classinfo object.Object) (bool, object.Object) {
	// Check against user-defined object.Class first
	if targetClass, ok := classinfo.(*object.Class); ok {
		// This uses the MRO, so it's the most robust check.
		isMatch, err := object.IsInstance(ctx, obj, targetClass)
		if err != nil {
			return false, err // Propagate internal errors from IsInstance
		}
		return isMatch, nil
	}

	// Handle built-in types by checking their *constructor* functions
	if builtinType, ok := classinfo.(*object.Builtin); ok {
		// checkBuiltinInstanceOf handles builtins like `int`, `str`, `list`, etc.
		result := checkBuiltinInstanceOf(obj, builtinType)
		return result == object.TRUE, nil
	}

	// If classinfo is some other kind of object that's not a type, it's an error.
	return false, object.NewError(constants.TypeError, constants.BuiltinsObjectIsInstanceArg2TypeError, classinfo.Type())
}

// Helper function to check if obj is an instance of a built-in type
// (No changes needed in this helper function)
func checkBuiltinInstanceOf(obj object.Object, builtinType *object.Builtin) object.Object {
	// In Pylearn, we use the constructor functions (like `int`, `str`) as the type objects.
	// The `Name` field of the Builtin struct helps us identify them.
	switch builtinType.Name {
	// Standard Types
	case constants.BuiltinsStrFuncName:
		_, ok := obj.(*object.String)
		return object.NativeBoolToBooleanObject(ok)
	case constants.BuiltinsIntFuncName:
		_, ok := obj.(*object.Integer)
		return object.NativeBoolToBooleanObject(ok)
	case constants.BuiltinsBoolFuncName:
		_, ok := obj.(*object.Boolean)
		return object.NativeBoolToBooleanObject(ok)
	case constants.BuiltinsFloatFuncName:
		_, ok := obj.(*object.Float)
		return object.NativeBoolToBooleanObject(ok)
	case constants.BuiltinsListFuncName:
		_, ok := obj.(*object.List)
		return object.NativeBoolToBooleanObject(ok)
	case constants.BuiltinsDictFuncName:
		_, ok := obj.(*object.Dict)
		return object.NativeBoolToBooleanObject(ok)
	case constants.BuiltinsTupleFuncName:
		_, ok := obj.(*object.Tuple)
		return object.NativeBoolToBooleanObject(ok)
	case constants.BuiltinsSetFuncName:
		_, ok := obj.(*object.Set)
		return object.NativeBoolToBooleanObject(ok)
	case constants.BuiltinsBytesFuncName:
		_, ok := obj.(*object.Bytes)
		return object.NativeBoolToBooleanObject(ok)

	// In Python, everything is an instance of object.
	case constants.BuiltinsObjectFuncName:
		return object.TRUE
	}
	return object.FALSE
}

// --- issubclass() ---
// Accepts ExecutionContext (may be needed for future __subclasscheck__)
func pyIsSubclassFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 2 {
		// Use object.NewError
		return object.NewError(constants.TypeError, constants.BuiltinsObjectIsSubclassArgCountError, len(args))
	}

	// Arg 1 must be a class (object.Class or vm.Class?) Assume object.Class for now.
	class1, ok1 := args[0].(*object.Class)
	if !ok1 {
		// TODO: Handle built-in type objects
		return object.NewError(constants.TypeError, constants.BuiltinsObjectIsSubclassArg1TypeError)
	}

	// Arg 2 must be a class or tuple of classes
	// TODO: Handle tuple for arg 2 using ctx to iterate
	class2, ok2 := args[1].(*object.Class)
	if !ok2 {
		// TODO: Handle built-in type objects
		return object.NewError(constants.TypeError, constants.BuiltinsObjectIsSubclassArg2TypeError)
	}

	// TODO: Call __subclasscheck__ on class2 if present (use ctx)

	if class1 == class2 {
		return object.TRUE // A class is a subclass of itself
	}

	// --- Basic Inheritance Check ---
	// TODO: Replace with proper MRO/superclass chain lookup later
	currentCls := class1
	for currentCls != nil {
		// if currentCls.Superclass == class2 { // Check immediate parent
		//    return object.TRUE
		// }
		// Recurse or loop up the chain
		// currentCls = currentCls.Superclass // Needs Superclass field
		break // Remove break when Superclass field is added
	}
	// --- End Inheritance Check ---

	return object.FALSE
}

// --- id() ---
// Returns memory address of the Go object backing the Pylearn object
// Accepts ExecutionContext (unused but required by signature)
func pyIdFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		// Use object.NewError
		return object.NewError(constants.TypeError, constants.BuiltinsObjectIdArgCountError, len(args))
	}
	// Use reflection to get the pointer address.
	// Note: Pointer address isn't guaranteed stable by Go GC.
	ptr := reflect.ValueOf(args[0]).Pointer()
	return &object.Integer{Value: int64(ptr)}
}

// --- hasattr() ---
// Accepts ExecutionContext (needed for potential __getattr__ call)
// Accepts ExecutionContext (needed for potential __getattr__ call)
func pyHasAttrFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 2 {
		// Use object.NewError
		return object.NewError(constants.TypeError, constants.BuiltinsObjectHasattrArgCountError)
	}
	obj := args[0]
	nameObj, ok := args[1].(*object.String)
	if !ok {
		// Use object.NewError
		return object.NewError(constants.TypeError, constants.BuiltinsObjectHasattrNameTypeError)
	}

	// We reuse the getattr logic but catch the AttributeError.
	result := pyGetAttrFn(ctx, obj, nameObj)

	if errResult, isErr := result.(*object.Error); isErr {
		// Check if the error is specifically an AttributeError.
		if errResult.ErrorClass == object.AttributeErrorClass {
			return object.FALSE // AttributeError means it doesn't have it
		}
		// Python's hasattr suppresses most errors. Return FALSE on any error for simplicity.
		return object.FALSE
	}

	// If no error occurred, the attribute exists
	return object.TRUE
}

// func pyHasAttrFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
// 	if len(args) != 2 {
// 		// Use object.NewError
// 		return object.NewError(constants.TypeError, constants.BuiltinsObjectHasattrArgCountError)
// 	}
// 	obj := args[0]
// 	nameObj, ok := args[1].(*object.String)
// 	if !ok {
// 		// Use object.NewError
// 		return object.NewError(constants.TypeError, constants.BuiltinsObjectHasattrNameTypeError)
// 	}
// 	name := nameObj.Value

// 	// --- Attempt getattr using a helper (avoids raising error) ---
// 	// We need a way to try getting the attribute without immediately failing.
// 	// Let's reuse the getattr logic but catch the AttributeError.
// 	result := pyGetAttrFn(ctx, obj, nameObj) // Pass nameObj directly

// 	if errResult, isErr := result.(*object.Error); isErr {
// 		// Check if the error message indicates AttributeError (fragile check)
// 		if strings.HasSuffix(errResult.Message, constants.SingleQuote+name+constants.SingleQuote) && strings.Contains(errResult.Message, constants.BuiltinsObjectHasNoAttribute) {
// 			return object.FALSE // AttributeError means it doesn't have it
// 		}
// 		// If it was some other error during attribute lookup, maybe propagate it?
// 		// Python's hasattr suppresses most errors. Let's return FALSE on any error for now.
// 		return object.FALSE
// 	}

// 	// If no error occurred, the attribute exists
// 	return object.TRUE
// }

// --- getattr() ---
// Accepts ExecutionContext (needed for __getattr__ and method binding)
// Updated to handle optional default value.
func pyGetAttrFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	var obj object.Object
	var nameObj *object.String
	var defaultVal object.Object
	var name string
	hasDefault := false
	// Use a generic placeholder token for errors originating from getattr logic itself
	errReportingToken := object.NoToken

	switch len(args) {
	case 2:
		obj = args[0]
		nameObjArg, ok := args[1].(*object.String)
		if !ok {
			// Error for wrong argument type to getattr itself
			return object.NewErrorWithLocation(errReportingToken, constants.TypeError, constants.BuiltinsObjectGetattrNameTypeError)
		}
		nameObj = nameObjArg
	case 3:
		obj = args[0]
		nameObjArg, ok := args[1].(*object.String)
		if !ok {
			return object.NewErrorWithLocation(errReportingToken, constants.TypeError, constants.BuiltinsObjectGetattrNameTypeError)
		}
		nameObj = nameObjArg
		defaultVal = args[2]
		hasDefault = true
	default:
		return object.NewErrorWithLocation(errReportingToken, constants.TypeError, constants.BuiltinsObjectGetattrArgCountError, len(args))
	}
	name = nameObj.Value

	var foundValue object.Object = nil

	// --- Primary Lookup: AttributeGetter interface ---
	if getter, isGetter := obj.(object.AttributeGetter); isGetter {
		attrVal, found := getter.GetObjectAttribute(ctx, name) // Pass context
		if found {
			if errObj, isActualError := attrVal.(*object.Error); isActualError {
				isSpecificAttributeError := strings.Contains(errObj.Message, constants.AttributeErrorColon) ||
					(strings.Contains(errObj.Message, constants.BuiltinsObjectHasNoAttribute) && strings.HasSuffix(errObj.Message, constants.SingleQuote+name+constants.SingleQuote))

				if hasDefault && isSpecificAttributeError {
					// Fall through to return defaultVal. foundValue remains nil.
				} else {
					return errObj // Propagate the error.
				}
			} else {
				foundValue = attrVal
			}
		}
	}
	// --- End AttributeGetter Lookup ---

	// --- Handle __getattr__ if attribute not found by AttributeGetter ---
	if foundValue == nil {
		if inst, isInst := obj.(*object.Instance); isInst && inst.Class != nil {
			if getattrDunderObj, hasGetattrDunder := inst.Class.Methods[constants.DunderGetAttr]; hasGetattrDunder {
				if getattrDunder, isFunc := getattrDunderObj.(*object.Function); isFunc { // <<< TYPE ASSERTION
					boundGetattr := &object.BoundMethod{Instance: inst, Method: getattrDunder}
					resultFromDunder := object.ApplyBoundMethod(ctx, boundGetattr, []object.Object{nameObj}, errReportingToken)
					if object.IsError(resultFromDunder) {
						if errObj, isActualError := resultFromDunder.(*object.Error); isActualError {
							isSpecificAttributeError := strings.Contains(errObj.Message, constants.AttributeErrorColon) ||
								(strings.Contains(errObj.Message, constants.BuiltinsObjectHasNoAttribute) && strings.HasSuffix(errObj.Message, constants.SingleQuote+name+constants.SingleQuote))
							if hasDefault && isSpecificAttributeError {
								// Fall through to return defaultVal. foundValue remains nil.
							} else {
								return resultFromDunder
							}
						} else {
							return resultFromDunder // Should not happen if IsError is true
						}
					} else {
						foundValue = resultFromDunder
					}
				}
			}
		}
	}
	// --- End __getattr__ Handling ---

	// --- Return Result ---
	if foundValue != nil {
		return foundValue
	} else if hasDefault {
		return defaultVal
	} else {
		typeName := string(obj.Type())
		if inst, ok := obj.(*object.Instance); ok && inst.Class != nil {
			typeName = inst.Class.Name
		}
		// The error is about 'obj' not having 'name', so errReportingToken (generic) is appropriate.
		return object.NewErrorWithLocation(errReportingToken, constants.AttributeError, constants.BuiltinsObjectHasNoAttribute, typeName, name)
	}
}

// --- setattr() ---
// Accepts ExecutionContext (needed for potential __setattr__)
func pySetAttrFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 3 {
		// Use object.NewError
		return object.NewError(constants.TypeError, constants.BuiltinsObjectSetattrArgCountError)
	}
	obj := args[0]
	nameObj, ok := args[1].(*object.String)
	if !ok {
		// Use object.NewError
		return object.NewError(constants.TypeError, constants.BuiltinsObjectSetattrNameTypeError)
	}
	value := args[2]
	name := nameObj.Value

	// --- Check for __setattr__ method ---
	// TODO: Implement __setattr__ lookup and call via context.
	// If __setattr__ exists, call it:
	// result := ctx.Execute(setattrMethod, nameObj, value)
	// if object.IsError(result) { return result }
	// return object.NULL // __setattr__ handles the setting

	// --- Default setattr behavior (if no __setattr__) ---
	// Simplification: Only allow setting on object.Instance for now
	instance, ok := obj.(*object.Instance)
	if !ok {
		// TODO: Allow setting on classes, modules?
		// Use object.NewError
		return object.NewError(constants.AttributeError, constants.BuiltinsObjectSetattrAttributeError, obj.Type(), name)
	}

	// Lazily initialize environment if needed
	if instance.Env == nil {
		instance.Env = object.NewEnvironment()
	}
	instance.Env.Set(name, value) // Set attribute directly on instance's env
	return object.NULL            // setattr returns None
}

// --- delattr() ---
// Accepts ExecutionContext (needed for potential __delattr__)
func pyDelAttrFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 2 {
		// Use object.NewError
		return object.NewError(constants.TypeError, constants.BuiltinsObjectDelattrArgCountError)
	}
	obj := args[0]
	nameObj, ok := args[1].(*object.String)
	if !ok {
		// Use object.NewError
		return object.NewError(constants.TypeError, constants.BuiltinsObjectDelattrNameTypeError)
	}
	name := nameObj.Value

	// --- Check for __delattr__ method ---
	// TODO: Implement __delattr__ lookup and call via context.
	// If __delattr__ exists, call it:
	// result := ctx.Execute(delattrMethod, nameObj)
	// if object.IsError(result) { return result }
	// return object.NULL // __delattr__ handles the deletion

	// --- Default delattr behavior (if no __delattr__) ---
	// Simplification: Only allow deleting from object.Instance env for now
	instance, ok := obj.(*object.Instance)
	if !ok {
		// Use object.NewError
		return object.NewError(constants.AttributeError, constants.BuiltinsObjectDelattrNotSupported, obj.Type())
	}

	// Try deleting from instance's environment
	// Requires an `Unset` or `Delete` method on Environment
	if instance.Env == nil { // Cannot delete if env doesn't exist
		return object.NewError(constants.AttributeError, constants.ErrorFormat, name)
	}
	deleted := instance.Env.Delete(name) // Assuming Env.Delete method exists and returns bool
	if !deleted {
		// Attribute didn't exist in the instance env
		return object.NewError(constants.AttributeError, constants.ErrorFormat, name)
	}

	return object.NULL // delattr returns None
}

// --- callable() ---
// Accepts ExecutionContext (needed for potential __call__ check on instances)
func pyCallableFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		// Use object.NewError
		return object.NewError(constants.TypeError, constants.BuiltinsObjectCallableArgCountError, len(args))
	}
	obj := args[0]

	switch obj.(type) {
	case *object.Function:
		return object.TRUE
	case *object.Builtin:
		return object.TRUE
	case *object.Class: // object.Class is callable
		return object.TRUE
	// case *vm.Class: // vm.Class should also be callable
	//	return object.TRUE
	case *object.BoundMethod:
		return object.TRUE
	// case *vm.BoundMethod: // vm.BoundMethod should also be callable
	//	return object.TRUE
	case *object.Instance:
		// Check if the instance's class has a __call__ method
		// This requires access to the class definition.
		if inst, ok := obj.(*object.Instance); ok && inst.Class != nil {
			if _, methodOk := inst.Class.Methods[constants.DunderCall]; methodOk {
				return object.TRUE
			}
		}
		// case *vm.Instance: // Also check vm.Instance if distinct type
		//	if inst, ok := obj.(*vm.Instance); ok && inst.Class != nil {
		//		if _, methodOk := inst.Class.Methods["__call__"]; methodOk {
		//			return object.TRUE
		//		}
		//	}
	}
	return object.FALSE
}

// Pylearn: property(fget=None, fset=None, fdel=None, doc=None)
func pyPropertyFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) > 4 {
		return object.NewError(constants.TypeError, constants.BuiltinsPropertyTakesAtMost4Arguments)
	}

	// This is a simplified implementation that only handles the first argument (fget),
	// which is the most common use case for the @property decorator.
	var fget, fset, fdel, doc object.Object = object.NULL, object.NULL, object.NULL, object.NULL

	if len(args) >= 1 {
		fget = args[0]
	}
	if len(args) >= 2 {
		fset = args[1]
	}
	if len(args) >= 3 {
		fdel = args[2]
	}
	if len(args) == 4 {
		doc = args[3]
	}

	// The decorated function is passed as the first argument.
	if fget != object.NULL && !object.IsCallable(fget) {
		return object.NewError(constants.TypeError, constants.BuiltinsPropertyFgetArgumentMustBeACallable)
	}

	return &object.Property{
		FGet: fget,
		FSet: fset,
		FDel: fdel,
		Doc:  doc,
	}
}

// --- PLACEHOLDERS for complex object functions ---
// Update signatures to accept context

func pyDirFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	return object.NewError(constants.NotImplementedError, constants.BuiltinsObjectDirNotImplemented)
}
func pyVarsFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	return object.NewError(constants.NotImplementedError, constants.BuiltinsObjectVarsNotImplemented)
}
func pyGlobalsFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	return object.NewError(constants.NotImplementedError, constants.BuiltinsObjectGlobalsNotImplemented)
}
func pyLocalsFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	return object.NewError(constants.NotImplementedError, constants.BuiltinsObjectLocalsNotImplemented)
}

// This implementation is HIGHLY dependent on how the interpreter/VM can provide
func pySuperFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	numArgs := len(args)

	var startClassArg *object.Class
	var selfInstanceArg object.Object

	if numArgs == 0 {
		// --- ZERO-ARGUMENT SUPER() LOGIC ---
		selfInstanceArg = ctx.GetSuperSelf()
		startClassArg = ctx.GetSuperClass()

		if selfInstanceArg == nil || startClassArg == nil {
			return object.NewError(constants.RuntimeError, constants.BuiltinsObjectSuperNoArgsNoContextError)
		}

	} else if numArgs == 2 {
		// --- TWO-ARGUMENT SUPER(TYPE, OBJ) LOGIC ---
		typeArg, okType := args[0].(*object.Class)
		if !okType {
			return object.NewError(constants.TypeError, constants.BuiltinsObjectSuperFirstArgTypeError)
		}
		startClassArg = typeArg
		selfInstanceArg = args[1] // This is the 'self' or 'cls' for the super call

	} else {
		return object.NewError(constants.TypeError, constants.BuiltinsObjectSuperArgCountError, numArgs)
	}

	// --- VALIDATION AND SUPER OBJECT CREATION ---
	var targetTypeForMRO *object.Class
	if inst, isInst := selfInstanceArg.(*object.Instance); isInst {
		if inst.Class == nil {
			return object.NewError(constants.RuntimeError, constants.BuiltinsObjectSuperInstanceHasNoClass)
		}
		targetTypeForMRO = inst.Class
	} else if cls, isCls := selfInstanceArg.(*object.Class); isCls {
		targetTypeForMRO = cls
	} else {
		return object.NewError(constants.TypeError, constants.BuiltinsObjectSuperSecondArgTypeError, selfInstanceArg.Type())
	}

	// Create and return the Super object which will be used for attribute lookup.
	return &object.Super{
		SelfInstance: selfInstanceArg,
		StartClass:   startClassArg,
		TargetType:   targetTypeForMRO,
	}
}

func pyStaticMethodFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, constants.BuiltinsObjectStaticMethodArgCountError, len(args))
	}
	// The argument must be a callable, typically a Pylearn function.
	// For simplicity, we'll expect an object.Function for now.
	// Python's staticmethod can wrap other callables too.
	fn, ok := args[0].(*object.Function)
	if !ok {
		// Could also check for Builtin or other callable types if we want to support them.
		return object.NewError(constants.TypeError, constants.BuiltinsObjectStaticMethodArgTypeError, args[0].Type())
	}
	return &object.StaticMethod{Function: fn}
}

func pyClassMethodFn(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, constants.BuiltinsObjectClassMethodArgCountError, len(args))
	}
	fn, ok := args[0].(*object.Function)
	if !ok {
		return object.NewError(constants.TypeError, constants.BuiltinsObjectClassMethodArgTypeError, args[0].Type())
	}
	return &object.ClassMethod{Function: fn}
}

// --- Registration ---
// Ensure function signatures match object.BuiltinFunction
func init() {
	registerBuiltin(constants.BuiltinsIsInstanceFuncName, &object.Builtin{Fn: pyIsInstanceFn})
	registerBuiltin(constants.BuiltinsIsSubclassFuncName, &object.Builtin{Fn: pyIsSubclassFn})
	registerBuiltin(constants.BuiltinsIdFuncName, &object.Builtin{Fn: pyIdFn})
	registerBuiltin(constants.BuiltinsHasAttrFuncName, &object.Builtin{Fn: pyHasAttrFn})
	registerBuiltin(constants.BuiltinsGetAttrFuncName, &object.Builtin{Fn: pyGetAttrFn})
	registerBuiltin(constants.BuiltinsSetAttrFuncName, &object.Builtin{Fn: pySetAttrFn})
	registerBuiltin(constants.BuiltinsDelAttrFuncName, &object.Builtin{Fn: pyDelAttrFn})
	registerBuiltin(constants.BuiltinsCallableFuncName, &object.Builtin{Fn: pyCallableFn})
	// registerBuiltin(constants.BuiltinsPropertyFuncName, &object.Builtin{Fn: pyPropertyFn})
	registerBuiltin(constants.BuiltinsPropertyFuncName, &object.Builtin{Name: constants.BuiltinsPropertyFuncName, Fn: pyPropertyFn})

	// Placeholders
	registerBuiltin(constants.BuiltinsDirFuncName, &object.Builtin{Fn: pyDirFn})
	registerBuiltin(constants.BuiltinsVarsFuncName, &object.Builtin{Fn: pyVarsFn})
	registerBuiltin(constants.BuiltinsGlobalsFuncName, &object.Builtin{Fn: pyGlobalsFn})
	registerBuiltin(constants.BuiltinsLocalsFuncName, &object.Builtin{Fn: pyLocalsFn})
	registerBuiltin(constants.BuiltinsSuperFuncName, &object.Builtin{Fn: pySuperFn, Name: constants.BuiltinsSuperFuncName})
	registerBuiltin(constants.BuiltinsStaticMethodFuncName, &object.Builtin{Fn: pyStaticMethodFn, Name: constants.BuiltinsStaticMethodFuncName})
	registerBuiltin(constants.BuiltinsClassMethodFuncName, &object.Builtin{Fn: pyClassMethodFn, Name: constants.BuiltinsClassMethodFuncName})
}
