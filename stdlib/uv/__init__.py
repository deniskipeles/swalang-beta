"""
A Pylearn wrapper for libuv.
The engine behind Node.js and Python's uvloop.
"""

import ffi
import sys

def _load_library():
    platform = sys.platform
    candidates =[]
    if platform == 'linux':
        candidates =["bin/x86_64-linux/libuv/libuv.so", "libuv.so", "libuv.so.1"]
    elif platform == 'windows':
        candidates =["bin/x86_64-windows-gnu/libuv/uv.dll", "libuv.dll", "uv.dll"]
    elif platform == 'darwin':
        candidates = ["libuv.dylib"]
    
    for name in candidates:
        try:
            return ffi.CDLL(name)
        except ffi.FFIError:
            pass
    raise ffi.FFIError("Could not load libuv shared library")

_lib = _load_library()

# --- libuv Constants ---
UV_RUN_DEFAULT = 0
UV_RUN_ONCE = 1
UV_RUN_NOWAIT = 2

# Handle types (from uv.h)
UV_TIMER = 14

# --- C Function Signatures ---
_uv_loop_size = _lib.uv_loop_size([], ffi.c_uint64)
_uv_handle_size = _lib.uv_handle_size([ffi.c_int32], ffi.c_uint64)

_uv_loop_init = _lib.uv_loop_init([ffi.c_void_p], ffi.c_int32)
_uv_loop_close = _lib.uv_loop_close([ffi.c_void_p], ffi.c_int32)
_uv_run = _lib.uv_run([ffi.c_void_p, ffi.c_int32], ffi.c_int32)

_uv_timer_init = _lib.uv_timer_init([ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_uv_timer_start = _lib.uv_timer_start([ffi.c_void_p, ffi.c_void_p, ffi.c_uint64, ffi.c_uint64], ffi.c_int32)
_uv_timer_stop = _lib.uv_timer_stop([ffi.c_void_p], ffi.c_int32)

_uv_close = _lib.uv_close([ffi.c_void_p, ffi.c_void_p], None)

class Loop:
    """The libuv Event Loop."""
    def __init__(self):
        size = _uv_loop_size()
        self.ptr = ffi.malloc(size)
        _uv_loop_init(self.ptr)
        self.callbacks =[]

    def run(self, mode=UV_RUN_DEFAULT):
        """Start the event loop."""
        _uv_run(self.ptr, mode)

    def close(self):
        """Close the loop and free memory."""
        if self.ptr:
            _uv_loop_close(self.ptr)
            ffi.free(self.ptr)
            self.ptr = None
            for cb in self.callbacks:
                ffi.free_callback(cb)
            self.callbacks =[]

class Timer:
    """An asynchronous timer using libuv."""
    def __init__(self, loop):
        self.loop = loop
        size = _uv_handle_size(UV_TIMER)
        self.ptr = ffi.malloc(size)
        _uv_timer_init(self.loop.ptr, self.ptr)

    def start(self, callback, timeout_ms, repeat_ms=0):
        """
        Start the timer. 
        callback: Python function to call
        timeout_ms: milliseconds until first execution
        repeat_ms: milliseconds between subsequent executions (0 for one-shot)
        """
        def c_timer_cb(timer_ptr):
            callback()

        # Create C callback
        cb = ffi.callback(c_timer_cb, None, [ffi.c_void_p])
        self.loop.callbacks.append(cb)

        _uv_timer_start(self.ptr, cb, timeout_ms, repeat_ms)

    def stop(self):
        """Stop the timer."""
        _uv_timer_stop(self.ptr)

    def close(self):
        """Close the handle and free memory."""
        if self.ptr:
            _uv_close(self.ptr, None)
            ffi.free(self.ptr)
            self.ptr = None
