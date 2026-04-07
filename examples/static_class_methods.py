print("--- Testing Class Methods and Static Methods ---")

class MyMath:
    class_val = 10 # Class variable

    def __init__(self, instance_val):
        self.instance_val = instance_val
        print("MyMath.__init__ called, self.instance_val =", self.instance_val)

    def regular_method(self, x):
        print("MyMath.regular_method called, self.instance_val =", self.instance_val, "x =", x)
        return self.instance_val + x + MyMath.class_val

    def cm_impl(cls, y): # cls will be MyMath class object
        print("MyMath.cm_impl called, cls.class_val =", cls.class_val, "y =", y)
        # Cannot access self.instance_val here
        return cls.class_val + y

    actual_classmethod = classmethod(cm_impl)

    def sm_impl(z): # No self or cls
        print("MyMath.sm_impl called, z =", z)
        # Cannot access self.instance_val or cls.class_val directly
        return MyMath.class_val + z # Can access class_val via class name

    actual_staticmethod = staticmethod(sm_impl)

print("MyMath.class_val =", MyMath.class_val)

# Test classmethod
print("Calling MyMath.actual_classmethod(5):")
res_cm_class = MyMath.actual_classmethod(5)  # Should pass MyMath class as cls
print("Result:", res_cm_class) # Expected: 10 + 5 = 15

# Test staticmethod
print("Calling MyMath.actual_staticmethod(7):")
res_sm_class = MyMath.actual_staticmethod(7) # No implicit first arg
print("Result:", res_sm_class) # Expected: 10 + 7 = 17

print("Creating instance m = MyMath(20)")
m = MyMath(20)
print("m.instance_val =", m.instance_val)
print("m.class_val (via instance) =", m.class_val) # Python allows this, accesses class_val

# Test regular method on instance
print("Calling m.regular_method(3):")
res_reg = m.regular_method(3)
print("Result:", res_reg) # Expected: 20 + 3 + 10 = 33

# Test classmethod on instance
print("Calling m.actual_classmethod(6):")
res_cm_instance = m.actual_classmethod(6) # Should pass m's class (MyMath) as cls
print("Result:", res_cm_instance) # Expected: 10 + 6 = 16

# Test staticmethod on instance
print("Calling m.actual_staticmethod(8):")
res_sm_instance = m.actual_staticmethod(8) # No implicit first arg
print("Result:", res_sm_instance) # Expected: 10 + 8 = 18


print("\n--- Test with *args and **kwargs with static/class methods ---")

# class DecoratorDemo:
#     shared = 100

#     def static_args_kwargs_impl(*args, **kwargs):
#         print("static_args_kwargs_impl: args=", args, "kwargs=", kwargs)
#         # return DecoratorDemo.shared + len(args) + len(kwargs) # Requires class name
#         return 1000 + len(args) + len(kwargs) # Simpler for now


#     decorated_static = staticmethod(static_args_kwargs_impl)

#     def class_args_kwargs_impl(cls, *args, **kwargs):
#         print("class_args_kwargs_impl: cls=", cls.shared, "args=", args, "kwargs=", kwargs)
#         return cls.shared + len(args) + len(kwargs)

#     decorated_class = classmethod(class_args_kwargs_impl)


# print("Calling DecoratorDemo.decorated_static(1, 2, x=3):")
# res_ds = DecoratorDemo.decorated_static(1, 2, x=3)
# print("Result:", res_ds) # Expected: 1000 + 2 + 1 = 1003

# print("Calling DecoratorDemo.decorated_class(1, 2, x=3):")
# res_dc = DecoratorDemo.decorated_class(1, 2, x=3)
# print("Result:", res_dc) # Expected: 100 + 2 + 1 = 103

# instance_dd = DecoratorDemo()
# print("Calling instance_dd.decorated_static('a', 'b', 'c', p=1, q=2):")
# res_ids = instance_dd.decorated_static("a", "b", "c", p=1, q=2)
# print("Result:", res_ids) # Expected: 1000 + 3 + 2 = 1005

# print("Calling instance_dd.decorated_class('a', 'b', 'c', p=1, q=2):")
# res_idc = instance_dd.decorated_class("a", "b", "c", p=1, q=2)
# print("Result:", res_idc) # Expected: 100 + 3 + 2 = 105

# print("--- End Testing ---")







class MyMath2:
    val = 10 # Class variable

    def regular_method(self, x):
        return self.val + x

    def cm(cls, y): # cls will be MyMath class object
        return cls.val + y
    
    classmethod_impl = classmethod(cm) # cm becomes a classmethod named classmethod_impl

    def sm(z): # No self or cls
        return z * 2
    
    staticmethod_impl = staticmethod(sm) # sm becomes a staticmethod

# Accessing via class
print(MyMath2.val)                 # Output: 10 (Class variable)
print(MyMath2.staticmethod_impl(5)) # Output: 10 (Calls sm(5))
print(MyMath2.classmethod_impl(3)) # Output: 13 (Calls cm(MyMath, 3))
# print(MyMath.regular_method(None, 1)) # Error or special handling for unbound instance methods

# Accessing via instance
m = MyMath2()
m.val = 20 # Instance variable shadows class variable for this instance
print(m.val)                     # Output: 20 (Instance variable)
print(m.regular_method(5))       # Output: 25 (self.val is 20)
print(m.staticmethod_impl(5))    # Output: 10 (Calls sm(5), no self)
print(m.classmethod_impl(3))     # Output: 13 (Calls cm(MyMath, 3), cls is MyMath, not m's class if m was subclass)


class MyClass:
    def __init__(self, name):
        self.name = name

    @classmethod
    def my_cm(cls, arg):
        print(format_str("Called from class: {cls.name} with arg: {arg}"))

MyClass.my_cm(123)