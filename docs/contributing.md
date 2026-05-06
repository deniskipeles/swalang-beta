# Contributing to Swalang

Thank you for your interest in contributing to Swalang! This guide provides an overview of the internal architecture to help you get started.

## Internal Architecture

Swalang is implemented in Go and follows a classic interpreter design.

### 1. Lexer (`internal/lexer`)

The lexer (or scanner) takes the raw source code and converts it into a stream of `Token` objects. It handles:
- **Indentation:** Swalang uses a stack-based approach to track indentation levels and emit `INDENT` and `DEDENT` tokens, similar to Python.
- **Literals:** Recognizing integers, floats, strings (including f-strings), and bytes.
- **Comments:** Stripping out single-line (`#`) and multiline comments.

### 2. Parser (`internal/parser`)

The parser takes the token stream and builds an Abstract Syntax Tree (AST). It is a recursive descent parser that handles:
- **Operator Precedence:** Using Pratt parsing for expressions.
- **Statement Structure:** Parsing `if`, `for`, `while`, `def`, `class`, and other high-level constructs.
- **Error Reporting:** Collecting and reporting syntax errors with line and column information.

### 3. AST (`internal/ast`)

The AST consists of node structures that represent the program. Each node implements the `Node` interface, with specific interfaces for `Statement` and `Expression`.

### 4. Object System (`internal/object`)

Every value in Swalang is represented by a Go struct that implements the `Object` interface.
- **Environment:** Stores variable bindings in a hierarchical structure (supporting closures).
- **Builtins:** Native Go functions that are exposed to Swalang (found in `internal/builtins`).

### 5. Interpreter (`internal/interpreter`)

The interpreter (often referred to as `Eval`) walks the AST and executes each node.
- **ExecutionContext:** Maintains the state of execution, including the current environment and async runtime.
- **Method Dispatch:** Handles dunder method lookups and method binding.

### 6. Standard Library (`stdlib/`)

Most of Swalang's standard library is written in Swalang itself, often acting as a high-level wrapper around C libraries via the FFI.

## Development Workflow

### Running Tests

Ensure you use the `en` build tag when running Go tests:

```bash
go test -tags="en" ./...
```

### Adding a Builtin

1. Define the function in a file within `internal/builtins/`.
2. Register it in the `init()` function using `registerBuiltin`.

### Modifying the Lexer/Parser

If you add new syntax, you will likely need to update both the lexer (to recognize new tokens) and the parser (to handle the new grammatical structure).
