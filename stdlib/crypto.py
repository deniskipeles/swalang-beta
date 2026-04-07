# pylearn/stdlib/crypto.py

"""
A Pylearn implementation of common cryptographic functions, including hashing
and secure random number generation.

This module provides a high-level, object-oriented interface for hashing,
similar to Python's hashlib, built on top of a native Go backend.
"""

# from pylearn_importlib import load_module_from_path
# _crypto_native = load_module_from_path("_crypto_native")
import _crypto_native

# --- Hash Algorithm Class ---

class Hash:
    """A generic hash object wrapper."""
    def __init__(self, name, data=b''):
        # Validate the algorithm name
        if name not in ('md5', 'sha1', 'sha256'):
            raise ValueError(format_str("unsupported hash type {name}"))
        
        self.name = name
        self.digest_size = {'md5': 16, 'sha1': 20, 'sha256': 32}.get(name)
        self.block_size = {'md5': 64, 'sha1': 64, 'sha256': 64}.get(name)
        
        # This implementation is simplified. A real hashlib object would
        # manage the hash state internally and allow incremental updates.
        # For now, we store the data and hash it all at once on digest/hexdigest.
        # This matches the simplified backend but can be extended later.
        if not isinstance(data, (str, bytes)):
             raise TypeError("data must be str or bytes")
        self._data = data

    def update(self, data):
        """
        Update the hash object with the bytes-like object. Repeated calls
        are equivalent to a single call with the concatenation of all the
        arguments.
        """
        # In our simplified model, we just append the data.
        if isinstance(self._data, str) and isinstance(data, str):
            self._data = self._data + data
        elif isinstance(self._data, bytes) and isinstance(data, bytes):
            self._data = self._data + data
        else:
            # For simplicity, we won't auto-convert str to bytes.
            raise TypeError("cannot update hash with different data types (str/bytes)")

    def hexdigest(self):
        """Return the digest as a string of hexadecimal digits."""
        hash_func = getattr(_crypto_native, self.name)
        return hash_func(self._data)

    # def digest(self):
    #     """Return the digest as a bytes object."""
    #     hex_digest = self.hexdigest()
    #     # Convert hex string to bytes (requires a helper, let's implement one)
    #     # This is a good candidate for a future `binascii` module.
        
    #     # Simple implementation for now:
    #     if len(hex_digest) % 2 != 0:
    #         hex_digest = '0' + hex_digest
        
    #     b = []
    #     i = 0
    #     while i < len(hex_digest):
    #         # A bit of a hack without a proper hex-to-int converter.
    #         # This demonstrates the need for more powerful built-ins or stdlibs.
    #         # A real implementation would have a native `binascii.unhexlify`.
    #         # For now, we'll return a placeholder.
    #         # Let's assume a native helper will be added.
    #         pass # Placeholder
        
    #     # Let's add a native helper for this. For now, we'll skip this implementation.
    #     raise NotImplementedError("digest() requires hex-to-bytes conversion, not yet implemented")
    def digest(self):
        """Return the digest as a bytes object."""
        hex_digest = self.hexdigest()
        # Call the new native helper function to convert hex to bytes.
        return _crypto_native.unhexlify(hex_digest)

    def copy(self):
        """Return a copy (clone) of the hash object."""
        # The copy contains the current data state.
        return Hash(self.name, self._data)


# --- Factory Functions ---

def new(name, data=b''):
    """
    Return a new hash object implementing the given hash function.
    `data` is an optional initial chunk of data to hash.
    """
    return Hash(name, data)

def md5(data=b''):
    """Return a new MD5 hash object."""
    return new('md5', data)

def sha1(data=b''):
    """Return a new SHA1 hash object."""
    return new('sha1', data)

def sha256(data=b''):
    """Return a new SHA256 hash object."""
    return new('sha256', data)


# --- Secure Randomness ---

def token_bytes(nbytes=None):
    """
    Return a string containing `nbytes` random bytes. If `nbytes` is None or
    not supplied, a reasonable default is used.
    """
    if nbytes is None:
        nbytes = 32
    return _crypto_native.rand_bytes(nbytes)

def token_hex(nbytes=None):
    """
    Return a random text string, in hexadecimal. The string has `nbytes`
    random bytes, each byte converted to two hex digits.
    """
    random_bytes = token_bytes(nbytes)
    # This requires bytes-to-hex conversion. A native helper is ideal.
    # For now, we'll implement a Pylearn version.
    hex_chars = "0123456789abcdef"
    result = ""
    for byte_val in random_bytes: # Iterating over bytes yields integers
        high_nibble = byte_val // 16
        low_nibble = byte_val % 16
        result = result + hex_chars[high_nibble] + hex_chars[low_nibble]
    return result