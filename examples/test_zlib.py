# pylearn/examples/test_zlib.py

import zlib

def run_compression_test():
    print("--- Running zlib Compression/Decompression Test ---")

    # Test with string data
    original_str = "hello, world! this is a test of the zlib wrapper in pylearn. " * 5
    print(format_str("Original String (length {len(original_str)}): '{original_str[:50]}...'"))
    
    compressed_str = zlib.compress(original_str)
    print(format_str("Compressed String (length {len(compressed_str)}): {compressed_str}"))
    
    decompressed_str = zlib.decompress(compressed_str).decode('utf-8')
    print(format_str("Decompressed String (length {len(decompressed_str)}): '{decompressed_str[:50]}...'"))
    
    assert original_str == decompressed_str, "String data did not match after decompression!"
    print("String test PASSED.\n")

    # Test with bytes data
    original_bytes = b"some binary data\x00\x01\x02\x03" * 10
    print(format_str("Original Bytes (length {len(original_bytes)}): {original_bytes[:50]}..."))

    compressed_bytes = zlib.compress(original_bytes, level=9) # Test with max compression
    print(format_str("Compressed Bytes (length {len(compressed_bytes)}): {compressed_bytes}"))

    decompressed_bytes = zlib.decompress(compressed_bytes)
    print(format_str("Decompressed Bytes (length {len(decompressed_bytes)}): {decompressed_bytes[:50]}..."))

    assert original_bytes == decompressed_bytes, "Bytes data did not match after decompression!"
    print("Bytes test PASSED.\n")


def run_crc32_test():
    print("--- Running zlib CRC32 Test ---")
    
    data = b"hello world"
    
    # Calculate checksum
    checksum = zlib.crc32(data)
    print(format_str("CRC32 of '{data}' is: {checksum}"))
    
    # Python's zlib.crc32(b"hello world") is 222957957
    assert checksum == 222957957, "CRC32 checksum did not match expected value!"
    
    # Test with initial value
    crc_part1 = zlib.crc32(b"hello", 0)
    crc_part2 = zlib.crc32(b" world", crc_part1)
    
    print(format_str("Incremental CRC32 is: {crc_part2}"))
    assert crc_part2 == checksum, "Incremental CRC32 did not match!"
    print("CRC32 test PASSED.\n")

# --- Main script ---
if not zlib.ZLIB_AVAILABLE:
    print("zlib is not available, cannot run tests.")
else:
    run_compression_test()
    run_crc32_test()
    print("All zlib tests completed successfully!")