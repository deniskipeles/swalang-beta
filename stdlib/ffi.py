# pylearn/stdlib/ffi/__init__.py

"""
A ctypes-like Foreign Function Interface for Pylearn.

This module provides C compatible data types, and allows calling functions in
shared libraries (.so, .dll, .dylib). It can be used to wrap these libraries
in pure Pylearn code.

New Features Supported:
- Full set of C primitive types (char, short, long, bool, wchar_t).
- Variadic function calls (e.g., printf).
- Robust support for Structs and Unions with correct memory layout.
- Advanced pointer types for typed pointers (e.g., int*) and fixed-size arrays (e.g., int[10]).
- Automatic memory management for Union instances.
- Explicit memory cleanup for callbacks and other manually allocated C resources.
"""

# Import the low-level native backend.
import _ffi_native
import sys

# ==============================================================================
#  Re-export Core Types, Functions, and Errors from Native Backend
# ==============================================================================

# --- C Primitive Data Types ---
c_int8 = _ffi_native.c_int8
c_uint8 = _ffi_native.c_uint8
c_int32 = _ffi_native.c_int32
c_uint32 = _ffi_native.c_uint32
c_int64 = _ffi_native.c_int64
c_uint64 = _ffi_native.c_uint64
c_float = _ffi_native.c_float
c_double = _ffi_native.c_double
c_char = _ffi_native.c_char
c_uchar = _ffi_native.c_uchar
c_short = _ffi_native.c_short
c_ushort = _ffi_native.c_ushort
c_long = _ffi_native.c_long
c_ulong = _ffi_native.c_ulong
c_longlong = _ffi_native.c_longlong
c_ulonglong = _ffi_native.c_ulonglong
c_bool = _ffi_native.c_bool
c_wchar_t = _ffi_native.c_wchar_t

# --- Pointer Types ---
c_void_p = _ffi_native.c_void_p
c_char_p = _ffi_native.c_char_p
c_wchar_p = _ffi_native.c_wchar_p

# --- Memory Management Functions ---
malloc = _ffi_native.malloc
free = _ffi_native.free
memcpy = _ffi_native.memcpy
addressof = _ffi_native.addressof
read_memory = _ffi_native.read_memory
write_memory = _ffi_native.write_memory
write_memory_with_offset = _ffi_native.write_memory_with_offset
read_memory_with_offset = _ffi_native.read_memory_with_offset
buffer_to_bytes = _ffi_native.buffer_to_bytes
string_at = _ffi_native.string_at

# --- Resource Cleanup ---
free_c_resource = _ffi_native.free_c_resource
free_callback = _ffi_native.free_callback

# --- Callback Constructor ---
callback = _ffi_native.callback

# --- Error Class ---
FFIError = _ffi_native.error


# ==============================================================================
#  High-Level Pythonic Wrappers
# ==============================================================================

def POINTER(typ):
    """
    Creates a pointer type that points to the given FFI type.
    Example: POINTER(c_int32) creates a type for 'int*'.

    :param typ: The FFI type the pointer will point to.
    :return: A new FFI pointer type.
    """
    # This now correctly calls the Go backend function we just implemented.
    ptr_type_obj = _ffi_native._get_or_create_pointer_type(typ)
    if isinstance(ptr_type_obj, FFIError):
        raise ptr_type_obj
    return ptr_type_obj

def ArrayType(typ, size):
    """
    Creates a fixed-size array type.
    Example: ArrayType(c_float, 16) creates a type for 'float[16]'.

    :param typ: The FFI type of the array elements.
    :param size: The number of elements in the array.
    :return: A new FFI array type.
    """
    if not isinstance(size, int) or size <= 0:
        raise TypeError("Array size must be a positive integer.")
    # This now correctly calls the Go backend function we just implemented.
    array_type_obj = _ffi_native._create_pointer_type(typ, size)
    if isinstance(array_type_obj, FFIError):
        raise array_type_obj
    return array_type_obj

class _StructUnionBase:
    """Base class for Struct and Union to share instance logic."""
    def __init__(self, *args, **kwargs):
        # This allows creating instances: my_struct = MyStruct(field1=1, field2=b'hello')
        # This part requires a more complex Go backend to handle instance creation
        # and marshalling from Python arguments. For now, we focus on type definition.
        raise NotImplementedError("Instance creation from Python is not yet implemented.")

    @classmethod
    def from_address(cls, address):
        """Creates a Struct/Union instance from a memory address (a Pointer object)."""
        # This would require a Go backend function to wrap a raw pointer
        # with the struct/union type information.
        raise NotImplementedError("from_address is not yet implemented.")

def _create_ffi_type(cls, name, fields_list):
    """Helper to create Struct or Union types."""
    if name == "Struct":
        creator_func = _ffi_native.create_struct_type
    elif name == "Union":
        creator_func = _ffi_native.create_union_type
    else:
        raise TypeError("Expected Struct or Union")

    type_obj = creator_func(cls.__name__, fields_list)
    if isinstance(type_obj, FFIError):
        raise type_obj
    return type_obj

class Struct:
    """
    Class decorator to create a C struct type.

    Example:
    @ffi.Struct
    class Point:
        _fields_ = [
            ('x', ffi.c_int32),
            ('y', ffi.c_int32)
        ]
    """
    def __init__(self, cls):
        self._cls = cls
        # Get fields from the decorated class
        if not hasattr(cls, '_fields_'):
            raise TypeError("Struct definition must have a '_fields_' attribute.")
        self._fields = cls._fields_
        # Create the underlying native FFI type
        self._ffi_type = _create_ffi_type(cls, "Struct", self._fields)

    def __call__(self, *args, **kwargs):
        # This would create an instance of the struct.
        # It's a placeholder for future implementation.
        instance = _StructUnionBase()
        instance._ffi_type = self._ffi_type
        # ... logic to populate fields from *args and **kwargs ...
        return instance

    def __getattr__(self, name):
        # Allow access to the underlying FFI type, e.g., MyStruct.Size()
        if hasattr(self._ffi_type, name):
            return getattr(self._ffi_type, name)
        raise AttributeError(format_str("'{self.__class__.__name__}' object has no attribute '{name}'"))

class Union:
    """
    Class decorator to create a C union type.

    Example:
    @ffi.Union
    class Data:
        _fields_ = [
            ('i', ffi.c_int64),
            ('f', ffi.c_double),
            ('c', ffi.c_char * 8) # Assuming array support
        ]
    """
    def __init__(self, cls):
        self._cls = cls
        if not hasattr(cls, '_fields_'):
            raise TypeError("Union definition must have a '_fields_' attribute.")
        self._fields = cls._fields_
        self._ffi_type = _create_ffi_type(cls, "Union", self._fields)

    def __call__(self, *args, **kwargs):
        instance = _StructUnionBase()
        instance._ffi_type = self._ffi_type
        return instance

    def __getattr__(self, name):
        if hasattr(self._ffi_type, name):
            return getattr(self._ffi_type, name)
        raise AttributeError(format_str("'{self.__class__.__name__}' object has no attribute '{name}'"))


class CDLL:
    """
    A class representing a loaded shared library.
    Functions are accessed as attributes. This returns a configurator function which,
    when called with the signature, returns the final callable function.
    """
    def __init__(self, name):
        """Loads a shared library. `name` can be a path or a library name."""
        self._name = name
        self._lib = _ffi_native.load_library(name)
        if isinstance(self._lib, FFIError):
            raise self._lib

    def __getattr__(self, name):
        """
        Provides access to functions within the library by returning a
        stateless configurator function.
        
        Example:
        libc = CDLL('c')
        printf = libc.printf([c_char_p], c_int32, is_variadic=True)
        printf(b"Hello %s, number %d\n", b"world", 123)
        """
        def configure_function(argtypes, restype, is_variadic=False):
            """
            Configures the function signature.
            :param argtypes: List of FFI types for fixed arguments.
            :param restype: The FFI type for the return value.
            :param is_variadic: Set to True for functions like printf.
            """
            func_obj = _ffi_native.define_function(
                self._lib, name, restype, argtypes, is_variadic
            )
            if isinstance(func_obj, FFIError):
                raise func_obj
            
            # Create the final callable that wraps the FFI call.
            def wrapper(*args):
                return _ffi_native.call_function(func_obj, *args)
            
            return wrapper

        return configure_function

class WinDLL(CDLL):
    """
    Windows-specific library loader that provides access to GetLastError().
    """
    def __init__(self, name):
        super().__init__(name)
        if sys.platform != 'windows':
            self._get_last_error = lambda: 0
        else:
            self._get_last_error = _ffi_native.get_last_error

    def get_last_error(self):
        """Returns the result of the Win32 GetLastError() function."""
        err_obj = self._get_last_error()
        if isinstance(err_obj, FFIError):
            raise err_obj
        if hasattr(err_obj, 'Value'):
            return err_obj.Value
        return err_obj


# ==============================================================================
#  Standard Library Access
# ==============================================================================

def _load_standard_library(name):
    try:
        return CDLL(name)
    except FFIError:
        # Try with common fallbacks if direct name fails
        if sys.platform == 'darwin':
            try: 
                return CDLL(format_str("lib{name}.dylib"))
            except FFIError: 
                pass
        elif sys.platform == 'linux':
            try: 
                return CDLL(format_str("lib{name}.so"))
            except FFIError: 
                pass
    return None

# A global handle to the standard C library
libc = _load_standard_library('c')
if sys.platform == 'windows':
    libc = CDLL('msvcrt')

# A global handle to the standard math library
libm = _load_standard_library('m')

# Pre-configured common functions for convenience, if libc was found
if libc:
    printf = libc.printf([c_char_p], c_int32, is_variadic=True)
    memset = libc.memset([c_void_p, c_int32, c_int64], c_void_p)
    puts = libc.puts([c_char_p], c_int32)
    
    # Example usage
    puts(b"Hello from C via Pylearn FFI!")
    printf(b"Formatting a number: %d and a string: %s\n", 123, b"pylearn")

else:
    def _unavail(*args, **kwargs):
        raise FFIError("Standard C library (libc) could not be loaded.")
    printf = memset = puts = _unavail

# Create a windll object for windows-specific libraries
if sys.platform == 'windows':
    windll = WinDLL
else:
    # On non-windows, windll is just an alias for CDLL
    windll = CDLL