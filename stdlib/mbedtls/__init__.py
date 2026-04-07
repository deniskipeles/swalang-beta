# pylearn/stdlib/mbedtls.py

"""
A Pylearn wrapper for the mbedtls cryptography library, built using the Pylearn FFI.
This module provides a Pythonic, object-oriented interface for cryptographic
hashing algorithms like SHA-256.
"""

import ffi
import sys

# ==============================================================================
#  Platform-Aware Library Loading
# ==============================================================================

def _load_library_with_fallbacks(base_name):
    """
    Tries to load a shared library using common platform-specific names,
    including a search within the project's 'bin' directory.
    """
    platform = sys.platform
    
    candidates = []
    if platform == 'linux':
        candidates.append(format_str("bin/x86_64-linux/mbedtls/lib{base_name}.so"))
        candidates.append(format_str("lib{base_name}.so"))
    elif platform == 'windows':
        candidates.append(format_str("bin/x86_64-windows-gnu/mbedtls/{base_name}.dll"))
        candidates.append(format_str('{base_name}.dll'))
    elif platform == 'darwin':
        candidates.append(format_str('lib{base_name}.dylib'))

    last_error = None
    for name in candidates:
        try:
            return ffi.CDLL(name)
        except ffi.FFIError as e:
            last_error = e
    
    raise last_error

_lib = None
MBEDTLS_AVAILABLE = False
try:
    # Target 'mbedtls' as that's the base name.
    _lib = _load_library_with_fallbacks("mbedtls")
    MBEDTLS_AVAILABLE = True
except ffi.FFIError as e:
    # The error message from the loop will be more specific now.
    print(format_str("Warning: Failed to load mbedtls library: {e}"))
    print("Cryptography functionality in the 'mbedtls' module will not be available.")
    
# ==============================================================================
#  Exception and Constants
# ==============================================================================

class MbedError(Exception):
    """Raised for mbedtls-related errors."""
    pass

class MD_TYPE:
    """Message Digest (Hash) Algorithm Types"""
    NONE = 0
    SHA256 = 5 # From mbedtls/md.h

# ==============================================================================
#  Define C Function Signatures
# ==============================================================================

if MBEDTLS_AVAILABLE:
    # --- Message Digest (md) API functions ---
    _mbedtls_md_info_from_type = _lib.mbedtls_md_info_from_type([ffi.c_int32], ffi.c_void_p)
    _mbedtls_md_get_size = _lib.mbedtls_md_get_size([ffi.c_void_p], ffi.c_uchar)
    _mbedtls_md_init = _lib.mbedtls_md_init([ffi.c_void_p], None)
    _mbedtls_md_setup = _lib.mbedtls_md_setup([ffi.c_void_p, ffi.c_void_p, ffi.c_int32], ffi.c_int32)
    _mbedtls_md_starts = _lib.mbedtls_md_starts([ffi.c_void_p], ffi.c_int32)
    _mbedtls_md_update = _lib.mbedtls_md_update([ffi.c_void_p, ffi.c_void_p, ffi.c_uint64], ffi.c_int32)
    _mbedtls_md_finish = _lib.mbedtls_md_finish([ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
    _mbedtls_md_free = _lib.mbedtls_md_free([ffi.c_void_p], None)
    
# ==============================================================================
#  High-Level Pylearn API
# ==============================================================================

class Hash:
    """A generic hash object wrapper, similar to hashlib."""
    def __init__(self, name, data=b''):
        if not MBEDTLS_AVAILABLE:
            raise MbedError("mbedcrypto library not available")

        self.name = name.lower()
        
        if self.name == 'sha256':
            md_type = MD_TYPE.SHA256
        else:
            raise MbedError(format_str("Unsupported hash type '{self.name}'"))

        self._md_info = _mbedtls_md_info_from_type(md_type)
        if self._md_info.Address == 0:
            raise MbedError("Could not get message digest info for the requested algorithm")

        # The C function returns a Pylearn Integer object.
        digest_size_obj = _mbedtls_md_get_size(self._md_info)
        # We must convert it to a primitive integer for use in other ffi calls.
        self.digest_size = int(str(digest_size_obj))

        # Allocate memory for the mbedtls_md_context_t struct.
        # Its size is not public, so we must allocate a safe amount. 256 bytes is sufficient.
        self._ctx = ffi.malloc(256)
        self._is_finished = False
        
        try:
            _mbedtls_md_init(self._ctx)
            
            res = _mbedtls_md_setup(self._ctx, self._md_info, 0) # 0 for no HMAC
            if res != 0:
                raise MbedError(format_str("Failed to set up mbedtls md context, error {res}"))

            res = _mbedtls_md_starts(self._ctx)
            if res != 0:
                raise MbedError(format_str("Failed to start mbedtls md operation, error {res}"))

            if data:
                self.update(data)
        except MbedError as e:
            # If any part of setup fails, ensure we free the allocated context
            self.free()
            raise

    def update(self, data):
        """Update the hash object with a bytes-like object."""
        if self._is_finished:
            raise MbedError("Hash has already been finalized.")
        if not isinstance(data, (bytes, str)):
            raise TypeError("a bytes-like object is required")
        
        data_bytes = data.encode('utf-8') if isinstance(data, str) else data

        res = _mbedtls_md_update(self._ctx, data_bytes, len(data_bytes))
        if res != 0:
            raise MbedError(format_str("Failed to update hash, error {res}"))

    def digest(self):
        """Return the digest as a bytes object."""
        # This operation finalizes the current hash. We will restart the context
        # afterwards so the object can be reused for a new hash.
        if self._ctx is None:
            raise MbedError("Hash object has been freed.")

        output_buffer = ffi.malloc(self.digest_size)
        try:
            # Finalize the current hash operation
            res = _mbedtls_md_finish(self._ctx, output_buffer)
            if res != 0:
                raise MbedError(format_str("Failed to finish hash, error {res}"))
            
            # Read the result before we modify the context again
            result_bytes = ffi.buffer_to_bytes(output_buffer, self.digest_size)

            # <<< START OF FIX >>>
            # Restart the context so the object can be used again for a new hash.
            # This mimics the behavior of Python's standard hashlib library.
            res = _mbedtls_md_starts(self._ctx)
            if res != 0:
                # This indicates a problem that should make the object unusable
                self.free()
                raise MbedError(format_str("Failed to restart mbedtls md context, error {res}"))
            # <<< END OF FIX >>>
            
            return result_bytes
        finally:
            ffi.free(output_buffer)

    def hexdigest(self):
        """Return the digest as a string of hexadecimal digits."""
        # Note: We are now removing the self._is_finished check because digest()
        # automatically resets the context, allowing for reuse.
        digest_bytes = self.digest()
        
        hex_chars = "0123456789abcdef"
        result = ""
        for byte_val in digest_bytes:
            high_nibble = byte_val // 16
            low_nibble = byte_val % 16
            result = result + hex_chars[high_nibble] + hex_chars[low_nibble]
        return result

    def free(self):
        """Releases the C memory used by the hash object."""
        if self._ctx:
            _mbedtls_md_free(self._ctx)
            ffi.free(self._ctx)
            self._ctx = None

def sha256(data=b''):
    """Returns a SHA256 hash object; an optional `data` can be provided."""
    h = None
    try:
        h = Hash('sha256', data)
        return h
    except MbedError as e:
        # If the constructor fails, ensure the half-created object is freed.
        if h:
            h.free()
        raise