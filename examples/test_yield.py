# examples/test_yield.py 
# 1) A generator that yields numbers and can receive feedback
def countdown(start):
    print("Generator started")
    current = start
    while current > 0:
        received = yield current          # <-- yield pauses here
        if received:
            print(format_str("Received {received} from caller"))
            current = received            # caller can change the next value
        current = current - 1
    print("Generator exhausted")

# 2) Demonstration -------------------------------------------------------
gen = countdown(5)

# First next() primes the generator up to the first yield
print("First next():", next(gen))          # -> 5

# Now we can send values back in
print("send(10) :", gen.send(10))          # -> 10
print("send(2)  :", gen.send(2))           # -> 2
print("next()   :", next(gen))             # -> 1
print("next()   :", next(gen))             # raises StopIteration
