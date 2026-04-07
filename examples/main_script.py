# examples/main_script.py

print("--- Inside main_script.py ---")
print(f("The value of __name__ is: {__name__}"))

# Import the other module
import imported_module

print("\n--- Back in main_script.py after import ---")

def run_main_logic():
    print("Now running the main logic of the script...")
    imported_module.module_function()
    print("Main logic finished.")

# This is the classic Python idiom for a runnable script
if __name__ == "__main__":
    print("\nExecuting code inside the `if __name__ == '__main__'` block...")
    run_main_logic()

print("\n--- End of main_script.py ---")