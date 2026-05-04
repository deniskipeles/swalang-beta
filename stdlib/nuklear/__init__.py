import ffi
import sys
import sdl2

def _load_library():
    platform = sys.platform
    candidates = []
    if platform == 'linux':
        candidates =["bin/x86_64-linux/nuklear/libnuklear.so", "libnuklear.so"]
    elif platform == 'windows':
        candidates =["bin/x86_64-windows-gnu/nuklear/nuklear.dll", "nuklear.dll"]
    
    for name in candidates:
        try:
            return ffi.CDLL(name)
        except ffi.FFIError:
            pass
    raise ffi.FFIError("Could not load Nuklear shared library. Ensure it is built and in the bin/ directory.")

_lib = _load_library()

# =============================================================================
# Constants
# =============================================================================

WINDOW_BORDER            = 1
WINDOW_MOVABLE           = 2
WINDOW_SCALABLE          = 4
WINDOW_CLOSABLE          = 8
WINDOW_MINIMIZABLE       = 16
WINDOW_NO_SCROLLBAR      = 32
WINDOW_TITLE             = 64
WINDOW_SCROLL_AUTO_HIDE  = 128
WINDOW_BACKGROUND        = 256
WINDOW_SCALE_LEFT        = 512
WINDOW_NO_INPUT          = 1024

TEXT_LEFT     = 0x11
TEXT_CENTERED = 0x12
TEXT_RIGHT    = 0x14

BUTTON_LEFT   = 0
BUTTON_MIDDLE = 1
BUTTON_RIGHT  = 2

# =============================================================================
# Structs
# =============================================================================

nk_rect = ffi.create_struct_type("nk_rect", [
    ['x', ffi.c_float],
    ['y', ffi.c_float],
    ['w', ffi.c_float],
    ['h', ffi.c_float]
])

nk_color = ffi.create_struct_type("nk_color", [
    ['r', ffi.c_uint8],
    ['g', ffi.c_uint8],
    ['b', ffi.c_uint8],
    ['a', ffi.c_uint8]
])

# =============================================================================
# C Signatures
# =============================================================================

_nk_init_default = _lib.nk_init_default([ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_nk_free = _lib.nk_free([ffi.c_void_p], None)

_nk_input_begin = _lib.nk_input_begin([ffi.c_void_p], None)
_nk_input_motion = _lib.nk_input_motion([ffi.c_void_p, ffi.c_int32, ffi.c_int32], None)
_nk_input_button = _lib.nk_input_button([ffi.c_void_p, ffi.c_int32, ffi.c_int32, ffi.c_int32, ffi.c_int32], None)
_nk_input_end = _lib.nk_input_end([ffi.c_void_p], None)

_nk_begin = _lib.nk_begin([ffi.c_void_p, ffi.c_char_p, nk_rect, ffi.c_uint32], ffi.c_bool)
_nk_end = _lib.nk_end([ffi.c_void_p], None)

_nk_layout_row_dynamic = _lib.nk_layout_row_dynamic([ffi.c_void_p, ffi.c_float, ffi.c_int32], None)
_nk_label = _lib.nk_label([ffi.c_void_p, ffi.c_char_p, ffi.c_uint32], None)
_nk_button_label = _lib.nk_button_label([ffi.c_void_p, ffi.c_char_p], ffi.c_bool)
_nk_clear = _lib.nk_clear([ffi.c_void_p], None)


# =============================================================================
# High-Level API
# =============================================================================

class Context:
    def __init__(self):
        # 1. Allocate 64KB for the Nuklear context (it's usually ~10KB, but safe is better)
        self.ptr = ffi.malloc(65536)
        
        # 2. We must provide a dummy font to Nuklear so it doesn't crash when calculating text width.
        # struct nk_user_font layout (x86_64):
        # 0: userdata (8 bytes)
        # 8: height (4 bytes)
        # 12: padding (4 bytes)
        # 16: width_func_ptr (8 bytes)
        self.font_ptr = ffi.malloc(24)
        
        # Define the Swalang callback that C will call to measure text!
        def _dummy_text_width(handle_ptr, h, text_ptr, length):
            return float(length * 8) # Assume 8 pixels per character
            
        self._width_cb = ffi.callback(_dummy_text_width, ffi.c_float,[ffi.c_void_p, ffi.c_float, ffi.c_char_p, ffi.c_int32])
        
        # Populate the font struct memory manually to bypass union padding issues
        ffi.write_memory_with_offset(self.font_ptr, 0, ffi.c_void_p, None)
        ffi.write_memory_with_offset(self.font_ptr, 8, ffi.c_float, 14.0) # Font height
        ffi.write_memory_with_offset(self.font_ptr, 16, ffi.c_void_p, self._width_cb)
        
        # 3. Initialize Nuklear
        _nk_init_default(self.ptr, self.font_ptr)

    def input_begin(self):
        _nk_input_begin(self.ptr)

    def input_end(self):
        _nk_input_end(self.ptr)

    def handle_event(self, ev):
        """Pass an SDL2 Event object here to feed input to Nuklear"""
        if ev.type == sdl2.MOUSEMOTION:
            _nk_input_motion(self.ptr, ev.motion_x, ev.motion_y)
            
        elif ev.type == sdl2.MOUSEBUTTONDOWN or ev.type == sdl2.MOUSEBUTTONUP:
            is_down = 1 if ev.type == sdl2.MOUSEBUTTONDOWN else 0
            btn = BUTTON_LEFT
            if ev.button == 1: btn = BUTTON_LEFT
            elif ev.button == 2: btn = BUTTON_MIDDLE
            elif ev.button == 3: btn = BUTTON_RIGHT
            _nk_input_button(self.ptr, btn, ev.button_x, ev.button_y, is_down)

    def begin(self, title, x, y, w, h, flags):
        """Begins a new window"""
        rect_dict = {"x": float(x), "y": float(y), "w": float(w), "h": float(h)}
        return _nk_begin(self.ptr, title.encode('utf-8'), rect_dict, flags)

    def end(self):
        """Ends the window"""
        _nk_end(self.ptr)

    def layout_row_dynamic(self, height, cols):
        _nk_layout_row_dynamic(self.ptr, float(height), int(cols))

    def label(self, text, alignment=TEXT_LEFT):
        _nk_label(self.ptr, text.encode('utf-8'), alignment)

    def button_label(self, text):
        """Returns True if clicked"""
        return _nk_button_label(self.ptr, text.encode('utf-8'))

    def clear(self):
        """Must be called at the end of the frame"""
        _nk_clear(self.ptr)

    def free(self):
        """Cleans up memory"""
        _nk_free(self.ptr)
        ffi.free(self.ptr)
        ffi.free(self.font_ptr)
        ffi.free_callback(self._width_cb)
        self.ptr = None