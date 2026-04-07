# examples/test_for_yield.py
# A generator function that yields from inside a for loop.
def number_generator():
    print("Generator started.")
    
    names = ["Alice", "Bob", "Charlie"]
    
    for i in range(len(names)):
        # Construct a message inside the loop
        message = format_str("Loop {i}: Yielding {names[i]}")
        
        # Yield the message. The generator should pause here.
        yield message
        
        # This print statement should execute *after* the generator is resumed.
        print(format_str("Generator resumed after yielding for index {i}."))
        
    print("Generator finished.")
    # Generators implicitly return after the loop.

# --- Main execution logic ---
def main():
    print("--- Testing For-Loop Yielding ---")
    
    # Create an instance of the generator.
    # The code inside number_generator() does NOT run yet.
    gen = number_generator()
    
    print("Generator object created. Now starting iteration...")
    
    # The for loop will drive the generator.
    # Each iteration will run the generator's code until the next `yield`.
    for yielded_value in gen:
        print(format_str("Main program received: '{yielded_value}'"))
        print("---")
        
    print("--- Iteration complete. ---")

if __name__ == "__main__":
    main()