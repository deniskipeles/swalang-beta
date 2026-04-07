#!/usr/bin/env bash
# Build script for libhv (shared library) on Linux with a C bridge
set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

# Dependencies
echo "⚙️ 1. Installing build dependencies"
sudo apt-get update -qq
sudo apt-get install -y build-essential git cmake g++ libssl-dev

# Submodule
LIBHV_DIR=".extensions/libhv"
echo "📥 2. Updating libhv submodule"
git submodule update --init --recursive --depth=1 -- "$LIBHV_DIR"

# Build libhv as a STATIC library FIRST
echo "🔨 3. Building libhv as a static library"
cd "$LIBHV_DIR"
INSTALL_DIR_STATIC="$REPO_ROOT/bin/linux-x86_64/libhv_static_install"
# Clean previous build artifacts to ensure a fresh build
rm -rf build_static "$INSTALL_DIR_STATIC"

cmake -B build_static \
  -DCMAKE_BUILD_TYPE=Release \
  -DCMAKE_INSTALL_PREFIX="$INSTALL_DIR_STATIC" \
  -DBUILD_SHARED_LIBS=OFF \
  -DBUILD_EXAMPLES=OFF \
  -DBUILD_TESTS=OFF \
  -DWITH_HTTP_CLIENT=ON \
  -DWITH_HTTP_SERVER=ON

cmake --build build_static -j"$(nproc)"
cmake --install build_static
cd "$REPO_ROOT" # Return to repo root

# C++ Bridge File (create it in the repo root temporarily)
BRIDGE_FILE="$REPO_ROOT/libhv_c_bridge.cpp"
echo "✍️ 4. Creating C++ bridge file at $BRIDGE_FILE"
cat > "$BRIDGE_FILE" <<'EOF'
#include <HttpServer.h>
#include <string.h> // For strcmp

// <<< FIX: Use the hv namespace
using namespace hv;

// Use C linkage to prevent C++ name mangling
extern "C" {
    // Define the C-style function pointer type that matches the Python FFI callback
    typedef int (*pylearn_http_handler_t)(HttpRequest* req, HttpResponse* resp);

    // Wrapper function that libhv will call. This wrapper will then call our C-style pointer.
    int pylearn_handler_bridge(HttpRequest* req, HttpResponse* resp) {
        // <<< FIX: Get userdata from req->service, not req->server
        HttpService* service = req->service;
        if (!service) return 500;

        pylearn_http_handler_t c_handler = (pylearn_http_handler_t)service->userdata;
        if (c_handler) {
            return c_handler(req, resp);
        }
        resp->status_code = HTTP_STATUS_INTERNAL_SERVER_ERROR;
        return 500;
    }

    HttpServer* new_HttpService() { return new HttpServer(); }
    void delete_HttpService(HttpServer* server) { delete server; }
    void set_port(HttpServer* server, int port) { server->setPort(port); }

    // Store the Pylearn callback pointer in userdata and register our C++ bridge
    void set_request_handler(HttpServer* server, pylearn_http_handler_t fn) {
        server->userdata = (void*)fn;
        server->registerHttpService(pylearn_handler_bridge);
    }

    int start_service(HttpServer* server) { return server->run(); }
    
    // <<< FIX: Compare method string and return integer enum
    int get_method(HttpRequest* req) {
        const char* method_str = req->Method();
        if (strcmp(method_str, "GET") == 0) return 1;
        if (strcmp(method_str, "POST") == 0) return 2;
        if (strcmp(method_str, "PUT") == 0) return 3;
        if (strcmp(method_str, "DELETE") == 0) return 4;
        if (strcmp(method_str, "HEAD") == 0) return 5;
        if (strcmp(method_str, "OPTIONS") == 0) return 6;
        if (strcmp(method_str, "PATCH") == 0) return 7;
        return 0; // UNKNOWN
    }

    const char* get_path(HttpRequest* req) { return req->Path().c_str(); }
    const char* get_header(HttpRequest* req, const char* name) { return req->GetHeader(name).c_str(); }
    
    // <<< FIX: Use std::string API correctly
    const char* get_body(HttpRequest* req) { return req->body.c_str(); }
    long long get_body_size(HttpRequest* req) { return req->body.length(); }
    
    void set_status_code(HttpResponse* resp, int code) { resp->status_code = (http_status)code; }
    void add_header(HttpResponse* resp, const char* name, const char* value) { resp->SetHeader(name, value); }

    // <<< FIX: Use std::string API correctly
    void set_body(HttpResponse* resp, const char* data, long long len) {
        resp->body.assign(data, len);
    }
    
    // Dummy functions to match python script (not implemented in this simple bridge)
    void* new_HttpResponseWriter(void* ctx) { return nullptr; }
    int write_chunk(void* w, const char* d, long long l) { return -1; }
    void close_writer(void* w) {}
    void* get_request(void* ctx) { return nullptr; }
    void* get_response_writer(void* ctx) { return nullptr; }
    void* new_AsyncHttpClient() { return nullptr; }
    void HThreadPool_commit(void* pool) {}
}
EOF

# Build the final shared library linking the bridge and the static lib
echo "🔗 5. Building final shared library with C bridge"
FINAL_INSTALL_DIR="$REPO_ROOT/bin/linux-x86_64/lib"
mkdir -p "$FINAL_INSTALL_DIR"

g++ -shared -fPIC -o "$FINAL_INSTALL_DIR/libhv.so" \
    "$BRIDGE_FILE" \
    -I"$INSTALL_DIR_STATIC/include" \
    -L"$INSTALL_DIR_STATIC/lib" \
    -lhv -lssl -lcrypto -lpthread

echo "✅ 6. Installation complete:"
ls -l "$FINAL_INSTALL_DIR/libhv.so"

# Cleanup
echo "🧹 7. Cleaning up temporary files"
rm -rf "$INSTALL_DIR_STATIC"
rm -f "$BRIDGE_FILE"