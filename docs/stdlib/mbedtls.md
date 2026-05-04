# `mbedtls` Module Reference

The `mbedtls` module provides cryptographic hashing support.

## Global Functions

- `sha256(data=b'')`: Creates a new SHA-256 `Hash` object.

## Constants

- `MD_TYPE.SHA256`

## Classes

### `Hash(name, data=b'')`
- `update(data)`: Updates the hash with new bytes.
- `digest()`: Returns the current digest as bytes and resets the hash.
- `hexdigest()`: Returns the current digest as a hex string and resets the hash.
- `free()`: Explicitly frees the underlying C memory context.
- `digest_size`: (Property) The size of the resulting hash in bytes.
