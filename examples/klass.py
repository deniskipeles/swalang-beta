class Greeter:
    def __init__(self, name):
        self.name = name
    
    def greet(self): # Method taking only self
        return "Hello, " + self.name + "!"

    def loop(self):
        for x in range(5):
            print(x)

g = Greeter("PyLearn")
x=g.greet() # Call the method
print(x)
g.loop()
print("Done.")