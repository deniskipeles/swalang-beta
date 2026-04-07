# This script is designed to test the robustness of the Pylearn parser,
# specifically its ability to correctly handle blank lines, lines with only
# whitespace, and comment-only lines inside various indented blocks.
#
# If this script runs to completion and produces the expected output,
# the parser's handling of these cases is correct.

print("--- Testing Robust Parsing of Indented Blocks ---")

# 1. Test inside a Class definition
print("\n[1] Testing Class Block:")
class RobustClass:

    # A blank line before the first method.
    def __init__(self):

        print("  - Inside class __init__ method")
        self.value = "initialized"
    
        # A line with only whitespace below.
    
    # A comment-only line between methods.
    # Another comment.
    def greet(self):
        
        # A blank line at the start of the method body.
        
        print("  - Inside class greet method")
        return format_str("Hello from {self.value} RobustClass")

        # A blank line at the end of the method body.

instance = RobustClass()
print(instance.greet())


# 2. Test inside a Function definition
print("\n[2] Testing Function Block:")
def robust_function(a):

    print("  - Inside robust_function")
    x = a * 10
    
    # This is a comment inside the function.
    
    y = x + 5
    
    # A line with only spaces below.
    
    return y

result = robust_function(4)
print(format_str("  - Result of robust_function: {result}"))


# 3. Test inside If/Elif/Else blocks
print("\n[3] Testing If/Elif/Else Blocks:")
test_val = "elif"

if test_val == "if":
    
    print("  - This should NOT print (if block)")

elif test_val == "elif":

    # Comment in elif.
    print("  - Correctly executed elif block")
    
else:
    
    print("  - This should NOT print (else block)")


# 4. Test inside a For loop
print("\n[4] Testing For Loop Block:")
for i in range(2):

    print(format_str("  - For loop iteration {i}"))
    
    # Comment in for loop.


# 5. Test inside a While loop
print("\n[5] Testing While Loop Block:")
counter = 2
while counter > 0:
    
    print(format_str("  - While loop counter: {counter}"))
    
    counter = counter - 1
    # Another comment.


# 6. Test inside Try/Except/Finally blocks
print("\n[6] Testing Try/Except/Finally Blocks:")
try:

    print("  - Inside try block (before error)")
    1 / 0
    print("  - This should NOT print (after error)")

except ZeroDivisionError:
    
    # Comment in except block.
    print("  - Correctly caught ZeroDivisionError in except block")
    
finally:
    
    # Comment in finally block.
    print("  - Executed finally block")
    

# 7. Test inside a With statement
print("\n[7] Testing With Statement Block:")

# A mock context manager for the test
class DummyContextManager:
    def __enter__(self):
        print("  - Entering context")
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        print("  - Exiting context")

with DummyContextManager() as d:

    # Comment inside with block
    print("  - Inside with block")
    

print("\n--- Test script completed successfully! ---")