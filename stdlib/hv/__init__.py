# ./stdlib/hv/__init__.py
"""
hv.py - A Pylearn wrapper for the hv (Hypertext Virtual) C++ HTTP framework.
Provides a Pythonic interface to create HTTP servers, clients, and handle requests.
"""

import ffi

# ==============================================================================
# Check Availability
# ==============================================================================

HV_AVAILABLE = False
try:
    # Try to load the shared library
    _lib = ffi.CDLL("libhv.so")  # Adjust name: libhv.dylib (macOS), hv.dll (Windows)
    HV_AVAILABLE = True
except Exception as e:
    _lib = None
    HV_AVAILABLE = False

if not HV_AVAILABLE:
    raise ImportError("hv module: libhv could not be loaded")

# ==============================================================================
# Core Type Definitions
# ==============================================================================

# Forward declarations for pointers
class HttpContext: 
    pass
class HttpService: 
    pass
class HttpRequest: 
    pass
class HttpResponse: 
    pass
class HttpResponseWriter: 
    pass

# Create pointer types
_PHttpContext = ffi.POINTER(HttpContext)
_PHttpService = ffi.POINTER(HttpService)
_PHttpRequest = ffi.POINTER(HttpRequest)
_PHttpResponse = ffi.POINTER(HttpResponse)
_PHttpResponseWriter = ffi.POINTER(HttpResponseWriter)

# <<< START FIX >>>
# The old code used a custom CFunctionType which is not supported.
# A function pointer argument in FFI is just a pointer. We use ffi.c_void_p for its type.
_FnRequestHandlerType = ffi.c_void_p
# <<< END FIX >>>

# ==============================================================================
# HTTP Request Methods (enum)
# ==============================================================================

HTTP_GET = 1
HTTP_POST = 2
HTTP_PUT = 3
HTTP_DELETE = 4
HTTP_HEAD = 5
HTTP_OPTIONS = 6
HTTP_PATCH = 7

# ==============================================================================
# HTTP Status Codes
# ==============================================================================

HTTP_STATUS_OK = 200
HTTP_STATUS_NOT_FOUND = 404
HTTP_STATUS_INTERNAL_ERROR = 500

# ==============================================================================
# Define C Function Signatures
# ==============================================================================

# --- HttpService lifecycle ---
_new_HttpService = _lib.new_HttpService([], _PHttpService) # <<< FIX: Correct restype
_delete_HttpService = _lib.delete_HttpService([_PHttpService], None) # <<< FIX: Correct restype

# --- HttpService configuration ---
_set_port = _lib.set_port([_PHttpService, ffi.c_int32], None) # <<< FIX: Correct restype
_set_request_handler = _lib.set_request_handler([_PHttpService, _FnRequestHandlerType], None) # <<< FIX: Use correct type and restype
_start_service = _lib.start_service([_PHttpService], ffi.c_int32)

# --- HttpRequest accessors ---
_get_method = _lib.get_method([_PHttpRequest], ffi.c_int32)
_get_path = _lib.get_path([_PHttpRequest], ffi.c_char_p)
_get_header = _lib.get_header([_PHttpRequest, ffi.c_char_p], ffi.c_char_p)
_get_body = _lib.get_body([_PHttpRequest], ffi.c_char_p) # Returns char*, not void*
_get_body_size = _lib.get_body_size([_PHttpRequest], ffi.c_int64)

# --- HttpResponse writers ---
_set_status_code = _lib.set_status_code([_PHttpResponse, ffi.c_int32], None) # <<< FIX: Correct restype
_add_header = _lib.add_header([_PHttpResponse, ffi.c_char_p, ffi.c_char_p], None) # <<< FIX: Correct restype
_set_body = _lib.set_body([_PHttpResponse, ffi.c_char_p, ffi.c_int64], None) # <<< FIX: Correct restype

# --- HttpResponseWriter (streaming) ---
_new_HttpResponseWriter = _lib.new_HttpResponseWriter([_PHttpContext], _PHttpResponseWriter)
_write_chunk = _lib.write_chunk([_PHttpResponseWriter, ffi.c_char_p, ffi.c_int64], ffi.c_int32)
_close_writer = _lib.close_writer([_PHttpResponseWriter], None) # <<< FIX: Correct restype

# --- HttpContext (for streaming) ---
_get_request = _lib.get_request([_PHttpContext], _PHttpRequest)
_get_response_writer = _lib.get_response_writer([_PHttpContext], _PHttpResponseWriter)

# --- Async: AsyncHttpClient ---
_new_AsyncHttpClient = _lib.new_AsyncHttpClient([], ffi.c_void_p)

# --- Threading ---
_HThreadPool_commit = _lib.HThreadPool_commit([ffi.c_void_p], None) # <<< FIX: Correct restype

# ==============================================================================
# Pythonic Wrapper Classes
# ==============================================================================

class Request:
    """Wrapper around HttpRequest."""
    def __init__(self, ptr):
        self._ptr = ptr

    @property
    def method(self):
        m = _get_method(self._ptr)
        return {1: "GET", 2: "POST", 3: "PUT", 4: "DELETE"}.get(m, "UNKNOWN")

    @property
    def path(self):
        path_ptr = _get_path(self._ptr)
        return ffi.string_at(path_ptr) if path_ptr else ""

    @property
    def headers(self):
        # This would require iterating headers; simplified: get common ones
        return {
            "User-Agent": self.get_header("User-Agent"),
            "Content-Type": self.get_header("Content-Type"),
        }

    def get_header(self, name):
        val_ptr = _get_header(self._ptr, name.encode('utf-8'))
        return ffi.string_at(val_ptr) if val_ptr else ""

    @property
    def body(self):
        size = _get_body_size(self._ptr)
        if size == 0:
            return b""
        body_ptr = _get_body(self._ptr)
        # Use a generic pointer type for buffer_to_bytes
        generic_ptr = ffi.POINTER(ffi.c_void_p)
        generic_ptr.Address = body_ptr.Address
        return ffi.buffer_to_bytes(generic_ptr, size)

    def __repr__(self):
        return format_str("<Request {self.method} {self.path}>")


class Response:
    """Wrapper around HttpResponse."""
    def __init__(self, ptr):
        self._ptr = ptr

    def status(self, code):
        _set_status_code(self._ptr, code)
        return self

    def header(self, name, value):
        _add_header(self._ptr, name.encode('utf-8'), value.encode('utf-8'))
        return self

    def body(self, data):
        if isinstance(data, str):
            data = data.encode('utf-8')
        _set_body(self._ptr, data, len(data))
        return self

    def json(self, data):
        import json
        body = json.dumps(data)
        self.header("Content-Type", "application/json")
        return self.body(body)


class StreamResponse:
    """Wrapper for streaming responses using HttpResponseWriter."""
    def __init__(self, writer_ptr):
        self._writer = writer_ptr

    def write(self, data):
        if isinstance(data, str):
            data = data.encode('utf-8')
        res = _write_chunk(self._writer, data, len(data))
        return res == 0

    def close(self):
        _close_writer(self._writer)


class HttpContextWrapper:
    """Wrapper for HttpContext to support streaming handlers."""
    def __init__(self, ctx_ptr):
        self._ctx = ctx_ptr
        self.request = Request(_get_request(self._ctx))
        self._writer_ptr = _get_response_writer(self._ctx)
        self.response = StreamResponse(self._writer_ptr)

    def __repr__(self):
        return format_str("<HttpContextWrapper request={self.request}>")


# ==============================================================================
# Server
# ==============================================================================

class Server:
    """A high-level HTTP server using hv."""
    def __init__(self, port=8080):
        self.port = port
        self.routes = {}
        self.service_ptr = _new_HttpService()
        if not self.service_ptr or self.service_ptr.Address == 0:
            raise RuntimeError("Failed to create HttpService")
        _set_port(self.service_ptr, port)
        self._c_handler_ref = None # To prevent GC

    def route(self, path, method="GET"):
        """Decorator to register a route."""
        def decorator(handler):
            key = (method.upper(), path)
            self.routes[key] = handler
            return handler
        return decorator

    def _default_handler(self, req_ptr, res_ptr):
        req = Request(req_ptr)
        key = (req.method, req.path)
        handler = self.routes.get(key)

        if not handler:
            Response(res_ptr).status(HTTP_STATUS_NOT_FOUND).body("Not Found")
            return 0

        try:
            resp = Response(res_ptr)
            result = handler(req, resp)
            if result is not None:
                # Handler returned a value; assume it's body
                resp.body(result)
        except Exception as e:
            Response(res_ptr).status(HTTP_STATUS_INTERNAL_ERROR).body(format_str("Error: {e}"))
        return 0

    def _streaming_handler(self, ctx_ptr):
        wrapper = HttpContextWrapper(ctx_ptr)
        path = wrapper.request.path
        handler = self.routes.get(("GET", path))  # Simplified

        if not handler:
            wrapper.response.write("Not Found")
            wrapper.response.close()
            return 0

        try:
            handler(wrapper)
        except Exception as e:
            wrapper.response.write(format_str("Error: {e}"))
        finally:
            wrapper.response.close()
        return 0

    def on(self, event, handler=None):
        """Register an event handler. Supported: 'request'"""
        if event == "request":
            # <<< START FIX >>>
            # Create the callback with the correct signature:
            # ffi.callback(python_func, return_type, [arg_types...])
            c_handler = ffi.callback(
                self._default_handler,
                ffi.c_int32,  # return type: int
                [_PHttpRequest, _PHttpResponse]  # arg types: HttpRequest*, HttpResponse*
            )
            self._c_handler_ref = c_handler # Keep a reference
            _set_request_handler(self.service_ptr, c_handler)
            # <<< END FIX >>>
        elif event == "stream":
            raise ValueError("Streaming 'on' event is not correctly implemented in this wrapper.")
        else:
            raise ValueError(format_str("Unknown event: {event}"))
        return self

    def start(self):
        """Start the server (blocks)."""
        print(format_str("Starting hv server on port {self.port}..."))
        code = _start_service(self.service_ptr)
        if code != 0:
            raise RuntimeError(format_str("Failed to start server, code={code}"))
        return self


# ==============================================================================
# Client (Bonus: AsyncHttpClient)
# ==============================================================================

class Client:
    """Simple HTTP client wrapper (stub - requires more bindings)."""
    def get(self, url):
        # TODO: implement using AsyncHttpClient
        return {"status": 200, "body": format_str("GET to {url} (client stub)")}


# ==============================================================================
# Global Instances
# ==============================================================================

# Default server and client
server = Server()
client = Client()


# ==============================================================================
# Example Usage (Uncomment to test)
# ==============================================================================
"""
@server.route("/", "GET")
def index(req, resp):
    return resp.status(200).header("X-Powered-By", "Pylearn+HV").body(
        "<h1>Hello from hv.py!</h1>"
    )

@server.route("/stream", "GET")
def stream(ctx):
    ctx.response.write("chunk 1\n")
    ctx.response.write("chunk 2\n")
    # ctx.response.close()  # optional, done by wrapper

if __name__ == "__main__":
    server.on("request").start()
"""