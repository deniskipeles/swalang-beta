# examples/json_test.py
# Test script for the built-in 'json' module in PyLearn

print("--- PyLearn JSON Module Test ---")

import json
import sys # For checking errors if needed

# Helper to print results neatly
def check_result(description, actual, expected_type_str, test_equality=None, expected_val=None):
    print("\nTesting:", description)
    print("  Actual Value:", actual)
    print("  Actual Type :", type(actual))
    print("  ExpectedType:", expected_type_str)
    if type(actual) != expected_type_str: # Compare string representation of type
        print("  TYPE MISMATCH!")
        return False # Early exit on type mismatch for some tests

    if test_equality != None:
        are_equal = actual == test_equality
        print("  Value Equals Expected:", are_equal)
        if not are_equal:
            print("  VALUE MISMATCH! Expected:", test_equality)
            return False
    elif expected_val != None: # For cases where direct == might be tricky (like complex dicts)
        print("  ExpectedVal :", expected_val)
        # This is more of a visual check unless we implement deep compare in pylearn
    return True

# --- Test json.dumps ---
print("\n=== Testing json.dumps ===")

# 1. Basic Types
check_result("dumps None", json.dumps(None), "<class 'str'>", "null")
check_result("dumps True", json.dumps(True), "<class 'str'>", "true")
check_result("dumps False", json.dumps(False), "<class 'str'>", "false")
check_result("dumps Integer", json.dumps(12345), "<class 'str'>", "12345")
check_result("dumps Negative Integer", json.dumps(-678), "<class 'str'>", "-678")
check_result("dumps Float", json.dumps(98.76), "<class 'str'>", "98.76") # Go might format float differently
check_result("dumps String", json.dumps("hello pylearn!"), "<class 'str'>", "\"hello pylearn!\"")
check_result("dumps String with quotes", json.dumps("it's \"quoted\""), "<class 'str'>", "\"it's \\\"quoted\\\"\"")

# 2. Collections
my_list = [1, "a", True, None, 3.14]
# Note: Python's json.dumps output for list with float might be 3.140000 or similar.
# Go's default float formatting for json.Marshal is usually precise.
# We'll compare visually first.
check_result("dumps List", json.dumps(my_list), "<class 'str'>", test_equality="[1,\"a\",true,null,3.14]")

my_dict = {"key1": "value1", "num": 100, "active": False, "inner_list": [8,9]}
# IMPORTANT: JSON object key order != guaranteed. Direct string comparison can fail.
# The test in Go should unmarshal and compare maps for robustness.
# For this Pylearn script, we'll print and visually inspect or check for substrings.
dumped_dict_str = json.dumps(my_dict)
print("\nTesting: dumps Dictionary")
print("  Actual Value:", dumped_dict_str)
print("  Actual Type :", type(dumped_dict_str))
print("  Expected Contains: \"key1\":\"value1\" and \"num\":100")
if "\"key1\":\"value1\"" in dumped_dict_str and "\"num\":100" in dumped_dict_str:
    print("  Substring checks PASSED.")
else:
    print("  Substring checks FAILED.")


# 3. Indentation
indented_str = json.dumps({"name": "PyLearn", "version": 0.1}, indent=2)
expected_indent_structure = """{
  "name": "PyLearn",
  "version": 0.1
}""" # Key order might differ from Go's output
print("\nTesting: dumps with indent=2")
print("  Actual Value:\n", indented_str)
print("  Expected Structure (order may vary):\n", expected_indent_structure)
if "\"name\": \"PyLearn\"" in indented_str and "\"version\": 0.1" in indented_str and "\n  " in indented_str :
    print("  Indent structure checks PASSED.")
else:
    print("  Indent structure checks FAILED.")



# 4. Errors with dumps (unserializable types)
print("\nTesting: dumps unserializable type (function)")
try:
    def my_func(): 
        pass
    json.dumps(my_func)
    print("  ERROR: json.dumps(function) did not raise error!")
except Exception as e: # Generic except
    # In Python this is TypeError. Your Pylearn might have a generic runtime error or specific JSON error.
    # Example: Check the error message if possible (requires 'as e' and error object introspection)
    print("  Successfully caught error for json.dumps(function) (expected).")


print("\n\n=== Testing json.loads ===")

# 1. Basic Types
check_result("loads null", json.loads("null"), "<class 'NoneType'>", None)
check_result("loads true", json.loads("true"), "<class 'bool'>", True)
check_result("loads false", json.loads("false"), "<class 'bool'>", False)
check_result("loads integer", json.loads("789"), "<class 'int'>", 789)
check_result("loads negative integer", json.loads("-101"), "<class 'int'>", -101)
check_result("loads float", json.loads("123.456"), "<class 'float'>", 123.456)
check_result("loads string", json.loads("\"hello from JSON\""), "<class 'str'>", "hello from JSON")

# 2. Collections
loaded_list = json.loads("[false, null, \"item\", 55]")
print("\nTesting: loads list")
print("  Actual Value:", loaded_list)
print("  Actual Type :", type(loaded_list))
if type(loaded_list) == "<class 'list'>" and len(loaded_list) == 4: # Assuming list type and len
    print("  List length check PASSED.")
    # Check elements (requires list indexing and type checks)
    print("  Item 0:", loaded_list[0], type(loaded_list[0])) # False
    print("  Item 1:", loaded_list[1], type(loaded_list[1])) # None
    print("  Item 2:", loaded_list[2], type(loaded_list[2])) # "item"
    print("  Item 3:", loaded_list[3], type(loaded_list[3])) # 55
else:
    print("  List structure check FAILED.")


loaded_dict = json.loads('{"name": "PyLearn", "score": 99.9, "tags": ["dev", "fun"], "valid": true}')
print("\nTesting: loads dictionary")
print("  Actual Value:", loaded_dict)
print("  Actual Type :", type(loaded_dict))
if type(loaded_dict) == "<class 'dict'>": # Assuming dict type
    print("  Dict type check PASSED.")
    # Check elements (requires dict indexing and type checks)
    print("  loaded_dict['name']:", loaded_dict["name"], type(loaded_dict["name"]))
    print("  loaded_dict['score']:", loaded_dict["score"], type(loaded_dict["score"]))
    print("  loaded_dict['tags']:", loaded_dict["tags"], type(loaded_dict["tags"])) # This will be a list
    if type(loaded_dict["tags"]) == "<class 'list'>" and len(loaded_dict["tags"]) == 2:
        print("    loaded_dict['tags'][0]:", loaded_dict["tags"][0])
else:
    print("  Dict structure check FAILED.")

# 3. Errors with loads (invalid JSON)
print("\nTesting: loads invalid JSON")
try:
    json.loads("this != valid json {")
    print("  ERROR: json.loads(invalid_json) did not raise error!")
except Exception as e: # Generic except
    # In Python this is JSONDecodeError.
    print("  Successfully caught error for json.loads(invalid_json) (expected).")

print("\nTesting: loads with wrong input type")
try:
    json.loads(12345) # Should be a string
    print("  ERROR: json.loads(integer) did not raise error!")
except Exception as e: # Generic except
    # In Python this is TypeError:
    print("  Successfully caught error for json.loads(integer) (expected).")


print("\n--- PyLearn JSON Module Test Done ---")

for x in [1,2,3]:
    print(x)

print('1' in ['12345'])