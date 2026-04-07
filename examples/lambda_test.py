async def main_program():
    print("--- Testing Lambda Functions ---")

    # Simple lambda
    add = lambda x, y: x + y
    result = add(5, 3)
    print(format_str("add(5, 3) = {result}")) # Expected: 8

    # Using lambda with map/filter (if they were implemented)
    # For now, let's just show it can be passed as an argument
    def operate(func, a, b):
        return func(a, b)

    mult_result = operate(lambda a, b: a * b, 6, 7)
    print(format_str("operate(lambda, 6, 7) = {mult_result}")) # Expected: 42

    # Lambda with a default argument
    power = lambda base, exp=2: base ** exp
    print(format_str("power(3) = {power(3)}")) # Expected: 9
    print(format_str("power(3, 3) = {power(3, 3)}")) # Expected: 27

    # Immediately invoked lambda
    imm_result = (lambda x: x + 1)(10)
    print(format_str("(lambda x: x + 1)(10) = {imm_result}")) # Expected: 11