import crypto
import sys

def test_hashing():
    print("--- Testing Hashing API ---")
    
    # Using the factory functions
    h1 = crypto.sha256(b"hello world")
    print(format_str("sha256('hello world'): {h1.hexdigest()}"))
    
    # Using the 'new' constructor and incremental updates
    h2 = crypto.new('sha256')
    h2.update(b"hello")
    h2.update(b" ")
    h2.update(b"world")
    print(format_str("sha256('hello' + ' ' + 'world'): {h2.hexdigest()}"))
    
    if h1.hexdigest() != h2.hexdigest():
        print("ERROR: Incremental and direct hashing produced different results!")
        sys.exit(1)
    else:
        print("SUCCESS: Incremental and direct hashing match.")
        
    # # Test a different algorithm
    # md5_hash = crypto.md5("some string data".encode("utf-8")) # Assuming str.encode exists
    # print(format_str("md5('some string data'): {md5_hash.hexdigest()}"))
    md5_hash = crypto.md5(b"some string data")
    print(format_str("md5('some string data'): {md5_hash.hexdigest()}"))


def test_randomness():
    print("\n--- Testing Secure Randomness ---")
    
    # Generate random bytes
    random_data = crypto.token_bytes(16)
    print(format_str("Generated 16 random bytes. Length: {len(random_data)}"))
    
    # Generate a secure hex token (e.g., for API keys, session IDs)
    hex_token = crypto.token_hex(16)
    print(format_str("Generated 32-character hex token: {hex_token}"))
    print(format_str("Token length: {len(hex_token)}"))


def main():
    test_hashing()
    test_randomness()

if __name__ == '__main__':
    main()