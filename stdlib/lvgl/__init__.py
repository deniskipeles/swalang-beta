# stdlib/lvgl/__init__.py

import ffi
import sys

def _load_library():
    platform = sys.platform
    candidates = []
    if platform == 'linux':
        candidates = ["bin/x86_64-linux/lvgl/liblvgl.so", "liblvgl.so.9", "liblvgl.so"]
    elif platform == 'windows':
        candidates = ["bin/x86_64-windows-gnu/lvgl/liblvgl.dll", "liblvgl.dll"]

    for name in candidates:
        try:
            return ffi.CDLL(name)
        except ffi.FFIError:
            pass
    raise ffi.FFIError("Could not load LVGL shared library")

_lib = _load_library()

# --- FFI Function Definitions ---
_lv_init = _lib.lv_init([], None)
_lv_is_initialized = _lib.lv_is_initialized([], ffi.c_bool)
_lv_deinit = _lib.lv_deinit([], None)

_lv_tick_inc = _lib.lv_tick_inc([ffi.c_uint32], None)
_lv_timer_handler = _lib.lv_timer_handler([], ffi.c_uint32)

_lv_display_create = _lib.lv_display_create([ffi.c_int32, ffi.c_int32], ffi.c_void_p)
_lv_sdl_window_create = _lib.lv_sdl_window_create([ffi.c_int32, ffi.c_int32], ffi.c_void_p)
_lv_sdl_mouse_create = _lib.lv_sdl_mouse_create([], ffi.c_void_p)
_lv_sdl_keyboard_create = _lib.lv_sdl_keyboard_create([], ffi.c_void_p)

_lv_screen_active = _lib.lv_screen_active([], ffi.c_void_p)

_lv_obj_create = _lib.lv_obj_create([ffi.c_void_p], ffi.c_void_p)
_lv_obj_set_size = _lib.lv_obj_set_size([ffi.c_void_p, ffi.c_int32, ffi.c_int32], None)
_lv_obj_set_pos = _lib.lv_obj_set_pos([ffi.c_void_p, ffi.c_int32, ffi.c_int32], None)
_lv_obj_align = _lib.lv_obj_align([ffi.c_void_p, ffi.c_int32, ffi.c_int32, ffi.c_int32], None)

_lv_button_create = _lib.lv_button_create([ffi.c_void_p], ffi.c_void_p)
_lv_label_create = _lib.lv_label_create([ffi.c_void_p], ffi.c_void_p)
_lv_label_set_text = _lib.lv_label_set_text([ffi.c_void_p, ffi.c_char_p], None)

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

# --- Classes ---

def init():
    _lv_init()

def is_initialized():
    return _lv_is_initialized()

def deinit():
    _lv_deinit()

def tick_inc(ms):
    _lv_tick_inc(ms)

def timer_handler():
    return _lv_timer_handler()

class Display:
    def __init__(self, w, h, use_sdl=True):
        if use_sdl:
            self.ptr = _lv_sdl_window_create(w, h)
        else:
            self.ptr = _lv_display_create(w, h)

class Object:
    def __init__(self, parent=None):
        p_ptr = parent.ptr if parent else _lv_screen_active()
        self.ptr = _lv_obj_create(p_ptr)

    def set_size(self, w, h):
        _lv_obj_set_size(self.ptr, w, h)

    def set_pos(self, x, y):
        _lv_obj_set_pos(self.ptr, x, y)

    def align(self, align, x_ofs=0, y_ofs=0):
        _lv_obj_align(self.ptr, align, x_ofs, y_ofs)

class Button(Object):
    def __init__(self, parent=None):
        p_ptr = parent.ptr if parent else _lv_screen_active()
        self.ptr = _lv_button_create(p_ptr)

class Label(Object):
    def __init__(self, parent=None):
        p_ptr = parent.ptr if parent else _lv_screen_active()
        self.ptr = _lv_label_create(p_ptr)

    def set_text(self, text):
        _lv_label_set_text(self.ptr, text)

def screen_active():
    ptr = _lv_screen_active()
    obj = Object.__new__(Object)
    obj.ptr = ptr
    return obj

def sdl_init_all():
    _lv_sdl_mouse_create()
    _lv_sdl_keyboard_create()
