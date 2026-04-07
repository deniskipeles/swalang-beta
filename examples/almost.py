# examples/test_suite.py
# A script to test various features of the PyLearn interpreter

print("--- PyLearn Interpreter Test Suite ---")

# 1. Literals and Basic Types
print("\n[1] Testing Literals & Types")
i = 100
f = 3.14
s = "hello pylearn"
b = True
n = None
print("Integer:", i, type(i))
print("Float:", f, type(f))
print("String:", s, type(s))
print("Boolean:", b, type(b))
print("None:", n, type(n))

# 2. Arithmetic Operators
print("\n[2] Testing Arithmetic Operators")
print("i + 50 =", i + 50)
print("f * 2 =", f * 2)
print("i / 8 =", i / 8)  # Expect float division
print("100 % 7 =", 100 % 7)
print("-i =", -i)
print("Precedence (10 + 2 * 3):", 10 + 2 * 3)
print("Parentheses ( (10 + 2) * 3 ):", (10 + 2) * 3)

# 3. Comparison Operators
print("\n[3] Testing Comparison Operators")
print("i > 50:", i > 50)
print("f <= 3.0:", f <= 3.0)
print("s == 'hello pylearn':", s == "hello pylearn")
print("s != 'goodbye':", s != "goodbye")
print("b == True:", b == True)
print("n == None:", n == None)
print("i == f:", i == f) # False (int vs float)
print("100 == 100.0:", 100 == 100.0) # True (numeric comparison)

# 4. Logical Operators
print("\n[4] Testing Logical Operators")
print("True and i > 0:", True and i > 0)
print("b or f < 0:", b or f < 0)
print("not b:", not b)
print("not n:", not n) # None is falsy
print("not '':", not "") # Empty string is falsy
print("not 'a':", not "a") # Non-empty string is truthy

# 5. Variables and Assignment
print("\n[5] Testing Variables & Assignment")
my_var = i + 1
print("my_var (init):", my_var)
my_var = my_var * 2
print("my_var (updated):", my_var)
a = b = 10 # Chained assignment (if supported by parser/eval)
print("a, b:", a, b)

# 6. Control Flow: If/Elif/Else
print("\n[6] Testing If/Elif/Else")
if i > 200:
    print("i is large")
elif i > 50:
    print("i is medium")
else:
    print("i is small")

if not n:
    print("None is Falsy - PASS")

# 7. Control Flow: While Loop
print("\n[7] Testing While Loop")
counter = 0
while counter < 4:
    print("While counter:", counter)
    counter = counter + 1
    if counter == 2:
        print("Continuing loop...")
        continue
    if counter == 3:
        print("Breaking loop...")
        break
# else: # Python's while-else (executes if loop finished without break)
#     print("While loop finished normally.") # This shouldn't print

# 8. Control Flow: For Loop (Range, List)
print("\n[8] Testing For Loop")
print("For loop over range(5):")
for x in range(5):
    print("  range x:", x)

print("For loop over list:")
items = ["apple", 1, True]
for item in items:
    print("  list item:", item, type(item))

# 9. Data Structures: List
print("\n[9] Testing Lists")
my_list = [10, "hi", 30.5]
print("Initial list:", my_list, "len:", len(my_list))
print("Element at index 1:", my_list[1])
my_list[0] = 99
print("Modified list:", my_list)
# print("Out of bounds:", my_list[5]) # Should cause IndexError

# 10. Data Structures: Dictionary
print("\n[10] Testing Dictionaries")
my_dict = {"name": "PyLearn", "version": 0.1, 1: "one"}
print("Initial dict:", my_dict, "len:", len(my_dict))
print("Value for 'name':", my_dict["name"])
my_dict["version"] = 0.2
my_dict[True] = "boolean key"
print("Modified dict:", my_dict)
# print("Missing key:", my_dict["unknown"]) # Should cause KeyError

# 11. Functions
print("\n[11] Testing Functions")
global_var = 1000

def calculate(p1, p2):
    print("  Inside calculate, global_var =", global_var) # Access global
    local_var = p1 * p2
    return local_var

res = calculate(5, 6)
print("calculate(5, 6) =", res)
# print("local_var outside function:", local_var) # Should cause NameError

# 12. Classes and OOP
print("\n[12] Testing Classes & OOP")
class Vehicle:
    # class_attr = "movable" # Optional: Test class attributes later

    def __init__(self, wheels):
        self.num_wheels = wheels
        print("Vehicle with", wheels, "wheels created.")

    def display(self):
        return "A vehicle with " + str(self.num_wheels) + " wheels."

    def __str__(self):
        return "<Vehicle wheels=" + str(self.num_wheels) + ">"

    def __len__(self):
        print("Called Vehicle.__len__")
        return self.num_wheels # Example: len is number of wheels

    def __add__(self, other):
        # Example: Adding vehicles adds their wheels
        if type(other) == Vehicle:
            return Vehicle(self.num_wheels + other.num_wheels)
        return None # Or raise TypeError

    def __eq__(self, other):
         if type(other) == Vehicle:
             return self.num_wheels == other.num_wheels
         return False


car = Vehicle(4)
bike = Vehicle(2)
trike = Vehicle(3)
car_copy = Vehicle(4)

print("Car object:", car) # Uses __str__
print("Car display:", car.display())
print("Length of bike (wheels):", len(bike)) # Uses __len__

print("car == bike:", car == bike) # Uses __eq__ -> False
print("car == car_copy:", car == car_copy) # Uses __eq__ -> True
print("car != bike:", car != bike) # Uses __eq__ negation -> True

big_vehicle = car + bike # Uses __add__
print("Car + Bike =", big_vehicle)

print("Type of car:", type(car))
print("Type of Vehicle class:", type(Vehicle))


# 13. Modules (Native & File-based)
print("\n[13] Testing Modules")

# Native 'sys'
import sys
print("sys.platform:", sys.platform)
print("len(sys.argv):", len(sys.argv))
if len(sys.argv) > 0:
    for arg in sys.argv:
        print("sys.argv:", arg)
    print("sys.argv[0]:", sys.argv[0])

# Native 'http' (Requires internet & assumes httpbin.org is up)
import http
print("Attempting HTTP GET...")
if True:
    response = http.get("https://httpbin.org/get")
    print("http.get status:", response.status_code)
    if response.status_code == 200:
        print("http.get response text contains 'httpbin.org':", "httpbin.org" + response.text) # Requires 'in' for strings
    # Test POST
    post_resp = http.post("https://httpbin.org/post", "test_data=hello")
    print("http.post status:", post_resp.status_code)
else:
    # No specific error types yet, catch generic runtime error
    # Need to check the actual error object type/message if interpreter provides it
    print("HTTP request failed (Network/httpbin issue or TypeError?). Check interpreter error output.")


# File-based 'mymath' (Create examples/mymath.py first)
# Contents of examples/mymath.py:
# pi = 3.14159
# def square(x): return x * x

if True: # Use try only if you want the script to continue past import error
    import mymath
    print("mymath.pi:", mymath.pi)
    print("mymath.square(9):", mymath.square(9))
else:
    print("Failed to import 'mymath'. Make sure examples/mymath.py exists.")


# 14. Runtime Errors (Uncomment to test reporting)
print("\n[14] Testing Runtime Error Reporting (Examples)")
# print(10 / 0)
# print(non_existent_variable)
# print("a" + 1)
# items[99]
# my_dict["nope"]
# car.non_existent_method()


print("\n--- Interpreter Test Suite Done ---")

# Optionally, use sys.exit to signal success/failure if desired
# import sys
# sys.exit(0) # Signal success

# Using the ternary operator
def compare(val1, val2, ret1, ret2):
    if val1 > val2:
        return ret1
    else:
        return ret2

    
x = 10
result = compare(x,5,'greater','lesser')

print(result)  # Output: Greater than 5

import os

files = os.listdir('../cmd')
for file in files:
    print(file)
