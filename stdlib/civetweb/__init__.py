# pylearn/stdlib/civetweb/__init__.py
"""
A Pylearn wrapper for the Civetweb C web server.

This module provides a simple, Pythonic interface for creating an embedded
web server, backed by the high-performance Civetweb library via Pylearn's FFI.
"""

import ffi
import sys

# ==============================================================================
#  Platform-Aware Library Loading
# ==============================================================================

_lib = None
CIVETWEB_AVAILABLE = False
try:
    _lib = ffi.CDLL("libcivetweb.so")
    CIVETWEB_AVAILABLE = True
except ffi.FFIError as e:
    print("Warning: Failed to load libcivetweb.so. The 'civetweb' module will be unavailable.")

# ==============================================================================
#  Exceptions and Constants
# ==============================================================================

class CivetwebError(Exception):
    """Exception for Civetweb-related errors."""
    pass

# Opaque C struct forward declarations for the FFI
class _mg_connection: 
    pass
class _mg_context: 
    pass
class _mg_request_info: 
    pass

# ==============================================================================
#  FFI C Type and Function Definitions
# ==============================================================================

if CIVETWEB_AVAILABLE:
    # --- Pointer Types ---
    _mg_connection_p = ffi.POINTER(_mg_connection)
    _mg_context_p = ffi.POINTER(_mg_context)
    _mg_request_info_p = ffi.POINTER(_mg_request_info)

    # --- THE FIX: Define the mg_callbacks struct that mg_start expects ---
    @ffi.Struct
    class mg_callbacks:
        # <<< START OF THE FIX >>>
        # The FFI backend expects a list of lists, not a list of tuples.
        _fields_ = [
            ['begin_request', ffi.c_void_p]
        ]
        # <<< END OF THE FIX >>>

    # --- Core Functions ---
    # The first argument is now a pointer to our mg_callbacks struct.
    _mg_start = _lib.mg_start(
        [ffi.POINTER(mg_callbacks), ffi.c_void_p, ffi.POINTER(ffi.c_char_p)],
        _mg_context_p
    )
    _mg_stop = _lib.mg_stop([_mg_context_p], None)

    # --- Request Info Functions (unchanged) ---
    _mg_get_request_info = _lib.mg_get_request_info([_mg_connection_p], _mg_request_info_p)
    _mg_get_header = _lib.mg_get_header([_mg_connection_p, ffi.c_char_p], ffi.c_char_p)
    _mg_read = _lib.mg_read([_mg_connection_p, ffi.c_void_p, ffi.c_uint64], ffi.c_int64)

    # --- Response Functions (unchanged) ---
    _mg_send_http_ok = _lib.mg_send_http_ok(
        [_mg_connection_p, ffi.c_char_p, ffi.c_void_p, ffi.c_uint64],
        ffi.c_int32
    )
    _mg_send_http_error = _lib.mg_send_http_error(
        [_mg_connection_p, ffi.c_int32, ffi.c_char_p],
        ffi.c_int32
    )

# ==============================================================================
#  Pythonic Wrapper Classes
# ==============================================================================
# (The rest of the file remains unchanged as it was correct in the previous step)

class Request:
    """A wrapper around the C mg_request_info struct."""
    def __init__(self, conn_ptr, info_ptr):
        self._conn = conn_ptr
        self._info = info_ptr

    @property
    def method(self):
        method_ptr = ffi.read_memory_with_offset(self._info, 0, ffi.c_char_p)
        return ffi.string_at(method_ptr) if method_ptr.Address != 0 else ""

    @property
    def uri(self):
        uri_ptr = ffi.read_memory_with_offset(self._info, 8, ffi.c_char_p)
        return ffi.string_at(uri_ptr) if uri_ptr.Address != 0 else ""

    @property
    def query_string(self):
        query_ptr = ffi.read_memory_with_offset(self._info, 24, ffi.c_char_p)
        return ffi.string_at(query_ptr) if query_ptr.Address != 0 else ""

    def get_header(self, name):
        val_ptr = _mg_get_header(self._conn, name.encode('utf-8'))
        return ffi.string_at(val_ptr) if val_ptr.Address != 0 else None
    
    def read(self, size=8192):
        buffer = ffi.malloc(size)
        try:
            bytes_read = _mg_read(self._conn, buffer, size)
            if bytes_read > 0:
                return ffi.buffer_to_bytes(buffer, bytes_read)
            return b''
        finally:
            ffi.free(buffer)

class Server:
    """A Pylearn wrapper for a Civetweb server instance."""
    def __init__(self, options=None):
        if not CIVETWEB_AVAILABLE:
            raise CivetwebError("Civetweb library is not loaded.")
        
        self.options = options if options is not None else {}
        self.routes = {}
        self._context = None
        self._c_callback_ref = None # To prevent garbage collection

    def route(self, uri, method="GET"):
        """Decorator to register a request handler for a specific route."""
        def decorator(handler):
            self.routes[(method.upper(), uri)] = handler
            return handler
        return decorator

    def _master_callback(self, conn_ptr, user_data_ptr):
        """The single C callback that dispatches to Pylearn handlers."""
        try:
            info_ptr = _mg_get_request_info(conn_ptr)
            if not info_ptr or info_ptr.Address == 0:
                return 0
            
            req = Request(conn_ptr, info_ptr)
            handler = self.routes.get((req.method, req.uri))

            if handler:
                response_body = handler(req)
                
                if isinstance(response_body, str):
                    body_bytes = response_body.encode('utf-8')
                    _mg_send_http_ok(conn_ptr, "text/html; charset=utf-8", body_bytes, len(body_bytes))
                elif isinstance(response_body, bytes):
                     _mg_send_http_ok(conn_ptr, "application/octet-stream", response_body, len(response_body))

                return 1 # Handled
            else:
                _mg_send_http_error(conn_ptr, 404, "Not Found")
                return 1 # Handled (with a 404)
        except Exception as e:
            print(format_str("Error in request handler: {e}"))
            _mg_send_http_error(conn_ptr, 500, "Internal Server Error")
            return 1

    def start(self):
        """Starts the Civetweb server. This call blocks."""
        if self._context:
            raise CivetwebError("Server is already running.")

        opts_list = []
        for item_tuple in self.options.items():
            opts_list.append(str(item_tuple[0]))
            opts_list.append(str(item_tuple[1]))

        num_opts = len(opts_list)
        c_options_array = ffi.malloc(ffi.c_char_p.Size() * (num_opts + 1))
        callbacks_struct_ptr = ffi.malloc(mg_callbacks.Size())
        
        pinned_byte_objects = []
        
        try:
            i = 0
            while i < num_opts:
                pylearn_bytes = opts_list[i].encode('utf-8')
                pinned_byte_objects.append(pylearn_bytes)
                offset = i * ffi.c_char_p.Size()
                ffi.write_memory_with_offset(c_options_array, offset, ffi.c_char_p, pylearn_bytes)
                i = i + 1
            null_term_offset = num_opts * ffi.c_char_p.Size()
            ffi.write_memory_with_offset(c_options_array, null_term_offset, ffi.c_char_p, None)

            self._c_callback_ref = ffi.callback(
                self._master_callback, 
                ffi.c_int32,
                [_mg_connection_p, ffi.c_void_p]
            )
            
            ffi.write_memory(callbacks_struct_ptr, ffi.c_void_p, self._c_callback_ref)
            
            print("Starting Civetweb server...")
            
            self._context = _mg_start(callbacks_struct_ptr, None, c_options_array)

            if not self._context or self._context.Address == 0:
                raise CivetwebError("Failed to start Civetweb server. Check server options and port availability.")
            
            print("Civetweb server stopped.")

        finally:
            ffi.free(c_options_array)
            ffi.free(callbacks_struct_ptr)

    def stop(self):
        """Stops the running server."""
        if self._context:
            print("Stopping Civetweb server...")
            _mg_stop(self._context)
            self._context = None