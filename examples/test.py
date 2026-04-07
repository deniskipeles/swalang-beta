# examples/test.py
empty_dict = {}
print(empty_dict)
empty_list = []
print(empty_list)
empty_set = set()
print(empty_set)
empty_tuple = ()
print(empty_tuple)
f_string = f"""
hello world
"""
print(f_string)

w = 1+1
print(w)
x = 10
x += 1
print(x)
for k in range(10):
    print(k)
# error.py

def error(value):
    if value == 1:
        raise Exception('One')
    elif value == 2:
        raise Exception('Two')
    elif value == 3:
        raise Exception('Three')
    elif value == 4:
        raise Exception('Four')
    else:
        print('no error here')
try:
    e = error(1)
except Exception as e:
    print(e)
list_x = [i+2 for i in range(10)]
print(list_x)
set_x = {j+10 for j in range(10)}
print(set_x)
for k,v in enumerate([1,2,3]):
    print(k,v)
list_x = tuple((1,2,3,4,5)) # or tuple
list_x.__add__((1,2,3,4,5))
print(list_x[0]," AND ",list_x[4])
print(1 in list_x)



value = 0
print(value)
value += 1
print(value+1)
print(0x200000 + 2)

get_first_vals_from_list_x = list_x[:2]
print(get_first_vals_from_list_x)

get_last_vals_from_list_x = list_x[-2:]
print(get_last_vals_from_list_x)

get_middle_vals_from_list_x = list_x[1:-1]
print(get_middle_vals_from_list_x)




# print(type('hi'))
# print(isinstance('True', str))
# print(isinstance(True, bool))
# print(isinstance((), bool))
# print(isinstance([], list))
# print(isinstance('True', bool))

# # from webbrowser import get

set_x = {1,2,3}
print(set_x)

obj = {
    "a":"a",
    "b":"b",
    "c":"c",
    "d":"d"
}
print(obj)
x = 'hello'
print(x.split(''))
print(obj.__len__())
print(1 not in obj)




# Test script specifically for the Compiler/VM features
print("--- VM Test Script ---")
# Test Constants and Basic Arithmetic
a = 10
b = 3
print("a =", a)
print("b =", b)
# --- ADD FUNCTION TEST ---
def add(x, y):
    # Simple addition  # << Use spaces before this comment too (though comments are often ignored for indent)
    z = x + y          # << Use 4 spaces here
    return z           # << Use 4 spaces here
    # Implicit return None if no return statement
res = add(5, 9)
print("add(5, 9) =", res)
# Test calling again
print("add(a, b) =", add(a, b))

# Test function returning None implicitly (optional)
def no_return():
    c = 55 # << Use 4 spaces if uncommented
print("no_return() =", no_return())
# --- END FUNCTION TEST ---
# --- FOR LOOP TEST ---
print("--- For Loop Test ---")
for x in range(1, 10, 2):
    print("x =", x)

for x2 in range(5,10):
    print("x2 =", x2)

for x3 in range(5):
    print("x3 =", x3)

# # for x in range(500000000):
# #     result = x * x + x / (x+1)
# #     if x % 10000000 == 0:
# #         print("result=",result)
print("--- End For Loop ---")
# --- END FOR LOOP TEST ---

# --- WHILE LOOP TEST ---
print("--- While Loop Test ---")
count = 0
while count < 3:
    print("count =", count)
    # Use simple assignment as += might not be implemented
    count = count + 1
print("After while, count =", count) # Should print 3
print("--- End While Loop ---")
# --- END WHILE LOOP TEST ---



# --- IF,ELIF and ELSE TEST ---
print("--- If, Elif and Else Test ---")
if a > b:
    print("a is greater than b")
elif a < b:
    print("a is less than b")
else:
    print("a is equal to b")
print("--- End If, Elif and Else ---")
# --- END IF,ELIF and ELSE TEST ---



# --- IF/ELIF/ELSE TEST ---
print("--- If/Elif/Else Test ---")
val = 15

if val < 10:
    print("val is less than 10")
elif val < 20:
    print("val is less than 20 but not less than 10") # Should print this
else:
    print("val is 20 or greater")

val = 5
if val < 10:
    print("val is less than 10") # Should print this
elif val < 20:
    print("val is less than 20 but not less than 10")
else:
    print("val is 20 or greater")

val = 25
if val < 10:
    print("val is less than 10")
elif val < 20:
    print("val is less than 20 but not less than 10")
else:
    print("val is 20 or greater") # Should print this

# Test without else
val = 100
if val == 50:
    print("val is 50")
elif val == 75:
    print("val is 75")
# Should print nothing here

print("--- End If/Elif/Else Test ---")
# --- END IF/ELIF/ELSE TEST ---



# Simplified examples/test.py (as before)
# print("--- VM Test Script ---")
class Greeter:
    def __init__(self, name):
        pass # Just pass for now

    def greet(self):
        return "Hello, world!" # Return simple string

    def loop(self):
        for x in range(5):
            print("loop:", x)

print("--- Class Test ---")
g = Greeter("PyLearn") # Instantiation -> calls __init__
x = g.greet()        # Method call
print(x)
g.loop()             # Comment out loop for now
print("Done.")
print("--- VM Test Script Done ---")




# --- AND/OR TEST ---
print("--- And/Or Test ---")
print(" 5 and 10 =", 5 and 10)    # Should print 10
print(" 0 and 10 =", 0 and 10)    # Should print 0
print(" 10 and 0 =", 10 and 0)    # Should print 0
print(" '' and 10 =", '' and 10) # Should print '' (empty string)
print(" 'a' and 10 =", 'a' and 10) # Should print 10

print(" 5 or 10 =", 5 or 10)     # Should print 5
print(" 0 or 10 =", 0 or 10)     # Should print 10
print(" 10 or 0 =", 10 or 0)     # Should print 10
print(" '' or 10 =", '' or 10)  # Should print 10
print(" 'a' or 10 =", 'a' or 10)  # Should print 'a'

print(" False and 5 =", False and 5) # Should print False
print(" True and 5 =", True and 5)   # Should print 5
print(" False or 5 =", False or 5)   # Should print 5
print(" True or 5 =", True or 5)     # Should print True

# Combining
a = 1
b = 0
c = 3
print(" a and b or c =", a and b or c) # (1 and 0) or 3 -> 0 or 3 -> 3
print(" a or b and c =", a or b and c) # 1 or (0 and 3) -> 1 or 0 -> 1
print("--- End And/Or Test ---")
# --- END AND/OR TEST ---

# examples/test.py
# --- BREAK/CONTINUE TEST ---
print("--- Break/Continue Test ---")

# Test break
i = 0
while i < 10:
    if i == 3:
        print("breaking at i =", i)
        break
    print("in while, i =", i)
    i = i + 1
print("after while (break), i =", i) # Should be 3

# Test continue
j = 0
result = 0
while j < 5:
    j = j + 1
    if j == 2 or j == 4:
        print("continuing at j =", j)
        continue
    print("adding j =", j)
    result = result + j
print("after while (continue), result =", result) # Should be 1 + 3 + 5 = 9

# Test in for loop
print("--- For loop break/continue ---")
evens = 0
for k in range(10):
    if k == 7:
        print("break at k =", k)
        break
    if k % 2 != 0: # If odd
        print("continue at k =", k)
        continue
    print("even k =", k)
    evens = evens + 1
print("evens count =", evens) # Should be 0, 2, 4, 6 -> count 4
print("--- End Break/Continue Test ---")
# --- END BREAK/CONTINUE TEST ---

list_x = [i+4 for i in range(10)]
print(list_x)
for k,v in enumerate(list_x):
    print(k,v)
multiline_set = {
    1,
    2,
    3,
}
print(multiline_set)
multiline_dict = {
    "a": 1,
    "b": 2,
    "c": 3,
}
print(multiline_dict)

single_line_tuple = (1, 2, 3)
print(single_line_tuple)

multiline_tuple = (
    1,
    2,
    3,
)

single_line_list = [1, 2, 3]
print(single_line_list)

multiline_list = [
    1,
    2,
    3,
]
print(multiline_tuple)
print(multiline_list)

print("add",1+1)
print("subtract",1-1)
print("multiply",1*1)
print("power",1**1)
print("modulo  ",1%1)
# try:
#     print("divide",1/0)
# except Exception as e:
#     print(e)

f_string = f"""
hello world
"""
print(f_string)

price = 49
txt = f"For only {price} dollars!"
print(txt) 

txt = "For only {price:.2f} dollars!"
print(txt.format(price = 49))