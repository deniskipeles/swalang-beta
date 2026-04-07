//go:build sw
// constants/interpreter.go
package constants

const (
	ContextGoExecuteCall                                 = "__call__"
	ContextGoExecuteTypeError                            = "TypeError"
	ContextGoExecuteReturnErrorValue1ObjectIsNotCallable = "'%s' object is not callable"
)

// pylearn/internal/interpreter/apply.go
const (
	InterpreterApplyFunctionKwargMultipleValuesError    = "%s() got multiple values for argument '%s'"
	InterpreterApplyFunctionUnexpectedKwargError        = "%s() got an unexpected keyword argument '%s'"
	InterpreterApplyFunctionMissingPositionalArgError   = "%s() missing 1 required positional argument: '%s'"
	InterpreterApplyFunctionTooManyPositionalArgsError  = "%s() takes %d positional arguments but %d were given"
	InterpreterApplyFunctionInitReturnNoneError         = "__init__() should return None, not %s"
	InterpreterApplyFunctionTakesNoArgumentsError       = "%s() takes no arguments"
	InterpreterApplyFunctionBoundGoMethodTypeError      = "BoundGoMethod has incorrect instance type"
	InterpreterApplyFunctionUnsupportedNativeMethodType = "Unsupported native method type"
	InterpreterApplyFunctionUnexpectedKwargInBuiltin    = "%s() got an unexpected keyword argument '%s'"
	InterpreterApplyFunctionObjectNotCallable           = "'%s' object is not callable"

	FunctionLiteralFunctionPlaceholder = "<function>"
)

// pylearn/internal/interpreter/async_setup.go
const (
	InterpreterAsyncSetupSleepArgCountError     = "async_builtins.sleep() takes 1 argument (duration_seconds)"
	InterpreterAsyncSetupSleepDurationTypeError = "async_builtins.sleep() duration must be an integer"
	InterpreterAsyncSetupRuntimeNotInitialized  = "Pylearn async runtime not initialized. Cannot use aio.sleep."
	InterpreterAsyncSetupModulePath             = "<builtin_async>"

	BuiltinsAsyncBuiltinsModule        = "async_builtins"
	BuiltinsAsyncBuiltinsSleepFuncName = "async_builtins.sleep"
	BuiltinsSleepFuncName              = "sleep"
)

// pylearn/internal/interpreter/eval_await.go
const (
	AwaitUsedButNoAsyncRuntimeIsAvailable = "await used, but no async runtime is available"
)

// pylearn/internal/interpreter/dynamic_loader.go
const (
	InterpreterDynamicLoaderLoadModulePathArgCountError = "load_module_from_path() takes exactly 1 argument (path), got %d"
	InterpreterDynamicLoaderLoadModulePathArgTypeError  = "load_module_from_path() argument must be a string, not %s"
	InterpreterDynamicLoaderRelativePathError           = "ImportError: cannot resolve relative path '%s'; current script directory not set"
	InterpreterDynamicLoaderAbsPathError                = "ImportError: could not determine absolute path for '%s': %v"
	InterpreterDynamicLoaderDynamicModule               = "__dynamic_module__"
	InterpreterDynamicLoaderCircularImportPlaceholder   = "ImportError: circular import detected for module at path '%s' (placeholder found)"
	InterpreterDynamicLoaderCircularImportBeingImported = "ImportError: circular import - module '%s' (%s) is already being imported (placeholder found)"
	InterpreterDynamicLoaderFailedToReadFile            = "ImportError: failed to read module '%s' from '%s': %v"
	InterpreterDynamicLoaderSyntaxErrorInModule         = "SyntaxError in module '%s' (%s):\n"
	InterpreterDynamicLoaderSyntaxErrorFormat           = "\t%s\n"
	InterpreterDynamicLoaderErrorExecutingModule        = "Error during execution of module '%s' (%s): %s"

	ErrorDuringExecutionModule = "Error during execution of module '%s'"
)

// pylearn/internal/interpreter/eval_expressions.go
const (
	EvalExpressionsNameNotDefined                     = "name '%s' is not defined"
	EvalExpressionsPropagatedFromIsTruthy             = " (propagated from IsTruthy): %v"
	EvalExpressionsBadOperandTypeUnaryMinus           = "bad operand type for unary -: '%s'"
	EvalExpressionsBadOperandTypeUnaryPlus            = "bad operand type for unary +: '%s'"
	EvalExpressionsUnknownPrefixOperator              = "unknown prefix operator: %s"
	EvalExpressionsStrAttrNotCallable                 = "__contains__ attribute of type %s is not callable"
	EvalExpressionsStrReturnedNonBool                 = "__contains__ returned non-bool (type %s)"
	EvalExpressionsInStringRequiresString             = "'in <string>' requires string as left operand, not %s"
	EvalExpressionsArgNotIterable                     = "argument of type '%s' is not iterable for 'in' operator (or __contains__ not defined)"
	EvalExpressionsListIndexMustBeInteger             = "list indices must be integers, not %s"
	EvalExpressionsTupleIndexMustBeInteger            = "tuple indices must be integers, not %s"
	EvalExpressionsBytesIndexMustBeInteger            = "bytes indices must be integers, not %s"
	EvalExpressionsStringIndexMustBeInteger           = "string indices must be integers, not %s"
	EvalExpressionsStringIndexOutOfRange              = "string index out of range"
	EvalExpressionsObjectNotSubscriptable             = "'%s' object is not subscriptable"
	EvalExpressionsListIndexOutOfRange                = "list index out of range"
	EvalExpressionsTupleIndexOutOfRange               = "tuple index out of range"
	EvalExpressionsBytesIndexOutOfRange               = "bytes index out of range"
	EvalExpressionsFailedToHashKey                    = "failed to hash key: %v"
	EvalExpressionsKeyError                           = "KeyError: %s"
	EvalExpressionsListAssignmentIndexOutOfRange      = "list assignment index out of range"
	EvalExpressionsTupleDoesNotSupportItemAssignment  = "'tuple' object does not support item assignment"
	EvalExpressionsBytesDoesNotSupportItemAssignment  = "'bytes' object does not support item assignment"
	EvalExpressionsStrDoesNotSupportItemAssignment    = "'str' object does not support item assignment"
	EvalExpressionsObjectDoesNotSupportItemAssignment = "'%s' object does not support item assignment"
	EvalExpressionsComprehensionIfConditionError      = "error in comprehension 'if' condition: %v"
	EvalExpressionsSetHashFailed                      = "failed to hash element for set: %v"
	EvalExpressionsZeroCannotBeRaisedNegativePower    = "0 cannot be raised to a negative power"
	EvalExpressionsIntegerPowerResultTooLarge         = "integer power result too large for int64"
	EvalExpressionsDivisionByZero                     = "division by zero"
	EvalExpressionsIntegerModuloByZero                = "integer modulo by zero"
	EvalExpressionsFloatModuloRequiresMath            = "object.Float modulo requires math import"
	EvalExpressionsStringUnsupportedOperandType       = "TypeError: unsupported operand type(s) for %s: 'str' and 'str'"
	EvalExpressionsSyntaxErrorKeywordArgumentRepeated = "keyword argument '%s' repeated"
	EvalExpressionsTypeErrorArgumentAfterStarStar     = "argument after ** must be a mapping, not %s"
	EvalExpressionsTypeErrorKeywordsMustBeStrings     = "keywords must be strings (in **%s mapping)"
	EvalExpressionsTypeErrorMultipleValuesKwarg       = "%s() got multiple values for keyword argument '%s'"
	EvalExpressionsPositionalArgFollowsKeywordArg     = "positional argument follows keyword argument"
	EvalExpressionsTypeErrorUnsupportedOperandType    = "unsupported operand type(s) for %s: '%s' and '%s'"
	EvalExpressionsTypeErrorCannotSetAttribute        = "'%s' object cannot set attribute '%s'"
	EvalExpressionsSyntaxErrorCannotAssignTo          = "cannot assign to %s"
	EvalExpressionsObjectNotAwaitable                 = "object %s is not awaitable"
	EvalExpressionsAsyncResultNilGoResult             = "Pylearn AsyncResultWrapper contains a nil GoAsyncResult"
	EvalExpressionsAwaitGoError                       = "(from await): %v"
	EvalExpressionsAwaitUnexpectedGoType              = "(await): Go async operation returned unexpected non-Pylearn type %T"
	EvalExpressionsFloatDivisionByZero                = "division by zero"
	EvalExpressionsUnhashableType                     = "unhashable type: '%s'"
	EvalExpressionsFailedToHashElementForSet          = "failed to hash element for set: %v"
)

// pylearn/internal/interpreter/eval_slices.go
const (
	InterpreterEvalSlicesObjectNotSliceable    = "'%s' object is not sliceable"
	InterpreterEvalSlicesIndexTypeError        = "slice indices must be integers or None, not %s"
	InterpreterEvalSlicesStepCannotBeZeroError = "slice step cannot be zero"
)

// pylearn/internal/interpreter/eval_statements.go
const (
	InterpreterEvalStatementsBreakOutsideLoop                               = "'%s' outside loop"
	InterpreterEvalStatementsIsTruthyPropagatedError                        = " (propagated from IsTruthy): %v"
	InterpreterEvalStatementsUnpackTooManyValues                            = "too many values to unpack (expected %d)"
	InterpreterEvalStatementsUnpackingError                                 = "unpacking error: expected %d values, got %d"
	InterpreterEvalStatementsNoLoopVariableSpecified                        = "no loop variable specified"
	InterpreterEvalStatementsSuperclassNotDefined                           = "superclass '%s' not defined for class '%s'"
	InterpreterEvalStatementsSuperclassNotAClass                            = "superclass '%s' for class '%s' is not a class (got %s)"
	InterpreterEvalStatementsObjectClassNotInitialized                      = "Pylearn's root 'object' class not initialized."
	InterpreterEvalStatementsWithEnterAttributeError                        = "'%s' object has no attribute '__enter__'"
	InterpreterEvalStatementsWithEnterNotCallable                           = "'%s' object's __enter__ is not callable"
	InterpreterEvalStatementsWithExitAttributeError                         = "'%s' object has no attribute '__exit__'"
	InterpreterEvalStatementsWithExitNotCallable                            = "'%s' object's __exit__ is not callable"
	InterpreterEvalStatementsTruthinessOfExitError                          = "(evaluating truthiness of __exit__ result): %v"
	InterpreterEvalStatementsBareRaiseNotImplemented                        = "bare 'raise' is not yet implemented (and only valid inside an except block)"
	InterpreterEvalStatementsExceptionsMustDerive                           = "exceptions must derive from BaseException"
	InterpreterEvalStatementsExpectedIndentedBlock                          = "expected indented block after 'try:'"
	InterpreterEvalStatementsExpectedExceptOrFinally                        = "expected 'except' or 'finally' block after 'try' body"
	STRINGFORMATER_ObjectDoesNotSupportItemDeletion                         = "'%s' object does not support item deletion"
	InvalidDeletionTarget                                                   = "invalid deletion target"
	PropagatedFromIsTruthyInAssert_VERBFORMATER                             = "(propagated from IsTruthy in assert): %v"
	StrBuiltinName                                                          = "str"
	UnsupportedOperandTypesFor_STRINGFORMATER_STRINGFORMATER_STRINGFORMATER = "unsupported operand type(s) for %s: '%s' and '%s'"
	NotEnoughValuesToUnpackExpected_NUMBERFORMATER_Got_NUMBERFORMATER       = "not enough values to unpack (expected %d, got %d)"
	TooManyValuesToUnpackExpected_NUMBERFORMATER                            = "too many values to unpack (expected %d)"
	STRINGFORMATER_ObectDoesNotSupportItemAssignment                        = "'%s' object does not support item assignment"
	STRINGFORMATER_ObectHasNoAttribute_STRINGFORMATER_OrCannotBeAssignedTo  = "'%s' object has no attribute '%s' or cannot be assigned to"
	CannotAssignTo_STRINGFORMATER                                           = "cannot assign to %s"
)

// pylearn/internal/interpreter/interpreter.go
const (
	InterpreterEvalParamDefaultError                                          = "Error evaluating default for parameter '%s': %s"
	InterpreterEvalObjectNotAwaitable                                         = "object %s is not awaitable"
	InterpreterEvalAsyncResultNilGoError                                      = "Pylearn AsyncResultWrapper contains a nil GoAsyncResult"
	InterpreterEvalAwaitFromAwaitError                                        = "(from await): %v"
	InterpreterEvalAwaitGoReturnError                                         = "(await): Go async operation returned unexpected non-Pylearn type %T"
	InterpreterEvalSyntaxErrorNode                                            = "evaluation not implemented for AST node type %T"
	ErrorEvaluatingDefaultForParameter_STRINGFORMATER_InLambda_STRINGFORMATER = "Error evaluating default for parameter '%s' in lambda: %s"
)

// pylearn/internal/interpreter/modules.go
const (
	InterpreterModulesWarnAbsolutePath                                                           = "Warning: could not get absolute path for %s: %v\n"
	InterpreterModulesComplexImportError                                                         = "complex module paths in 'from' import not yet supported (e.g., 'from a.b import c')"
	InterpreterModulesLoadedNotModuleError                                                       = "InternalError: loaded module '%s' is not a Module object (type %s)"
	InterpreterModulesImportNameError                                                            = "ImportError: cannot import name '%s' from module '%s'"
	InterpreterModulesCannotDetermineCWD                                                         = "cannot determine current directory to resolve module '%s'"
	InterpreterModulesNoModuleFoundPathResolution                                                = "No module named '%s' (path resolution failed: %s)"
	InterpreterModulesCircularImportDetected                                                     = "Circular import detected for module '%s' at %s"
	InterpreterModulesCircularImportBeingImported                                                = "circular import - module '%s' (%s) is already being imported"
	InterpreterModulesNoModuleFoundFileNotFound                                                  = "No module named '%s' (file not found: %s)"
	InterpreterModulesCouldNotReadModule                                                         = "Could not read module '%s' from '%s': %v"
	InterpreterModulesSyntaxErrorImportedModule                                                  = "SyntaxError in imported module '%s' (%s):\n"
	InterpreterModulesSyntaxErrorFormat                                                          = "\t%s\n"
	InterpreterModulesErrorImportModuleFull                                                      = "Error during import of module '%s' (%s): %s"
	InterpreterModulesErrorImportModuleLine                                                      = "Error during import of module '%s' (File \"%s\", line %d): %s"
	WarningFoundPluginFor_STRINGFORMATER_At_STRINGFORMATER_ButFailedToLoad_VERBFORMATER_NEXTLINE = "Warning: Found plugin for '%s' at '%s' but failed to load: %v\n"
	NoModuleNamed_STRINGFORMATER                                                                 = "No module named '%s'"
	CouldNotDetermineAbsolutePathForModule_STRINGFORMATER                                        = "Could not determine absolute path for module '%s'"

	Init_DOT_Py                   = "__init__.py"
	DOT_Py                        = ".py"
	PluginPathEnvironmentVariable = "PYLEARN_PLUGIN_PATH"
	PluginsDirectory              = "plugins"
	DOT_OurLanguageDirectory      = ".pylearn"

	USR_SLASH_LOCAL_SLASH_LIB_SLASH_OurLanguageDirectory_SLASH_PLUGINS = "/usr/local/lib/pylearn/plugins"
	ModulesDirectoryForThirdPartyPackagesInstalled                     = "modules"
	LibDirectoryForProjectSpecificModules                              = "lib"
)
