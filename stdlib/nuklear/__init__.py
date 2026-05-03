# stdlib/nuklear/__init__.py

import ffi
import sys

def _load_library():
    platform = sys.platform
    candidates = []
    if platform == 'linux':
        candidates = ["bin/x86_64-linux/nuklear/libnuklear.so", "libnuklear.so"]
    elif platform == 'windows':
        candidates = ["bin/x86_64-windows-gnu/nuklear/nuklear.dll", "nuklear.dll"]

    for name in candidates:
        try:
            return ffi.CDLL(name)
        except ffi.FFIError:
            pass
    raise ffi.FFIError("Could not load Nuklear shared library")

_lib = _load_library()

# --- Nuklear Constants ---
NK_WINDOW_BORDER = 1 << 0
NK_WINDOW_MOVABLE = 1 << 1
NK_WINDOW_SCALABLE = 1 << 2
NK_WINDOW_CLOSABLE = 1 << 3
NK_WINDOW_MINIMIZABLE = 1 << 4
NK_WINDOW_NO_SCROLLBAR = 1 << 5
NK_WINDOW_TITLE = 1 << 6
NK_WINDOW_SCROLL_AUTO_HIDE = 1 << 7
NK_WINDOW_BACKGROUND = 1 << 8
NK_WINDOW_SCALE_LEFT = 1 << 9
NK_WINDOW_NO_INPUT = 1 << 10

NK_TEXT_ALIGN_LEFT = 0x01
NK_TEXT_ALIGN_CENTERED = 0x02
NK_TEXT_ALIGN_RIGHT = 0x04
NK_TEXT_ALIGN_TOP = 0x08
NK_TEXT_ALIGN_MIDDLE = 0x10
NK_TEXT_ALIGN_BOTTOM = 0x20

NK_TEXT_LEFT = 0x11 # NK_TEXT_ALIGN_MIDDLE | NK_TEXT_ALIGN_LEFT
NK_TEXT_CENTERED = 0x12 # NK_TEXT_ALIGN_MIDDLE | NK_TEXT_ALIGN_CENTERED
NK_TEXT_RIGHT = 0x14 # NK_TEXT_ALIGN_MIDDLE | NK_TEXT_ALIGN_RIGHT

# --- FFI Function Definitions ---
_nk_begin = _lib.nk_begin([ffi.c_void_p, ffi.c_char_p, ffi.c_float, ffi.c_float, ffi.c_float, ffi.c_float, ffi.c_uint32], ffi.c_bool)
_nk_end = _lib.nk_end([ffi.c_void_p], None)

_nk_layout_row_dynamic = _lib.nk_layout_row_dynamic([ffi.c_void_p, ffi.c_float, ffi.c_int32], None)
_nk_label = _lib.nk_label([ffi.c_void_p, ffi.c_char_p, ffi.c_uint32], None)
_nk_button_label = _lib.nk_button_label([ffi.c_void_p, ffi.c_char_p], ffi.c_bool)

# --- Classes ---

class Context:
    def __init__(self, ptr):
        self.ptr = ptr

    def begin(self, title, x, y, w, h, flags):
        return _nk_begin(self.ptr, title, x, y, w, h, flags)

    def end(self):
        _nk_end(self.ptr)

    def layout_row_dynamic(self, height, cols):
        _nk_layout_row_dynamic(self.ptr, height, cols)

    def label(self, text, align=NK_TEXT_LEFT):
        _nk_label(self.ptr, text, align)

    def button_label(self, title):
        return _nk_button_label(self.ptr, title)
