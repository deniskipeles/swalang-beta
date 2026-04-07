# pylearn/examples/test_pcre2.py

import pcre2

def run_pcre2_tests():
    print("--- Running PCRE2 Wrapper Tests ---")

    # Test 1: Simple compilation and successful match
    print("Test 1: Simple Match")
    p = pcre2.compile("hello (\\w+)", pcre2.PCRE2_CASELESS)
    m = p.match("Hello World!")
    assert m is not None, "Test 1 Failed: Should have matched"
    assert m.group(0) == "Hello World", "Test 1 Failed: Group 0 mismatch"
    assert m.group(1) == "World", "Test 1 Failed: Group 1 mismatch"
    assert m.groups() == ("World",), "Test 1 Failed: groups() mismatch"
    m.free()
    p.free()
    print("Test 1 PASSED.\n")

    # Test 2: Search for a pattern
    print("Test 2: Search")
    text = "The quick brown fox jumps over the lazy dog."
    m = pcre2.search("fox", text)
    assert m is not None, "Test 2 Failed: Search should find 'fox'"
    assert m.group(0) == "fox", "Test 2 Failed: Search result mismatch"
    m.free()
    print("Test 2 PASSED.\n")

    # Test 3: No match
    print("Test 3: No Match")
    m = pcre2.match("goodbye", text)
    assert m is None, "Test 3 Failed: Should not have matched"
    print("Test 3 PASSED.\n")

    # Test 4: Findall
    print("Test 4: Findall")
    text_numbers = "Find all numbers: 123, 45, and 9876."
    numbers = pcre2.findall("(\\d+)", text_numbers)
    assert numbers == ["123", "45", "9876"], "Test 4 Failed: Findall with one group failed"
    
    pairs = pcre2.findall("(\\w+): (\\d+)", "item1: 10, item2: 20")
    assert pairs == [("item1", "10"), ("item2", "20")], "Test 4 Failed: Findall with multiple groups failed"
    print("Test 4 PASSED.\n")
    
    # Test 5: Error Handling
    print("Test 5: Error Handling")
    try:
        pcre2.compile("unclosed_paren(")
        # This line should not be reached
        assert False, "Test 5 Failed: Invalid pattern should raise PCREError"
    except pcre2.PCREError as e:
        print(format_str("Caught expected error: {e}"))
        assert e.code != 0, "Test 5 Failed: Error code should not be zero"
    print("Test 5 PASSED.\n")

# --- Main script ---
if not pcre2.PCRE2_AVAILABLE:
    print("libpcre2-8 is not available, cannot run tests.")
else:
    run_pcre2_tests()
    print("All pcre2 tests completed successfully!")