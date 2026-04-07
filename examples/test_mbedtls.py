# pylearn/examples/test_mbedtls.py

import mbedtls

def run_mbedtls_tests():
    print("--- Running mbedtls SHA-256 Wrapper Tests ---")

    # Known SHA-256 hash for "hello world"
    expected_hex = "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
    expected_bytes = b'\xb9M\'\xb9\x93M>\x08\xa5.R\xd7\xda}\xbf\xac\x84\xef\xe3zS\x80\xee\x90\x88\xf7\xac\xe2\xef\xcd\xe9'

    # Test 1: Hashing in one go using the factory function
    print("Test 1: Hashing all at once")
    h1 = mbedtls.sha256(b"hello world")
    assert h1.hexdigest() == expected_hex, "Hexdigest mismatch for single update"
    
    # Test that the object is reusable after hexdigest()
    h1.update(b"another test")
    print("Object reusable after hexdigest()")
    h1.free()
    print("Test 1 PASSED.\n")

    # Test 2: Incremental hashing
    print("Test 2: Incremental hashing with update()")
    h2 = mbedtls.Hash("sha256")
    h2.update(b"hello")
    h2.update(b" ")
    h2.update(b"world")
    
    assert h2.digest() == expected_bytes, "Digest mismatch for incremental update"
    
    # Test object reuse after digest()
    h2.update(b"new data")
    new_hash = h2.hexdigest()
    assert new_hash == "9a312da82c5a935216ce78835878f4482ea847427a20a440133c75a4a5814144"
    print("Object reusable after digest()")
    h2.free()
    print("Test 2 PASSED.\n")

    # Test 3: Hashing an empty string
    print("Test 3: Hashing empty data")
    empty_hex = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
    h4 = mbedtls.sha256(b"")
    assert h4.hexdigest() == empty_hex, "Hexdigest mismatch for empty data"
    h4.free()
    print("Test 3 PASSED.\n")


# --- Main script ---
if not mbedtls.MBEDTLS_AVAILABLE:
    print("mbedcrypto library is not available, cannot run tests.")
else:
    run_mbedtls_tests()
    print("All mbedtls tests completed successfully!")