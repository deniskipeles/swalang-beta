### How to Run

1.  **Install `tinyxml2`**: You must install the TinyXML2 development library.
    *   **On Ubuntu/Debian**: `sudo apt-get install libtinyxml2-dev`
    *   **On macOS (with Homebrew)**: `brew install tinyxml2`
    *   **On Windows (with MSYS2/MinGW)**: `pacman -S mingw-w64-x86_64-tinyxml2`

### Build For C
```bash
g++ -shared -fPIC -o ../../bin/libtinyxml2_bridge.so tinyxml2_shim.cpp -ltinyxml2
```