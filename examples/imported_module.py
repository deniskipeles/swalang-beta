# examples/imported_module.py

print(f("--- Inside imported_module.py ---"))
print(f("The value of __name__ is: {__name__}"))

def module_function():
    print("This is a function from the imported module.")

if __name__ == "__main__":
    # This block will NOT run when this file is imported.
    # It would only run if you executed `go run ... ./examples/imported_module.py`
    print("This should not be printed during an import!")
    