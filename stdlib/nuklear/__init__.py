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
    raise ffi.FFIError("Could not load Nuklear shared library.")

_lib = _load_library()

# =============================================================================
# Constants
# =============================================================================

WINDOW_BORDER = 1
WINDOW_MOVABLE = 2
WINDOW_TITLE = 64
TEXT_LEFT = 0x11
TEXT_CENTERED = 0x12

BUTTON_LEFT   = 0
BUTTON_MIDDLE = 1
BUTTON_RIGHT  = 2

# =============================================================================
# Structs
# =============================================================================

nk_rect = ffi.create_struct_type("nk_rect", [['x', ffi.c_float], ['y', ffi.c_float], ['w', ffi.c_float],['h', ffi.c_float]])

# =============================================================================
# C Signatures
# =============================================================================

_nk_init_fixed = _lib.nk_init_fixed([ffi.c_void_p, ffi.c_void_p, ffi.c_uint64, ffi.c_void_p], ffi.c_bool)
_nk_begin = _lib.nk_begin([ffi.c_void_p, ffi.c_char_p, nk_rect, ffi.c_uint32], ffi.c_bool)
_nk_end = _lib.nk_end([ffi.c_void_p], None)
_nk_layout_row_dynamic = _lib.nk_layout_row_dynamic([ffi.c_void_p, ffi.c_float, ffi.c_int32], None)
_nk_label = _lib.nk_label([ffi.c_void_p, ffi.c_char_p, ffi.c_uint32], None)
_nk_button_label = _lib.nk_button_label([ffi.c_void_p, ffi.c_char_p], ffi.c_bool)
_nk_clear = _lib.nk_clear([ffi.c_void_p], None)

_nk_input_begin = _lib.nk_input_begin([ffi.c_void_p], None)
_nk_input_motion = _lib.nk_input_motion([ffi.c_void_p, ffi.c_int32, ffi.c_int32], None)
_nk_input_button = _lib.nk_input_button([ffi.c_void_p, ffi.c_int32, ffi.c_int32, ffi.c_int32, ffi.c_int32], None)
_nk_input_end = _lib.nk_input_end([ffi.c_void_p], None)

# =============================================================================
# High-Level API
# =============================================================================

class Context:
    def __init__(self):
        # Allocation for the context and a working buffer
        self.ptr = ffi.malloc(4096) 
        self.memory = ffi.malloc(32768) 
        
        # Setup dummy font to prevent crash
        self.font_ptr = ffi.malloc(32)
        def _dummy_text_width(handle_ptr, h, text_ptr, length):
            return float(length * 8)
        self._width_cb = ffi.callback(_dummy_text_width, ffi.c_float,[ffi.c_void_p, ffi.c_float, ffi.c_char_p, ffi.c_int32])
        
        ffi.write_memory_with_offset(self.font_ptr, 16, ffi.c_void_p, self._width_cb)
        
        # Initialize
        _nk_init_fixed(self.ptr, self.memory, 32768, self.font_ptr)

    def input_begin(self):
        _nk_input_begin(self.ptr)

    def input_end(self):
        _nk_input_end(self.ptr)

    def handle_event(self, ev):
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
        rect = {"x": float(x), "y": float(y), "w": float(w), "h": float(h)}
        return _nk_begin(self.ptr, title.encode('utf-8'), rect, flags)

    def end(self): _nk_end(self.ptr)
    def layout_row_dynamic(self, h, cols): _nk_layout_row_dynamic(self.ptr, float(h), cols)
    def label(self, text, align=TEXT_LEFT): _nk_label(self.ptr, text.encode('utf-8'), align)
    def button_label(self, text): return _nk_button_label(self.ptr, text.encode('utf-8'))
    def clear(self): _nk_clear(self.ptr)

    def free(self):
        ffi.free(self.ptr)
        ffi.free(self.memory)
        ffi.free(self.font_ptr)
        ffi.free_callback(self._width_cb)