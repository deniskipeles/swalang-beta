# Building Real-World Applications with Swalang

This guide covers the best practices for structuring, developing, and deploying applications using Swalang.

## 1. Project Structure

A typical Swalang project should be organized to keep code, assets, and dependencies clean.

```text
my_awesome_app/
├── main.py           # Application entry point
├── models/           # Data models and database logic
│   ├── __init__.py
│   └── user.py
├── ui/               # UI components (SDL2/LVGL)
│   ├── main_window.py
│   └── widgets.py
├── assets/           # Images, fonts, and data files
│   └── logo.png
├── lib/              # Shared C libraries (.so or .dll)
└── tests/            # Test scripts
    └── test_logic.py
```

## 2. Best Practices

### Code Style
- **Indentation**: Use 4 spaces for indentation.
- **Naming**: Use `snake_case` for variables and functions, and `PascalCase` for classes.
- **F-Strings**: Use `format_str()` or the `f"..."` syntax for readable string interpolation.

### Error Handling
Be specific with exceptions. Avoid bare `except:` blocks.

```python
try:
    data = http.get(url).json()
except ValueError as e:
    print("Failed to parse JSON")
except Exception as e:
    print(format_str("Network error: {e}"))
```

### Module Organization
Use `import` and `from ... import ...` to keep your code modular.

```python
# In models/user.py
class User:
    def __init__(self, name):
        self.name = name

# In main.py
from models.user import User
alice = User("Alice")
```

## 3. Working with Native Libraries

If you are building something that requires a custom C library, place the shared library in a `lib/` directory and use `ffi` to load it.

```python
import ffi
import os

# Get path to your local lib
lib_path = os.path.join(os.getcwd(), "lib", "my_native_lib.so")
my_lib = ffi.CDLL(lib_path)

# Define and use functions
# (See FFI Guide for details)
```

## 4. Performance Optimization

1.  **Avoid Global Variables**: Accessing local variables is faster than global lookups in the Swalang VM.
2.  **Use Built-ins**: Built-in functions like `len()`, `range()`, and `sum()` are implemented in Go and are much faster than pure Swalang equivalents.
3.  **Batch Database Operations**: Use `executemany()` or transactions in SQLite to handle large data updates.
4.  **JSON Choice**: Use `yyjson` if you are processing massive amounts of JSON data frequently.

## 5. Deployment

To deploy a Swalang application:
1.  **Interpreter**: Provide the `swalang` binary for the target platform (Linux/Windows).
2.  **Shared Libraries**: Include all necessary `.so` or `.dll` files from the Swalang `bin/` directory.
3.  **Standard Library**: Include the `stdlib/` directory so the interpreter can find core modules like `os`, `http`, and `sqlite`.

A common distribution structure:
```text
dist/
├── bin/
│   ├── swalang       # The interpreter
│   └── *.so / *.dll  # Core C libraries
├── stdlib/           # Swalang standard library
└── app/              # Your application code
```

You can run your app from the root using:
```bash
./bin/swalang app/main.py
```
