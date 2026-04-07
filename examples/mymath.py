# examples/mymath.py
pi = 3.14159
PI = 3.14159

def add(a, b):
    return a + b

def subtract(a, b):
    return a - b

print("mymath.py: Module initialized!")
# print('=========*args and **kwargs=========')
# def func(a, b, *args, **kwargs):
#     print(a, b)       # 1 2
#     print(args)       # (3, 4)
#     print(kwargs)     # {'x': 5, 'y': 6}
# func(1, 2, 3, 4, x=5, y=6)
# print('ends=========*args and **kwargs=========')

def square(x):
  return x * x

class Helper:
  def __init__(self, factor):
    self.factor = factor
  def multiply(self, val):
    return val * self.factor