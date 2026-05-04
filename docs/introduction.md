# Introduction to Swalang

Swalang is a simplified, Python-like programming language designed for both educational purposes and practical utility in systems that require easy integration with native C libraries.

## Goals

1.  **Readability:** Maintain a clean, indentation-based syntax similar to Python.
2.  **Native Integration:** Provide a first-class Foreign Function Interface (FFI) that makes using C libraries as natural as possible.
3.  **Modern Features:** Support essential modern programming constructs like `async/await`, `yield`, and robust exception handling.
4.  **Go-powered:** Leverage Go's performance and concurrency primitives for the interpreter's implementation.

## Project History

Originally conceived as "Pylearn," the project evolved into Swalang to better reflect its identity as a versatile scripting language. While it borrows heavily from Python's design, it is an independent implementation optimized for its specific goals of FFI and UI development.

## Core Components

- **Lexer:** Converts source code into a stream of tokens, handling indentation and complex literals.
- **Parser:** Transforms tokens into an Abstract Syntax Tree (AST), performing syntax validation.
- **Interpreter:** A tree-walking interpreter that executes the AST, managing memory and execution flow.
- **Object System:** A Go-based representation of Swalang types (integers, strings, lists, classes, etc.).
- **Standard Library:** A collection of modules, many of which are high-level wrappers around C libraries using the internal FFI.

## A Simple Example

Here's a taste of Swalang:

```python
def greet(name):
    return f"Hello, {name}!"

names = ["Alice", "Bob", "World"]
for n in names:
    print(greet(n))
```
