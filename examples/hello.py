# Simple PyLearn example

print("Hello from PyLearn!")
print('power',2**4)
x = 10
y = 20.5
result = x + y * 2
print("Result:", result)

is_greater = result > 50
print("Is result > 50?", is_greater)

if is_greater:
    print("It's greater!")
else:
    print("It's not greater.")

# Function definition and call
def greet(name):
    msg = "Hello, " + name + "!"
    return msg

message = greet("World")
print(message)

# List and loop
items = [1, "two", 3.0, True]
print("List items:",len(items))
for item in items:
    print("-", item)

# Dictionary
my_dict = {"a": 1, "b": "foo"}
print("Dictionary:", my_dict.get("a"), my_dict.get("b"), my_dict.get("c", "default_value"))
# print("Dict value for 'a':", my_dict["a"]) # Uncomment when dict indexing is fully tested

# Simple while loop
count = 0
while count < 3:
    print("Count:", count)
    count = count + 1

for x in range(5):
    print(x)

print("\nTesting range with continue:")
for y in range(50,100,10):
    print("y1 =", y)
    if y == 80:
        break
    print("y =", y)
print("After continue loop")

print("Done.")

class Greeter:
  def __init__(self, name):
    self.name = name

  def greet(self): # Method taking only self
    return "Hello, " + self.name + "!"

g = Greeter("PyLearn")
g.greet() # Call the method

class Human:
  def __init__(self, name, age):
    self.name = name
    self.age = age
  def introduce(self):
    return "My name is " + self.name + " and I am " + str(self.age) + " years old."

h = Human("Alice", 30)
print(h.introduce())
print(len(h.name,""))
print("Done with classes.")

# TODO: to be implemented to achieve the go native iteration speed/looping
# for.native() x in range(500000000).native():
#     result = x * x + x / (x+1)
#     if x % 10000000 == 0:
#         print("result=", result)

for x in range(500000000):
    result = x * x + x / (x+1)
    if x % 10000000 == 0:
        print("result=",result)


print("\nTesting range with continue:")
for y in range(50,100,10):
    print("y =", y)