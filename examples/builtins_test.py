# examples/builtins_test.py
import os
print("--- Testing type() ---")
print(type(10))
print("--- Testing os.listdir() ---")
print(os.listdir())
# /*
# well 
# done 
# this is a multiline comment
# */
x = """
hello world
this is a multiline string
"""
print(x)
print(type(x))


print(type(3.14))
print(type("hello"))
print(type(True))
print(type(None))
print(type([1, 2]))
print(type({"a":1}))
print(type(range(3)))
class MyClass:
  def __init__(self):
    pass
print(MyClass())
print(type(MyClass))
print(type(print)) # Type of a built-in

def my_func():
  return 42
print(my_func())
print(type(my_func))
print(type(print)) # Type of a built-in

print("\n--- Testing str() ---")
print(str(123))
print(str(-45.6))
print(str("string")) # Should not add extra quotes
print(str(False))
print(str(None))
print(str([1, "two"]))
s = str({"x": 99})
print(s)
print(type(s))

print("\n--- Testing int() ---")
print(int(10))
print(int(99.9)) # Truncates
print(int(-1.2)) # Truncates
print(int(True))
print(int(False))
print(int("12345"))
# print(int("-99")) # Uncomment when negative string parsing works if needed
# print(int("  5 ")) # Uncomment when whitespace handling needed
# print(int("10.5")) # Should cause ValueError
# print(int("abc")) # Should cause ValueError

# print("\n--- Testing input() ---")
# name = input("Enter your name: ")
# print("Hello,", name)
# age_str = input("Enter your age: ")
# age_int = int(age_str) # Test combination
# print("Next year you will be:", age_int + 1)
# print(type(age_str))
# print(type(age_int))

print("Done with built-ins test.")