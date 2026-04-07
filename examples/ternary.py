# examples/ternary.py

a = 10
b = 20

# Test Case 1: Condition is True
max_val = a if a > b else b
print(f("The max value is: {max_val}"))  # Expected: 20

# Test Case 2: Condition is False
status = "active" if True else "inactive"
print(f("Status is: {status}"))  # Expected: active

# Test Case 3: Nested ternary and complex expressions
result = (a + 5) if a == 10 and b == 20 else ("b is not 20" if b != 20 else "something else")
print(f("Complex result: {result}"))  # Expected: 15

# Test Case 4: Using a function call in the condition
def is_even(n):
    return n % 2 == 0

num = 7
parity = "even" if is_even(num) else "odd"
print(f("The number {num} is {parity}")) # Expected: odd

