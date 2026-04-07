# pylearn/stdlib/py

"""
A ctypes-like Foreign Function Interface for Pylearn.

This module provides C compatible data types, and allows calling functions in
shared libraries (.so, .dll, .dylib). It can be used to wrap these libraries
in pure Pylearn code.
"""

# Import the low-level native backend.
import _ffi_native

# Re-export the core C data types from the native module.
c_int8 = _ffi_native.c_int8
c_uint8 = _ffi_native.c_uint8
c_int32 = _ffi_native.c_int32
c_uint32 = _ffi_native.c_uint32
c_int64 = _ffi_native.c_int64
c_uint64 = _ffi_native.c_uint64
c_float = _ffi_native.c_float
c_double = _ffi_native.c_double
c_void_p = _ffi_native.c_void_p
c_char_p = _ffi_native.c_char_p

# Re-export memory management functions and the custom error class.
# Note: The additional memory functions will not be used by this
# simple wrapper but are available if needed.
malloc = _ffi_native.malloc
free = _ffi_native.free
memcpy = _ffi_native.memcpy
addressof = _ffi_native.addressof
read_memory = _ffi_native.read_memory
write_memory = _ffi_native.write_memory
write_memory_with_offset = _ffi_native.write_memory_with_offset
read_memory_with_offset = _ffi_native.read_memory_with_offset
callback = _ffi_native.callback
FFIError = _ffi_native.error

buffer_to_bytes = _ffi_native.buffer_to_bytes


class CDLL:
    """
    A class representing a loaded shared library.

    Functions are accessed as attributes. This returns a temporary configurator
    function which, when called with the signature, returns the final
    callable function pointer.
    """
    def __init__(self, name):
        """Loads a shared library. `name` can be a path or a library name."""
        self._name = name
        self._lib = _ffi_native.load_library(name)
        if isinstance(self._lib, FFIError):
            raise self._lib
        # NO CACHE HERE. The cache in the original simple wrapper was the bug.
        # The Go backend handles caching the prepared Function objects.

    def __getattr__(self, name):
        """
        Provides access to functions within the library by returning a
        configurator function.
        """
        # Always return a new configurator function. It is stateless.
        def configure_function(argtypes, restype):
            # Call the native Go function to prepare the libffi interface.
            # This is cached on the Go side, so it's efficient.
            func_obj = _ffi_native.define_function(self._lib, name, restype, argtypes)
            if isinstance(func_obj, FFIError):
                raise func_obj
            
            # Create the final callable Pylearn function that wraps the FFI call.
            def wrapper(*args):
                return _ffi_native.call_function(func_obj, *args)
            
            return wrapper

        # Return the configurator function. The user must call this.
        return configure_function



# Represents a C function pointer that must be configured before use.
class _FuncPtr:
    def __init__(self, lib, name):
        self._lib = lib
        self._name = name
        self._native_func = None
        self._callable_wrapper = None
        
    def configure(self, argtypes, restype):
        """
        Configures the function signature. This must be called before the
        function pointer can be invoked.
        
        :param argtypes: A list of FFI types for the arguments.
        :param restype: The FFI type for the return value.
        """
        if not isinstance(argtypes, list):
            raise TypeError("argtypes must be a list of FFI types")

        if self._native_func is not None:
            # Optionally allow reconfiguration, or raise an error.
            # For simplicity, we'll just re-prepare.
            pass

        func_obj = _ffi_native.define_function(self._lib, self._name, restype, argtypes)
        if isinstance(func_obj, FFIError):
            raise func_obj
        self._native_func = func_obj

        # Create the Pylearn callable that invokes the native backend.
        def wrapper(*args):
            native_func = self._native_func
            return _ffi_native.call_function(native_func, *args)
        
        self._callable_wrapper = wrapper

    def __call__(self, *args):
        """
        Calls the C function. Raises an error if .configure() has not been called.
        """
        if self._callable_wrapper is None:
            raise FFIError(format_str("Function '{self._name}' has not been configured. Call .configure(argtypes, restype) first."))
        
        return self._callable_wrapper(*args)


# A global handle to the standard C library for functions like memset
# This is a common pattern for FFIs.
try:
    _libc = CDLL("libc.so.6")
except FFIError:
    try:
        # Fallback for macOS
        _libc = CDLL("libc.dylib")
    except FFIError:
        _libc = None

# Define memset if libc was loaded successfully
if _libc is not None:
    # C Signature: void *memset(void *s, int c, size_t n);
    memset = _libc.memset(
        [c_void_p, c_int32, c_int64], # s, c, n
        c_void_p # returns the pointer s
    )
else:
    # Define a placeholder if libc couldn't be loaded
    def memset(ptr, val, size):
        raise RuntimeError("Standard C library (libc) could not be loaded, memset is unavailable.")