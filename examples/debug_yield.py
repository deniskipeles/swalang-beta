# examples/debug_yield.py - Simple test to isolate the issue
def simple_gen():
    print("Generator starting")
    received = yield 1
    print("After first yield, received:", received)
    received = yield 2  
    print("After second yield, received:", received)
    print("Generator ending")

print("=== Testing simple generator ===")
gen = simple_gen()

print("1. Calling next()...")
result1 = next(gen)
print("   Returned:", result1)

print("2. Calling send(10)...")
result2 = gen.send(10)
print("   Returned:", result2)

print("3. Calling send(20)...")
try:
    result3 = gen.send(20)
    print("   Returned:", result3)
except Exception:
    print("   Generator exhausted")