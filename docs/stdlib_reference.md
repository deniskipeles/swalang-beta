# Standard Library Reference

Swalang's standard library provides a mix of pure Swalang modules and high-level wrappers around native C libraries.

## Core Modules

### `os`
Operating system dependent functionality.

- **`os.getcwd()`**: Returns the current working directory.
- **`os.mkdir(path)`**: Creates a directory.
- **`os.remove(path)` / `os.unlink(path)`**: Deletes a file.
- **`os.rename(old, new)`**: Renames a file or directory.
- **`os.system(command)`**: Executes a command in a subshell.
- **`os.getenv(key, default=None)`**: Returns the value of an environment variable.
- **`os.listdir(path='.')`**: Returns a list of entries in a directory.
- **`os.path.exists(path)`**: Returns `True` if the path exists.
- **`os.path.isdir(path)`**: Returns `True` if the path is a directory.
- **`os.path.isfile(path)`**: Returns `True` if the path is a regular file.
- **`os.path.getsize(path)`**: Returns the size of a file in bytes.
- **`os.path.join(*paths)`**: Joins path components.
- **`os.path.split(path)`**: Splits the path into `(dirname, basename)`.
- **`os.path.abspath(path)`**: Returns the absolute version of a path.
- **`os.walk(top)`**: Directory tree generator.

### `time`
- **`time.time()`**: Returns the current time in seconds since the epoch (high precision).
- **`time.sleep(seconds)`**: Suspends execution for the given number of seconds (blocking).

### `ffi`
The Foreign Function Interface module. See the [FFI Guide](ffi_guide.md) for details.

### `asyncio` / `aio`
Asynchronous I/O and task management. See [Async & Concurrency](async_concurrency.md).

---

## Data & Communication

### [`sqlite`](stdlib/sqlite.md)
DB-API 2.0 compatible interface for SQLite.

### [`cjson`](stdlib/cjson.md)
JSON encoding and decoding using the cJSON library.

### [`mongoose`](stdlib/mongoose.md)
Embedded networking library.

### [`pcre2`](stdlib/pcre2.md)
Regular expression interface using PCRE2.

---

## Cryptography & Compression

### [`mbedtls`](stdlib/mbedtls.md)
Cryptography and hashing.

### [`zlib`](stdlib/zlib.md)
Compression/decompression.

---

## Graphics & UI

### [`sdl2`](stdlib/sdl2.md)
Low-level multimedia library.

### [`lvgl`](stdlib/lvgl.md)
Embedded graphics library.

---

## Other Modules

- **`lzma`**: Compression using the LZMA algorithm.
- **`uv`**: Low-level bindings to libuv.
- **`yyjson`**: High-performance JSON library.
