# Swalang Syntax Guide

Swalang's syntax is heavily inspired by Python. It uses indentation to define code blocks and focuses on readability.

## Basic Types

- **Integers:** `42`, `-7`
- **Floats:** `3.14`, `2.0`
- **Booleans:** `True`, `False`
- **None:** `None` (representing the absence of a value)
- **Strings:** `"Hello"`, `'World'`, `f"Value: {x}"` (f-strings)
- **Bytes:** `b"data"`
- **Lists:** `[1, 2, 3]`
- **Tuples:** `(1, 2, 3)`
- **Dictionaries:** `{"key": "value"}`
- **Sets:** `{1, 2, 3}`

## Variables

Variables are assigned using the `=` operator. Swalang is dynamically typed.

```python
x = 10
y = "Swalang"
```

## Control Flow

### If Statements

```python
if x > 10:
    print("Greater than 10")
elif x == 10:
    print("Exactly 10")
else:
    print("Less than 10")
```

### For Loops

```python
for i in [1, 2, 3]:
    print(i)
```

### While Loops

```python
count = 0
while count < 5:
    print(count)
    count = count + 1
```

## Functions

Functions are defined using the `def` keyword.

```python
def add(a, b=0):
    return a + b

result = add(5, 3)
```

### Lambdas

```python
square = lambda x: x * x
print(square(4)) # 16
```

## Classes and OOP

Swalang supports object-oriented programming with classes and inheritance.

```python
class Animal:
    def __init__(self, name):
        self.name = name

    def speak(self):
        pass

class Dog(Animal):
    def speak(self):
        return "Woof!"

my_dog = Dog("Buddy")
print(my_dog.speak())
```

### Advanced OOP

Swalang supports advanced OOP features like dictionary-like access via dunder methods:

```python
class SimpleDict:
  def __init__(self):
    self._data = {}
  def __setitem__(self, key, value):
    self._data[str(key)] = value
  def __getitem__(self, key):
    return self._data[str(key)]

sd = SimpleDict()
sd["foo"] = 100
print(sd["foo"]) # 100
```

### Dunder Methods

Swalang supports many Python-style dunder methods for operator overloading and special behavior:
- `__init__`: Constructor
- `__str__`: String representation
- `__getitem__`, `__setitem__`: Indexing support
- `__call__`: Making instances callable
- `__getattr__`, `__setattr__`: Attribute access control

## Exception Handling

```python
try:
    risky_operation()
except ValueError as e:
    print("Caught a value error")
finally:
    print("Cleanup")
```

## Import System

```python
import os
from time import sleep
import sdl2 as sdl
```
