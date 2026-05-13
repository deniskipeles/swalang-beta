# Swalang VSCode Extension

This is the official Language Server Protocol (LSP) extension for **Swalang** (and Python syntax compatibility).

## Features
- **Real-time Syntax Validation:** Detects errors as you type.
- **Semantic Highlighting:** Highlights keywords, strings, numbers, variables, and comments natively using the Swalang compiler.
- **Auto-Activation:** Activates automatically on `.swa` and `.py` files.

## Installation (Cloud IDEs like Lightning.ai)
1. In the terminal, run `npm run build` inside the `vscode-ext` folder.
2. Run `npm run package` to generate the `.vsix` file.
3. Open the Extensions sidebar in VSCode, click the `...` menu, and select **"Install from VSIX..."**.
4. Select the generated `swalang-vscode-0.0.1.vsix` file.

## Configuration
The extension relies on the `swalang-lsp` binary. 
It will attempt to look in the `SWALANG_PATH` environment variable. If you are on a cloud IDE and variables are wiped, it will automatically fallback to the default build directory: `/some-path-to-swalang-folder/bin`.

To manually set the path, run the `set-swalang` utility included in the bin folder!