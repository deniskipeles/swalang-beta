
import ffi
import sys

def _load_library():
    platform = sys.platform
    candidates = []
    if platform == 'linux':
        candidates = ["bin/x86_64-linux/lvgl/liblvgl.so", "liblvgl.so"]
    elif platform == 'windows':
        candidates = ["bin/x86_64-windows-gnu/lvgl/lvgl.dll", "lvgl.dll"]
    
    for name in candidates:
        try:
            return ffi.CDLL(name)
        except ffi.FFIError:
            pass
    raise ffi.FFIError("Could not load LVGL shared library.")

_lib = _load_library()

# =============================================================================
# Constants (LVGL v9)
# =============================================================================

# Alignments
LV_ALIGN_DEFAULT = 0
LV_ALIGN_TOP_LEFT = 1
LV_ALIGN_TOP_MID = 2
LV_ALIGN_TOP_RIGHT = 3
LV_ALIGN_BOTTOM_LEFT = 4
LV_ALIGN_BOTTOM_MID = 5
LV_ALIGN_BOTTOM_RIGHT = 6
LV_ALIGN_LEFT_MID = 7
LV_ALIGN_RIGHT_MID = 8
LV_ALIGN_CENTER = 9

# Input Device Types
LV_INDEV_TYPE_NONE = 0
LV_INDEV_TYPE_POINTER = 1
LV_INDEV_TYPE_KEYPAD = 2
LV_INDEV_TYPE_BUTTON = 3
LV_INDEV_TYPE_ENCODER = 4

# Input Device States
LV_INDEV_STATE_RELEASED = 0
LV_INDEV_STATE_PRESSED = 1

# Event Codes (Partial list)
LV_EVENT_PRESSED = 1
LV_EVENT_CLICKED = 14
LV_EVENT_VALUE_CHANGED = 28

# Display Render Modes
LV_DISPLAY_RENDER_MODE_PARTIAL = 0
LV_DISPLAY_RENDER_MODE_DIRECT = 1
LV_DISPLAY_RENDER_MODE_FULL = 2

# =============================================================================
# Structs
# =============================================================================

@ffi.Struct
class Area:
    _fields_ = [
        ('x1', ffi.c_int32),
        ('y1', ffi.c_int32),
        ('x2', ffi.c_int32),
        ('y2', ffi.c_int32)
    ]

# =============================================================================
# C Signatures
# =============================================================================

_lv_init = _lib.lv_init([], None)
_lv_tick_inc = _lib.lv_tick_inc([ffi.c_uint32], None)
_lv_timer_handler = _lib.lv_timer_handler([], ffi.c_uint32)

# Display API (v9)
_lv_display_create = _lib.lv_display_create([ffi.c_int32, ffi.c_int32], ffi.c_void_p)
_lv_display_set_buffers = _lib.lv_display_set_buffers([ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_uint32, ffi.c_int32], None)
_lv_display_set_flush_cb = _lib.lv_display_set_flush_cb([ffi.c_void_p, ffi.c_void_p], None)
_lv_display_flush_ready = _lib.lv_display_flush_ready([ffi.c_void_p], None)

# Input API (v9)
_lv_indev_create = _lib.lv_indev_create([], ffi.c_void_p)
_lv_indev_set_type = _lib.lv_indev_set_type([ffi.c_void_p, ffi.c_int32], None)
_lv_indev_set_read_cb = _lib.lv_indev_set_read_cb([ffi.c_void_p, ffi.c_void_p], None)

# Widget API
_lv_screen_active = _lib.lv_screen_active([], ffi.c_void_p)

_lv_button_create = _lib.lv_button_create([ffi.c_void_p], ffi.c_void_p)
_lv_label_create = _lib.lv_label_create([ffi.c_void_p], ffi.c_void_p)
_lv_label_set_text = _lib.lv_label_set_text([ffi.c_void_p, ffi.c_char_p], None)

_lv_obj_align = _lib.lv_obj_align([ffi.c_void_p, ffi.c_int32, ffi.c_int32, ffi.c_int32], None)
_lv_obj_set_size = _lib.lv_obj_set_size([ffi.c_void_p, ffi.c_int32, ffi.c_int32], None)
_lv_obj_add_event_cb = _lib.lv_obj_add_event_cb([ffi.c_void_p, ffi.c_void_p, ffi.c_int32, ffi.c_void_p], None)

# =============================================================================
# High-Level API
# =============================================================================

def init():
    """Initialize the LVGL library."""
    _lv_init()

def tick_inc(ms):
    """Tell LVGL how much time has passed."""
    _lv_tick_inc(ms)

def timer_handler():
    """Run LVGL tasks (timers, redrawing, input reading)."""
    return _lv_timer_handler()

class Display:
    def __init__(self, w, h):
        self.ptr = _lv_display_create(w, h)
        self.w = w
        self.h = h
        self._flush_cb_keepalive = None

    def set_buffers(self, buf1_ptr, buf2_ptr, buf_size_bytes, render_mode=LV_DISPLAY_RENDER_MODE_PARTIAL):
        _lv_display_set_buffers(self.ptr, buf1_ptr, buf2_ptr, buf_size_bytes, render_mode)

    def set_flush_cb(self, py_callback):
        """
        py_callback should take: (disp_ptr, area_ptr, px_map_ptr)
        """
        cb = ffi.callback(py_callback, None, [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p])
        self._flush_cb_keepalive = cb
        _lv_display_set_flush_cb(self.ptr, cb)

    def flush_ready(self):
        """Must be called at the end of the flush callback."""
        _lv_display_flush_ready(self.ptr)


class InputDevice:
    def __init__(self, type_=LV_INDEV_TYPE_POINTER):
        self.ptr = _lv_indev_create()
        _lv_indev_set_type(self.ptr, type_)
        self._read_cb_keepalive = None

    def set_read_cb(self, py_callback):
        """
        py_callback should take: (indev_ptr, data_ptr)
        data_ptr points to lv_indev_data_t. You must write x, y, and state to it.
        """
        cb = ffi.callback(py_callback, None, [ffi.c_void_p, ffi.c_void_p])
        self._read_cb_keepalive = cb
        _lv_indev_set_read_cb(self.ptr, cb)

    @staticmethod
    def write_pointer_data(data_ptr, x, y, state):
        """Helper to write x, y, and state into lv_indev_data_t in memory."""
        # Offset 0: int32 x
        # Offset 4: int32 y
        # Offset 28: uint32 state (in LVGL v9 64-bit builds)
        ffi.write_memory_with_offset(data_ptr, 0, ffi.c_int32, x)
        ffi.write_memory_with_offset(data_ptr, 4, ffi.c_int32, y)
        ffi.write_memory_with_offset(data_ptr, 28, ffi.c_uint32, state)


class Screen:
    @staticmethod
    def active():
        return _lv_screen_active()

class Button:
    def __init__(self, parent_ptr):
        self.ptr = _lv_button_create(parent_ptr)
        self._event_cbs = []

    def align(self, alignment, x_ofs=0, y_ofs=0):
        _lv_obj_align(self.ptr, alignment, x_ofs, y_ofs)

    def set_size(self, w, h):
        _lv_obj_set_size(self.ptr, w, h)

    def add_event_cb(self, py_callback, event_filter):
        """py_callback should take: (event_ptr)"""
        cb = ffi.callback(py_callback, None, [ffi.c_void_p])
        self._event_cbs.append(cb)
        _lv_obj_add_event_cb(self.ptr, cb, event_filter, None)

class Label:
    def __init__(self, parent_ptr):
        self.ptr = _lv_label_create(parent_ptr)

    def set_text(self, text):
        _lv_label_set_text(self.ptr, text.encode('utf-8'))

    def align(self, alignment, x_ofs=0, y_ofs=0):
        _lv_obj_align(self.ptr, alignment, x_ofs, y_ofs)