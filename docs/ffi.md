
## Pylearn FFI Module: A Developer's Guide

The `ffi` module is a powerful tool for calling functions in external shared libraries (like `.dll` files on Windows, `.so` on Linux, or `.dylib` on macOS) directly from your Pylearn code. It acts as a bridge, allowing Pylearn to interact with native C-compatible code. This guide explains how to use it effectively, including newly added features for structs, unions, and variadic functions.

### Table of Contents

1.  **Core Concepts**
2.  **Loading a Library (`ffi.CDLL`, `ffi.windll`)**
3.  **Defining a C Function's Signature**
4.  **Calling C Functions**
5.  **Working with C Data Types**
    *   Primitives
    *   Pointers
    *   Arrays
    *   Structs and Unions
6.  **Advanced: Variadic Functions (like `printf`)**
7.  **Working with Memory**
8.  **Creating and Using Callbacks**
9.  **Complete Example: Using `libc`**

### 1. Core Concepts

The `ffi` module works in three main steps:

1.  **Load**: You load a shared library into memory using `ffi.CDLL("library_name")`.
2.  **Define**: You access a function by its name (e.g., `mylib.my_c_function`) and immediately tell the FFI its signature—what types of arguments it takes (`argtypes`) and what it returns (`restype`).
3.  **Call**: You can then call the configured function just like a regular Pylearn function.

The FFI backend handles the complex task of converting Pylearn objects (like integers and strings) into a format that the C function understands, and converting the C function's return value back into a Pylearn object.

### 2. Loading a Library (`ffi.CDLL`, `ffi.windll`)

The entry point to using a library is the `ffi.CDLL` class. You create an instance of it by passing the name of the shared library.

```python
import ffi

# On Linux
libc = ffi.CDLL("c") # Automatically finds libc.so.6

# On macOS
libm = ffi.CDLL("m") # Automatically finds libm.dylib

# On Windows
kernel32 = ffi.CDLL("kernel32.dll")
```

The `ffi` module is smart enough to search in standard system locations and common fallback paths. If the library cannot be found, an `ffi.FFIError` is raised.

For Windows-specific libraries, you can use `ffi.windll`, which behaves like `ffi.CDLL` but also provides access to `get_last_error()`.

```python
# On Windows only
user32 = ffi.windll("user32.dll")
# ... call a user32 function ...
error_code = user32.get_last_error()
print(format_str("Last Win32 error code: {error_code}"))
```

### 3. Defining a C Function's Signature

Once you have a library object, you access functions within it like attributes. This returns a special **configurator function**. You must immediately call this configurator with the function's signature.

The signature consists of:
*   `argtypes`: A **list** of FFI data types corresponding to the C function's parameters.
*   `restype`: The single FFI data type that the C function returns. Use `None` for functions that return `void`.
*   `is_variadic` (optional): A boolean, `True` if the function is variadic (see section 6).

```python
import ffi

# C function signature: long int random(void);
# Assume 'libc' is already loaded via ffi.CDLL('c')
random = libc.random([], ffi.c_long)
#           ^         ^    ^
#           |         |    restype
#           |         argtypes (an empty list for no arguments)
#           Access the 'random' function and call the configurator
```

### 4. Calling C Functions

After a function has been configured, the result is a callable Pylearn function.

```python
# Continuing the example from above
random_number = random()
print(format_str("A random number from libc: {random_number}"))

# C function signature: int puts(const char *str);
puts = libc.puts([ffi.c_char_p], ffi.c_int32)
puts(b"Hello from C via Pylearn FFI!")
# This will print the string to your console.
```
**Note:** When passing strings to `c_char_p`, you must use a `bytes` literal (e.g., `b"my string"`).

### 5. Working with C Data Types

#### Primitives

The `ffi` module provides objects that represent common C types.

| FFI Type      | C Equivalent     | Pylearn Type for Calling/Receiving |
|---------------|------------------|------------------------------------|
| `ffi.c_bool`    | `_Bool`          | `bool`                             |
| `ffi.c_char`    | `char`           | `bytes` (of length 1)              |
| `ffi.c_wchar_t` | `wchar_t`        | `str` (of length 1)                |
| `ffi.c_short`   | `short`          | `int`                              |
| `ffi.c_int`     | `int`            | `int`                              |
| `ffi.c_long`    | `long`           | `int`                              |
| `ffi.c_longlong`| `long long`      | `int`                              |
| `ffi.c_float`   | `float`          | `float`                            |
| `ffi.c_double`  | `double`         | `float`                            |
...and their `unsigned` counterparts (`c_ushort`, `c_uint`, etc.).

#### Pointers

Pointers are fundamental. The `ffi.POINTER` function is the primary way to create pointer types.

| FFI Type        | C Equivalent | Notes |
|-----------------|--------------|-------|
| `ffi.c_void_p`  | `void*`      | Generic untyped pointer. |
| `ffi.c_char_p`  | `const char*`| Pointer to a null-terminated byte string. |
| `ffi.c_wchar_p` | `const wchar_t*`| Pointer to a null-terminated wide string. |
| `ffi.POINTER(T)`| `T*`         | A typed pointer, e.g., `ffi.POINTER(ffi.c_int)`. |

#### Arrays

Fixed-size C arrays are defined using `ffi.ArrayType`.

```python
# C signature: void process_array(float numbers[10]);
FloatArray10 = ffi.ArrayType(ffi.c_float, 10)

process_array = mylib.process_array([FloatArray10], None)

# Pass a Pylearn list, which will be converted to a C array.
data = [1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0, 10.0]
process_array(data)
```

#### Structs and Unions

Use the `@ffi.Struct` and `@ffi.Union` class decorators for an elegant, `ctypes`-like syntax.

*   **Defining Types:** Define a class with a `_fields_` attribute, which is a list of `(name, type)` tuples.

```python
import ffi

@ffi.Struct
class Point:
    _fields_ = [
        ('x', ffi.c_int32),
        ('y', ffi.c_int32)
    ]

@ffi.Union
class DataPacket:
    _fields_ = [
        ('as_int', ffi.c_int64),
        ('as_float', ffi.c_double),
        ('as_bytes', ffi.ArrayType(ffi.c_char, 8))
    ]
```

*   **Using Types:** These new types (`Point`, `DataPacket`) can now be used in `argtypes` and `restype` definitions. Creating and passing instances from Pylearn is an advanced topic that requires manual memory management with `ffi.malloc` and `ffi.write_memory`.

### 6. Advanced: Variadic Functions (like `printf`)

To define a function that accepts a variable number of arguments, set the `is_variadic=True` flag in the configurator.

You only need to define the types for the **fixed** arguments. The FFI will automatically promote the types of the extra arguments you pass (e.g., Pylearn `int` becomes C `long long`, `float` becomes `double`).

```python
# C signature: int printf(const char *format, ...);
printf = libc.printf(
    [ffi.c_char_p],      # argtypes for the fixed part (the format string)
    ffi.c_int32,         # restype
    is_variadic=True     # The magic flag!
)

# Now you can call it with any number of extra arguments.
printf(b"Hello, %s!\n", b"Pylearn")
printf(b"Number: %d, Float: %f\n", 42, 3.14)
```

### 7. Working with Memory

For functions that require you to pass pointers to memory buffers, you must manage the memory yourself.

*   `ffi.malloc(size)`: Allocates `size` bytes of memory and returns a `Pointer` object.
*   `ffi.free(pointer)`: Frees memory that was allocated with `malloc`.
*   `ffi.string_at(pointer, [size])`: Reads a null-terminated string or `size` bytes from a pointer.
*   `ffi.memset(dest, value, size)`: Fills a block of memory with a specific byte value.

### 8. Creating and Using Callbacks

You can pass Pylearn functions to C functions that expect function pointers (callbacks). Use `ffi.callback` to create a C-compatible function pointer.

**Example: A C function expecting a callback `int (*cb)(int)`**
```python
# Assume a C library has this function:
# void set_int_processor(int (*my_callback)(int));

# In Pylearn:
def my_pylearn_adder(num):
    print(format_str("Pylearn callback received: {num}"))
    return num + 10

# 1. Create the C-compatible function pointer from the Pylearn function.
adder_callback = ffi.callback(my_pylearn_adder, ffi.c_int32, [ffi.c_int32])
#                                ^             ^            ^
#                                Pylearn func  restype      argtypes

# 2. Configure and call the C function that accepts the callback.
# The callback itself is passed as a void pointer.
set_processor = mylib.set_int_processor([ffi.c_void_p], None)
set_processor(adder_callback)
```
**Important**: The FFI creates a durable C-level object for the callback. To prevent memory leaks, you must explicitly free it with `ffi.free_callback(adder_callback)` when it is no longer needed by the C library.

### 9. Complete Example: Using `libc`

```python
import ffi

# Load the standard C library
libc = ffi.libc
if not libc:
    print("Could not load standard C library.")
    # In a real script, you might exit here.
else:
    # --- Define and call puts ---
    puts = libc.puts([ffi.c_char_p], ffi.c_int32)
    puts(b"--- FFI Test Start ---")

    # --- Define and call a variadic function: printf ---
    printf = libc.printf([ffi.c_char_p], ffi.c_int32, is_variadic=True)
    
    name = b"Pylearn"
    version = 1
    printf(b"Welcome to %s FFI v%d.\n", name, version)

    # --- Use memset ---
    buffer = ffi.malloc(20)
    # C signature: void *memset(void *s, int c, size_t n);
    # Note: size_t is mapped to c_int64 for portability
    memset_func = libc.memset([ffi.c_void_p, ffi.c_int32, ffi.c_int64], ffi.c_void_p)

    # Fill the buffer with the character 'A' (ASCII 65)
    memset_func(buffer, 65, 19)

    # Read it back and print it
    result_bytes = ffi.string_at(buffer, 19)
    puts(b"memset result: " + result_bytes)

    # Clean up
    ffi.free(buffer)
    puts(b"--- FFI Test End ---")

```