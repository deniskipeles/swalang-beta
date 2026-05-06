# `zlib` Module Reference

The `zlib` module provides compression and decompression functionality using the zlib library.

## Global Functions

- `compress(data, level=-1)`: Compresses a bytes-like object. `level` ranges from 0 (no compression) to 9 (max compression). `-1` is the default.
- `decompress(data)`: Decompresses a compressed bytes-like object.
- `crc32(data, value=0)`: Computes a CRC32 checksum.

## Constants

- `Z_OK`, `Z_STREAM_END`, `Z_NEED_DICT`, `Z_ERRNO`, `Z_STREAM_ERROR`, `Z_DATA_ERROR`, `Z_MEM_ERROR`, `Z_BUF_ERROR`, `Z_VERSION_ERROR`
- `Z_DEFAULT_COMPRESSION`
