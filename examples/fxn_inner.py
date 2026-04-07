# function call itself
def factorial(n):
    # Base case (stop condition)
    if n == 0:
        return 1
    # Recursive step (function calls itself)
    else:
        return n * factorial(n - 1)


print(factorial(30))  # Output: 265252859812191058636308480000000

# class ParentA:
#     def __init__(self, a, **kwargs):
#         super().__init__(**kwargs)  # Call next in MRO first
#         self.a = a
#         print("ParentA initialized with a =", a)

# class ParentB:
#     def __init__(self, b, **kwargs):
#         super().__init__(**kwargs)  # Call next in MRO first
#         self.b = b
#         print("ParentB initialized with b =", b)

# class Child(ParentA, ParentB):
#     def __init__(self, a, b):
#         super().__init__(a=a, b=b)  # Pass all parameters
#         print("Child initialized")

# c = Child(a=10, b="hello")




# examples/inheritance_test.py
print("--- Testing Single Inheritance ---")

class Animal:
    class_var_animal = "I am an animal"

    def __init__(self, name):
        print("Animal __init__ called for", name)
        self.name = name

    def speak(self):
        return "Generic animal sound"

    def get_name(self):
        return self.name

class YoungOne:
    def __init__(self,young_one):
        print("YoungOne __init__ called")
        self.young_one = young_one



class Dog(Animal):
    class_var_dog = "I am a dog"

    def __init__(self, name, breed):
        # Manual parent __init__ call (NO super() YET)
        Animal.__init__(self, name) # Pass 'self' explicitly
        print("Dog __init__ called for", name, "of breed", breed)
        self.breed = breed

    def speak(self): # Override
        return "Woof!"

    def get_breed(self):
        return self.breed

# Test Parent Class
generic_animal = Animal("Creature")
print("generic_animal.name:", generic_animal.get_name())
print("generic_animal.speak():", generic_animal.speak())
print("generic_animal class_var:", generic_animal.class_var_animal)


# Test Child Class
my_dog = Dog("Buddy", "Golden Retriever")
print("my_dog.name (from Animal):", my_dog.get_name())      # Inherited method
print("my_dog.breed (from Dog):", my_dog.get_breed())      # Own method
print("my_dog.speak() (overridden):", my_dog.speak())      # Overridden method
print("my_dog class_var_animal (from Animal):", my_dog.class_var_animal) # Inherited class var
print("my_dog class_var_dog (from Dog):", my_dog.class_var_dog)       # Own class var

# Test attribute setting on instance doesn't affect parent class var
my_dog.class_var_animal = "Instance specific animal var"
print("my_dog.class_var_animal (instance):", my_dog.class_var_animal)
another_animal = Animal("Else")
print("another_animal.class_var_animal (parent):", another_animal.class_var_animal)


# Test accessing attribute not present anywhere
try:
    print(my_dog.non_existent_attr)
except: # Generic except for now
    print("Caught error trying to access my_dog.non_existent_attr (expected)")

print("--- Done Testing Single Inheritance ---")
