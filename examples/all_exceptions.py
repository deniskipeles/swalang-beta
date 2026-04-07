
# examples/all_exceptions.py
tasks = {i for i in range(10)}
array = [
    'hello',
    'world',
    1
]
print(*tasks)
print(*array)
# A simple custom class to demonstrate AttributeError
class MyObject:
    def __init__(self):
        self.attribute = "I exist"

# ----------------------------------------------------------------------
# 1. TypeError: Operation on an inappropriate type.
# ----------------------------------------------------------------------
def test_type_error():
    print("--- 1. Testing TypeError ---")
    try:
        result = "hello" + 5
        print(format_str("FAILURE: This line should not be reached. Result: {result}"))
    except TypeError as e:
        print(format_str("SUCCESS: Caught TypeError: {e}"))
    print("")

# ----------------------------------------------------------------------
# 2. ValueError: Correct type, but inappropriate value.
# We will manually raise it to show it can be caught.
# ----------------------------------------------------------------------
def test_value_error():
    print("--- 2. Testing ValueError ---")
    try:
        # In a more complete `int()` implementation, `int("abc")` would raise this.
        # For now, we raise it manually to test the `except` clause.
        raise ValueError("a custom value error message")
    except ValueError as e:
        print(format_str("SUCCESS: Caught ValueError: {e}"))
    print("")

# ----------------------------------------------------------------------
# 3. NameError: Using a variable that has not been defined.
# ----------------------------------------------------------------------
def test_name_error():
    print("--- 3. Testing NameError ---")
    try:
        print(an_undefined_variable)
        print("FAILURE: This line should not be reached.")
    except NameError as e:
        print(format_str("SUCCESS: Caught NameError: {e}"))
    print("")

# ----------------------------------------------------------------------
# 4. IndexError: Sequence index is out of range.
# ----------------------------------------------------------------------
def test_index_error():
    print("--- 4. Testing IndexError ---")
    my_list = [10, 20, 30]
    try:
        item = my_list[100]
        print(format_str("FAILURE: This line should not be reached. Item: {item}"))
    except IndexError as e:
        print(format_str("SUCCESS: Caught IndexError: {e}"))
    print("")

# ----------------------------------------------------------------------
# 5. KeyError: Dictionary key not found.
# ----------------------------------------------------------------------
def test_key_error():
    print("--- 5. Testing KeyError ---")
    my_dict = {"a": 1}
    try:
        value = my_dict["b"]
        print(format_str("FAILURE: This line should not be reached. Value: {value}"))
    except KeyError as e:
        print(format_str("SUCCESS: Caught KeyError: {e}"))
    print("")


# ----------------------------------------------------------------------
# 6. AttributeError: Attribute reference or assignment fails.
# ----------------------------------------------------------------------
def test_attribute_error():
    print("--- 6. Testing AttributeError ---")
    obj = MyObject()
    try:
        val = obj.non_existent_attribute
        print(format_str("FAILURE: This line should not be reached. Value: {val}"))
    except AttributeError as e:
        print(format_str("SUCCESS: Caught AttributeError: {e}"))
    print("")

# ----------------------------------------------------------------------
# 7. ImportError: The `import` statement fails to find the module.
# ----------------------------------------------------------------------
def test_import_error():
    print("--- 7. Testing ImportError ---")
    try:
        import a_module_that_really_does_not_exist
        print("FAILURE: This line should not be reached.")
    except ImportError as e:
        print(format_str("SUCCESS: Caught ImportError: {e}"))
    print("")

# ----------------------------------------------------------------------
# 8. Multiple 'except' Clauses & Specificity
# ----------------------------------------------------------------------
def test_multiple_excepts(case):
    print(format_str("--- 8. Testing Multiple Excepts (case: {case}) ---"))
    try:
        if case == "key":
            d = {}
            print(d["nokey"])
        elif case == "index":
            l = []
            print(l[0])
        else:
            print("No error raised in this case.")
            
    except KeyError as e:
        print(format_str("SUCCESS: Caught KeyError as expected: {e}"))
    except IndexError as e:
        print(format_str("SUCCESS: Caught IndexError as expected: {e}"))
    except Exception as e:
        print(format_str("FAILURE: Caught generic Exception, but a specific one should have matched: {e}"))
    print("")

# ----------------------------------------------------------------------
# 9. Catching a Tuple of Exceptions
# ----------------------------------------------------------------------
def test_tuple_except():
    print("--- 9. Testing Tuple of Exceptions ---")
    try:
        result = 1 / "string"
    except (ValueError, TypeError) as e:
        print(format_str("SUCCESS: Caught error in tuple (ValueError, TypeError): {e}"))
    print("")

# ----------------------------------------------------------------------
# 10. Nested Try...Except (Uncaught inner exception)
# ----------------------------------------------------------------------
def test_nested_except():
    print("--- 10. Testing Nested Try...Except ---")
    try:
        print("Outer try block starts.")
        try:
            print("Inner try block starts.")
            raise ValueError("An error from the inner try.")
        except TypeError as e:
            print("Inner except (TypeError) - should not be hit.")
        print("This line in outer try should not be printed.")
    except ValueError as e:
        print(format_str("SUCCESS: Outer except caught the error from the inner try: {e}"))
    print("")


# --- Main Execution ---
test_type_error()
test_value_error()
test_name_error()
test_index_error()
test_key_error()
test_attribute_error()
test_import_error()
test_multiple_excepts("key")
test_multiple_excepts("index")
test_tuple_except()
test_nested_except()

print("--- All exception tests completed ---")