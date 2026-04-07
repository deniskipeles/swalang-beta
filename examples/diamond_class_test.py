# examples/diamond_class_test.py

class A:
    def __init__(self,var):
        print("Call from>>>",var)
    def who(self):
        print("I am an A")

class B(A):
    def who(self):
        print("I am a B")
        super().__init__("Call From>>>B")
        super().who()

class C(A):
    def who(self):
        print("I am a C")
        super().who()

class D(B, C):
    def who(self):
        print("I am a D")
        super().who()

d_instance = D("D")
d_instance.who()