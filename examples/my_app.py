# my_app.py
import pyzlib  as zstd # No flags, no special commands

data = b"some data to compress with zstd"
compressed = zstd.compress(data, 3)
print(format_str("Compressed {len(data)} bytes to {len(compressed)} bytes."))