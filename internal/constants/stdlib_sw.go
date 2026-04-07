//go:build sw
// pylearn/internal/constants/stdlib.go
package constants

// internal/stdlib/pyaio/module.go
const (
	StdlibAioModuleName = "aio"
	StdlibAioModulePath = "<builtin_aio>"
	StdlibAioSleepName  = "sleep"
	StdlibAioGatherName = "gather"
)

// internal/stdlib/pyaio/aio_funcs.go
const (
	StdlibAioDotSleepFuncName = "aio.sleep"
	StdlibAioDotGatherFuncName = "aio.gather"

	StdlibAioSleepArgCountError        = "aio.sleep() takes exactly 1 argument (duration_seconds), got %d"
	StdlibAioSleepDurationTypeError    = "aio.sleep() argument 'duration_seconds' must be a number, not %s"
	StdlibAioSleepNegativeDurationError = "sleep length must be non-negative"
	StdlibAioSleepRuntimeNotInitialized = "Pylearn async runtime not initialized. Cannot use aio.sleep."
	StdlibAioGatherAwaitableTypeError  = "all arguments to aio.gather() must be awaitables (AsyncResult objects), got %s at position %d"
	StdlibAioGatherNilGoAsyncResult    = "encountered a Pylearn AsyncResultWrapper with a nil GoAsyncResult at position %d"
	StdlibAioGatherRuntimeNotInitialized = "Pylearn async runtime not initialized. Cannot use aio.gather."
	StdlibAioGatherUnexpectedGoType    = "aio.gather: unexpected Go type '%T' in gathered results at index %d. Expected Pylearn object."
)