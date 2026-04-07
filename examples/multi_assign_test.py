# examples/multi_assign_test.py

print("--- Testing basic multiple assignment ---")
x, y = 10, 20
print(format_str("x is {x}, y is {y}")) # Expected: x is 10, y is 20

# Test swapping variables
x, y = y, x
print(format_str("After swap, x is {x}, y is {y}")) # Expected: After swap, x is 20, y is 10

print("\n--- Testing unpacking from a list ---")
data_list = [100, "hello", True]
a, b, c = data_list
print(format_str("a = {a} (type: {type(a)})"))
print(format_str("b = {b} (type: {type(b)})"))
print(format_str("c = {c} (type: {type(c)})"))

print("\n--- Testing unpacking from a tuple ---")
DELIMITERS = {
    "variable": ("{{", "}}"),
    "block": ("{%", "%}"),
}
var_start, var_end = DELIMITERS["variable"]
print(format_str("Variable start: '{var_start}', Variable end: '{var_end}'"))

print("\n--- Testing error: not enough values ---")
try:
    j, k, l = [1, 2]
except ValueError as e:
    print(format_str("Caught expected error: {e}"))

print("\n--- Testing error: too many values ---")
try:
    m, n = (5, 6, 7)
except ValueError as e:
    print(format_str("Caught expected error: {e}"))

print("\n--- Testing error: unpacking non-iterable ---")
try:
    p, q = 123
except TypeError as e:
    print(format_str("Caught expected error: {e}"))