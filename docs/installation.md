# Installation & Setup

This guide will help you set up the Swalang development environment and build the interpreter from source.

## Prerequisites

- **Go:** Version 1.23 or later.
- **Wget:** To download C library source code.
- **Git:** For cloning the repository.
- **Zig:** Swalang uses `zig cc` as a cross-compiler for its C dependencies.

## Setup Steps

### 1. Install Zig

Swalang provides a script to install the required version of Zig:

```bash
scripts/setup/install-zig.sh
```

### 2. Build Shared Libraries

Download and build the core C libraries (mongoose, mbedtls, pcre2, yyjson, libuv, xz):

```bash
scripts/build-shared-libs.sh
```

### 3. Build UI Libraries

Build the UI-related C libraries (SDL2, LVGL, Nuklear):

```bash
scripts/build-ui-libs.sh
```

### 4. Build the Swalang Interpreter

Finally, build the Swalang interpreter binary:

```bash
scripts/build-interpreter.sh
```

The resulting binary will be located at `bin/swalang` (or `bin/swalang.exe` on Windows).

## Verifying the Installation

Run the "Hello World" example to ensure everything is working correctly:

```bash
./bin/swalang examples/hello.py
```

## Troubleshooting

- **Build Tag Error:** If you are running tests manually using `go test`, ensure you include the `-tags="en"` tag (e.g., `go test -tags="en" ./...`).
- **Missing Shared Libraries:** Ensure the `bin/` directory contains the `.so` or `.dll` files generated during the build steps. The FFI system searches these locations by default.
