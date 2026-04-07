# pylearn/stdlib/lzma.py

"""
A Pylearn wrapper for the liblzma compression library (XZ Utils),
built using the Pylearn FFI. This module provides functions to compress
and decompress data using the LZMA and XZ formats.
"""

import ffi
import sys

# ==============================================================================
#  Platform-Aware Library Loading
# ==============================================================================

def _load_library_with_fallbacks(base_name):
    """Tries to load liblzma using common platform-specific names."""
    platform = sys.platform
    
    candidates = []
    if platform == 'windows':
        candidates = ['liblzma.dll']
    elif platform == 'darwin':
        candidates = ['liblzma.5.dylib', 'liblzma.dylib']
    else: # Linux and other Unix-like
        candidates = ['liblzma.so.5', 'liblzma.so']

    last_error = None
    for name in candidates:
        try:
            return ffi.CDLL(name)
        except ffi.FFIError as e:
            last_error = e
    
    raise last_error

_lib = None
LZMA_AVAILABLE = False
try:
    _lib = _load_library_with_fallbacks("lzma")
    LZMA_AVAILABLE = True
except ffi.FFIError as e:
    print(format_str("Warning: Failed to load liblzma: {e}"))
    print("LZMA compression functionality will not be available.")

# ==============================================================================
#  Exception and Constants
# ==============================================================================

class LZMAError(Exception):
    """Exception raised for lzma-related errors."""
    pass

# liblzma C return codes
LZMA_OK = 0
LZMA_STREAM_END = 1
LZMA_NO_CHECK = 2
LZMA_UNSUPPORTED_CHECK = 3
LZMA_GET_CHECK = 4
LZMA_MEM_ERROR = 5
LZMA_MEMLIMIT_ERROR = 6
LZMA_FORMAT_ERROR = 7
LZMA_OPTIONS_ERROR = 8
LZMA_DATA_ERROR = 9
LZMA_BUF_ERROR = 10

# Compression presets
PRESET_DEFAULT = 6

# Integrity checks
CHECK_NONE = 0
CHECK_CRC32 = 1
CHECK_CRC64 = 4 # Default for xz
CHECK_SHA256 = 10

# A very large memory limit for decompression by default.
# Users can override this if they need to handle untrusted files safely.
MEMLIMIT_DEFAULT = 9223372036854775807 # 2**63 - 1

# ==============================================================================
#  Define C Function Signatures
# ==============================================================================

# ==============================================================================
#  Define C Function Signatures
# ==============================================================================

if LZMA_AVAILABLE:
    # For one-shot compression of a buffer
    _lzma_easy_buffer_encode = _lib.lzma_easy_buffer_encode(
        [
            ffi.c_uint32,         # preset
            ffi.c_int32,          # check
            ffi.c_void_p,         # allocator (NULL)
            ffi.c_void_p,         # in (source buffer)
            ffi.c_uint64,         # in_size
            ffi.c_void_p,         # out (destination buffer)
            ffi.POINTER(ffi.c_uint64),  # out_pos (using uint64 for size_t*)
            ffi.c_uint64          # out_size
        ],
        ffi.c_int32  # return type
    )

    # For one-shot decompression of a buffer
    _lzma_stream_buffer_decode = _lib.lzma_stream_buffer_decode(
        [
            ffi.POINTER(ffi.c_uint64),  # memlimit
            ffi.c_uint32,               # flags
            ffi.c_void_p,               # allocator (NULL)
            ffi.c_void_p,               # in (source buffer)
            ffi.POINTER(ffi.c_uint64),  # in_pos (using uint64 for size_t*)
            ffi.c_uint64,               # in_size
            ffi.c_void_p,               # out (destination buffer)
            ffi.POINTER(ffi.c_uint64),  # out_pos (using uint64 for size_t*)
            ffi.c_uint64                # out_size
        ],
        ffi.c_int32  # return type
    )

# ==============================================================================
#  High-Level Pylearn API
# ==============================================================================

def compress(data, preset=PRESET_DEFAULT, check=CHECK_CRC64):
    """
    Compresses the data, returning a bytes object in the .xz format.
    
    :param data: The input data, as bytes or a string (which will be UTF-8 encoded).
    :param preset: An integer from 0 to 9 for compression level.
    :param check: The integrity check to use (e.g., lzma.CHECK_CRC64).
    :return: A bytes object containing the compressed data.
    """
    if not LZMA_AVAILABLE:
        raise LZMAError("liblzma library not available")

    if isinstance(data, str):
        source_bytes = data.encode('utf-8')
    elif isinstance(data, bytes):
        source_bytes = data
    else:
        raise TypeError("a bytes-like object is required, not a '" + str(type(data)) + "'")

    source_len = len(source_bytes)
    
    # Use a more accurate buffer size estimation
    # The maximum expansion for LZMA is about 0.5% + 12 bytes for very small inputs
    # But for safety, we'll use a larger bound
    bound = max(source_len + (source_len // 4) + 1024, 8192)
    
    out_buf = ffi.malloc(bound)
    out_pos_ptr = ffi.malloc(ffi.c_uint64.Size())
    
    try:
        # Initialize out_pos to 0
        ffi.write_memory(out_pos_ptr, ffi.c_uint64, 0)
        
        res = _lzma_easy_buffer_encode(
            preset,           # preset
            check,            # check
            None,             # allocator (NULL)
            source_bytes,     # in
            source_len,       # in_size
            out_buf,          # out
            out_pos_ptr,      # out_pos
            bound             # out_size
        )

        if res != LZMA_OK:
            raise LZMAError("lzma compression error: {}".format(res))

        actual_size = ffi.read_memory(out_pos_ptr, ffi.c_uint64)
        compressed_data = ffi.buffer_to_bytes(out_buf, actual_size)
        return compressed_data
    finally:
        ffi.free(out_buf)
        ffi.free(out_pos_ptr)

def decompress(data, memlimit=MEMLIMIT_DEFAULT):
    """
    Decompresses the data, returning a bytes object.
    This function can decompress both .xz and .lzma formats.
    
    :param data: The compressed data, as a bytes object.
    :param memlimit: A memory usage limit in bytes to prevent decompression bombs.
    :return: A bytes object containing the decompressed data.
    """
    if not LZMA_AVAILABLE:
        raise LZMAError("liblzma library not available")

    if not isinstance(data, bytes):
        raise TypeError("a bytes-like object is required, not a '" + str(type(data)) + "'")
        
    source_len = len(data)
    
    # Pointers required by the C function
    memlimit_ptr = ffi.malloc(ffi.c_uint64.Size())
    in_pos_ptr = ffi.malloc(ffi.c_uint64.Size())
    out_pos_ptr = ffi.malloc(ffi.c_uint64.Size())
    out_buf = None

    try:
        ffi.write_memory(memlimit_ptr, ffi.c_uint64, memlimit)
        ffi.write_memory(in_pos_ptr, ffi.c_uint64, 0)  # Start reading from the beginning
        ffi.write_memory(out_pos_ptr, ffi.c_uint64, 0)  # Start writing to beginning

        # We don't know the output size, so we guess and resize if needed.
        buffer_size = max(source_len * 4, 8192)

        while True:
            out_buf = ffi.malloc(buffer_size)
            ffi.write_memory(out_pos_ptr, ffi.c_uint64, 0)  # Reset for each attempt
            
            # Reset in_pos for each attempt
            ffi.write_memory(in_pos_ptr, ffi.c_uint64, 0)
            
            # The `flags` argument is reserved and should be 0.
            res = _lzma_stream_buffer_decode(
                memlimit_ptr,     # memlimit
                0,                # flags
                None,             # allocator (NULL)
                data,             # in
                in_pos_ptr,       # in_pos
                source_len,       # in_size
                out_buf,          # out
                out_pos_ptr,      # out_pos
                buffer_size       # out_size
            )

            if res == LZMA_OK or res == LZMA_STREAM_END:
                # Success!
                break
            elif res == LZMA_BUF_ERROR:
                # Output buffer was too small. Double it and try again.
                ffi.free(out_buf)
                out_buf = None
                buffer_size = buffer_size * 2
                continue
            else:
                raise LZMAError("lzma decompression error: {}".format(res))

        actual_size = ffi.read_memory(out_pos_ptr, ffi.c_uint64)
        decompressed_data = ffi.buffer_to_bytes(out_buf, actual_size)
        return decompressed_data

    finally:
        if out_buf:
            ffi.free(out_buf)
        ffi.free(memlimit_ptr)
        ffi.free(in_pos_ptr)
        ffi.free(out_pos_ptr)