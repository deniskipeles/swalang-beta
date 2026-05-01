# pylearn/stdlib/mongoose/__init__.py

import ffi
import sys

def _load_library():
    platform = sys.platform
    candidates = []
    if platform == 'linux':
        candidates = ["bin/x86_64-linux/mongoose/libmongoose.so", "libmongoose.so"]
    elif platform == 'windows':
        candidates = ["bin/x86_64-windows-gnu/mongoose/mongoose.dll", "mongoose.dll"]
    elif platform == 'darwin':
        candidates = ["libmongoose.dylib"]
    
    for name in candidates:
        try:
            return ffi.CDLL(name)
        except ffi.FFIError:
            pass
    raise ffi.FFIError("Could not load mongoose shared library")

_lib = _load_library()

# --- Mongoose Constants ---
MG_EV_OPEN = 1
MG_EV_POLL = 0
MG_EV_RESOLVE = 2
MG_EV_CONNECT = 3
MG_EV_ACCEPT = 4
MG_EV_READ = 5
MG_EV_WRITE = 6
MG_EV_CLOSE = 7
MG_EV_ERROR = 8
MG_EV_HTTP_MSG = 9
MG_EV_HTTP_CHUNK = 10

# --- C Function Signatures ---
_mg_mgr_init = _lib.mg_mgr_init([ffi.c_void_p], None)
_mg_mgr_free = _lib.mg_mgr_free([ffi.c_void_p], None)
_mg_mgr_poll = _lib.mg_mgr_poll([ffi.c_void_p, ffi.c_int32], None)

_mg_http_listen = _lib.mg_http_listen([ffi.c_void_p, ffi.c_char_p, ffi.c_void_p, ffi.c_void_p], ffi.c_void_p)
_mg_http_connect = _lib.mg_http_connect([ffi.c_void_p, ffi.c_char_p, ffi.c_void_p, ffi.c_void_p], ffi.c_void_p)
_mg_http_reply = _lib.mg_http_reply([ffi.c_void_p, ffi.c_int32, ffi.c_char_p, ffi.c_char_p], None, is_variadic=True)

# --- Memory Parsers ---
def _read_mg_str(base_ptr, offset):
    """Helper to read Mongoose's non-null-terminated string struct (mg_str)"""
    str_ptr = ffi.read_memory_with_offset(base_ptr, offset, ffi.c_void_p)
    str_len = ffi.read_memory_with_offset(base_ptr, offset + 8, ffi.c_uint64)
    
    # If pointer is NULL or length is 0, return empty string
    if not str_ptr or str_ptr.Address == 0 or str_len == 0:
        return ""
        
    return ffi.string_at(str_ptr, str_len)

class HttpMessage:
    """Wraps the C mg_http_message struct passed during MG_EV_HTTP_MSG"""
    def __init__(self, ptr):
        self.ptr = ptr

    @property
    def method(self):
        # offset 64 in mg_http_message is the 'method' mg_str
        return _read_mg_str(self.ptr, 64)

    @property
    def uri(self):
        # offset 16 in mg_http_message is the 'uri' mg_str
        return _read_mg_str(self.ptr, 16)
        
    @property
    def query(self):
        # offset 32 is 'query'
        return _read_mg_str(self.ptr, 32)

    @property
    def body(self):
        # offset 1360 in mg_http_message is the 'body' mg_str (after 40 header structs)
        return _read_mg_str(self.ptr, 1360)


class Manager:
    def __init__(self):
        self.mgr_ptr = ffi.malloc(1024)
        _mg_mgr_init(self.mgr_ptr)
        self.callbacks = []

    def poll(self, ms):
        _mg_mgr_poll(self.mgr_ptr, ms)

    def http_listen(self, url, handler_func):
        def c_handler(c_ptr, ev, ev_data, fn_data):
            # We only pass parsed HTTP messages to the user for simplicity
            if ev == MG_EV_HTTP_MSG:
                req = HttpMessage(ev_data)
                handler_func(c_ptr, req)

        cb = ffi.callback(c_handler, None, [ffi.c_void_p, ffi.c_int32, ffi.c_void_p, ffi.c_void_p])
        self.callbacks.append(cb)

        conn_ptr = _mg_http_listen(self.mgr_ptr, url, cb, None)
        if not conn_ptr or conn_ptr.Address == 0:
            raise Exception(format_str("Failed to listen on {url}"))
            
        print(format_str("🚀 Mongoose Server listening on {url}"))

    def http_connect(self, url, handler_func):
        def c_handler(c_ptr, ev, ev_data, fn_data):
            if ev == MG_EV_HTTP_MSG:
                resp = HttpMessage(ev_data)
                handler_func(c_ptr, resp)
            elif ev == MG_EV_ERROR:
                print("Client Connection Error!")

        cb = ffi.callback(c_handler, None, [ffi.c_void_p, ffi.c_int32, ffi.c_void_p, ffi.c_void_p])
        self.callbacks.append(cb)

        conn_ptr = _mg_http_connect(self.mgr_ptr, url, cb, None)
        if not conn_ptr or conn_ptr.Address == 0:
            raise Exception(format_str("Failed to connect to {url}"))

    def free(self):
        if self.mgr_ptr:
            _mg_mgr_free(self.mgr_ptr)
            ffi.free(self.mgr_ptr)
            self.mgr_ptr = None
            for cb in self.callbacks:
                ffi.free_callback(cb)
            self.callbacks = []

def http_reply(conn_ptr, status_code, headers, body):
    if headers:
        headers = headers + "\r\n"
    else:
        headers = ""
    _mg_http_reply(conn_ptr, status_code, headers, "%s", body)