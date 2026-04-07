import numpy_like

print("--- Testing numpy_like module (via FFI) ---")

# Create two arrays
# This will allocate C memory and copy the data under the hood.
a = numpy_like.array([1.0, 2.0, 3.0, 4.0])
b = numpy_like.array([5.0, 6.0, 7.0, 8.0])

print("Array a:", a)
print("Array b:", b)
print("Shape of a:", a.shape)

# Calculate the dot product
# This calls the high-performance OpenBLAS C function via FFI.
# Expected: (1*5) + (2*6) + (3*7) + (4*8) = 5 + 12 + 21 + 32 = 70
dot_product = numpy_like.dot(a, b)

print("Dot product of a and b:", dot_product)

# Assert the result is correct
assert dot_product == 70.0

# --- IMPORTANT: Manual Memory Management ---
# Because we used `ffi.malloc`, we must `ffi.free` the memory.
# A real-world language would handle this with a garbage collector and finalizers.
a.free()
b.free()

print("\n--- numpy_like test successful! ---")