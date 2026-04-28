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

# --- Mongoose C Function Signatures ---

# void mg_mgr_init(struct mg_mgr *mgr);
_mg_mgr_init = _lib.mg_mgr_init([ffi.c_void_p], None)

# void mg_mgr_free(struct mg_mgr *mgr);
_mg_mgr_free = _lib.mg_mgr_free([ffi.c_void_p], None)

# void mg_mgr_poll(struct mg_mgr *mgr, int ms);
_mg_mgr_poll = _lib.mg_mgr_poll([ffi.c_void_p, ffi.c_int32], None)

# struct mg_connection *mg_http_listen(struct mg_mgr *mgr, const char *url, mg_event_handler_t fn, void *fn_data);
_mg_http_listen = _lib.mg_http_listen([ffi.c_void_p, ffi.c_char_p, ffi.c_void_p, ffi.c_void_p], ffi.c_void_p)

# void mg_http_reply(struct mg_connection *c, int status_code, const char *headers, const char *body_fmt, ...);
# Notice we set is_variadic=True!
_mg_http_reply = _lib.mg_http_reply([ffi.c_void_p, ffi.c_int32, ffi.c_char_p, ffi.c_char_p], None, is_variadic=True)


class Manager:
    def __init__(self):
        # Allocate enough memory for the mg_mgr struct (1024 bytes is extremely safe)
        self.mgr_ptr = ffi.malloc(1024)
        _mg_mgr_init(self.mgr_ptr)
        self.callbacks = [] # Keep callbacks alive to prevent Garbage Collection

    def poll(self, ms):
        """Blocks for up to `ms` milliseconds to process network events."""
        _mg_mgr_poll(self.mgr_ptr, ms)

    def http_listen(self, url, handler_func):
        """
        Starts an HTTP server on the given URL.
        handler_func should accept: (connection_ptr, event_id, event_data_ptr)
        """
        # Mongoose event handler signature: 
        # void fn(struct mg_connection *c, int ev, void *ev_data, void *fn_data)
        
        def c_handler(c_ptr, ev, ev_data, fn_data):
            # We wrap the user's Python function inside this C callback
            handler_func(c_ptr, ev, ev_data)

        # Create the C callback pointer using our FFI
        cb = ffi.callback(
            c_handler, 
            None, # return type (void)
            [ffi.c_void_p, ffi.c_int32, ffi.c_void_p, ffi.c_void_p] # arg types
        )
        
        # Store the callback so Go's Garbage Collector doesn't delete it
        self.callbacks.append(cb)

        # Tell mongoose to start listening
        conn_ptr = _mg_http_listen(self.mgr_ptr, url, cb, None)
        if not conn_ptr or conn_ptr.Address == 0:
            raise Exception(format_str("Failed to listen on {url}"))
            
        print(format_str("🚀 Mongoose is now listening on {url}"))

    def free(self):
        if self.mgr_ptr:
            _mg_mgr_free(self.mgr_ptr)
            ffi.free(self.mgr_ptr)
            self.mgr_ptr = None
            for cb in self.callbacks:
                ffi.free_callback(cb)
            self.callbacks = []

def http_reply(conn_ptr, status_code, headers, body):
    """
    Helper to send an HTTP response back to the client.
    """
    # If no headers provided, pass empty string. Add \r\n at the end of headers.
    if headers:
        headers = headers + "\r\n"
    else:
        headers = ""

    # Use the variadic C function to send the response. 
    # We pass %s as the format, and the body as the variadic argument.
    _mg_http_reply(conn_ptr, status_code, headers, "%s", body)
