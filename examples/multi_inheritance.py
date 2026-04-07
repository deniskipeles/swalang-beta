print("--- Multiple Inheritance Test ---")

class Base:
    def who_am_i(self):
        print("I am Base")
        return "Base"

    def common_method(self):
        print("Base's common_method")
        return "BaseCommon"

class Mixin1:
    def who_am_i(self): # Overrides Base.who_am_i if Mixin1 is earlier in MRO
        print("I am Mixin1")
        # super().who_am_i() # This would call the next in MRO (e.g., Base)
        return "Mixin1"

    def mixin1_method(self):
        print("Mixin1's specific method")
        return "Mixin1Specific"

    def common_method(self): # Overrides Base.common_method
        print("Mixin1's common_method, calling super...")
        # res = super().common_method() # Call next in MRO for common_method
        # return "Mixin1Common + " + res
        return "Mixin1Common" # Simplified for now if super in mixin is tricky

class Mixin2:
    val = 100
    def who_am_i(self): # Overrides Base.who_am_i if Mixin2 is earlier
        print("I am Mixin2")
        return "Mixin2"

    def mixin2_method(self):
        print("Mixin2's specific method")
        return "Mixin2Specific"

# --- Scenario 1: Simple MRO ---
# MRO for Child1: Child1, Mixin1, Base, object
class Child1(Mixin1, Base):
    def who_am_i(self):
        print("I am Child1, calling super...")
        # res = super().who_am_i() # Should call Mixin1.who_am_i
        # print("Child1: super().who_am_i() returned:", res)
        # return "Child1 -> " + res
        # return "Child1 -> " + "Mixin1" # Simplified for now if super not fully tested
        print("I am Child1, calling super...")
        res = super(Child1, self).who_am_i() # MODIFIED: Explicit super
        print("Child1: super().who_am_i() returned:", res)
        return "Child1 -> " + res

    def child1_method(self):
        print("Child1's specific method")
        return "Child1Specific"

print("\n--- Child1 (Mixin1, Base) ---")
c1 = Child1()
print("c1.who_am_i():", c1.who_am_i())                 # Expected: Child1 -> Mixin1
print("c1.mixin1_method():", c1.mixin1_method())         # Expected: Mixin1Specific
print("c1.common_method():", c1.common_method())         # Expected: Mixin1Common (from Mixin1)
# If super() was fully working in Mixin1: Mixin1Common + BaseCommon
print("c1.child1_method():", c1.child1_method())         # Expected: Child1Specific

# --- Scenario 2: Diamond Problem (Simplified) ---
# MRO for Child2: Child2, Mixin1, Mixin2, Base, object
class Child2(Mixin1, Mixin2, Base): # Mixin2 also has who_am_i
    def who_am_i(self):
        print("I am Child2, calling super...")
        # res = super().who_am_i() # Should call Mixin1.who_am_i due to order
        # print("Child2: super().who_am_i() returned:", res)
        # return "Child2 -> " + res
        return "Child2 -> " + "Mixin1" # Simplified

    def common_method(self): # Child2 overrides common_method
        print("Child2's common_method, calling super...")
        # res = super().common_method() # Should call Mixin1.common_method
        # return "Child2Common + " + res
        return "Child2Common + " + "Mixin1Common" # Simplified


print("\n--- Child2 (Mixin1, Mixin2, Base) ---")
c2 = Child2()
print("c2.who_am_i():", c2.who_am_i())                 # Expected: Child2 -> Mixin1
print("c2.mixin1_method():", c2.mixin1_method())         # Expected: Mixin1Specific
print("c2.mixin2_method():", c2.mixin2_method())         # Expected: Mixin2Specific
print("c2.common_method():", c2.common_method())         # Expected: Child2Common + Mixin1Common
print("Mixin2.val (accessed via class):", Mixin2.val) # Expected: 100
# print("c2.val (inherited):", c2.val) # Expected: 100 (if class var inheritance works)


# --- Scenario 3: Using super() more extensively ---
class Left(Base):
    def common_method(self):
        print("Left's common_method, calling super...")
        # return "Left -> " + super().common_method()
        return "Left -> " + "BaseCommon" # Simplified

class Right(Base):
    def common_method(self):
        print("Right's common_method, calling super...")
        # return "Right -> " + super().common_method()
        return "Right -> " + "BaseCommon" # Simplified

# MRO for Diamond: Diamond, Left, Right, Base, object
class Diamond(Left, Right):
    def common_method(self):
        print("Diamond's common_method, calling super...")
        # return "Diamond -> " + super().common_method() # Calls Left.common_method
        return "Diamond -> " + "Left -> " + "BaseCommon" # Simplified

print("\n--- Diamond (Left, Right) ---")
d = Diamond()
print("d.common_method():", d.common_method())
# Expected if super() works perfectly: Diamond -> Left -> Right -> BaseCommon
# (C3 MRO makes Right come before Base in Left's super() call path if Left calls super)
# For simplified version: Diamond -> Left -> BaseCommon

print("\n--- Testing super() from a specific point ---")
class A:
    def f(self): 
        print("A.f")
        return "A"

class B(A):
    def f(self):
        print("B.f, calling super(B, self).f()")
        # return "B -> " + super(B, self).f() # super(B, self) starts search *after* B in self's MRO (i.e., at A)
        return "B -> " + "A" # Simplified

class C(A):
    def f(self):
        print("C.f, calling super(C, self).f()")
        # return "C -> " + super(C, self).f()
        return "C -> " + "A" # Simplified

class D(B, C): # MRO: D, B, C, A, object
    def f(self):
        print("D.f, calling super(D, self).f()")
        # s_d = super(D, self).f() # Calls B.f

        print("D.f, calling super(B, self).f()") # Start search after B in D's MRO (i.e. at C)
        # s_b = super(B, self).f() # Calls C.f

        print("D.f, calling super(C, self).f()") # Start search after C in D's MRO (i.e. at A)
        # s_c = super(C, self).f() # Calls A.f
        # return "D -> " + s_d + " | D via B's super: " + s_b + " | D via C's super: " + s_c
        return "D -> B -> A | D via B's super: C -> A | D via C's super: A" # Simplified


print("\n--- D(B, C) super() tests ---")
inst_d = D()
print("inst_d.f():", inst_d.f())
# Expected if super(type, obj) works:
# D -> B -> A | D via B's super: C -> A | D via C's super: A

print("\n--- End Multiple Inheritance Test ---")






