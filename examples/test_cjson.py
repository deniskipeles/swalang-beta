# pylearn/examples/test_cjson.py

import cjson

def run_cjson_tests():
    print("--- Running cJSON Wrapper Tests ---")

    # Test 1: loads (string to Pylearn object)
    print("Test 1: loads()")
    json_string = '{"name": "pylearn", "version": 0.1, "is_awesome": true, "features": ["ffi", "jit", null]}'
    data = cjson.loads(json_string)

    assert isinstance(data, dict), "loads() should return a dictionary"
    assert data["name"] == "pylearn", "String value mismatch"
    assert data["version"] == 0.1, "Float value mismatch"
    assert data["is_awesome"] is True, "Boolean value mismatch"
    assert data["features"][0] == "ffi", "Array value mismatch"
    assert data["features"][2] is None, "Null value mismatch"
    print("Test 1 PASSED.\n")

    # Test 2: dumps (Pylearn object to string)
    print("Test 2: dumps()")
    pylearn_obj = {
        "project": "pylearn",
        "active": True,
        "version": 1,
        "libs": ["uv", "zlib", "cjson"]
    }
    json_output = cjson.dumps(pylearn_obj, True)
    
    # We load it back to verify the content, as key order is not guaranteed
    reloaded_obj = cjson.loads(json_output)
    assert reloaded_obj["project"] == "pylearn"
    assert reloaded_obj["libs"][1] == "zlib"
    print("Test 2 PASSED.\n")
    
    # Test 3: Error handling
    print("Test 3: Error Handling")
    try:
        cjson.loads('{"key": "value",}') # Invalid JSON with trailing comma
        assert False, "Test 3 Failed: Invalid JSON should raise CJSONError"
    except cjson.CJSONError as e:
        print(format_str("Caught expected parsing error: {e}"))

    try:
        # A function object is not serializable
        cjson.dumps({"func": cjson.loads})
        assert False, "Test 3 Failed: Non-serializable type should raise CJSONError"
    except cjson.CJSONError as e:
        print(format_str("Caught expected serialization error: {e}"))
    
    print("Test 3 PASSED.\n")

# --- Main script ---
if not cjson.CJSON_AVAILABLE:
    print("cJSON library is not available, cannot run tests.")
else:
    print("Running cJSON Wrapper Tests...")
    run_cjson_tests()
    print("All cjson tests completed successfully!")