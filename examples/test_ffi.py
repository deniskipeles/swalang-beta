# examples/test_ffi.py

# This script demonstrates how to use the FFI module to call C functions.

import ffi

# On Linux/macOS, libc can often be loaded by its name.
# On some systems, you may need the full path like '/lib/x86_64-linux-gnu/libc.so.6'
try:
    libc = ffi.CDLL("/lib/x86_64-linux-gnu/libc.so.6")
except Exception:
    try:
        libc = ffi.CDLL("libc.dylib")
    except Exception:
        # Fallback for systems where just 'c' works, or other names.
        libc = ffi.CDLL("c")

print("--- Testing FFI with libc ---")

# --- 1. Define and call 'printf' ---
# Configure the 'printf' function.
# Signature: int printf(char* format, ...)
# We will call it with one string and one int argument.
printf = libc.printf([ffi.c_char_p, ffi.c_int32], ffi.c_int32)

# Call the function
message = "Hello from C! The magic number is %d\n"
number = 42
bytes_written = printf(message, number)

print(format_str("(Pylearn: printf returned that it wrote {bytes_written} bytes)"))
print("")


# --- 2. Define and call 'abs' ---
# Signature: int abs(int n)
c_abs = libc.abs([ffi.c_int32], ffi.c_int32)

neg_val = -123
abs_val = c_abs(neg_val)

print(format_str("The absolute value of {neg_val} is {abs_val}."))
assert abs_val == 123
print("Assertion successful!")