# examples/archiver.py

import zlib  # The interpreter will find and load zlib.so automatically!
import os
import sys

def create_compressed_archive(input_path, output_path):
    """
    Reads a file, compresses its content, and writes it to a new file.
    """
    print(f"--- Compressing '{input_path}' to '{output_path}' ---")

    # 1. Read the source file content as bytes
    try:
        with open(input_path, "rb") as f:
            original_data = f.read()
        print(f"Read {len(original_data)} bytes from source file.")
    except OSError as e:
        print(f"Error reading source file: {e}")
        return False

    # 2. Compress the data using our zlib FFI
    try:
        compressed_data = zlib.compress(original_data, zlib.Z_BEST_COMPRESSION)
        print(f"Compressed data to {len(compressed_data)} bytes.")
    except zlib.error as e:
        print(f"Error during compression: {e}")
        return False

    # 3. Write the compressed bytes to the output file
    try:
        with open(output_path, "wb") as f:
            bytes_written = f.write(compressed_data)
        print(f"Wrote {bytes_written} bytes to archive.")
    except OSError as e:
        print(f"Error writing archive file: {e}")
        return False
        
    return True

def extract_compressed_archive(archive_path):
    """
    Reads a compressed file, decompresses its content, and returns it.
    """
    print(f"\n--- Extracting '{archive_path}' ---")
    
    # 1. Read the compressed archive as bytes
    try:
        with open(archive_path, "rb") as f:
            compressed_data = f.read()
        print(f"Read {len(compressed_data)} bytes from archive.")
    except OSError as e:
        print(f"Error reading archive file: {e}")
        return None

    # 2. Decompress the data
    try:
        decompressed_data = zlib.decompress(compressed_data)
        print(f"Decompressed data to {len(decompressed_data)} bytes.")
    except zlib.error as e:
        print(f"Error during decompression: {e}")
        return None
        
    return decompressed_data

def main_program():
    # Define file paths
    source_file = "my_document.txt"
    archive_file = "my_document.txt.z"

    # Create some sample data to compress
    sample_content = "This is a document about the Pylearn FFI.\n"
    sample_content += "The Foreign Function Interface allows Pylearn to call C code.\n"
    sample_content += "By dynamically loading plugins, we can extend the language without recompiling the interpreter.\n"
    sample_content *= 50 # Make the file large enough to show good compression ratio

    with open(source_file, "wb") as f:
        f.write(sample_content.encode("utf-8"))
    
    # --- Main Logic ---
    
    success = create_compressed_archive(source_file, archive_file)
    if not success:
        sys.exit(1)
        
    extracted_data = extract_compressed_archive(archive_file)
    if extracted_data is None:
        sys.exit(1)
        
    # --- Verification ---
    print("\n--- Verifying Integrity ---")
    
    original_data = sample_content.encode("utf-8")
    if extracted_data == original_data:
        print("SUCCESS: Decompressed data matches the original content perfectly!")
    else:
        print("FAILURE: Decompressed data does not match the original.")
        sys.exit(1)
        
    # --- Clean up ---
    os.remove(source_file)
    os.remove(archive_file)
    print("\nCleaned up temporary files.")