# pylearn/stdlib/zlib.py

"""
A Pylearn wrapper for the zlib compression library, built using the Pylearn FFI.
This module provides functions to compress and decompress data, as well as
calculate CRC32 checksums, similar to Python's standard zlib library.
"""

import ffi
import sys

# ==============================================================================
#  Platform-Aware Library Loading
# ==============================================================================

def _load_library_with_fallbacks(base_name):
    """Tries to load zlib using common platform-specific names."""
    platform = sys.platform
    
    candidates = []
    if platform == 'windows':
        candidates = ['zlib1.dll', 'zlib.dll']
    elif platform == 'darwin':
        candidates = ['libz.1.dylib', 'libz.dylib']
    else: # Linux and other Unix-like
        candidates = ['libz.so.1', 'libz.so']

    last_error = None
    for name in candidates:
        try:
            return ffi.CDLL(name)
        except ffi.FFIError as e:
            last_error = e
    
    raise last_error

_zlib = None
ZLIB_AVAILABLE = False
try:
    _zlib = _load_library_with_fallbacks("z")
    ZLIB_AVAILABLE = True
except ffi.FFIError as e:
    print(format_str("Warning: Failed to load zlib: {e}"))
    print("Compression functionality will not be available.")

# ==============================================================================
#  Exception and Constants
# ==============================================================================

class ZlibError(Exception):
    """Exception raised for zlib-related errors."""
    pass

# zlib C return codes
Z_OK = 0
Z_STREAM_END = 1
Z_NEED_DICT = 2
Z_ERRNO = -1
Z_STREAM_ERROR = -2
Z_DATA_ERROR = -3
Z_MEM_ERROR = -4
Z_BUF_ERROR = -5
Z_VERSION_ERROR = -6

# Default compression level
Z_DEFAULT_COMPRESSION = -1

# ==============================================================================
#  Define C Function Signatures
# ==============================================================================

if ZLIB_AVAILABLE:
    # int compress2(Bytef *dest, uLongf *destLen, const Bytef *source, uLong sourceLen, int level);
    _compress2 = _zlib.compress2(
        [ffi.c_void_p, ffi.POINTER(ffi.c_ulong), ffi.c_void_p, ffi.c_ulong, ffi.c_int32],
        ffi.c_int32
    )

    # int uncompress(Bytef *dest, uLongf *destLen, const Bytef *source, uLong sourceLen);
    _uncompress = _zlib.uncompress(
        [ffi.c_void_p, ffi.POINTER(ffi.c_ulong), ffi.c_void_p, ffi.c_ulong],
        ffi.c_int32
    )
    
    # uLong crc32(uLong crc, const Bytef *buf, uInt len);
    _crc32 = _zlib.crc32(
        [ffi.c_ulong, ffi.c_void_p, ffi.c_uint32],
        ffi.c_ulong
    )

# ==============================================================================
#  High-Level Pylearn API
# ==============================================================================

def compress(data, level=Z_DEFAULT_COMPRESSION):
    """
    Compresses the data, returning a bytes object.
    
    :param data: The input data, as bytes or a string (which will be UTF-8 encoded).
    :param level: An integer from 0 to 9, or -1 for the default compression level.
    :return: A bytes object containing the compressed data.
    """
    if not ZLIB_AVAILABLE:
        raise ZlibError("zlib library not available")

    if isinstance(data, str):
        source_bytes = data.encode('utf-8')
    elif isinstance(data, bytes):
        source_bytes = data
    else:
        raise TypeError("a bytes-like object is required, not a '" + type(data) + "'")

    source_len = len(source_bytes)
    # Estimate the maximum buffer size needed for the compressed data
    bound = source_len + (source_len // 1000) + 12
    
    # Allocate memory for the destination buffer and the destination length
    dest_buf = ffi.malloc(bound)
    dest_len_ptr = ffi.malloc(ffi.c_ulong.Size())

    try:
        # Write the initial size of our buffer into the dest_len_ptr
        ffi.write_memory(dest_len_ptr, ffi.c_ulong, bound)

        # Call the C function
        res = _compress2(dest_buf, dest_len_ptr, source_bytes, source_len, level)
        if res != Z_OK:
            raise ZlibError(format_str("zlib compression error: {res}"))

        # Read the actual size of the compressed data
        actual_size = ffi.read_memory(dest_len_ptr, ffi.c_ulong)
        
        # Read the compressed data from the buffer
        compressed_data = ffi.buffer_to_bytes(dest_buf, actual_size)
        return compressed_data

    finally:
        # CRITICAL: Always free the C memory we allocated
        ffi.free(dest_buf)
        ffi.free(dest_len_ptr)

def decompress(data):
    """
    Decompresses the data, returning a bytes object.
    
    :param data: The compressed data, as a bytes object.
    :return: A bytes object containing the decompressed data.
    """
    if not ZLIB_AVAILABLE:
        raise ZlibError("zlib library not available")

    if not isinstance(data, bytes):
        raise TypeError("a bytes-like object is required, not a '" + type(data) + "'")
        
    source_len = len(data)
    
    # We don't know the uncompressed size, so we start with a guess and loop if it's too small.
    buffer_size = source_len * 2
    if buffer_size == 0:
        buffer_size = 16 # Handle empty input

    dest_buf = None
    dest_len_ptr = ffi.malloc(ffi.c_ulong.Size())

    try:
        while True:
            dest_buf = ffi.malloc(buffer_size)
            ffi.write_memory(dest_len_ptr, ffi.c_ulong, buffer_size)

            res = _uncompress(dest_buf, dest_len_ptr, data, source_len)

            if res == Z_OK:
                # Success!
                break
            elif res == Z_BUF_ERROR:
                # Buffer was too small. Free it, double the size, and try again.
                ffi.free(dest_buf)
                buffer_size = buffer_size * 2
                continue
            else:
                # A real data error occurred.
                raise ZlibError(format_str("zlib decompression error: {res}"))

        actual_size = ffi.read_memory(dest_len_ptr, ffi.c_ulong)
        decompressed_data = ffi.buffer_to_bytes(dest_buf, actual_size)
        return decompressed_data

    finally:
        if dest_buf:
            ffi.free(dest_buf)
        ffi.free(dest_len_ptr)

def crc32(data, value=0):
    """
    Computes a CRC (Cyclic Redundancy Check) checksum of data.
    
    :param data: The input data, as bytes or a string.
    :param value: An optional starting value for the checksum.
    :return: An unsigned 32-bit integer.
    """
    if not ZLIB_AVAILABLE:
        raise ZlibError("zlib library not available")

    if isinstance(data, str):
        source_bytes = data.encode('utf-8')
    elif isinstance(data, bytes):
        source_bytes = data
    else:
        raise TypeError("a bytes-like object is required, not a '" + type(data) + "'")

    # The result from C is an unsigned long.
    result = _crc32(value, source_bytes, len(source_bytes))

    # In Python, crc32 is always an unsigned 32-bit integer.
    # We must apply a bitmask to handle potential sign extension from the FFI.
    return result & 0xFFFFFFFF
