# Swalang

Swalang is a lightweight, Python-like programming language implemented in Go. It is designed for simplicity, ease of use, and seamless integration with C libraries through a powerful Foreign Function Interface (FFI).

## Features

- **Pythonic Syntax:** Familiar and clean syntax, making it easy to learn and write.
- **Async/Await:** Built-in support for asynchronous programming and cooperative multitasking.
- **Powerful FFI:** Call functions in shared libraries (.so, .dll, .dylib) directly from Swalang.
- **Standard Library:** Includes wrappers for popular C libraries like SDL2, LVGL, SQLite, Mongoose, and more.
- **Object-Oriented:** Supports classes, inheritance, and Python-style dunder methods.

## Documentation

Comprehensive documentation can be found in the [docs/](docs/) directory:

- [Introduction](docs/introduction.md)
- [Installation & Setup](docs/installation.md)
- [Syntax Guide](docs/syntax.md)
- [Async & Concurrency](docs/async_concurrency.md)
- [FFI Guide](docs/ffi_guide.md)
- [Standard Library Reference](docs/stdlib_reference.md)
- [Contributing Guide](docs/contributing.md)

## Quick Start

### Build the Interpreter

```bash
scripts/setup/install-zig.sh
scripts/build-shared-libs.sh
scripts/build-ui-libs.sh
scripts/build-interpreter.sh
```

### Run a Script

```bash
./bin/swalang examples/hello.py
```

## License

This project is licensed under the MIT License.
