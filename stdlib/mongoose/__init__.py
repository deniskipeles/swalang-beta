import ffi
import sys

def _load_library():
    platform = sys.platform
    candidates = []
    if platform == 'linux':
        candidates =["bin/x86_64-linux/mongoose/libmongoose.so", "libmongoose.so"]
    elif platform == 'windows':
        candidates =["bin/x86_64-windows-gnu/mongoose/mongoose.dll", "mongoose.dll"]
    elif platform == 'darwin':
        candidates = ["libmongoose.dylib"]
    
    for name in candidates:
        try:
            return ffi.CDLL(name)
        except ffi.FFIError:
            pass
    raise ffi.FFIError("Could not load mongoose shared library")

_lib = _load_library()

# --- Mongoose Constants (Matched exactly to Mongoose 7.x/8.x) ---
MG_EV_ERROR = 0
MG_EV_OPEN = 1
MG_EV_POLL = 2
MG_EV_RESOLVE = 3
MG_EV_CONNECT = 4
MG_EV_ACCEPT = 5
MG_EV_READ = 6
MG_EV_WRITE = 7
MG_EV_CLOSE = 8
MG_EV_HTTP_MSG = 9
MG_EV_HTTP_CHUNK = 10
MG_EV_WS_OPEN = 11
MG_EV_WS_MSG = 12
MG_EV_WS_CTL = 13

# --- C Function Signatures ---
_mg_mgr_init = _lib.mg_mgr_init([ffi.c_void_p], None)
_mg_mgr_free = _lib.mg_mgr_free([ffi.c_void_p], None)
_mg_mgr_poll = _lib.mg_mgr_poll([ffi.c_void_p, ffi.c_int32], None)

_mg_http_listen = _lib.mg_http_listen([ffi.c_void_p, ffi.c_char_p, ffi.c_void_p, ffi.c_void_p], ffi.c_void_p)
_mg_http_connect = _lib.mg_http_connect([ffi.c_void_p, ffi.c_char_p, ffi.c_void_p, ffi.c_void_p], ffi.c_void_p)
_mg_http_reply = _lib.mg_http_reply([ffi.c_void_p, ffi.c_int32, ffi.c_char_p, ffi.c_char_p], None, is_variadic=True)

_mg_send = _lib.mg_send([ffi.c_void_p, ffi.c_void_p, ffi.c_int32], None)

_mg_ws_connect = _lib.mg_ws_connect([ffi.c_void_p, ffi.c_char_p, ffi.c_void_p, ffi.c_void_p, ffi.c_char_p], ffi.c_void_p, is_variadic=True)
_mg_ws_send = _lib.mg_ws_send([ffi.c_void_p, ffi.c_void_p, ffi.c_uint64, ffi.c_int32], ffi.c_uint64)
_mg_ws_upgrade = _lib.mg_ws_upgrade([ffi.c_void_p, ffi.c_void_p, ffi.c_char_p], None)

# --- Memory Parsers ---
def _read_mg_str(base_ptr, offset):
    str_ptr = ffi.read_memory_with_offset(base_ptr, offset, ffi.c_void_p)
    str_len = ffi.read_memory_with_offset(base_ptr, offset + 8, ffi.c_uint64)
    if not str_ptr or not getattr(str_ptr, "Address", None) or str_len == 0:
        return ""
    return ffi.string_at(str_ptr, str_len)

def _read_mg_bytes(base_ptr, offset):
    str_ptr = ffi.read_memory_with_offset(base_ptr, offset, ffi.c_void_p)
    str_len = ffi.read_memory_with_offset(base_ptr, offset + 8, ffi.c_uint64)
    if not str_ptr or not getattr(str_ptr, "Address", None) or str_len == 0:
        return b""
    return ffi.buffer_to_bytes(str_ptr, str_len)

class HttpMessage:
    def __init__(self, ptr):
        self.ptr = ptr

    @property
    def message(self): return _read_mg_str(self.ptr, 0)
    
    @property
    def body(self): return _read_mg_str(self.ptr, 16)
    
    @property
    def body_bytes(self): return _read_mg_bytes(self.ptr, 16)
    
    @property
    def head(self): return _read_mg_str(self.ptr, 32)

    @property
    def method(self): return _read_mg_str(self.ptr, 48)

    @property
    def uri(self): return _read_mg_str(self.ptr, 64)
        
    @property
    def query(self): return _read_mg_str(self.ptr, 80)
    
    @property
    def proto(self): return _read_mg_str(self.ptr, 96)

    def get_headers(self):
        headers = {}
        curr_off = 112
        for _ in range(40):
            name = _read_mg_str(self.ptr, curr_off)
            if not name:
                break
            value = _read_mg_str(self.ptr, curr_off + 16)
            headers[name.lower()] = value
            curr_off = curr_off + 32
        return headers

class WsMessage:
    def __init__(self, ptr):
        self.ptr = ptr

    @property
    def data(self): return _read_mg_str(self.ptr, 0)
        
    @property
    def data_bytes(self): return _read_mg_bytes(self.ptr, 0)

    @property
    def flags(self): return ffi.read_memory_with_offset(self.ptr, 16, ffi.c_uint8)

class Manager:
    def __init__(self):
        self.mgr_ptr = ffi.malloc(1024)
        _mg_mgr_init(self.mgr_ptr)
        self.callbacks = []

    def poll(self, ms):
        _mg_mgr_poll(self.mgr_ptr, ms)

    def http_listen(self, url, handler_func):
        def c_handler(c_ptr, ev, ev_data, fn_data):
            try:
                # Add strict pointer safety check
                if not ev_data or not getattr(ev_data, "Address", None):
                    is_safe = False
                else:
                    is_safe = True

                if ev == MG_EV_HTTP_MSG and is_safe:
                    req = HttpMessage(ev_data)
                    handler_func(c_ptr, "HTTP_REQUEST", req)
                elif ev == MG_EV_WS_OPEN:
                    handler_func(c_ptr, "WS_OPEN", None)
                elif ev == MG_EV_WS_MSG and is_safe:
                    msg = WsMessage(ev_data)
                    handler_func(c_ptr, "WS_MSG", msg)
                elif ev == MG_EV_CLOSE:
                    handler_func(c_ptr, "CLOSE", None)
            except Exception as e:
                print(format_str("🔥 Mongoose Server Handler Error: {e}"))

        cb = ffi.callback(c_handler, None, [ffi.c_void_p, ffi.c_int32, ffi.c_void_p, ffi.c_void_p])
        self.callbacks.append(cb)

        conn_ptr = _mg_http_listen(self.mgr_ptr, url.encode('utf-8'), cb, None)
        if not conn_ptr or not getattr(conn_ptr, "Address", None):
            raise Exception(format_str("Failed to listen on {url}"))
            
        print(format_str("🚀 Mongoose Server listening on {url}"))
        return conn_ptr

    def http_connect(self, url, handler_func):
        def c_handler(c_ptr, ev, ev_data, fn_data):
            try:
                if not ev_data or not getattr(ev_data, "Address", None):
                    is_safe = False
                else:
                    is_safe = True

                if ev == MG_EV_CONNECT:
                    handler_func(c_ptr, "CONNECT", None)
                elif ev == MG_EV_HTTP_MSG and is_safe:
                    resp = HttpMessage(ev_data)
                    handler_func(c_ptr, "RESPONSE", resp)
                elif ev == MG_EV_ERROR:
                    err_msg = ""
                    if is_safe:
                        err_msg = ffi.string_at(ev_data)
                    handler_func(c_ptr, "ERROR", err_msg)
                elif ev == MG_EV_CLOSE:
                    handler_func(c_ptr, "CLOSE", None)
            except Exception as e:
                print(format_str("🔥 Mongoose Client Handler Error: {e}"))

        cb = ffi.callback(c_handler, None, [ffi.c_void_p, ffi.c_int32, ffi.c_void_p, ffi.c_void_p])
        self.callbacks.append(cb)

        conn_ptr = _mg_http_connect(self.mgr_ptr, url.encode('utf-8'), cb, None)
        if not conn_ptr or not getattr(conn_ptr, "Address", None):
            raise Exception(format_str("Failed to connect to {url}"))
        return conn_ptr

    def ws_connect(self, url, handler_func):
        def c_handler(c_ptr, ev, ev_data, fn_data):
            try:
                if not ev_data or not getattr(ev_data, "Address", None):
                    is_safe = False
                else:
                    is_safe = True

                if ev == MG_EV_OPEN:
                    handler_func(c_ptr, "OPEN", None)
                elif ev == MG_EV_WS_OPEN:
                    handler_func(c_ptr, "WS_OPEN", None)
                elif ev == MG_EV_WS_MSG and is_safe:
                    msg = WsMessage(ev_data)
                    handler_func(c_ptr, "WS_MSG", msg)
                elif ev == MG_EV_ERROR:
                    err_msg = ""
                    if is_safe:
                        err_msg = ffi.string_at(ev_data)
                    handler_func(c_ptr, "ERROR", err_msg)
                elif ev == MG_EV_CLOSE:
                    handler_func(c_ptr, "CLOSE", None)
            except Exception as e:
                print(format_str("🔥 Mongoose WS Handler Error: {e}"))

        cb = ffi.callback(c_handler, None, [ffi.c_void_p, ffi.c_int32, ffi.c_void_p, ffi.c_void_p])
        self.callbacks.append(cb)

        conn_ptr = _mg_ws_connect(self.mgr_ptr, url.encode('utf-8'), cb, None, b"")
        if not conn_ptr or not getattr(conn_ptr, "Address", None):
            raise Exception(format_str("Failed to connect to WS {url}"))
        return conn_ptr

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
        if not headers.endswith("\r\n"):
            headers = headers + "\r\n"
        headers_bytes = headers.encode('utf-8')
    else:
        headers_bytes = b""
        
    if not isinstance(body, (str, bytes)):
        body = str(body)
        
    body_bytes = body.encode('utf-8') if isinstance(body, str) else body
    _mg_http_reply(conn_ptr, status_code, headers_bytes, b"%s", body_bytes)

def send(conn_ptr, data):
    data_bytes = data.encode('utf-8') if isinstance(data, str) else data
    _mg_send(conn_ptr, data_bytes, len(data_bytes))

def ws_send(conn_ptr, data, op=1): 
    data_bytes = data.encode('utf-8') if isinstance(data, str) else data
    _mg_ws_send(conn_ptr, data_bytes, len(data_bytes), op)

def ws_upgrade(conn_ptr, req_message, extra_headers=""):
    _mg_ws_upgrade(conn_ptr, req_message.ptr, extra_headers.encode('utf-8'))