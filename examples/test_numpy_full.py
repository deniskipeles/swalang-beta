import numpy_like

print("--- Testing numpy_like module (Full) ---")
print(2*4)

a = numpy_like.array([1.0, 2.0, 3.0])
b = numpy_like.array([10.0, 20.0, 30.0])

# --- Test Element Access ---
print("Accessing element a[1]:", a[1])
assert a[1] == 2.0
print("Element access successful.")

# --- Test Vector Addition ---
# Using the module function
c_add_func = numpy_like.add(a, b)
print("a + b (using add()):", c_add_func)
assert c_add_func[0] == 11.0
assert c_add_func[2] == 33.0

# Using the overloaded operator
c_add_op = a + b
print("a + b (using + op):", c_add_op)
assert c_add_op[1] == 22.0
print("Vector addition successful.")


# --- Test Scalar Multiplication ---
# Using the module function
d_mul_func = numpy_like.multiply(a, 5)
print("a * 5 (using multiply()):", d_mul_func)
assert d_mul_func[0] == 5.0
assert d_mul_func[2] == 15.0

# Using the overloaded operator
d_mul_op = a * 10
print("a * 10 (using * op):", d_mul_op)
assert d_mul_op[1] == 20.0

# Test reflected operator
d_rmul_op = 3 * b
print("3 * b (using * op):", d_rmul_op)
assert d_rmul_op[0] == 30.0
print("Scalar multiplication successful.")

# --- Test Dot Product (already working) ---
dot_product = numpy_like.dot(a, b)
print("Dot product of a and b:", dot_product)
assert dot_product == 140.0 # 10 + 40 + 90
print("Dot product successful.")


print("\n--- numpy_like full test successful! ---")

b = numpy_like.zeros((3, 3, 3,2))
print(b)
ones = numpy_like.ones((4,4))
print(ones)
print("=" * 40)