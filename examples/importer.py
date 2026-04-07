# examples/importer.py
from mymath import square, pi, Helper,PI as PY,add


list_with_values = [1,2,3]
print(list_with_values)
list_with_no_values = {
    1,2,3
}
print(list_with_no_values)


print(pi*PY)
print(add(1,2))

import mymath
import os
import http
print(os.listdir())



async def fetch_data(url):
    # async operation
    http.get(url)
    data = await http.get(url)
    return data

async def main_program():
    data1 = await fetch_data("jsonplaceholder.typicode.com/todos/1")  # These can run concurrently
    data2 = await fetch_data("jsonplaceholder.typicode.com/todos/2")
    return [data1, data2]




def test():
    obj = {'a': 1, 'b': 2 }
    y = [1,2,3]

    print('outer fxn',obj,y)
    def inner():
        print('inner fxn')
    inner()
test()
print("Imported mymath!")
print("Pi is:", mymath.pi)

area = mymath.pi * mymath.square(10)
print("Area:", area)

h = mymath.Helper(5)
print("Multiplied:", h.multiply(7))

# Try importing again (should use cache)
import mymath
print("Imported again, should be fast.")




# This is importer.py
import pylearn_importlib # Assuming this is how we access the new module

print("Importer: Trying to load mymath.py dynamically...")

# Construct the path relative to importer.py
# In a real framework, path resolution would be more robust.
# For this example, let's assume mymath.py is in the same directory.
math_module_path = "./mymath.py" 
# Or, if your SetCurrentScriptDir works well, just "mymath.py" might be enough
# if importer.py and mymath.py are in the same directory from where you run.
# To be safe, let's assume SetCurrentScriptDir correctly sets context to examples/

try:
    # Use the dynamic loading function
    math = pylearn_importlib.load_module_from_path(math_module_path)
    print(math)
    x = {
        "a":1,
        "b":2
    }

    if math: # Check if loading was successful (not an Error object)
        print("Importer: mymath module loaded successfully.")
        print("Importer: Type of math:", type(math)) # Should be <class 'module'>
        
        # Access attributes from the dynamically loaded module
        print("Importer: PI from math:", math.PI)
        
        sum_result = math.add(10, 5)
        print("Importer: math.add(10, 5) =", sum_result)
        
        diff_result = math.subtract(10, 3)
        print("Importer: math.subtract(10, 3) =", diff_result)
    else:
        print("Importer: Failed to load mymath module (math is None or an error occurred before assignment).")

except Exception as e: # Generic Pylearn exception
    print("Importer: An error occurred during import or usage:", 'e')

print("Importer: Script finished.")

list_x = []
print(list_x.__len__())
list_x.append(1)
print(list_x.__len__())
list_x.append(2)
print(list_x.__len__())
list_x.append(3)
print(list_x.__len__())

print(list_x)

x='hi'
print(x.__len__())

y = {"a":"b","c":"d"}
print(y.__contains__('a'))


def func(a, b=2, *args, **kwargs):
    print(a, b, args, kwargs)

func(1)           # a=1, b=2, args=[], kwargs={}
func(1, 3, 4, 5, x=6)  # a=1, b=3, args=[4, 5], kwargs={}


class BaseHTTPMiddleware:
    def __init__(self, app):
        self.app = app

    def dispatch(self, request, call_next):
        # Base implementation, maybe some logging
        print("BaseHTTPMiddleware dispatch called",self.app)
        response = 'call_next(request)'
        return response

class LoggingMiddleware(BaseHTTPMiddleware):
    def __init__(self, app):
        super(LoggingMiddleware, self).__init__(app) # Uses super(Type, self)
        print("LoggingMiddleware initialized")

    def dispatch(self, request, call_next):
        print(request)
        hello = "hello"
        print(f("Request (Logging): {request['url']} {hello}"))
        # response = super().dispatch(request, call_next) # Needs zero-arg super()
        # For now, with two-arg super:
        response = super(LoggingMiddleware, self).dispatch(request, call_next)
        print(f("Response (Logging): ..."))
        return response

app = {"request":"hi","response":"bye","url":"http://google.com"}
c = LoggingMiddleware(app)
print(c.dispatch(app, None))


import static_class_methods
# y=(1,2,3)
# print(y)

# def two_vals():
#     return 1,"two"

print(two_vals)