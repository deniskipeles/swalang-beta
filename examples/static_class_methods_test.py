# examples/static_class_methods_test.py
# A comprehensive test for static and class methods.

print("--- Test 1: Basic Static Method ---")
class MyMath:
    @staticmethod
    def add(x, y):
        # This method knows nothing about the class or an instance.
        # It's just a function namespaced inside the class.
        return x + y

# Call on the class
print(format_str("MyMath.add(10, 5) = {MyMath.add(10, 5)}"))

# Call on an instance
m = MyMath()
print(format_str("instance.add(20, 7) = {m.add(20, 7)}"))
print("")


print("--- Test 2: Basic Class Method ---")
class Pizza:
    radius = 10 # A class variable

    def __init__(self, ingredients):
        self.ingredients = ingredients

    @classmethod
    def from_hawaiian_style(cls):
        # `cls` is the Pizza class itself.
        # It can access class variables and call other class methods or the constructor.
        print(format_str("Creating pizza from class '{cls.name}' with radius {cls.radius}"))
        return cls(["ham", "pineapple"]) # Calls Pizza(...)

# Call on the class
hawaiian_pizza = Pizza.from_hawaiian_style()
print(format_str("Hawaiian pizza ingredients: {hawaiian_pizza.ingredients}"))

# Call on an instance (less common, but should work)
regular_pizza = Pizza(["cheese", "tomato"])
other_hawaiian = regular_pizza.from_hawaiian_style()
print(format_str("Other Hawaiian pizza ingredients: {other_hawaiian.ingredients}"))
print("")


print("--- Test 3: Inheritance and super() with Class Methods ---")
class BetterPizza(Pizza):
    radius = 12 # Override class variable

    @classmethod
    def from_deluxe_style(cls):
        # `cls` here will be the BetterPizza class.
        # It correctly uses the overridden radius.
        print(format_str("Creating deluxe pizza from class '{cls.name}' with radius {cls.radius}"))
        # We can call a parent's class method via super()
        # but super() inside a classmethod needs the (cls, self) form.
        # Let's just call the constructor of the specific subclass `cls`.
        return cls(["pepperoni", "mushrooms", "peppers"])

deluxe = BetterPizza.from_deluxe_style()
print(format_str("Deluxe pizza ingredients: {deluxe.ingredients}"))

# Calling the PARENT's classmethod from the CHILD class
# The `cls` argument will correctly be `BetterPizza`
child_hawaiian = BetterPizza.from_hawaiian_style()
print(format_str("Child Hawaiian pizza ingredients: {child_hawaiian.ingredients}"))
print(format_str("Is child_hawaiian an instance of BetterPizza? {isinstance(child_hawaiian, BetterPizza)}"))
print("")


print("--- Test 4: Decorator Stacking (less common, but good test) ---")
# A simple decorator for demonstration
def verbose(fn):
    def wrapper(*args, **kwargs):
        print(format_str("Calling {fn.__name__}..."))
        result = fn(*args, **kwargs)
        print(format_str("{fn.__name__} finished."))
        return result
    return wrapper
    

class Calculator:
    @staticmethod
    @verbose
    def multiply(a, b):
        return a * b

# The decorators are applied bottom-up: staticmethod is applied first,
# then verbose is applied to the result.
result = Calculator.multiply(7, 6)
print(format_str("7 * 6 = {result}"))

print("power", 2**3)

list_x = [1, 2, 3, 4, 5]
print(*list_x)
dictionary_x = {'a': 1, 'b': 2, 'c': 3}
print(*dictionary_x)
