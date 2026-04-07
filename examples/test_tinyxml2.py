# pylearn/examples/test_tinyxml2.py

import tinyxml2

XML_DATA = """
<root version="1.0">
    <item id="A">
        <name>First Item</name>
        <value>100</value>
    </item>
    <item id="B">
        <name>Second Item</name>
        <value>200</value>
    </item>
    <empty_item/>
</root>
"""

INVALID_XML_DATA = "<root><item>Mismatched tag</wrong_item></root>"

def run_tinyxml2_tests():
    print("--- Running TinyXML2 Wrapper Tests ---")
    
    doc = tinyxml2.XMLDocument()
    try:
        # Test 1: Successful Parsing
        doc.parse(XML_DATA)
        print("Test 1: XML parsing... PASSED")

        # Test 2: Navigating the tree
        root = doc.first_child_element("root")
        assert root is not None, "Test 2 Failed: Root element not found"
        assert root.name == "root", "Test 2 Failed: Root element name is incorrect"
        print("Test 2: Root element navigation... PASSED")

        # Test 3: Reading attributes
        assert root.attribute("version") == "1.0", "Test 3 Failed: Attribute read failed"
        assert root.attribute("nonexistent") is None, "Test 3 Failed: Non-existent attribute should be None"
        print("Test 3: Attribute reading... PASSED")

        # Test 4: Reading child elements and text
        item1 = root.first_child_element("item")
        assert item1 is not None, "Test 4 Failed: First 'item' not found"
        assert item1.attribute("id") == "A", "Test 4 Failed: item1 id attribute mismatch"
        
        name1 = item1.first_child_element("name")
        assert name1.text == "First Item", "Test 4 Failed: item1 name text mismatch"
        print("Test 4: Child and text reading... PASSED")

        # Test 5: Navigating siblings
        item2 = item1.next_sibling_element("item")
        assert item2 is not None, "Test 5 Failed: Second 'item' sibling not found"
        assert item2.attribute("id") == "B", "Test 5 Failed: item2 id attribute mismatch"
        assert item2.first_child_element("value").text == "200", "Test 5 Failed: item2 value text mismatch"
        print("Test 5: Sibling navigation... PASSED")
        
        # Test 6: Handling empty elements
        empty_item = item2.next_sibling_element("empty_item")
        assert empty_item is not None, "Test 6 Failed: Empty item not found"
        assert empty_item.text is None, "Test 6 Failed: Empty item's text should be None"
        print("Test 6: Empty element handling... PASSED")

        # Test 7: Handling parsing errors
        try:
            doc.parse(INVALID_XML_DATA)
            assert False, "Test 7 Failed: Should have raised XMLError for invalid XML"
        except tinyxml2.XMLError as e:
            print(format_str("Caught expected parsing error: {e}"))
        print("Test 7: Error handling... PASSED")

    finally:
        # CRITICAL: Always free the C++ object memory
        print("Freeing XML Document...")
        doc.free()

# --- Main script ---
run_tinyxml2_tests()
print("\nAll TinyXML2 tests completed successfully!")