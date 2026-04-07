// In internal/stdlib/pyre/re_funcs.go

package pyre

import (
	"regexp"

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)

// =========================== START: New Function ===========================
// pyReEscape implements re.escape(string)
func pyReEscape(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, "re.escape() takes exactly 1 argument (string)")
	}
	patternStr, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, "re.escape() argument must be a string")
	}
	// Go's regexp.QuoteMeta does exactly what Python's re.escape does.
	return &object.String{Value: regexp.QuoteMeta(patternStr.Value)}
}
// =========================== END: New Function ===========================

func pyReCompile(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 { // Simplified: no flags yet
		return object.NewError(constants.TypeError, "re.compile() takes 1 argument (pattern)")
	}
	patternStr, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, "re.compile() argument must be a string")
	}

	// =========================== START: Modified Block ===========================
	// THE FIX: To support re.DOTALL (which makes '.' match newlines), we
	// prepend the `(?s)` flag to the Go regex pattern. This is a common way
	// to enable flags directly in the pattern string itself.
	goPattern := "(?s)" + patternStr.Value
	re, err := regexp.Compile(goPattern)
	// =========================== END: Modified Block ===========================

	if err != nil {
		return object.NewError(constants.ValueError, "invalid regular expression: %s", err)
	}

	return &object.Pattern{Regex: re, Pattern: patternStr.Value}
}

// ... (compileAndCall, pyReSearch, pyReMatch, pyReFindAll remain the same) ...
func compileAndCall(methodName string, ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) < 2 {
		return object.NewError(constants.TypeError, "re.%s() requires at least 2 arguments (pattern, string)", methodName)
	}
	patternStr, ok := args[0].(*object.String)
	if !ok { return object.NewError(constants.TypeError, "re.%s() pattern must be a string", methodName) }
	
	stringToSearch, ok := args[1].(*object.String)
	if !ok { return object.NewError(constants.TypeError, "re.%s() second argument must be a string", methodName) }

	compiledPatternObj := pyReCompile(ctx, patternStr)
	if err, isErr := compiledPatternObj.(*object.Error); isErr {
		return err
	}
	
	compiledPattern := compiledPatternObj.(*object.Pattern)

	method, found := compiledPattern.GetObjectAttribute(ctx, methodName)
	if !found {
		return object.NewError(constants.InternalError, "internal re error: method '%s' not found on Pattern object", methodName)
	}
	
	methodBuiltin, ok := method.(*object.Builtin)
	if !ok {
		return object.NewError(constants.InternalError, "internal re error: attribute '%s' is not a callable method", methodName)
	}
	return methodBuiltin.Fn(ctx, compiledPattern, stringToSearch)
}

func pyReSearch(ctx object.ExecutionContext, args ...object.Object) object.Object {
	return compileAndCall("search", ctx, args...)
}

func pyReMatch(ctx object.ExecutionContext, args ...object.Object) object.Object {
	return compileAndCall("match", ctx, args...)
}

func pyReFindAll(ctx object.ExecutionContext, args ...object.Object) object.Object {
	return compileAndCall("findall", ctx, args...)
}


var (
	// =========================== START: Add new Builtin ===========================
	ReEscape  = &object.Builtin{Name: "re.escape", Fn: pyReEscape}
	// =========================== END: Add new Builtin =============================
	ReCompile = &object.Builtin{Name: "re.compile", Fn: pyReCompile}
	ReSearch  = &object.Builtin{Name: "re.search", Fn: pyReSearch}
	ReMatch   = &object.Builtin{Name: "re.match", Fn: pyReMatch}
	ReFindAll = &object.Builtin{Name: "re.findall", Fn: pyReFindAll}
)
