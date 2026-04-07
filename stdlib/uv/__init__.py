# stdlib/uv/__init__.py

"""
A Pylearn wrapper for the libuv asynchronous I/O library.

This module provides a Pythonic, object-oriented interface for building
event-driven applications using libuv's event loop.
"""

import ffi
import sys

# ==============================================================================
#  Platform-Aware Library Loading
# ==============================================================================

def _load_library_with_fallbacks(base_name):
    """Tries to load libuv using common platform-specific names."""
    platform = sys.platform
    
    # Common names for libuv shared library
    candidates = []
    if platform == 'windows':
        candidates = ['libuv.dll', 'uv.dll']
    elif platform == 'darwin': # macOS
        candidates = ['libuv.1.dylib', 'libuv.dylib']
    else: # Linux and other Unix-like
        candidates = ['libuv.so.1', 'libuv.so']

    last_error = None
    for name in candidates:
        try:
            # Use the ffi's robust loader
            return ffi.CDLL(name)
        except ffi.FFIError as e:
            last_error = e
    
    # If all candidates failed, raise the last error
    raise last_error

# =======================================================
# helper function
# =======================================================
def _get_handle_address(ptr):
    """Converts the FFI Pointer's Address object to a primitive integer."""
    if ptr is None or ptr.Address is None:
        return 0
    # str(ptr.Address) calls the object's Inspect(), which returns the number as a string.
    # int() converts that string to a primitive integer.
    return int(str(ptr.Address))

_lib = None
UV_AVAILABLE = False
try:
    _lib = _load_library_with_fallbacks("uv")
    UV_AVAILABLE = True
except ffi.FFIError as e:
    print(format_str("Warning: Failed to load libuv: {e}"))
    print("Async I/O functionality in the 'uv' module will not be available.")

# ==============================================================================
#  Exception and Constants
# ==============================================================================

class UVError(Exception):
    """Raised for libuv-specific errors."""
    def __init__(self, code, func_name):
        self.code = code
        # Get the error message directly from libuv
        err_name_ptr = _lib.uv_err_name([ffi.c_int32], ffi.c_char_p)(code)
        err_msg_ptr = _lib.uv_strerror([ffi.c_int32], ffi.c_char_p)(code)
        
        self.err_name = ffi.string_at(err_name_ptr)
        self.message = ffi.string_at(err_msg_ptr)
        
        super().__init__(format_str("[{func_name}] {self.err_name}: {self.message}"))

class UVRunMode:
    DEFAULT = 0
    ONCE = 1
    NOWAIT = 2

class UVHandleType:
    TIMER = 14 # From uv.h, value for UV_TIMER

# ==============================================================================
#  Define C Function Signatures
# ==============================================================================

if UV_AVAILABLE:
    # --- Loop Functions ---
    _uv_loop_size = _lib.uv_loop_size([], ffi.c_uint64)
    _uv_loop_init = _lib.uv_loop_init([ffi.c_void_p], ffi.c_int32)
    _uv_loop_close = _lib.uv_loop_close([ffi.c_void_p], ffi.c_int32)
    _uv_run = _lib.uv_run([ffi.c_void_p, ffi.c_int32], ffi.c_int32)

    # --- Handle Functions ---
    _uv_handle_size = _lib.uv_handle_size([ffi.c_int32], ffi.c_uint64)
    _uv_close = _lib.uv_close([ffi.c_void_p, ffi.c_void_p], None) # Takes an optional close_cb

    # --- Timer Functions ---
    _uv_timer_init = _lib.uv_timer_init([ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
    _uv_timer_start = _lib.uv_timer_start([ffi.c_void_p, ffi.c_void_p, ffi.c_uint64, ffi.c_uint64], ffi.c_int32)
    _uv_timer_stop = _lib.uv_timer_stop([ffi.c_void_p], ffi.c_int32)

# Global registry to map C handle pointers back to our Python objects
_active_handles = {}

# ==============================================================================
#  Pythonic Wrapper Classes
# ==============================================================================

class Handle:
    """Base class for all libuv handle wrappers."""
    def __init__(self, ptr):
        self._ptr = ptr
        self._c_callback = None # To prevent garbage collection
        
        # <<< FIX: Use the primitive integer address as the key
        address = _get_handle_address(self._ptr)
        _active_handles[address] = self

    def close(self):
        """Request that the handle be closed."""
        if self._ptr:
            # <<< START OF FIX >>>
            # Capture the pointer's value *before* it gets nullified.
            # This value will be enclosed in the callback.
            ptr_to_free = self._ptr

            def on_close_cb(handle_ptr):
                # This callback now uses 'ptr_to_free', which holds the
                # original Pointer object, instead of 'py_handle._ptr'.
                address = _get_handle_address(handle_ptr)
                if address in _active_handles:
                    # We still need to remove the handle from the registry
                    del _active_handles[address]
                
                # Free the captured pointer
                ffi.free(ptr_to_free)
            
            c_cb = ffi.callback(on_close_cb, None, [ffi.c_void_p])
            self._c_callback = c_cb # Keep reference
            
            # Pass the original pointer to libuv
            _uv_close(self._ptr, c_cb)
            
            # Now it is safe to nullify the instance attribute because the
            # value we need has been saved in the 'ptr_to_free' variable.
            self._ptr = None
            # <<< END OF FIX >>>

class Loop:
    """Represents a libuv event loop."""
    def __init__(self):
        if not UV_AVAILABLE:
            raise UVError(-1, "Loop()")
        
        loop_size = _uv_loop_size()
        self._ptr = ffi.malloc(loop_size)
        
        res = _uv_loop_init(self._ptr)
        if res != 0:
            ffi.free(self._ptr)
            raise UVError(res, "uv_loop_init")

    def run(self, mode=UVRunMode.DEFAULT):
        """Starts the event loop. This call blocks until the loop is stopped."""
        res = _uv_run(self._ptr, mode)
        if res != 0:
            # This can happen if the loop is closed while running, which is fine.
            # A real app would check the specific error code.
            pass

    def close(self):
        """Closes the event loop and frees its resources."""
        if self._ptr:
            res = _uv_loop_close(self._ptr)
            ffi.free(self._ptr)
            self._ptr = None
            if res != 0:
                # EBUSY means there are still active handles.
                raise UVError(res, "uv_loop_close")
    
    def __del__(self):
        # A simple cleanup attempt when the object is garbage collected.
        self.close()

class Timer(Handle):
    """Represents a timer handle."""
    def __init__(self, loop):
        if not isinstance(loop, Loop):
            raise TypeError("Timer must be initialized with a Loop object.")

        self._loop = loop
        timer_size = _uv_handle_size(UVHandleType.TIMER)
        ptr = ffi.malloc(timer_size)
        super().__init__(ptr)

        res = _uv_timer_init(loop._ptr, self._ptr)
        if res != 0:
            self.close()
            raise UVError(res, "uv_timer_init")
        
        self.pylearn_callback = None

    def start(self, callback, timeout, repeat):
        """
        Starts the timer.
        :param callback: A Pylearn function to call when the timer expires. It takes no arguments.
        :param timeout: The initial delay in milliseconds.
        :param repeat: The repeat interval in milliseconds. If 0, the timer is a one-shot.
        """
        self.pylearn_callback = callback

        def c_callback_wrapper(timer_handle_ptr):
            # <<< FIX: Use the primitive integer address for the lookup
            address = _get_handle_address(timer_handle_ptr)
            py_timer = _active_handles.get(address)
            
            if py_timer and py_timer.pylearn_callback:
                py_timer.pylearn_callback()

        c_cb = ffi.callback(c_callback_wrapper, None, [ffi.c_void_p])
        self._c_callback = c_cb

        res = _uv_timer_start(self._ptr, c_cb, timeout, repeat)
        if res != 0:
            raise UVError(res, "uv_timer_start")

    def stop(self):
        """Stops the timer."""
        res = _uv_timer_stop(self._ptr)
        if res != 0:
            raise UVError(res, "uv_timer_stop")

            