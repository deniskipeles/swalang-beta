# Swalang FFI Guide

The `ffi` module is a powerful tool for calling functions in external shared libraries directly from Swalang. It acts as a bridge, allowing Swalang to interact with native C-compatible code.

## Loading a Library

Use `ffi.CDLL` to load a shared library (.so, .dll, or .dylib).

```python
import ffi
import sys

# Load the standard C library
libc = ffi.CDLL("libc.so.6") # Linux
# or use the shorthand if available
# libc = ffi.libc
```

For Windows-specific libraries, you can use `ffi.windll`, which behaves like `ffi.CDLL` but also provides access to `get_last_error()`.

```python
# On Windows only
kernel32 = ffi.windll("kernel32.dll")
# ... call a kernel32 function ...
error_code = kernel32.get_last_error()
```

## Defining Function Signatures

Before calling a C function, you must define its signature (argument types and return type).

```python
# C signature: int puts(const char *s);
puts = libc.puts([ffi.c_char_p], ffi.c_int32)

# Now call it
puts(b"Hello from C!")
```

## Supported Data Types

### Primitives

| FFI Type | C Type | Swalang Type |
| :--- | :--- | :--- |
| `ffi.c_bool` | `_Bool` | `bool` |
| `ffi.c_char` | `char` | `bytes` (len 1) |
| `ffi.c_uint8` | `uint8_t` | `int` |
| `ffi.c_int32` | `int32_t` | `int` |
| `ffi.c_uint32` | `uint32_t` | `int` |
| `ffi.c_int64` | `int64_t` | `int` |
| `ffi.c_uint64` | `uint64_t` | `int` |
| `ffi.c_float` | `float` | `float` |
| `ffi.c_double` | `double` | `float` |
| `ffi.c_char_p` | `char*` | `bytes` / `str` |
| `ffi.c_void_p` | `void*` | `Pointer` |

### Typed Pointers

Use `ffi.POINTER(T)` to create a typed pointer.

```python
IntPtr = ffi.POINTER(ffi.c_int32)
```

### Arrays

Fixed-size C arrays are defined using `ffi.ArrayType`.

```python
# C: float numbers[10]
FloatArray10 = ffi.ArrayType(ffi.c_float, 10)
```

### Structs and Unions

Define C-compatible structs and unions using decorators.

```python
@ffi.Struct
class Point:
    _fields_ = [
        ('x', ffi.c_int32),
        ('y', ffi.c_int32)
    ]

@ffi.Union
class Data:
    _fields_ = [
        ('i', ffi.c_int32),
        ('f', ffi.c_float)
    ]
```

## Variadic Functions

To define a function that accepts a variable number of arguments (like `printf`), set `is_variadic=True`.

```python
printf = libc.printf([ffi.c_char_p], ffi.c_int32, is_variadic=True)
printf(b"Hello %s, score: %d\n", b"Swalang", 100)
```

## Memory Management

Swalang provides tools for manual memory management when interacting with C.

- `ffi.malloc(size)`: Allocates `size` bytes of memory.
- `ffi.free(ptr)`: Frees allocated memory.
- `ffi.read_memory(ptr, type)`: Reads a value of the given type from memory.
- `ffi.read_memory_with_offset(ptr, offset, type)`: Reads a value at a specific offset.
- `ffi.write_memory(ptr, type, value)`: Writes a value to memory.
- `ffi.write_memory_with_offset(ptr, offset, type, value)`: Writes a value at a specific offset.
- `ffi.string_at(ptr, [len], [offset])`: Reads a string from memory. If `len` is -1, reads until null terminator.
- `ffi.buffer_to_bytes(ptr, size)`: Converts a memory buffer to a Swalang `bytes` object.
- `ffi.memset(ptr, value, size)`: Fills memory with a byte value.

## Callbacks

You can pass Swalang functions to C as callbacks.

```python
def my_callback(val):
    print(f"C called Swalang with: {val}")
    return 0

# Create a C-compatible function pointer
cb_ptr = ffi.callback(my_callback, ffi.c_int32, [ffi.c_int32])

# Pass cb_ptr to a C function
register_callback(cb_ptr)

# When done, free the callback to prevent memory leaks
ffi.free_callback(cb_ptr)
```
