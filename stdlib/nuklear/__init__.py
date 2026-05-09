"""
Full Nuklear immediate-mode GUI wrapper for Swalang/Pylearn.

Build requirements — nuklear_impl.c must define before including nuklear.h:
    #define NK_INCLUDE_DEFAULT_ALLOCATOR
    #define NK_INCLUDE_FIXED_TYPES
    #define NK_INCLUDE_FONT_BAKING
    #define NK_INCLUDE_DEFAULT_FONT
    #define NK_INCLUDE_STANDARD_VARARGS
    #define NK_INCLUDE_VERTEX_BUFFER_OUTPUT
    #define NK_IMPLEMENTATION
    #include "nuklear.h"

Subsystems covered:
    Context · Input · Windows · Layout · Widgets · Style · Canvas ·
    Font Baking · Command Rendering · SDL2 Backend
"""

import ffi
import sys

# ==============================================================================
#  Library Loading
# ==============================================================================

def _try_load(names):
    for name in names:
        try:
            return ffi.CDLL(name)
        except ffi.FFIError:
            pass
    return None

if sys.platform == "linux":
    _lib = _try_load([
        "bin/x86_64-linux/nuklear/libnuklear.so",
        "libnuklear.so",
    ])
elif sys.platform == "windows":
    _lib = _try_load([
        "bin/x86_64-windows-gnu/nuklear/nuklear.dll",
        "nuklear.dll",
    ])
else:
    _lib = _try_load(["libnuklear.dylib"])

if _lib is None:
    raise ffi.FFIError("Could not load Nuklear shared library.")

# ==============================================================================
#  Constants
# ==============================================================================

# --- Boolean ---
nk_false = 0
nk_true  = 1

# --- Keys ---
NK_KEY_NONE              = 0
NK_KEY_SHIFT             = 1
NK_KEY_CTRL              = 2
NK_KEY_DEL               = 3
NK_KEY_ENTER             = 4
NK_KEY_TAB               = 5
NK_KEY_BACKSPACE         = 6
NK_KEY_COPY              = 7
NK_KEY_CUT               = 8
NK_KEY_PASTE             = 9
NK_KEY_UP                = 10
NK_KEY_DOWN              = 11
NK_KEY_LEFT              = 12
NK_KEY_RIGHT             = 13
NK_KEY_TEXT_INSERT_MODE  = 14
NK_KEY_TEXT_REPLACE_MODE = 15
NK_KEY_TEXT_RESET_MODE   = 16
NK_KEY_TEXT_LINE_START   = 17
NK_KEY_TEXT_LINE_END     = 18
NK_KEY_TEXT_START        = 19
NK_KEY_TEXT_END          = 20
NK_KEY_TEXT_UNDO         = 21
NK_KEY_TEXT_REDO         = 22
NK_KEY_TEXT_SELECT_ALL   = 23
NK_KEY_TEXT_WORD_LEFT    = 24
NK_KEY_TEXT_WORD_RIGHT   = 25
NK_KEY_SCROLL_START      = 26
NK_KEY_SCROLL_END        = 27
NK_KEY_SCROLL_DOWN       = 28
NK_KEY_SCROLL_UP         = 29
NK_KEY_MAX               = 30

# --- Buttons ---
NK_BUTTON_LEFT   = 0
NK_BUTTON_MIDDLE = 1
NK_BUTTON_RIGHT  = 2
NK_BUTTON_DOUBLE = 3
NK_BUTTON_MAX    = 4

# --- Window flags ---
NK_WINDOW_BORDER           = 1 << 0
NK_WINDOW_MOVABLE          = 1 << 1
NK_WINDOW_SCALABLE         = 1 << 2
NK_WINDOW_CLOSABLE         = 1 << 3
NK_WINDOW_MINIMIZABLE      = 1 << 4
NK_WINDOW_NO_SCROLLBAR     = 1 << 5
NK_WINDOW_TITLE            = 1 << 6
NK_WINDOW_SCROLL_AUTO_HIDE = 1 << 7
NK_WINDOW_BACKGROUND       = 1 << 8
NK_WINDOW_SCALE_LEFT       = 1 << 9
NK_WINDOW_NO_INPUT         = 1 << 10

# --- Text alignment ---
NK_TEXT_ALIGN_LEFT     = 0x01
NK_TEXT_ALIGN_CENTERED = 0x02
NK_TEXT_ALIGN_RIGHT    = 0x04
NK_TEXT_ALIGN_TOP      = 0x08
NK_TEXT_ALIGN_MIDDLE   = 0x10
NK_TEXT_ALIGN_BOTTOM   = 0x20
NK_TEXT_LEFT           = NK_TEXT_ALIGN_MIDDLE | NK_TEXT_ALIGN_LEFT
NK_TEXT_CENTERED       = NK_TEXT_ALIGN_MIDDLE | NK_TEXT_ALIGN_CENTERED
NK_TEXT_RIGHT          = NK_TEXT_ALIGN_MIDDLE | NK_TEXT_ALIGN_RIGHT

# --- Edit flags ---
NK_EDIT_DEFAULT                 = 0
NK_EDIT_READ_ONLY               = 1 << 0
NK_EDIT_AUTO_SELECT             = 1 << 1
NK_EDIT_SIG_ENTER               = 1 << 2
NK_EDIT_ALLOW_TAB               = 1 << 3
NK_EDIT_NO_CURSOR               = 1 << 4
NK_EDIT_SELECTABLE              = 1 << 5
NK_EDIT_CLIPBOARD               = 1 << 6
NK_EDIT_CTRL_ENTER_NEWLINE      = 1 << 7
NK_EDIT_NO_HORIZONTAL_SCROLL    = 1 << 8
NK_EDIT_ALWAYS_INSERT_MODE      = 1 << 9
NK_EDIT_MULTILINE               = 1 << 10
NK_EDIT_GOTO_END_ON_ACTIVATE    = 1 << 11
NK_EDIT_SIMPLE = NK_EDIT_ALWAYS_INSERT_MODE
NK_EDIT_FIELD  = NK_EDIT_SIMPLE | NK_EDIT_SELECTABLE | NK_EDIT_CLIPBOARD
NK_EDIT_BOX    = (NK_EDIT_ALWAYS_INSERT_MODE | NK_EDIT_SELECTABLE |
                  NK_EDIT_MULTILINE | NK_EDIT_ALLOW_TAB | NK_EDIT_CLIPBOARD)
NK_EDIT_EDITOR = NK_EDIT_SELECTABLE | NK_EDIT_MULTILINE | NK_EDIT_ALLOW_TAB | NK_EDIT_CLIPBOARD

# --- Edit events (return from nk_edit_string) ---
NK_EDIT_ACTIVE      = 1 << 0
NK_EDIT_INACTIVE    = 1 << 1
NK_EDIT_ACTIVATED   = 1 << 2
NK_EDIT_DEACTIVATED = 1 << 3
NK_EDIT_COMMITED    = 1 << 4

# --- Symbol types ---
NK_SYMBOL_NONE           = 0
NK_SYMBOL_X              = 1
NK_SYMBOL_UNDERSCORE     = 2
NK_SYMBOL_CIRCLE_SOLID   = 3
NK_SYMBOL_CIRCLE_OUTLINE = 4
NK_SYMBOL_RECT_SOLID     = 5
NK_SYMBOL_RECT_OUTLINE   = 6
NK_SYMBOL_TRIANGLE_UP    = 7
NK_SYMBOL_TRIANGLE_DOWN  = 8
NK_SYMBOL_TRIANGLE_LEFT  = 9
NK_SYMBOL_TRIANGLE_RIGHT = 10
NK_SYMBOL_PLUS           = 11
NK_SYMBOL_MINUS          = 12
NK_SYMBOL_MAX            = 13

# --- Tree types ---
NK_TREE_NODE = 0
NK_TREE_TAB  = 1

# --- Collapse states ---
NK_MINIMIZED = 0
NK_MAXIMIZED = 1

# --- Show states ---
NK_HIDDEN = 0
NK_SHOWN  = 1

# --- Popup types ---
NK_POPUP_STATIC  = 0
NK_POPUP_DYNAMIC = 1

# --- Layout formats ---
NK_DYNAMIC = 0
NK_STATIC  = 1

# --- Chart types ---
NK_CHART_LINES  = 0
NK_CHART_COLUMN = 1
NK_CHART_MAX    = 2

# --- Chart events ---
NK_CHART_HOVERING = 0x01
NK_CHART_CLICKED  = 0x02

# --- Color format ---
NK_RGB  = 0
NK_RGBA = 1

# --- Button behavior ---
NK_BUTTON_DEFAULT  = 0
NK_BUTTON_REPEATER = 1

# --- Anti-aliasing ---
NK_ANTI_ALIASING_OFF = 0
NK_ANTI_ALIASING_ON  = 1

# --- Style colors (nk_style_colors enum) ---
NK_COLOR_TEXT                    = 0
NK_COLOR_WINDOW                  = 1
NK_COLOR_HEADER                  = 2
NK_COLOR_BORDER                  = 3
NK_COLOR_BUTTON                  = 4
NK_COLOR_BUTTON_HOVER            = 5
NK_COLOR_BUTTON_ACTIVE           = 6
NK_COLOR_TOGGLE                  = 7
NK_COLOR_TOGGLE_HOVER            = 8
NK_COLOR_TOGGLE_CURSOR           = 9
NK_COLOR_SELECT                  = 10
NK_COLOR_SELECT_ACTIVE           = 11
NK_COLOR_SLIDER                  = 12
NK_COLOR_SLIDER_CURSOR           = 13
NK_COLOR_SLIDER_CURSOR_HOVER     = 14
NK_COLOR_SLIDER_CURSOR_ACTIVE    = 15
NK_COLOR_PROPERTY                = 16
NK_COLOR_EDIT                    = 17
NK_COLOR_EDIT_CURSOR             = 18
NK_COLOR_COMBO                   = 19
NK_COLOR_CHART                   = 20
NK_COLOR_CHART_COLOR             = 21
NK_COLOR_CHART_COLOR_HIGHLIGHT   = 22
NK_COLOR_SCROLLBAR               = 23
NK_COLOR_SCROLLBAR_CURSOR        = 24
NK_COLOR_SCROLLBAR_CURSOR_HOVER  = 25
NK_COLOR_SCROLLBAR_CURSOR_ACTIVE = 26
NK_COLOR_TAB_HEADER              = 27
NK_COLOR_COUNT                   = 28

# --- Command types (for renderer) ---
NK_COMMAND_NOP               = 0
NK_COMMAND_SCISSOR           = 1
NK_COMMAND_LINE              = 2
NK_COMMAND_CURVE             = 3
NK_COMMAND_RECT              = 4
NK_COMMAND_RECT_FILLED       = 5
NK_COMMAND_RECT_MULTI_COLOR  = 6
NK_COMMAND_CIRCLE            = 7
NK_COMMAND_CIRCLE_FILLED     = 8
NK_COMMAND_ARC               = 9
NK_COMMAND_ARC_FILLED        = 10
NK_COMMAND_TRIANGLE          = 11
NK_COMMAND_TRIANGLE_FILLED   = 12
NK_COMMAND_POLYGON           = 13
NK_COMMAND_POLYGON_FILLED    = 14
NK_COMMAND_POLYLINE          = 15
NK_COMMAND_TEXT              = 16
NK_COMMAND_IMAGE             = 17
NK_COMMAND_CUSTOM            = 18

# --- Memory sizes ---
_CTX_SIZE   = 256 * 1024      # 256 KB for nk_context struct
_POOL_SIZE  = 4 * 1024 * 1024  # 4 MB for nk_init_fixed memory pool
_FONT_SIZE  = 8 * 1024         # 8 KB for nk_user_font
_ATLAS_SIZE = 64 * 1024        # 64 KB for font atlas struct

# --- nk_command header layout (64-bit, no NK_INCLUDE_COMMAND_USERDATA) ---
# type (int32) @ 0, padding @ 4, next (uint64) @ 8  → 16 bytes total
_CMD_HDR = 16

# ==============================================================================
#  Struct helpers
# ==============================================================================

def _r8(buf, off):
    return ffi.read_memory_with_offset(buf, off, ffi.c_uint8)

def _r16(buf, off):
    return ffi.read_memory_with_offset(buf, off, ffi.c_uint16)

def _rs16(buf, off):
    return ffi.read_memory_with_offset(buf, off, ffi.c_short)

def _r32(buf, off):
    return ffi.read_memory_with_offset(buf, off, ffi.c_int32)

def _r64(buf, off):
    return ffi.read_memory_with_offset(buf, off, ffi.c_uint64)

def _rf(buf, off):
    return ffi.read_memory_with_offset(buf, off, ffi.c_float)

def _w8(buf, off, v):   ffi.write_memory_with_offset(buf, off, ffi.c_uint8,  v)
def _w16(buf, off, v):  ffi.write_memory_with_offset(buf, off, ffi.c_uint16, v)
def _w32(buf, off, v):  ffi.write_memory_with_offset(buf, off, ffi.c_int32,  v)
def _w64(buf, off, v):  ffi.write_memory_with_offset(buf, off, ffi.c_uint64, v)
def _wf(buf, off, v):   ffi.write_memory_with_offset(buf, off, ffi.c_float,  v)

def _enc(s):
    return s.encode("utf-8") if isinstance(s, str) else s

def _read_color(buf, off):
    """Read nk_color {r,g,b,a} from offset → (r,g,b,a) tuple."""
    return (_r8(buf, off), _r8(buf, off+1), _r8(buf, off+2), _r8(buf, off+3))

def make_color_buf(r, g, b, a=255):
    """Allocate a 4-byte nk_color buffer. Caller must ffi.free() it."""
    buf = ffi.malloc(4)
    _w8(buf, 0, r); _w8(buf, 1, g); _w8(buf, 2, b); _w8(buf, 3, a)
    return buf

def make_rect_buf(x, y, w, h):
    """Allocate a 16-byte nk_rect (float) buffer. Caller must ffi.free() it."""
    buf = ffi.malloc(16)
    _wf(buf,  0, x); _wf(buf,  4, y)
    _wf(buf,  8, w); _wf(buf, 12, h)
    return buf

def make_vec2_buf(x, y):
    """Allocate an 8-byte nk_vec2 buffer. Caller must ffi.free() it."""
    buf = ffi.malloc(8)
    _wf(buf, 0, x); _wf(buf, 4, y)
    return buf

def make_int_buf(v=0):
    buf = ffi.malloc(4)
    _w32(buf, 0, v)
    return buf

def make_float_buf(v=0.0):
    buf = ffi.malloc(4)
    _wf(buf, 0, v)
    return buf

def read_int_buf(buf):
    b = ffi.buffer_to_bytes(buf, 4)
    v = b[0] | (b[1] << 8) | (b[2] << 16) | (b[3] << 24)
    if v >= 0x80000000:
        v = v - 0x100000000
    return v

def read_float_buf(buf):
    return _rf(buf, 0)

# ==============================================================================
#  C Function Bindings
# ==============================================================================

# --- Initialization ---
_nk_init_fixed     = _lib.nk_init_fixed([ffi.c_void_p, ffi.c_void_p, ffi.c_uint64, ffi.c_void_p], ffi.c_int32)
# _nk_init_default   = _lib.nk_init_default([ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_nk_clear          = _lib.nk_clear([ffi.c_void_p], None)
_nk_free           = _lib.nk_free([ffi.c_void_p], None)

# --- Input ---
_nk_input_begin    = _lib.nk_input_begin(   [ffi.c_void_p], None)
_nk_input_end      = _lib.nk_input_end(     [ffi.c_void_p], None)
_nk_input_motion   = _lib.nk_input_motion(  [ffi.c_void_p, ffi.c_int32, ffi.c_int32], None)
_nk_input_key      = _lib.nk_input_key(     [ffi.c_void_p, ffi.c_int32, ffi.c_int32], None)
_nk_input_button   = _lib.nk_input_button([ffi.c_void_p, ffi.c_int32, ffi.c_int32, ffi.c_int32, ffi.c_int32], None)
_nk_input_scroll   = _lib.nk_input_scroll(  [ffi.c_void_p, ffi.c_void_p], None)
_nk_input_char     = _lib.nk_input_char(    [ffi.c_void_p, ffi.c_char], None)
_nk_input_unicode  = _lib.nk_input_unicode( [ffi.c_void_p, ffi.c_uint32], None)
_nk_input_is_mouse_hovering_rect = _lib.nk_input_is_mouse_hovering_rect([ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_nk_input_is_mouse_prev_hovering_rect = _lib.nk_input_is_mouse_prev_hovering_rect([ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_nk_input_is_mouse_click_in_rect = _lib.nk_input_is_mouse_click_in_rect([ffi.c_void_p, ffi.c_int32, ffi.c_void_p], ffi.c_int32)
_nk_input_is_key_pressed = _lib.nk_input_is_key_pressed([ffi.c_void_p, ffi.c_int32], ffi.c_int32)
_nk_input_is_key_released = _lib.nk_input_is_key_released([ffi.c_void_p, ffi.c_int32], ffi.c_int32)
_nk_input_is_key_down = _lib.nk_input_is_key_down([ffi.c_void_p, ffi.c_int32], ffi.c_int32)

# --- Window ---
_nk_begin          = _lib.nk_begin([ffi.c_void_p, ffi.c_char_p, ffi.c_void_p, ffi.c_uint32], ffi.c_int32)
_nk_begin_titled   = _lib.nk_begin_titled([ffi.c_void_p, ffi.c_char_p, ffi.c_char_p, ffi.c_void_p, ffi.c_uint32], ffi.c_int32)
_nk_end            = _lib.nk_end([ffi.c_void_p], None)
_nk_window_find    = _lib.nk_window_find([ffi.c_void_p, ffi.c_char_p], ffi.c_void_p)
_nk_window_get_bounds = _lib.nk_window_get_bounds([ffi.c_void_p], ffi.c_void_p)
_nk_window_get_position = _lib.nk_window_get_position([ffi.c_void_p], ffi.c_void_p)
_nk_window_get_size    = _lib.nk_window_get_size([ffi.c_void_p], ffi.c_void_p)
_nk_window_get_width   = _lib.nk_window_get_width([ffi.c_void_p], ffi.c_float)
_nk_window_get_height  = _lib.nk_window_get_height([ffi.c_void_p], ffi.c_float)
_nk_window_get_content_region = _lib.nk_window_get_content_region([ffi.c_void_p], ffi.c_void_p)
_nk_window_get_content_region_min = _lib.nk_window_get_content_region_min([ffi.c_void_p], ffi.c_void_p)
_nk_window_get_content_region_max = _lib.nk_window_get_content_region_max([ffi.c_void_p], ffi.c_void_p)
_nk_window_get_content_region_size = _lib.nk_window_get_content_region_size([ffi.c_void_p], ffi.c_void_p)
_nk_window_get_canvas  = _lib.nk_window_get_canvas([ffi.c_void_p], ffi.c_void_p)
_nk_window_has_focus   = _lib.nk_window_has_focus([ffi.c_void_p], ffi.c_int32)
_nk_window_is_hovered  = _lib.nk_window_is_hovered([ffi.c_void_p], ffi.c_int32)
_nk_window_is_collapsed = _lib.nk_window_is_collapsed([ffi.c_void_p, ffi.c_char_p], ffi.c_int32)
_nk_window_is_closed   = _lib.nk_window_is_closed([ffi.c_void_p, ffi.c_char_p], ffi.c_int32)
_nk_window_is_hidden   = _lib.nk_window_is_hidden([ffi.c_void_p, ffi.c_char_p], ffi.c_int32)
_nk_window_is_active   = _lib.nk_window_is_active([ffi.c_void_p, ffi.c_char_p], ffi.c_int32)
_nk_window_set_bounds  = _lib.nk_window_set_bounds([ffi.c_void_p, ffi.c_char_p, ffi.c_void_p], None)
_nk_window_set_position = _lib.nk_window_set_position([ffi.c_void_p, ffi.c_char_p, ffi.c_void_p], None)
_nk_window_set_size    = _lib.nk_window_set_size([ffi.c_void_p, ffi.c_char_p, ffi.c_void_p], None)
_nk_window_set_focus   = _lib.nk_window_set_focus([ffi.c_void_p, ffi.c_char_p], None)
_nk_window_set_scroll  = _lib.nk_window_set_scroll([ffi.c_void_p, ffi.c_uint32, ffi.c_uint32], None)
_nk_window_close       = _lib.nk_window_close([ffi.c_void_p, ffi.c_char_p], None)
_nk_window_collapse    = _lib.nk_window_collapse([ffi.c_void_p, ffi.c_char_p, ffi.c_int32], None)
_nk_window_show        = _lib.nk_window_show([ffi.c_void_p, ffi.c_char_p, ffi.c_int32], None)

# --- Layout ---
_nk_layout_row_dynamic    = _lib.nk_layout_row_dynamic([ffi.c_void_p, ffi.c_float, ffi.c_int32], None)
_nk_layout_row_static     = _lib.nk_layout_row_static([ffi.c_void_p, ffi.c_float, ffi.c_int32, ffi.c_int32], None)
_nk_layout_row_begin      = _lib.nk_layout_row_begin([ffi.c_void_p, ffi.c_int32, ffi.c_float, ffi.c_int32], None)
_nk_layout_row_push       = _lib.nk_layout_row_push([ffi.c_void_p, ffi.c_float], None)
_nk_layout_row_end        = _lib.nk_layout_row_end([ffi.c_void_p], None)
_nk_layout_row            = _lib.nk_layout_row([ffi.c_void_p, ffi.c_int32, ffi.c_float, ffi.c_int32, ffi.c_void_p], None)
_nk_layout_row_template_begin   = _lib.nk_layout_row_template_begin([ffi.c_void_p, ffi.c_float], None)
_nk_layout_row_template_push_dynamic = _lib.nk_layout_row_template_push_dynamic([ffi.c_void_p], None)
_nk_layout_row_template_push_variable = _lib.nk_layout_row_template_push_variable([ffi.c_void_p, ffi.c_float], None)
_nk_layout_row_template_push_static  = _lib.nk_layout_row_template_push_static([ffi.c_void_p, ffi.c_float], None)
_nk_layout_row_template_end   = _lib.nk_layout_row_template_end([ffi.c_void_p], None)
_nk_layout_space_begin        = _lib.nk_layout_space_begin([ffi.c_void_p, ffi.c_int32, ffi.c_float, ffi.c_int32], None)
_nk_layout_space_push         = _lib.nk_layout_space_push([ffi.c_void_p, ffi.c_void_p], None)
_nk_layout_space_end          = _lib.nk_layout_space_end([ffi.c_void_p], None)
_nk_layout_space_bounds       = _lib.nk_layout_space_bounds([ffi.c_void_p], ffi.c_void_p)
_nk_layout_widget_bounds      = _lib.nk_layout_widget_bounds([ffi.c_void_p], ffi.c_void_p)
_nk_layout_ratio_from_pixel   = _lib.nk_layout_ratio_from_pixel([ffi.c_void_p, ffi.c_float], ffi.c_float)
_nk_spacer                    = _lib.nk_spacer([ffi.c_void_p], None)

# --- Group ---
_nk_group_begin        = _lib.nk_group_begin([ffi.c_void_p, ffi.c_char_p, ffi.c_uint32], ffi.c_int32)
_nk_group_begin_titled = _lib.nk_group_begin_titled([ffi.c_void_p, ffi.c_char_p, ffi.c_char_p, ffi.c_uint32], ffi.c_int32)
_nk_group_end          = _lib.nk_group_end([ffi.c_void_p], None)
_nk_group_scrolled_begin = _lib.nk_group_scrolled_begin([ffi.c_void_p, ffi.c_void_p, ffi.c_char_p, ffi.c_uint32], ffi.c_int32)
_nk_group_scrolled_end = _lib.nk_group_scrolled_end([ffi.c_void_p], None)
_nk_group_get_scroll   = _lib.nk_group_get_scroll([ffi.c_void_p, ffi.c_char_p, ffi.c_void_p, ffi.c_void_p], None)
_nk_group_set_scroll   = _lib.nk_group_set_scroll([ffi.c_void_p, ffi.c_char_p, ffi.c_uint32, ffi.c_uint32], None)

# --- Tree ---
_nk_tree_push_hashed   = _lib.nk_tree_push_hashed([ffi.c_void_p, ffi.c_int32, ffi.c_char_p, ffi.c_int32, ffi.c_char_p, ffi.c_int32, ffi.c_int32], ffi.c_int32)
_nk_tree_image_push_hashed = _lib.nk_tree_image_push_hashed([ffi.c_void_p, ffi.c_int32, ffi.c_void_p, ffi.c_char_p, ffi.c_int32, ffi.c_char_p, ffi.c_int32, ffi.c_int32], ffi.c_int32)
_nk_tree_pop           = _lib.nk_tree_pop([ffi.c_void_p], None)
_nk_tree_element_push_hashed = _lib.nk_tree_element_push_hashed([ffi.c_void_p, ffi.c_int32, ffi.c_char_p, ffi.c_int32, ffi.c_void_p, ffi.c_char_p, ffi.c_int32, ffi.c_int32], ffi.c_int32)
_nk_tree_element_pop   = _lib.nk_tree_element_pop([ffi.c_void_p], None)

# --- Widgets: Label / Text ---
_nk_text               = _lib.nk_text([ffi.c_void_p, ffi.c_char_p, ffi.c_int32, ffi.c_uint32], None)
_nk_text_colored       = _lib.nk_text_colored([ffi.c_void_p, ffi.c_char_p, ffi.c_int32, ffi.c_uint32, ffi.c_void_p], None)
_nk_text_wrap          = _lib.nk_text_wrap([ffi.c_void_p, ffi.c_char_p, ffi.c_int32], None)
_nk_text_wrap_colored  = _lib.nk_text_wrap_colored([ffi.c_void_p, ffi.c_char_p, ffi.c_int32, ffi.c_void_p], None)
_nk_label              = _lib.nk_label([ffi.c_void_p, ffi.c_char_p, ffi.c_uint32], None)
_nk_label_colored      = _lib.nk_label_colored([ffi.c_void_p, ffi.c_char_p, ffi.c_uint32, ffi.c_void_p], None)
_nk_label_wrap         = _lib.nk_label_wrap([ffi.c_void_p, ffi.c_char_p], None)
_nk_label_colored_wrap = _lib.nk_label_colored_wrap([ffi.c_void_p, ffi.c_char_p, ffi.c_void_p], None)

# --- Widgets: Button ---
_nk_button_text        = _lib.nk_button_text([ffi.c_void_p, ffi.c_char_p, ffi.c_int32], ffi.c_int32)
_nk_button_label       = _lib.nk_button_label([ffi.c_void_p, ffi.c_char_p], ffi.c_int32)
_nk_button_color       = _lib.nk_button_color([ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_nk_button_symbol      = _lib.nk_button_symbol([ffi.c_void_p, ffi.c_int32], ffi.c_int32)
_nk_button_symbol_label = _lib.nk_button_symbol_label([ffi.c_void_p, ffi.c_int32, ffi.c_char_p, ffi.c_uint32], ffi.c_int32)
_nk_button_set_behavior = _lib.nk_button_set_behavior([ffi.c_void_p, ffi.c_int32], None)
_nk_button_push_behavior = _lib.nk_button_push_behavior([ffi.c_void_p, ffi.c_int32], ffi.c_int32)
_nk_button_pop_behavior = _lib.nk_button_pop_behavior([ffi.c_void_p], ffi.c_int32)

# --- Widgets: Checkbox ---
_nk_check_label        = _lib.nk_check_label([ffi.c_void_p, ffi.c_char_p, ffi.c_int32], ffi.c_int32)
_nk_check_flags_label  = _lib.nk_check_flags_label([ffi.c_void_p, ffi.c_char_p, ffi.c_uint32, ffi.c_uint32], ffi.c_uint32)
_nk_checkbox_label     = _lib.nk_checkbox_label([ffi.c_void_p, ffi.c_char_p, ffi.c_void_p], ffi.c_int32)
_nk_checkbox_flags_label = _lib.nk_checkbox_flags_label([ffi.c_void_p, ffi.c_char_p, ffi.c_void_p, ffi.c_uint32], ffi.c_int32)

# --- Widgets: Radio ---
_nk_radio_label        = _lib.nk_radio_label([ffi.c_void_p, ffi.c_char_p, ffi.c_void_p], ffi.c_int32)
_nk_option_label       = _lib.nk_option_label([ffi.c_void_p, ffi.c_char_p, ffi.c_int32], ffi.c_int32)

# --- Widgets: Selectable ---
_nk_selectable_label   = _lib.nk_selectable_label([ffi.c_void_p, ffi.c_char_p, ffi.c_uint32, ffi.c_void_p], ffi.c_int32)
_nk_selectable_text    = _lib.nk_selectable_text([ffi.c_void_p, ffi.c_char_p, ffi.c_int32, ffi.c_uint32, ffi.c_void_p], ffi.c_int32)
_nk_select_label       = _lib.nk_select_label([ffi.c_void_p, ffi.c_char_p, ffi.c_uint32, ffi.c_int32], ffi.c_int32)

# --- Widgets: Slider ---
_nk_slide_float        = _lib.nk_slide_float([ffi.c_void_p, ffi.c_float, ffi.c_float, ffi.c_float, ffi.c_float], ffi.c_float)
_nk_slide_int          = _lib.nk_slide_int([ffi.c_void_p, ffi.c_int32, ffi.c_int32, ffi.c_int32, ffi.c_int32], ffi.c_int32)
_nk_slider_float       = _lib.nk_slider_float([ffi.c_void_p, ffi.c_float, ffi.c_void_p, ffi.c_float, ffi.c_float], ffi.c_int32)
_nk_slider_int         = _lib.nk_slider_int([ffi.c_void_p, ffi.c_int32, ffi.c_void_p, ffi.c_int32, ffi.c_int32], ffi.c_int32)

# --- Widgets: Knob ---
_nk_knob_float         = _lib.nk_knob_float([ffi.c_void_p, ffi.c_float, ffi.c_void_p, ffi.c_float, ffi.c_float, ffi.c_int32, ffi.c_float], ffi.c_int32)
_nk_knob_int           = _lib.nk_knob_int([ffi.c_void_p, ffi.c_int32, ffi.c_void_p, ffi.c_int32, ffi.c_int32, ffi.c_int32, ffi.c_int32], ffi.c_int32)

# --- Widgets: Progress ---
_nk_progress           = _lib.nk_progress([ffi.c_void_p, ffi.c_void_p, ffi.c_uint64, ffi.c_int32], ffi.c_int32)
_nk_prog               = _lib.nk_prog([ffi.c_void_p, ffi.c_uint64, ffi.c_uint64, ffi.c_int32], ffi.c_uint64)

# --- Widgets: Color picker ---
_nk_color_picker       = _lib.nk_color_picker([ffi.c_void_p, ffi.c_void_p, ffi.c_int32], ffi.c_void_p)
_nk_color_pick         = _lib.nk_color_pick([ffi.c_void_p, ffi.c_void_p, ffi.c_int32], ffi.c_int32)

# --- Widgets: Property ---
_nk_property_int       = _lib.nk_property_int([ffi.c_void_p, ffi.c_char_p, ffi.c_int32, ffi.c_void_p, ffi.c_int32, ffi.c_int32, ffi.c_float], None)
_nk_property_float     = _lib.nk_property_float([ffi.c_void_p, ffi.c_char_p, ffi.c_float, ffi.c_void_p, ffi.c_float, ffi.c_float, ffi.c_float], None)
_nk_property_double    = _lib.nk_property_double([ffi.c_void_p, ffi.c_char_p, ffi.c_double, ffi.c_void_p, ffi.c_double, ffi.c_double, ffi.c_float], None)
_nk_propertyi          = _lib.nk_propertyi([ffi.c_void_p, ffi.c_char_p, ffi.c_int32, ffi.c_int32, ffi.c_int32, ffi.c_int32, ffi.c_float], ffi.c_int32)
_nk_propertyf          = _lib.nk_propertyf([ffi.c_void_p, ffi.c_char_p, ffi.c_float, ffi.c_float, ffi.c_float, ffi.c_float, ffi.c_float], ffi.c_float)
_nk_propertyd          = _lib.nk_propertyd([ffi.c_void_p, ffi.c_char_p, ffi.c_double, ffi.c_double, ffi.c_double, ffi.c_double, ffi.c_float], ffi.c_double)

# --- Widgets: Edit ---
_nk_edit_string        = _lib.nk_edit_string([ffi.c_void_p, ffi.c_uint32, ffi.c_char_p, ffi.c_void_p, ffi.c_int32, ffi.c_void_p], ffi.c_uint32)
_nk_edit_string_zero_terminated = _lib.nk_edit_string_zero_terminated([ffi.c_void_p, ffi.c_uint32, ffi.c_char_p, ffi.c_int32, ffi.c_void_p], ffi.c_uint32)
_nk_edit_focus         = _lib.nk_edit_focus([ffi.c_void_p, ffi.c_uint32], None)
_nk_edit_unfocus       = _lib.nk_edit_unfocus([ffi.c_void_p], None)

# --- Widgets: Chart ---
_nk_chart_begin        = _lib.nk_chart_begin([ffi.c_void_p, ffi.c_int32, ffi.c_int32, ffi.c_float, ffi.c_float], ffi.c_int32)
_nk_chart_begin_colored = _lib.nk_chart_begin_colored([ffi.c_void_p, ffi.c_int32, ffi.c_void_p, ffi.c_void_p, ffi.c_int32, ffi.c_float, ffi.c_float], ffi.c_int32)
_nk_chart_add_slot     = _lib.nk_chart_add_slot([ffi.c_void_p, ffi.c_int32, ffi.c_int32, ffi.c_float, ffi.c_float], None)
_nk_chart_add_slot_colored = _lib.nk_chart_add_slot_colored([ffi.c_void_p, ffi.c_int32, ffi.c_void_p, ffi.c_void_p, ffi.c_int32, ffi.c_float, ffi.c_float], None)
_nk_chart_push         = _lib.nk_chart_push([ffi.c_void_p, ffi.c_float], ffi.c_uint32)
_nk_chart_push_slot    = _lib.nk_chart_push_slot([ffi.c_void_p, ffi.c_float, ffi.c_int32], ffi.c_uint32)
_nk_chart_end          = _lib.nk_chart_end([ffi.c_void_p], None)
_nk_plot               = _lib.nk_plot([ffi.c_void_p, ffi.c_int32, ffi.c_void_p, ffi.c_int32, ffi.c_int32], None)

# --- Widgets: Combo ---
_nk_combo              = _lib.nk_combo([ffi.c_void_p, ffi.c_void_p, ffi.c_int32, ffi.c_int32, ffi.c_int32, ffi.c_void_p], ffi.c_int32)
_nk_combo_begin_label  = _lib.nk_combo_begin_label([ffi.c_void_p, ffi.c_char_p, ffi.c_void_p], ffi.c_int32)
_nk_combo_begin_text   = _lib.nk_combo_begin_text([ffi.c_void_p, ffi.c_char_p, ffi.c_int32, ffi.c_void_p], ffi.c_int32)
_nk_combo_begin_color  = _lib.nk_combo_begin_color([ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_nk_combo_begin_symbol = _lib.nk_combo_begin_symbol([ffi.c_void_p, ffi.c_int32, ffi.c_void_p], ffi.c_int32)
_nk_combo_begin_symbol_label = _lib.nk_combo_begin_symbol_label([ffi.c_void_p, ffi.c_char_p, ffi.c_int32, ffi.c_void_p], ffi.c_int32)
_nk_combo_item_label   = _lib.nk_combo_item_label([ffi.c_void_p, ffi.c_char_p, ffi.c_uint32], ffi.c_int32)
_nk_combo_item_text    = _lib.nk_combo_item_text([ffi.c_void_p, ffi.c_char_p, ffi.c_int32, ffi.c_uint32], ffi.c_int32)
_nk_combo_item_symbol_label = _lib.nk_combo_item_symbol_label([ffi.c_void_p, ffi.c_int32, ffi.c_char_p, ffi.c_uint32], ffi.c_int32)
_nk_combo_close        = _lib.nk_combo_close([ffi.c_void_p], None)
_nk_combo_end          = _lib.nk_combo_end([ffi.c_void_p], None)

# --- Contextual ---
_nk_contextual_begin   = _lib.nk_contextual_begin([ffi.c_void_p, ffi.c_uint32, ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_nk_contextual_item_label = _lib.nk_contextual_item_label([ffi.c_void_p, ffi.c_char_p, ffi.c_uint32], ffi.c_int32)
_nk_contextual_item_text  = _lib.nk_contextual_item_text([ffi.c_void_p, ffi.c_char_p, ffi.c_int32, ffi.c_uint32], ffi.c_int32)
_nk_contextual_item_symbol_label = _lib.nk_contextual_item_symbol_label([ffi.c_void_p, ffi.c_int32, ffi.c_char_p, ffi.c_uint32], ffi.c_int32)
_nk_contextual_close   = _lib.nk_contextual_close([ffi.c_void_p], None)
_nk_contextual_end     = _lib.nk_contextual_end([ffi.c_void_p], None)

# --- Tooltip ---
_nk_tooltip            = _lib.nk_tooltip([ffi.c_void_p, ffi.c_char_p], None)
_nk_tooltip_begin      = _lib.nk_tooltip_begin([ffi.c_void_p, ffi.c_float], ffi.c_int32)
_nk_tooltip_end        = _lib.nk_tooltip_end([ffi.c_void_p], None)

# --- Menubar ---
_nk_menubar_begin      = _lib.nk_menubar_begin([ffi.c_void_p], None)
_nk_menubar_end        = _lib.nk_menubar_end([ffi.c_void_p], None)
_nk_menu_begin_label   = _lib.nk_menu_begin_label([ffi.c_void_p, ffi.c_char_p, ffi.c_uint32, ffi.c_void_p], ffi.c_int32)
_nk_menu_begin_text    = _lib.nk_menu_begin_text([ffi.c_void_p, ffi.c_char_p, ffi.c_int32, ffi.c_uint32, ffi.c_void_p], ffi.c_int32)
_nk_menu_begin_symbol  = _lib.nk_menu_begin_symbol([ffi.c_void_p, ffi.c_char_p, ffi.c_int32, ffi.c_void_p], ffi.c_int32)
_nk_menu_begin_symbol_label = _lib.nk_menu_begin_symbol_label([ffi.c_void_p, ffi.c_char_p, ffi.c_int32, ffi.c_char_p, ffi.c_uint32, ffi.c_void_p], ffi.c_int32)
_nk_menu_item_label    = _lib.nk_menu_item_label([ffi.c_void_p, ffi.c_char_p, ffi.c_uint32], ffi.c_int32)
_nk_menu_item_text     = _lib.nk_menu_item_text([ffi.c_void_p, ffi.c_char_p, ffi.c_int32, ffi.c_uint32], ffi.c_int32)
_nk_menu_item_symbol_label = _lib.nk_menu_item_symbol_label([ffi.c_void_p, ffi.c_int32, ffi.c_char_p, ffi.c_uint32], ffi.c_int32)
_nk_menu_close         = _lib.nk_menu_close([ffi.c_void_p], None)
_nk_menu_end           = _lib.nk_menu_end([ffi.c_void_p], None)

# --- Popup ---
_nk_popup_begin        = _lib.nk_popup_begin([ffi.c_void_p, ffi.c_int32, ffi.c_char_p, ffi.c_uint32, ffi.c_void_p], ffi.c_int32)
_nk_popup_close        = _lib.nk_popup_close([ffi.c_void_p], None)
_nk_popup_end          = _lib.nk_popup_end([ffi.c_void_p], None)
_nk_popup_get_scroll   = _lib.nk_popup_get_scroll([ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], None)
_nk_popup_set_scroll   = _lib.nk_popup_set_scroll([ffi.c_void_p, ffi.c_uint32, ffi.c_uint32], None)

# --- Style ---
_nk_style_push_font          = _lib.nk_style_push_font([ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_nk_style_pop_font           = _lib.nk_style_pop_font([ffi.c_void_p], ffi.c_int32)
_nk_style_push_float         = _lib.nk_style_push_float([ffi.c_void_p, ffi.c_void_p, ffi.c_float], ffi.c_int32)
_nk_style_pop_float          = _lib.nk_style_pop_float([ffi.c_void_p], ffi.c_int32)
_nk_style_push_vec2          = _lib.nk_style_push_vec2([ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_nk_style_pop_vec2           = _lib.nk_style_pop_vec2([ffi.c_void_p], ffi.c_int32)
_nk_style_push_flags         = _lib.nk_style_push_flags([ffi.c_void_p, ffi.c_void_p, ffi.c_uint32], ffi.c_int32)
_nk_style_pop_flags          = _lib.nk_style_pop_flags([ffi.c_void_p], ffi.c_int32)
_nk_style_push_color         = _lib.nk_style_push_color([ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_nk_style_pop_color          = _lib.nk_style_pop_color([ffi.c_void_p], ffi.c_int32)
_nk_style_set_font           = _lib.nk_style_set_font([ffi.c_void_p, ffi.c_void_p], None)
_nk_style_default            = _lib.nk_style_default([ffi.c_void_p], None)
_nk_style_from_table         = _lib.nk_style_from_table([ffi.c_void_p, ffi.c_void_p], None)
_nk_style_load_all_cursors   = _lib.nk_style_load_all_cursors([ffi.c_void_p, ffi.c_void_p], None)
_nk_style_get_color_by_name  = _lib.nk_style_get_color_by_name([ffi.c_int32], ffi.c_char_p)
_nk_style_item_color         = _lib.nk_style_item_color([ffi.c_void_p], ffi.c_void_p)
_nk_style_item_hide          = _lib.nk_style_item_hide([], ffi.c_void_p)

# --- Color conversions ---
_nk_rgb              = _lib.nk_rgb([ffi.c_int32, ffi.c_int32, ffi.c_int32], ffi.c_void_p)
_nk_rgb_iv           = _lib.nk_rgb_iv([ffi.c_void_p], ffi.c_void_p)
_nk_rgb_fv           = _lib.nk_rgb_fv([ffi.c_void_p], ffi.c_void_p)
_nk_rgb_hex          = _lib.nk_rgb_hex([ffi.c_char_p], ffi.c_void_p)
_nk_rgba             = _lib.nk_rgba([ffi.c_int32, ffi.c_int32, ffi.c_int32, ffi.c_int32], ffi.c_void_p)
_nk_rgba_u32         = _lib.nk_rgba_u32([ffi.c_uint32], ffi.c_void_p)
_nk_rgba_hex         = _lib.nk_rgba_hex([ffi.c_char_p], ffi.c_void_p)
_nk_hsva_colorf      = _lib.nk_hsva_colorf([ffi.c_float, ffi.c_float, ffi.c_float, ffi.c_float], ffi.c_void_p)
_nk_color_hex_rgba   = _lib.nk_color_hex_rgba([ffi.c_char_p, ffi.c_void_p], None)
_nk_color_hex_rgb    = _lib.nk_color_hex_rgb([ffi.c_char_p, ffi.c_void_p], None)
_nk_color_d          = _lib.nk_color_d([ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], None)
# _nk_color_u8         = _lib.nk_color_u8([ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], None)

# --- Canvas / draw list ---
# _nk_draw_list_init   = _lib.nk_draw_list_init([ffi.c_void_p], None)
# _nk_draw_list_setup  = _lib.nk_draw_list_setup([ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_int32, ffi.c_int32], None)
# _nk__draw_list_begin = _lib.nk__draw_list_begin([ffi.c_void_p, ffi.c_void_p], ffi.c_void_p)
# _nk__draw_list_next  = _lib.nk__draw_list_next([ffi.c_void_p, ffi.c_void_p], ffi.c_void_p)
# _nk__draw_list_is_empty = _lib.nk__draw_list_is_empty([ffi.c_void_p], ffi.c_int32)

# nk_draw_* (onto draw list)
# _nk_draw_list_path_clear          = _lib.nk_draw_list_path_clear([ffi.c_void_p], None)
# _nk_draw_list_path_line_to        = _lib.nk_draw_list_path_line_to([ffi.c_void_p, ffi.c_void_p], None)
# _nk_draw_list_path_arc_to_fast    = _lib.nk_draw_list_path_arc_to_fast([ffi.c_void_p, ffi.c_void_p, ffi.c_float, ffi.c_int32, ffi.c_int32], None)
# _nk_draw_list_path_arc_to         = _lib.nk_draw_list_path_arc_to([ffi.c_void_p, ffi.c_void_p, ffi.c_float, ffi.c_float, ffi.c_float, ffi.c_uint32], None)
# _nk_draw_list_path_rect_to        = _lib.nk_draw_list_path_rect_to([ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_float], None)
# _nk_draw_list_path_curve_to       = _lib.nk_draw_list_path_curve_to([ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_uint32], None)
# _nk_draw_list_path_fill           = _lib.nk_draw_list_path_fill([ffi.c_void_p, ffi.c_void_p], None)
# _nk_draw_list_path_stroke         = _lib.nk_draw_list_path_stroke([ffi.c_void_p, ffi.c_void_p, ffi.c_int32, ffi.c_float], None)
# _nk_draw_list_stroke_line         = _lib.nk_draw_list_stroke_line([ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_float], None)
# _nk_draw_list_stroke_rect         = _lib.nk_draw_list_stroke_rect([ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_float, ffi.c_float], None)
# _nk_draw_list_stroke_triangle     = _lib.nk_draw_list_stroke_triangle([ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_float], None)
# _nk_draw_list_stroke_circle       = _lib.nk_draw_list_stroke_circle([ffi.c_void_p, ffi.c_void_p, ffi.c_float, ffi.c_void_p, ffi.c_uint32, ffi.c_float], None)
# _nk_draw_list_stroke_curve        = _lib.nk_draw_list_stroke_curve([ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_uint32, ffi.c_float], None)
# _nk_draw_list_stroke_poly_line    = _lib.nk_draw_list_stroke_poly_line([ffi.c_void_p, ffi.c_void_p, ffi.c_uint32, ffi.c_void_p, ffi.c_float, ffi.c_int32, ffi.c_int32], None)
# _nk_draw_list_fill_rect           = _lib.nk_draw_list_fill_rect([ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_float], None)
# _nk_draw_list_fill_rect_multi_color = _lib.nk_draw_list_fill_rect_multi_color([ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], None)
# _nk_draw_list_fill_triangle       = _lib.nk_draw_list_fill_triangle([ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], None)
# _nk_draw_list_fill_circle         = _lib.nk_draw_list_fill_circle([ffi.c_void_p, ffi.c_void_p, ffi.c_float, ffi.c_void_p, ffi.c_uint32], None)
# _nk_draw_list_fill_poly_convex    = _lib.nk_draw_list_fill_poly_convex([ffi.c_void_p, ffi.c_void_p, ffi.c_uint32, ffi.c_void_p, ffi.c_int32], None)
# _nk_draw_list_add_image           = _lib.nk_draw_list_add_image([ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], None)

# --- Window canvas (nk_command_buffer) draw functions ---
_nk_fill_rect              = _lib.nk_fill_rect([ffi.c_void_p, ffi.c_void_p, ffi.c_float, ffi.c_void_p], None)
_nk_fill_rect_multi_color  = _lib.nk_fill_rect_multi_color([ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], None)
_nk_fill_circle            = _lib.nk_fill_circle([ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], None)
_nk_fill_arc               = _lib.nk_fill_arc([ffi.c_void_p, ffi.c_float, ffi.c_float, ffi.c_float, ffi.c_float, ffi.c_float, ffi.c_void_p], None)
_nk_fill_triangle          = _lib.nk_fill_triangle([ffi.c_void_p, ffi.c_float, ffi.c_float, ffi.c_float, ffi.c_float, ffi.c_float, ffi.c_float, ffi.c_void_p], None)
_nk_fill_polygon           = _lib.nk_fill_polygon([ffi.c_void_p, ffi.c_void_p, ffi.c_int32, ffi.c_void_p], None)
_nk_stroke_line            = _lib.nk_stroke_line([ffi.c_void_p, ffi.c_float, ffi.c_float, ffi.c_float, ffi.c_float, ffi.c_float, ffi.c_void_p], None)
_nk_stroke_curve           = _lib.nk_stroke_curve([ffi.c_void_p, ffi.c_float, ffi.c_float, ffi.c_float, ffi.c_float, ffi.c_float, ffi.c_float, ffi.c_float, ffi.c_float, ffi.c_float, ffi.c_void_p], None)
_nk_stroke_rect            = _lib.nk_stroke_rect([ffi.c_void_p, ffi.c_void_p, ffi.c_float, ffi.c_float, ffi.c_void_p], None)
_nk_stroke_circle          = _lib.nk_stroke_circle([ffi.c_void_p, ffi.c_void_p, ffi.c_float, ffi.c_void_p], None)
_nk_stroke_arc             = _lib.nk_stroke_arc([ffi.c_void_p, ffi.c_float, ffi.c_float, ffi.c_float, ffi.c_float, ffi.c_float, ffi.c_float, ffi.c_void_p], None)
_nk_stroke_triangle        = _lib.nk_stroke_triangle([ffi.c_void_p, ffi.c_float, ffi.c_float, ffi.c_float, ffi.c_float, ffi.c_float, ffi.c_float, ffi.c_float, ffi.c_void_p], None)
_nk_stroke_polyline        = _lib.nk_stroke_polyline([ffi.c_void_p, ffi.c_void_p, ffi.c_int32, ffi.c_float, ffi.c_void_p], None)
_nk_stroke_polygon         = _lib.nk_stroke_polygon([ffi.c_void_p, ffi.c_void_p, ffi.c_int32, ffi.c_float, ffi.c_void_p], None)
_nk_draw_image             = _lib.nk_draw_image([ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], None)
_nk_draw_text              = _lib.nk_draw_text([ffi.c_void_p, ffi.c_void_p, ffi.c_char_p, ffi.c_int32, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], None)
_nk_push_scissor           = _lib.nk_push_scissor([ffi.c_void_p, ffi.c_void_p], None)

# --- Command iteration ---
_nk__begin             = _lib.nk__begin([ffi.c_void_p], ffi.c_void_p)
_nk__next              = _lib.nk__next([ffi.c_void_p, ffi.c_void_p], ffi.c_void_p)

# --- Font baking ---
_nk_font_atlas_init_default = _lib.nk_font_atlas_init_default([ffi.c_void_p], None)
_nk_font_atlas_begin    = _lib.nk_font_atlas_begin([ffi.c_void_p], None)
_nk_font_atlas_add_default = _lib.nk_font_atlas_add_default([ffi.c_void_p, ffi.c_float, ffi.c_void_p], ffi.c_void_p)
_nk_font_atlas_add_from_file = _lib.nk_font_atlas_add_from_file([ffi.c_void_p, ffi.c_char_p, ffi.c_float, ffi.c_void_p], ffi.c_void_p)
_nk_font_atlas_bake     = _lib.nk_font_atlas_bake([ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_int32], ffi.c_void_p)
_nk_font_atlas_end      = _lib.nk_font_atlas_end([ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], None)
_nk_font_atlas_cleanup  = _lib.nk_font_atlas_cleanup([ffi.c_void_p], None)
_nk_font_atlas_clear    = _lib.nk_font_atlas_clear([ffi.c_void_p], None)
_nk_font_handle         = _lib.nk_font_handle([ffi.c_void_p], ffi.c_void_p)

# --- Utility ---
_nk_image_id            = _lib.nk_image_id([ffi.c_int32], ffi.c_void_p)
_nk_image_ptr           = _lib.nk_image_ptr([ffi.c_void_p], ffi.c_void_p)
_nk_image_is_subimage   = _lib.nk_image_is_subimage([ffi.c_void_p], ffi.c_int32)
_nk_subimage_id         = _lib.nk_subimage_id([ffi.c_int32, ffi.c_uint16, ffi.c_uint16, ffi.c_void_p], ffi.c_void_p)
_nk_subimage_ptr        = _lib.nk_subimage_ptr([ffi.c_void_p, ffi.c_uint16, ffi.c_uint16, ffi.c_void_p], ffi.c_void_p)
_nk_murmur_hash         = _lib.nk_murmur_hash([ffi.c_void_p, ffi.c_int32, ffi.c_uint32], ffi.c_uint32)
_nk_strmatch_fuzzy_text = _lib.nk_strmatch_fuzzy_text([ffi.c_char_p, ffi.c_int32, ffi.c_char_p, ffi.c_void_p], ffi.c_int32)
_nk_utf_decode          = _lib.nk_utf_decode([ffi.c_char_p, ffi.c_void_p, ffi.c_int32], ffi.c_int32)
_nk_utf_encode          = _lib.nk_utf_encode([ffi.c_uint32, ffi.c_char_p, ffi.c_int32], ffi.c_int32)
_nk_utf_len             = _lib.nk_utf_len([ffi.c_char_p, ffi.c_int32], ffi.c_int32)

# ==============================================================================
#  Color helper class
# ==============================================================================

class Color:
    """
    Immutable RGBA color.

    Usage:
        c = Color(255, 128, 0)
        c = Color(255, 128, 0, 200)
        c = Color.hex("#FF8800")
        c = Color.hsv(0.1, 0.8, 1.0)
        buf = c.to_buf()    # 4-byte nk_color buffer — ffi.free() when done
    """
    def __init__(self, r, g, b, a=255):
        self.r = r & 0xFF
        self.g = g & 0xFF
        self.b = b & 0xFF
        self.a = a & 0xFF

    @staticmethod
    def hex(s):
        s = s.lstrip("#")
        if len(s) == 6:
            r = int(s[0:2], 16); g = int(s[2:4], 16); b = int(s[4:6], 16)
            return Color(r, g, b)
        r = int(s[0:2], 16); g = int(s[2:4], 16)
        b = int(s[4:6], 16); a = int(s[6:8], 16)
        return Color(r, g, b, a)

    @staticmethod
    def u32(v):
        """From packed 0xAARRGGBB."""
        a = (v >> 24) & 0xFF; r = (v >> 16) & 0xFF
        g = (v >> 8)  & 0xFF; b = v & 0xFF
        return Color(r, g, b, a)

    @staticmethod
    def hsv(h, s, v, a=1.0):
        """h,s,v,a in [0..1]."""
        if s == 0:
            i = int(v * 255)
            return Color(i, i, i, int(a * 255))
        i  = int(h * 6)
        f  = h * 6 - i
        p  = v * (1 - s)
        q  = v * (1 - f * s)
        t  = v * (1 - (1 - f) * s)
        i  = i % 6
        if i == 0:   rv, gv, bv = v, t, p
        elif i == 1: rv, gv, bv = q, v, p
        elif i == 2: rv, gv, bv = p, v, t
        elif i == 3: rv, gv, bv = p, q, v
        elif i == 4: rv, gv, bv = t, p, v
        else:        rv, gv, bv = v, p, q
        return Color(int(rv*255), int(gv*255), int(bv*255), int(a*255))

    def to_buf(self):
        """Return a 4-byte nk_color buffer. Caller must ffi.free()."""
        return make_color_buf(self.r, self.g, self.b, self.a)

    def with_alpha(self, a):
        return Color(self.r, self.g, self.b, a)

    def __repr__(self):
        return format_str("Color({self.r},{self.g},{self.b},{self.a})")

# Common colors
WHITE   = Color(255, 255, 255)
BLACK   = Color(0,   0,   0)
RED     = Color(255, 0,   0)
GREEN   = Color(0,   255, 0)
BLUE    = Color(0,   0,   255)
YELLOW  = Color(255, 255, 0)
CYAN    = Color(0,   255, 255)
MAGENTA = Color(255, 0,   255)
GREY    = Color(128, 128, 128)
ORANGE  = Color(255, 165, 0)
TRANSPARENT = Color(0, 0, 0, 0)

# ==============================================================================
#  UserFont (stub / SDL2_ttf integration point)
# ==============================================================================

class UserFont:
    """
    Wraps nk_user_font.  For basic use, the default baked font is recommended.
    This class lets you supply custom font metrics via callbacks if needed.

    Layout of nk_user_font (64-bit):
        userdata (nk_handle = void*)  @ 0   (8 bytes)
        height   (float)              @ 8   (4 bytes)
        padding                       @ 12  (4 bytes)
        width    (function pointer)   @ 16  (8 bytes)
        query    (function pointer)   @ 24  (8 bytes)  [if NK_INCLUDE_VERTEX_BUFFER_OUTPUT]
        texture  (nk_handle = void*)  @ 32  (8 bytes)  [if NK_INCLUDE_VERTEX_BUFFER_OUTPUT]
    """
    _SZ = 64   # generous padding

    def __init__(self, height=13.0, width_fn=None, query_fn=None, texture_ptr=None):
        self._buf       = ffi.malloc(self._SZ)
        self._width_cb  = None
        self._query_cb  = None

        # Zero the buffer
        for i in range(self._SZ):
            _w8(self._buf, i, 0)

        # userdata = NULL
        _w64(self._buf, 0, 0)
        # height
        _wf(self._buf, 8, height)

        if width_fn is not None:
            cb = ffi.callback(width_fn, ffi.c_float,
                              [ffi.c_void_p, ffi.c_float, ffi.c_char_p, ffi.c_int32])
            self._width_cb = cb
            # write function pointer at offset 16
            _w64(self._buf, 16, ffi.addressof(cb))

        if query_fn is not None:
            cb2 = ffi.callback(query_fn, None,
                               [ffi.c_void_p, ffi.c_float, ffi.c_void_p,
                                ffi.c_uint32, ffi.c_void_p])
            self._query_cb = cb2
            _w64(self._buf, 24, ffi.addressof(cb2))

        if texture_ptr is not None:
            _w64(self._buf, 32, texture_ptr)

    @property
    def ptr(self):
        return self._buf

    def free(self):
        if self._buf:
            ffi.free(self._buf)
            self._buf = None

# ==============================================================================
#  FontAtlas class
# ==============================================================================

class FontAtlas:
    """
    Bakes fonts into a texture atlas.

    Usage:
        atlas = FontAtlas()
        atlas.begin()
        default_font = atlas.add_default(13.0)
        # optional:
        # my_font = atlas.add_from_file("arial.ttf", 16.0)
        image_bytes, w, h = atlas.bake()

        # Upload image_bytes to your renderer texture here
        # e.g. with SDL2:
        #   tex = Texture(renderer, SDL_PIXELFORMAT_RGBA8888,
        #                 SDL_TEXTUREACCESS_STATIC, w, h)
        #   tex.update(None, image_bytes, w * 4)

        atlas.end(texture_handle)   # pass your texture id/pointer
        ctx.set_font(default_font)
        # later at shutdown:
        atlas.clear()
    """
    _SZ_ATLAS = 512 * 1024   # 512 KB for nk_font_atlas struct

    def __init__(self):
        self._buf     = ffi.malloc(self._SZ_ATLAS)
        self._default_font = None
        _nk_font_atlas_init_default(self._buf)

    def begin(self):
        _nk_font_atlas_begin(self._buf)

    def add_default(self, height=13.0):
        """Add the built-in default font at given pixel height."""
        ptr = _nk_font_atlas_add_default(self._buf, height, ffi.c_void_p(0))
        f   = _FontHandle(ptr)
        self._default_font = f
        return f

    def add_from_file(self, path, height=16.0):
        """Add a TrueType font from a file path."""
        ptr = _nk_font_atlas_add_from_file(
            self._buf, _enc(path), height, ffi.c_void_p(0))
        return _FontHandle(ptr)

    def bake(self, fmt=NK_RGBA):
        """
        Bake all added fonts into a pixel buffer.
        Returns (raw_bytes, width, height).
        """
        wb  = ffi.malloc(4)
        hb  = ffi.malloc(4)
        try:
            img_ptr = _nk_font_atlas_bake(self._buf, wb, hb, fmt)
            w = read_int_buf(wb)
            h = read_int_buf(hb)
            raw = ffi.buffer_to_bytes(img_ptr, w * h * 4)
            return (raw, w, h)
        except Exception:
            pass
        finally:
            ffi.free(wb)
            ffi.free(hb)

    def end(self, texture_handle_int):
        """
        Finalise the atlas.  texture_handle_int: integer ID of your GPU texture
        (e.g. OpenGL texture object, or SDL2 texture pointer cast to int).
        """
        # nk_handle is a union { void *ptr; int id; }
        # Simplest: pass as void* = texture_handle_int
        handle = ffi.malloc(8)
        _w64(handle, 0, texture_handle_int)
        try:
            _nk_font_atlas_end(self._buf, handle, ffi.c_void_p(0))
        except Exception:
            pass
        finally:
            ffi.free(handle)

    def cleanup(self):
        _nk_font_atlas_cleanup(self._buf)

    def clear(self):
        _nk_font_atlas_clear(self._buf)
        if self._buf:
            ffi.free(self._buf)
            self._buf = None

class _FontHandle:
    """Internal wrapper around a nk_font* pointer."""
    def __init__(self, ptr):
        self._ptr = ptr

    @property
    def user_font_ptr(self):
        """Returns the nk_user_font* for use with nk_style_set_font."""
        return _nk_font_handle(self._ptr)

    @property
    def ptr(self):
        return self._ptr

# ==============================================================================
#  Canvas (nk_command_buffer wrapper)
# ==============================================================================

class Canvas:
    """
    Direct drawing onto a window's command buffer.

    Usage (inside a begin/end block):
        if ctx.begin("win", ...):
            canvas = ctx.get_canvas()
            canvas.fill_rect(10, 10, 100, 50, 0, Color(255,0,0))
            canvas.stroke_circle(200, 200, 40, 40, 2, Color(0,255,0))
            canvas.stroke_line(0,0,100,100, 1, Color.hex("#FFFFFF"))
        ctx.end()
    """
    def __init__(self, ptr):
        self.ptr = ptr

    def _cbuf(self, color):
        return color.to_buf() if isinstance(color, Color) else make_color_buf(*color)

    def push_scissor(self, x, y, w, h):
        buf = make_rect_buf(x, y, w, h)
        try:
            _nk_push_scissor(self.ptr, buf)
        except Exception:
            pass
        finally:
            ffi.free(buf)

    def fill_rect(self, x, y, w, h, rounding, color):
        rb  = make_rect_buf(x, y, w, h)
        cb  = self._cbuf(color)
        try:
            _nk_fill_rect(self.ptr, rb, rounding, cb)
        except Exception:
            pass
        finally:
            ffi.free(rb); ffi.free(cb)

    def fill_rect_multicolor(self, x, y, w, h, top_left, top_right, bottom_right, bottom_left):
        rb  = make_rect_buf(x, y, w, h)
        tl  = self._cbuf(top_left)
        tr  = self._cbuf(top_right)
        br  = self._cbuf(bottom_right)
        bl  = self._cbuf(bottom_left)
        try:
            _nk_fill_rect_multi_color(self.ptr, rb, tl, tr, br, bl)
        except Exception:
            pass
        finally:
            ffi.free(rb); ffi.free(tl); ffi.free(tr); ffi.free(br); ffi.free(bl)

    def fill_circle(self, x, y, w, h, color):
        rb = make_rect_buf(x, y, w, h)
        cb = self._cbuf(color)
        try:
            _nk_fill_circle(self.ptr, rb, cb)
        except Exception:
            pass
        finally:
            ffi.free(rb); ffi.free(cb)

    def fill_arc(self, cx, cy, radius, a_min, a_max, color):
        cb = self._cbuf(color)
        try:
            _nk_fill_arc(self.ptr, cx, cy, radius, a_min, a_max, cb)
        except Exception:
            pass
        finally:
            ffi.free(cb)

    def fill_triangle(self, x0, y0, x1, y1, x2, y2, color):
        cb = self._cbuf(color)
        try:
            _nk_fill_triangle(self.ptr, x0, y0, x1, y1, x2, y2, cb)
        except Exception:
            pass
        finally:
            ffi.free(cb)

    def stroke_line(self, x0, y0, x1, y1, thickness, color):
        cb = self._cbuf(color)
        try:
            _nk_stroke_line(self.ptr, x0, y0, x1, y1, thickness, cb)
        except Exception:
            pass
        finally:
            ffi.free(cb)

    def stroke_curve(self, ax, ay, ctrl0x, ctrl0y, ctrl1x, ctrl1y, bx, by, thickness, color):
        cb = self._cbuf(color)
        try:
            _nk_stroke_curve(self.ptr, ax, ay, ctrl0x, ctrl0y,ctrl1x, ctrl1y, bx, by, thickness, cb)
        except Exception:
            pass
        finally:
            ffi.free(cb)

    def stroke_rect(self, x, y, w, h, rounding, thickness, color):
        rb = make_rect_buf(x, y, w, h)
        cb = self._cbuf(color)
        try:
            _nk_stroke_rect(self.ptr, rb, rounding, thickness, cb)
        except Exception:
            pass
        finally:
            ffi.free(rb); ffi.free(cb)

    def stroke_circle(self, x, y, w, h, thickness, color):
        rb = make_rect_buf(x, y, w, h)
        cb = self._cbuf(color)
        try:
            _nk_stroke_circle(self.ptr, rb, thickness, cb)
        except Exception:
            pass
        finally:
            ffi.free(rb); ffi.free(cb)

    def stroke_arc(self, cx, cy, radius, a_min, a_max, thickness, color):
        cb = self._cbuf(color)
        try:
            _nk_stroke_arc(self.ptr, cx, cy, radius, a_min, a_max, thickness, cb)
        except Exception:
            pass
        finally:
            ffi.free(cb)

    def stroke_triangle(self, x0, y0, x1, y1, x2, y2, thickness, color):
        cb = self._cbuf(color)
        try:
            _nk_stroke_triangle(self.ptr, x0, y0, x1, y1, x2, y2, thickness, cb)
        except Exception:
            pass
        finally:
            ffi.free(cb)

    def stroke_polyline(self, points, thickness, color):
        """points: list of (x,y) float tuples."""
        n   = len(points)
        buf = ffi.malloc(n * 8)
        cb  = self._cbuf(color)
        try:
            for i in range(n):
                _wf(buf, i*8,   points[i][0])
                _wf(buf, i*8+4, points[i][1])
            _nk_stroke_polyline(self.ptr, buf, n, thickness, cb)
        except Exception:
            pass
        finally:
            ffi.free(buf); ffi.free(cb)

    def stroke_polygon(self, points, thickness, color):
        n   = len(points)
        buf = ffi.malloc(n * 8)
        cb  = self._cbuf(color)
        try:
            for i in range(n):
                _wf(buf, i*8,   points[i][0])
                _wf(buf, i*8+4, points[i][1])
            _nk_stroke_polygon(self.ptr, buf, n, thickness, cb)
        except Exception:
            pass
        finally:
            ffi.free(buf); ffi.free(cb)

    def draw_image(self, x, y, w, h, image_ptr, tint_color=None):
        rb  = make_rect_buf(x, y, w, h)
        col = tint_color if tint_color else WHITE
        cb  = self._cbuf(col)
        try:
            _nk_draw_image(self.ptr, rb, image_ptr, cb)
        except Exception:
            pass
        finally:
            ffi.free(rb); ffi.free(cb)

    def draw_text(self, x, y, w, h, text, font_ptr, bg_color, fg_color):
        rb  = make_rect_buf(x, y, w, h)
        bgb = self._cbuf(bg_color)
        fgb = self._cbuf(fg_color)
        raw = _enc(text)
        try:
            _nk_draw_text(self.ptr, rb, raw, len(raw), font_ptr, bgb, fgb)
        except Exception:
            pass
        finally:
            ffi.free(rb); ffi.free(bgb); ffi.free(fgb)

# ==============================================================================
#  Context class (main API)
# ==============================================================================

class Context:
    """
    Central Nuklear context.

    Usage:
        ctx = Context()
        ctx.init()                  # uses nk_init_fixed with internal memory pool

        # Main loop:
        ctx.input_begin()
        # feed events...
        ctx.input_end()

        if ctx.begin("Demo", 50, 50, 400, 300,
                     NK_WINDOW_BORDER | NK_WINDOW_TITLE | NK_WINDOW_MOVABLE):
            ctx.layout_row_dynamic(30, 2)
            if ctx.button_label("Click me!"):
                print("clicked!")
            ctx.label("Hello Nuklear", NK_TEXT_LEFT)
        ctx.end()

        renderer.render(ctx)        # NuklearSDLRenderer
        ctx.clear()
    """
    def __init__(self):
        self._ctx_buf  = ffi.malloc(_CTX_SIZE)
        self._pool_buf = ffi.malloc(_POOL_SIZE)
        self._freed    = False
        self._font     = None
        self._edit_bufs = {}   # name → (buf, len_buf, capacity)

        # Zero context buffer
        for i in range(_CTX_SIZE):
            _w8(self._ctx_buf, i, 0)

    def init(self, font_ptr=None):
        """
        Initialise the context with a fixed memory pool.
        font_ptr: nk_user_font* (from UserFont.ptr or FontAtlas result).
                  Pass None to use a null font (layout only, no text).
        """
        fp  = font_ptr if font_ptr is not None else ffi.c_void_p(0)
        ret = _nk_init_fixed(self._ctx_buf, self._pool_buf, _POOL_SIZE, fp)
        if ret == 0:
            raise RuntimeError("nk_init_fixed failed — pool may be too small")
        self._font = font_ptr

    def set_font(self, font_handle_or_ptr):
        """Set the active font. Pass a _FontHandle or a raw nk_user_font* pointer."""
        if isinstance(font_handle_or_ptr, _FontHandle):
            ptr = font_handle_or_ptr.user_font_ptr
        else:
            ptr = font_handle_or_ptr
        _nk_style_set_font(self._ctx_buf, ptr)
        self._font = ptr

    # --- Input ---
    def input_begin(self):              _nk_input_begin(self._ctx_buf)
    def input_end(self):                _nk_input_end(self._ctx_buf)
    def input_motion(self, x, y):       _nk_input_motion(self._ctx_buf, x, y)
    def input_key(self, key, down):
        _nk_input_key(self._ctx_buf, key, 1 if down else 0)
    def input_button(self, btn, x, y, down):
        _nk_input_button(self._ctx_buf, btn, x, y, 1 if down else 0)
    def input_scroll(self, sx, sy):
        v = make_vec2_buf(sx, sy)
        try:
            _nk_input_scroll(self._ctx_buf, v)
        except Exception:
            pass
        finally:
            ffi.free(v)
    def input_char(self, ch):
        _nk_input_char(self._ctx_buf, ord(ch) if isinstance(ch, str) else ch)
    def input_unicode(self, codepoint):
        _nk_input_unicode(self._ctx_buf, codepoint)
    def input_is_key_pressed(self, key):
        return _nk_input_is_key_pressed(self._ctx_buf, key) != 0
    def input_is_key_released(self, key):
        return _nk_input_is_key_released(self._ctx_buf, key) != 0
    def input_is_key_down(self, key):
        return _nk_input_is_key_down(self._ctx_buf, key) != 0
    def input_is_mouse_hovering_rect(self, x, y, w, h):
        rb = make_rect_buf(x, y, w, h)
        try:
            return _nk_input_is_mouse_hovering_rect(self._ctx_buf, rb) != 0
        except Exception:
            pass
        finally:
            ffi.free(rb)

    # --- Window lifecycle ---
    def begin(self, title, x, y, w, h, flags=NK_WINDOW_BORDER | NK_WINDOW_TITLE | NK_WINDOW_MOVABLE):
        rb = make_rect_buf(x, y, w, h)
        try:
            return _nk_begin(self._ctx_buf, _enc(title), rb, flags) != 0
        except Exception:
            pass
        finally:
            ffi.free(rb)

    def begin_titled(self, name, title, x, y, w, h, flags=NK_WINDOW_BORDER | NK_WINDOW_TITLE | NK_WINDOW_MOVABLE):
        rb = make_rect_buf(x, y, w, h)
        try:
            return _nk_begin_titled(self._ctx_buf, _enc(name), _enc(title), rb, flags) != 0
        except Exception:
            pass
        finally:
            ffi.free(rb)

    def end(self):                      _nk_end(self._ctx_buf)

    def window_is_hovered(self):        return _nk_window_is_hovered(self._ctx_buf) != 0
    def window_has_focus(self):         return _nk_window_has_focus(self._ctx_buf) != 0
    def window_is_collapsed(self, name): return _nk_window_is_collapsed(self._ctx_buf, _enc(name)) != 0
    def window_is_closed(self, name):   return _nk_window_is_closed(self._ctx_buf, _enc(name)) != 0
    def window_is_hidden(self, name):   return _nk_window_is_hidden(self._ctx_buf, _enc(name)) != 0
    def window_is_active(self, name):   return _nk_window_is_active(self._ctx_buf, _enc(name)) != 0
    def window_close(self, name):       _nk_window_close(self._ctx_buf, _enc(name))
    def window_collapse(self, name, state): _nk_window_collapse(self._ctx_buf, _enc(name), state)
    def window_show(self, name, state): _nk_window_show(self._ctx_buf, _enc(name), state)
    def window_set_focus(self, name):   _nk_window_set_focus(self._ctx_buf, _enc(name))
    def window_set_scroll(self, x, y):  _nk_window_set_scroll(self._ctx_buf, x, y)

    def window_get_width(self):         return _nk_window_get_width(self._ctx_buf)
    def window_get_height(self):        return _nk_window_get_height(self._ctx_buf)

    def get_canvas(self):
        ptr = _nk_window_get_canvas(self._ctx_buf)
        return Canvas(ptr)

    # --- Layout ---
    def layout_row_dynamic(self, height, cols):
        _nk_layout_row_dynamic(self._ctx_buf, height, cols)

    def layout_row_static(self, height, item_width, cols):
        _nk_layout_row_static(self._ctx_buf, height, item_width, cols)

    def layout_row_begin(self, fmt, height, cols):
        _nk_layout_row_begin(self._ctx_buf, fmt, height, cols)

    def layout_row_push(self, ratio_or_width):
        _nk_layout_row_push(self._ctx_buf, ratio_or_width)

    def layout_row_end(self):           _nk_layout_row_end(self._ctx_buf)

    def layout_row(self, fmt, height, ratios):
        """ratios: list of floats (ratios for NK_DYNAMIC, widths for NK_STATIC)."""
        n   = len(ratios)
        buf = ffi.malloc(n * 4)
        try:
            for i in range(n):
                _wf(buf, i*4, ratios[i])
            _nk_layout_row(self._ctx_buf, fmt, height, n, buf)
        except Exception:
            pass
        finally:
            ffi.free(buf)

    def layout_row_template_begin(self, height):
        _nk_layout_row_template_begin(self._ctx_buf, height)

    def layout_row_template_push_dynamic(self):
        _nk_layout_row_template_push_dynamic(self._ctx_buf)

    def layout_row_template_push_variable(self, min_width):
        _nk_layout_row_template_push_variable(self._ctx_buf, min_width)

    def layout_row_template_push_static(self, width):
        _nk_layout_row_template_push_static(self._ctx_buf, width)

    def layout_row_template_end(self):
        _nk_layout_row_template_end(self._ctx_buf)

    def layout_space_begin(self, fmt, height, widget_count):
        _nk_layout_space_begin(self._ctx_buf, fmt, height, widget_count)

    def layout_space_push(self, x, y, w, h):
        rb = make_rect_buf(x, y, w, h)
        try:
            _nk_layout_space_push(self._ctx_buf, rb)
        except Exception:
            pass
        finally:
            ffi.free(rb)

    def layout_space_end(self):         _nk_layout_space_end(self._ctx_buf)
    def spacer(self):                   _nk_spacer(self._ctx_buf)
    def layout_ratio_from_pixel(self, pixel_width):
        return _nk_layout_ratio_from_pixel(self._ctx_buf, pixel_width)

    # --- Widgets: Label / Text ---
    def label(self, text, align=NK_TEXT_LEFT):
        _nk_label(self._ctx_buf, _enc(text), align)

    def label_colored(self, text, align, color):
        cb = color.to_buf() if isinstance(color, Color) else make_color_buf(*color)
        try:
            _nk_label_colored(self._ctx_buf, _enc(text), align, cb)
        except Exception:
            pass
        finally:
            ffi.free(cb)

    def label_wrap(self, text):
        _nk_label_wrap(self._ctx_buf, _enc(text))

    def label_colored_wrap(self, text, color):
        cb = color.to_buf() if isinstance(color, Color) else make_color_buf(*color)
        try:
            _nk_label_colored_wrap(self._ctx_buf, _enc(text), cb)
        except Exception:
            pass
        finally:
            ffi.free(cb)

    def text(self, text, align=NK_TEXT_LEFT):
        raw = _enc(text)
        _nk_text(self._ctx_buf, raw, len(raw), align)

    def text_wrap(self, text):
        raw = _enc(text)
        _nk_text_wrap(self._ctx_buf, raw, len(raw))

    # --- Widgets: Button ---
    def button_label(self, text):
        return _nk_button_label(self._ctx_buf, _enc(text)) != 0

    def button_text(self, text):
        raw = _enc(text)
        return _nk_button_text(self._ctx_buf, raw, len(raw)) != 0

    def button_color(self, color):
        cb = color.to_buf() if isinstance(color, Color) else make_color_buf(*color)
        try:
            return _nk_button_color(self._ctx_buf, cb) != 0
        except Exception:
            pass
        finally:
            ffi.free(cb)

    def button_symbol(self, symbol):
        return _nk_button_symbol(self._ctx_buf, symbol) != 0

    def button_symbol_label(self, symbol, text, align=NK_TEXT_LEFT):
        return _nk_button_symbol_label(
            self._ctx_buf, symbol, _enc(text), align) != 0

    def button_set_behavior(self, behavior):
        _nk_button_set_behavior(self._ctx_buf, behavior)

    # --- Widgets: Checkbox ---
    def check_label(self, text, active):
        return _nk_check_label(self._ctx_buf, _enc(text), 1 if active else 0) != 0

    def checkbox_label(self, text, active_ref):
        """active_ref: [bool] list — modified in place."""
        ib = make_int_buf(1 if active_ref[0] else 0)
        try:
            changed = _nk_checkbox_label(self._ctx_buf, _enc(text), ib) != 0
            active_ref[0] = read_int_buf(ib) != 0
            return changed
        except Exception:
            pass
        finally:
            ffi.free(ib)

    # --- Widgets: Radio / Option ---
    def option_label(self, text, active):
        return _nk_option_label(self._ctx_buf, _enc(text), 1 if active else 0) != 0

    def radio_label(self, text, active_ref):
        """active_ref: [bool] list — modified in place."""
        ib = make_int_buf(1 if active_ref[0] else 0)
        try:
            changed = _nk_radio_label(self._ctx_buf, _enc(text), ib) != 0
            active_ref[0] = read_int_buf(ib) != 0
            return changed
        except Exception:
            pass
        finally:
            ffi.free(ib)

    # --- Widgets: Selectable ---
    def selectable_label(self, text, align, value_ref):
        """value_ref: [bool]."""
        ib = make_int_buf(1 if value_ref[0] else 0)
        try:
            changed = _nk_selectable_label(self._ctx_buf, _enc(text), align, ib) != 0
            value_ref[0] = read_int_buf(ib) != 0
            return changed
        except Exception:
            pass
        finally:
            ffi.free(ib)

    def select_label(self, text, align, value):
        return _nk_select_label(self._ctx_buf, _enc(text), align, 1 if value else 0) != 0

    # --- Widgets: Slider ---
    def slide_float(self, min_, val, max_, step):
        return _nk_slide_float(self._ctx_buf, min_, val, max_, step)

    def slide_int(self, min_, val, max_, step):
        return _nk_slide_int(self._ctx_buf, min_, val, max_, step)

    def slider_float(self, min_, val_ref, max_, step):
        """val_ref: [float]."""
        fb = make_float_buf(val_ref[0])
        try:
            changed = _nk_slider_float(self._ctx_buf, min_, fb, max_, step) != 0
            val_ref[0] = read_float_buf(fb)
            return changed
        except Exception:
            pass
        finally:
            ffi.free(fb)

    def slider_int(self, min_, val_ref, max_, step):
        """val_ref: [int]."""
        ib = make_int_buf(val_ref[0])
        try:
            changed = _nk_slider_int(self._ctx_buf, min_, ib, max_, step) != 0
            val_ref[0] = read_int_buf(ib)
            return changed
        except Exception:
            pass
        finally:
            ffi.free(ib)

    # --- Widgets: Progress ---
    def progress(self, cur_ref, max_, modifiable=True):
        """cur_ref: [int]."""
        ib = ffi.malloc(8)   # nk_size = uint64
        _w64(ib, 0, cur_ref[0])
        try:
            changed = _nk_progress(self._ctx_buf, ib, max_, 1 if modifiable else 0) != 0
            b = ffi.buffer_to_bytes(ib, 8)
            cur_ref[0] = (b[0] | (b[1]<<8) | (b[2]<<16) | (b[3]<<24) | (b[4]<<32) | (b[5]<<40) | (b[6]<<48) | (b[7]<<56))
            return changed
        except Exception:
            pass
        finally:
            ffi.free(ib)

    def prog(self, cur, max_, modifiable=True):
        return _nk_prog(self._ctx_buf, cur, max_, 1 if modifiable else 0)

    # --- Widgets: Property ---
    def property_int(self, name, min_, val_ref, max_, step, inc_per_pixel=1.0):
        ib = make_int_buf(val_ref[0])
        try:
            _nk_property_int(self._ctx_buf, _enc(name), min_, ib, max_, step, inc_per_pixel)
            val_ref[0] = read_int_buf(ib)
        except Exception:
            pass
        finally:
            ffi.free(ib)

    def property_float(self, name, min_, val_ref, max_, step, inc_per_pixel=1.0):
        fb = make_float_buf(val_ref[0])
        try:
            _nk_property_float(self._ctx_buf, _enc(name), min_, fb, max_, step, inc_per_pixel)
            val_ref[0] = read_float_buf(fb)
        except Exception:
            pass
        finally:
            ffi.free(fb)

    def propertyi(self, name, min_, val, max_, step, inc_per_pixel=1.0):
        return _nk_propertyi(self._ctx_buf, _enc(name), min_, val, max_, step, inc_per_pixel)

    def propertyf(self, name, min_, val, max_, step, inc_per_pixel=1.0):
        return _nk_propertyf(self._ctx_buf, _enc(name), min_, val, max_, step, inc_per_pixel)

    # --- Widgets: Edit (text input) ---
    def edit_string(self, flags, buf_name, max_len=256, filter_fn=None):
        """
        Stateful text edit.  buf_name: string key used to persist buffer.
        Returns (event_flags, current_text_string).
        """
        if buf_name not in self._edit_bufs:
            cbuf = ffi.malloc(max_len)
            lbuf = make_int_buf(0)
            for i in range(max_len):
                _w8(cbuf, i, 0)
            self._edit_bufs[buf_name] = (cbuf, lbuf, max_len)

        cbuf, lbuf, capacity = self._edit_bufs[buf_name]
        fp = filter_fn if filter_fn else ffi.c_void_p(0)
        ev = _nk_edit_string(self._ctx_buf, flags, cbuf, lbuf, capacity, fp)

        length = read_int_buf(lbuf)
        raw    = ffi.buffer_to_bytes(cbuf, length)
        text   = ""
        for b in raw:
            if b == 0: break
            text = text + chr(b)
        return (ev, text)

    def edit_set_text(self, buf_name, text, max_len=256):
        """Pre-fill an edit buffer by name."""
        if buf_name not in self._edit_bufs:
            cbuf = ffi.malloc(max_len)
            lbuf = make_int_buf(0)
            for i in range(max_len):
                _w8(cbuf, i, 0)
            self._edit_bufs[buf_name] = (cbuf, lbuf, max_len)
        cbuf, lbuf, capacity = self._edit_bufs[buf_name]
        raw = _enc(text)
        n   = min(len(raw), capacity - 1)
        for i in range(n):
            _w8(cbuf, i, raw[i])
        _w8(cbuf, n, 0)
        _w32(lbuf, 0, n)

    def edit_focus(self, flags=NK_EDIT_DEFAULT):
        _nk_edit_focus(self._ctx_buf, flags)

    def edit_unfocus(self):
        _nk_edit_unfocus(self._ctx_buf)

    # --- Widgets: Chart ---
    def chart_begin(self, chart_type, count, min_, max_):
        return _nk_chart_begin(self._ctx_buf, chart_type, count, min_, max_) != 0

    def chart_begin_colored(self, chart_type, color, active_color, count, min_, max_):
        cb = color.to_buf() if isinstance(color, Color) else make_color_buf(*color)
        ab = active_color.to_buf() if isinstance(active_color, Color) else make_color_buf(*active_color)
        try:
            return _nk_chart_begin_colored(self._ctx_buf, chart_type, cb, ab, count, min_, max_) != 0
        except Exception:
            pass
        finally:
            ffi.free(cb); ffi.free(ab)

    def chart_push(self, value):        return _nk_chart_push(self._ctx_buf, value)
    def chart_push_slot(self, value, slot): return _nk_chart_push_slot(self._ctx_buf, value, slot)
    def chart_end(self):                _nk_chart_end(self._ctx_buf)

    def plot(self, chart_type, values, offset=0):
        n   = len(values)
        buf = ffi.malloc(n * 4)
        try:
            for i in range(n):
                _wf(buf, i*4, values[i])
            _nk_plot(self._ctx_buf, chart_type, buf, n, offset)
        except Exception:
            pass
        finally:
            ffi.free(buf)

    # --- Combo ---
    def combo(self, items, selected, item_height, w, h):
        """items: list of strings. Returns selected index."""
        ptrs_buf = ffi.malloc(len(items) * 8)
        enc_items = []
        try:
            for i in range(len(items)):
                enc = _enc(items[i])
                enc_items.append(enc)
                _w64(ptrs_buf, i*8, ffi.addressof(enc))
            sz = make_vec2_buf(w, h)
            result = _nk_combo(self._ctx_buf, ptrs_buf, len(items), selected, item_height, sz)
            ffi.free(sz)
            return result
        except Exception:
            pass
        finally:
            ffi.free(ptrs_buf)

    def combo_begin_label(self, text, w, h):
        sz = make_vec2_buf(w, h)
        try:
            return _nk_combo_begin_label(self._ctx_buf, _enc(text), sz) != 0
        except Exception:
            pass
        finally:
            ffi.free(sz)

    def combo_begin_color(self, color, w, h):
        cb = color.to_buf() if isinstance(color, Color) else make_color_buf(*color)
        sz = make_vec2_buf(w, h)
        try:
            return _nk_combo_begin_color(self._ctx_buf, cb, sz) != 0
        except Exception:
            pass
        finally:
            ffi.free(cb); ffi.free(sz)

    def combo_begin_symbol(self, symbol, w, h):
        sz = make_vec2_buf(w, h)
        try:
            return _nk_combo_begin_symbol(self._ctx_buf, symbol, sz) != 0
        except Exception:
            pass
        finally:
            ffi.free(sz)

    def combo_item_label(self, text, align=NK_TEXT_LEFT):
        return _nk_combo_item_label(self._ctx_buf, _enc(text), align) != 0

    def combo_item_symbol_label(self, symbol, text, align=NK_TEXT_LEFT):
        return _nk_combo_item_symbol_label(self._ctx_buf, symbol, _enc(text), align) != 0

    def combo_close(self):              _nk_combo_close(self._ctx_buf)
    def combo_end(self):                _nk_combo_end(self._ctx_buf)

    # --- Contextual ---
    def contextual_begin(self, flags, w, h, trigger_bounds_x, trigger_bounds_y, trigger_bounds_w, trigger_bounds_h):
        sz  = make_vec2_buf(w, h)
        trb = make_rect_buf(trigger_bounds_x, trigger_bounds_y, trigger_bounds_w, trigger_bounds_h)
        try:
            return _nk_contextual_begin(self._ctx_buf, flags, sz, trb) != 0
        except Exception:
            pass
        finally:
            ffi.free(sz); ffi.free(trb)

    def contextual_item_label(self, text, align=NK_TEXT_LEFT):
        return _nk_contextual_item_label(self._ctx_buf, _enc(text), align) != 0

    def contextual_close(self):         _nk_contextual_close(self._ctx_buf)
    def contextual_end(self):           _nk_contextual_end(self._ctx_buf)

    # --- Tooltip ---
    def tooltip(self, text):            _nk_tooltip(self._ctx_buf, _enc(text))

    def tooltip_begin(self, width):
        return _nk_tooltip_begin(self._ctx_buf, width) != 0

    def tooltip_end(self):              _nk_tooltip_end(self._ctx_buf)

    # --- Menubar ---
    def menubar_begin(self):            _nk_menubar_begin(self._ctx_buf)
    def menubar_end(self):              _nk_menubar_end(self._ctx_buf)

    def menu_begin_label(self, text, align, w, h):
        sz = make_vec2_buf(w, h)
        try:
            return _nk_menu_begin_label(self._ctx_buf, _enc(text), align, sz) != 0
        except Exception:
            pass
        finally:
            ffi.free(sz)

    def menu_begin_symbol_label(self, text, symbol, align, w, h):
        sz = make_vec2_buf(w, h)
        try:
            return _nk_menu_begin_symbol_label(self._ctx_buf, _enc(text), symbol, _enc(text), align, sz) != 0
        except Exception:
            pass
        finally:
            ffi.free(sz)

    def menu_item_label(self, text, align=NK_TEXT_LEFT):
        return _nk_menu_item_label(self._ctx_buf, _enc(text), align) != 0

    def menu_item_symbol_label(self, symbol, text, align=NK_TEXT_LEFT):
        return _nk_menu_item_symbol_label(self._ctx_buf, symbol, _enc(text), align) != 0

    def menu_close(self):               _nk_menu_close(self._ctx_buf)
    def menu_end(self):                 _nk_menu_end(self._ctx_buf)

    # --- Group ---
    def group_begin(self, title, flags=NK_WINDOW_BORDER):
        return _nk_group_begin(self._ctx_buf, _enc(title), flags) != 0

    def group_begin_titled(self, name, title, flags=NK_WINDOW_BORDER):
        return _nk_group_begin_titled(self._ctx_buf, _enc(name), _enc(title), flags) != 0

    def group_end(self):                _nk_group_end(self._ctx_buf)

    def group_set_scroll(self, name, x_offset, y_offset):
        _nk_group_set_scroll(self._ctx_buf, _enc(name), x_offset, y_offset)

    # --- Popup ---
    def popup_begin(self, popup_type, title, flags, x, y, w, h):
        rb = make_rect_buf(x, y, w, h)
        try:
            return _nk_popup_begin(self._ctx_buf, popup_type, _enc(title), flags, rb) != 0
        except Exception:
            pass
        finally:
            ffi.free(rb)

    def popup_close(self):              _nk_popup_close(self._ctx_buf)
    def popup_end(self):                _nk_popup_end(self._ctx_buf)

    # --- Tree ---
    def tree_push(self, tree_type, title, state=NK_MINIMIZED, line=0):
        h = hash(title) & 0x7FFFFFFF
        raw = _enc(title)
        return _nk_tree_push_hashed(self._ctx_buf, tree_type, raw, state, raw, len(raw), h + line) != 0

    def tree_pop(self):                 _nk_tree_pop(self._ctx_buf)

    def tree_element_push(self, tree_type, title, state, selected_ref, line=0):
        """selected_ref: [bool]."""
        ib  = make_int_buf(1 if selected_ref[0] else 0)
        raw = _enc(title)
        h   = hash(title) & 0x7FFFFFFF
        try:
            opened = _nk_tree_element_push_hashed(self._ctx_buf, tree_type, raw, state, ib, raw, len(raw), h + line) != 0
            selected_ref[0] = read_int_buf(ib) != 0
            return opened
        except Exception:
            pass
        finally:
            ffi.free(ib)

    def tree_element_pop(self):         _nk_tree_element_pop(self._ctx_buf)

    # --- Style ---
    def style_default(self):            _nk_style_default(self._ctx_buf)
    def style_set_font(self, font_ptr): _nk_style_set_font(self._ctx_buf, font_ptr)

    def style_push_color(self, color_field_ptr, color):
        cb = color.to_buf() if isinstance(color, Color) else make_color_buf(*color)
        try:
            return _nk_style_push_color(self._ctx_buf, color_field_ptr, cb) != 0
        except Exception:
            pass
        finally:
            ffi.free(cb)

    def style_pop_color(self):          return _nk_style_pop_color(self._ctx_buf) != 0
    def style_push_float(self, field_ptr, val):
        return _nk_style_push_float(self._ctx_buf, field_ptr, val) != 0
    def style_pop_float(self):          return _nk_style_pop_float(self._ctx_buf) != 0
    def style_push_flags(self, field_ptr, flags):
        return _nk_style_push_flags(self._ctx_buf, field_ptr, flags) != 0
    def style_pop_flags(self):          return _nk_style_pop_flags(self._ctx_buf) != 0

    def style_set_colors(self, color_table):
        """
        color_table: dict mapping NK_COLOR_* → Color.
        Sets all colors at once using nk_style_from_table.
        """
        buf = ffi.malloc(NK_COLOR_COUNT * 4)
        try:
            # Default: all black
            for i in range(NK_COLOR_COUNT * 4):
                _w8(buf, i, 0)
            for idx, color in color_table.items():
                off = idx * 4
                c   = color if isinstance(color, Color) else Color(*color)
                _w8(buf, off,   c.r)
                _w8(buf, off+1, c.g)
                _w8(buf, off+2, c.b)
                _w8(buf, off+3, c.a)
            _nk_style_from_table(self._ctx_buf, buf)
        except Exception:
            pass
        finally:
            ffi.free(buf)

    # --- Frame ---
    def clear(self):                    _nk_clear(self._ctx_buf)

    def free(self):
        if not self._freed:
            # Free edit buffers
            for name in self._edit_bufs:
                cbuf, lbuf, _ = self._edit_bufs[name]
                ffi.free(cbuf)
                ffi.free(lbuf)
            self._edit_bufs = {}
            _nk_free(self._ctx_buf)
            ffi.free(self._ctx_buf)
            ffi.free(self._pool_buf)
            self._ctx_buf  = None
            self._pool_buf = None
            self._freed    = True

    @property
    def ptr(self):
        return self._ctx_buf

# ==============================================================================
#  SDL2 Command Renderer
# ==============================================================================

class NuklearSDLRenderer:
    """
    Renders a Nuklear context into an SDL2 Renderer by iterating the
    nk_command buffer and translating each draw command to SDL2 calls.

    Command struct offsets (64-bit, no NK_INCLUDE_COMMAND_USERDATA):
        Header (16 bytes): type@0(i32), next@8(u64)

    Usage:
        renderer = NuklearSDLRenderer(sdl_renderer)
        # per frame:
        renderer.render(ctx)
    """
    def __init__(self, sdl_renderer):
        """sdl_renderer: sdl.Renderer instance."""
        self._ren = sdl_renderer

    def render(self, ctx):
        """Iterate nk_commands and draw each to SDL2."""
        cmd = _nk__begin(ctx.ptr)

        while cmd is not None and cmd.Address != 0:
            cmd_type = _r32(cmd, 0)

            if cmd_type == NK_COMMAND_NOP:
                pass

            elif cmd_type == NK_COMMAND_SCISSOR:
                x = _rs16(cmd, _CMD_HDR)
                y = _rs16(cmd, _CMD_HDR + 2)
                w = _r16( cmd, _CMD_HDR + 4)
                h = _r16( cmd, _CMD_HDR + 6)
                if x < 0 or y < 0 or w <= 0 or h <= 0:
                    self._ren.disable_clip()
                else:
                    self._ren.set_clip_rect(x, y, w, h)

            elif cmd_type == NK_COMMAND_LINE:
                t  = _r16(cmd, _CMD_HDR)
                x0 = _rs16(cmd, _CMD_HDR + 2)
                y0 = _rs16(cmd, _CMD_HDR + 4)
                x1 = _rs16(cmd, _CMD_HDR + 6)
                y1 = _rs16(cmd, _CMD_HDR + 8)
                r, g, b, a = _read_color(cmd, _CMD_HDR + 10)
                self._ren.set_draw_color(r, g, b, a)
                self._ren.draw_line(x0, y0, x1, y1)

            elif cmd_type == NK_COMMAND_RECT:
                # rounding@H, line_thickness@H+2, x@H+4, y@H+6, w@H+8, h@H+10, color@H+12
                rounding = _r16(cmd, _CMD_HDR)
                lt       = _r16(cmd, _CMD_HDR + 2)
                x  = _rs16(cmd, _CMD_HDR + 4)
                y  = _rs16(cmd, _CMD_HDR + 6)
                w  = _r16( cmd, _CMD_HDR + 8)
                h  = _r16( cmd, _CMD_HDR + 10)
                r, g, b, a = _read_color(cmd, _CMD_HDR + 12)
                self._ren.set_draw_color(r, g, b, a)
                self._ren.draw_rect(x, y, w, h)

            elif cmd_type == NK_COMMAND_RECT_FILLED:
                # rounding@H, x@H+2, y@H+4, w@H+6, h@H+8, color@H+10
                x  = _rs16(cmd, _CMD_HDR + 2)
                y  = _rs16(cmd, _CMD_HDR + 4)
                w  = _r16( cmd, _CMD_HDR + 6)
                h  = _r16( cmd, _CMD_HDR + 8)
                r, g, b, a = _read_color(cmd, _CMD_HDR + 10)
                self._ren.set_draw_color(r, g, b, a)
                self._ren.fill_rect(x, y, w, h)

            elif cmd_type == NK_COMMAND_RECT_MULTI_COLOR:
                # x@H, y@H+2, w@H+4, h@H+6, top_left@H+8, top_right@H+12,
                # bottom_right@H+16, bottom_left@H+20
                x  = _rs16(cmd, _CMD_HDR)
                y  = _rs16(cmd, _CMD_HDR + 2)
                w  = _r16( cmd, _CMD_HDR + 4)
                h  = _r16( cmd, _CMD_HDR + 6)
                # Approximate: fill with top-left color
                r, g, b, a = _read_color(cmd, _CMD_HDR + 8)
                self._ren.set_draw_color(r, g, b, a)
                self._ren.fill_rect(x, y, w, h)

            elif cmd_type == NK_COMMAND_CIRCLE:
                x  = _rs16(cmd, _CMD_HDR)
                y  = _rs16(cmd, _CMD_HDR + 2)
                lt = _r16( cmd, _CMD_HDR + 4)
                w  = _r16( cmd, _CMD_HDR + 6)
                h  = _r16( cmd, _CMD_HDR + 8)
                r, g, b, a = _read_color(cmd, _CMD_HDR + 10)
                self._ren.set_draw_color(r, g, b, a)
                self._draw_ellipse_outline(x, y, w, h)

            elif cmd_type == NK_COMMAND_CIRCLE_FILLED:
                x  = _rs16(cmd, _CMD_HDR)
                y  = _rs16(cmd, _CMD_HDR + 2)
                w  = _r16( cmd, _CMD_HDR + 4)
                h  = _r16( cmd, _CMD_HDR + 6)
                r, g, b, a = _read_color(cmd, _CMD_HDR + 8)
                self._ren.set_draw_color(r, g, b, a)
                self._draw_ellipse_filled(x, y, w, h)

            elif cmd_type == NK_COMMAND_TRIANGLE:
                lt = _r16( cmd, _CMD_HDR)
                ax = _rs16(cmd, _CMD_HDR + 2);  ay = _rs16(cmd, _CMD_HDR + 4)
                bx = _rs16(cmd, _CMD_HDR + 6);  by = _rs16(cmd, _CMD_HDR + 8)
                cx = _rs16(cmd, _CMD_HDR + 10); cy = _rs16(cmd, _CMD_HDR + 12)
                r, g, b, a = _read_color(cmd, _CMD_HDR + 14)
                self._ren.set_draw_color(r, g, b, a)
                self._ren.draw_line(ax, ay, bx, by)
                self._ren.draw_line(bx, by, cx, cy)
                self._ren.draw_line(cx, cy, ax, ay)

            elif cmd_type == NK_COMMAND_TRIANGLE_FILLED:
                ax = _rs16(cmd, _CMD_HDR);      ay = _rs16(cmd, _CMD_HDR + 2)
                bx = _rs16(cmd, _CMD_HDR + 4);  by = _rs16(cmd, _CMD_HDR + 6)
                cx = _rs16(cmd, _CMD_HDR + 8);  cy = _rs16(cmd, _CMD_HDR + 10)
                r, g, b, a = _read_color(cmd, _CMD_HDR + 12)
                self._ren.set_draw_color(r, g, b, a)
                self._draw_triangle_filled(ax, ay, bx, by, cx, cy)

            elif cmd_type == NK_COMMAND_TEXT:
                # font_ptr@H(8 bytes), bg@H+8(4), fg@H+12(4),
                # x@H+16(s16), y@H+18(s16), w@H+20(u16), h@H+22(u16),
                # height@H+24(f32), length@H+28(i32), string@H+32
                fg_r, fg_g, fg_b, fg_a = _read_color(cmd, _CMD_HDR + 12)
                x      = _rs16(cmd, _CMD_HDR + 16)
                y      = _rs16(cmd, _CMD_HDR + 18)
                length = _r32( cmd, _CMD_HDR + 28)
                # Text rendering requires a font — we draw a placeholder rect
                # Integrate with SDL2_ttf or your font system here
                self._ren.set_draw_color(fg_r, fg_g, fg_b, 64)
                # (no actual text rendering without SDL2_ttf)

            elif cmd_type == NK_COMMAND_IMAGE:
                # nk_image@H(24 bytes), x@H+24, y@H+26, w@H+28, h@H+30, color@H+32
                x = _rs16(cmd, _CMD_HDR + 24)
                y = _rs16(cmd, _CMD_HDR + 26)
                w = _r16( cmd, _CMD_HDR + 28)
                h = _r16( cmd, _CMD_HDR + 30)
                # image handle is at cmd+_CMD_HDR: void* ptr or int id
                img_id = _r64(cmd, _CMD_HDR)
                # Draw placeholder rect — integrate with your texture system
                self._ren.set_draw_color(128, 128, 128, 255)
                self._ren.fill_rect(x, y, w, h)

            cmd_next_off = _r64(cmd, 8)
            if cmd_next_off == 0:
                break
            # Advance: nk__next handles pointer arithmetic internally
            cmd = _nk__next(ctx.ptr, cmd)

        # Restore clip
        self._ren.disable_clip()

    # --- Shape rasterizers ---

    def _draw_ellipse_outline(self, rx, ry, w, h):
        """Midpoint ellipse algorithm for outline."""
        cx  = rx + w // 2
        cy  = ry + h // 2
        a   = w // 2
        b   = h // 2
        if a == 0 or b == 0:
            return None
        x = 0; y = b
        a2 = a * a; b2 = b * b
        d  = b2 - a2 * b + a2 // 4
        points = []
        while b2 * x < a2 * y:
            points.append((cx+x, cy+y)); points.append((cx-x, cy+y))
            points.append((cx+x, cy-y)); points.append((cx-x, cy-y))
            if d < 0:
                d = d + b2 * (2*x + 3)
            else:
                d = d + b2*(2*x+3) + a2*(2 - 2*y)
                y = y - 1
            x = x + 1
        while y >= 0:
            points.append((cx+x, cy+y)); points.append((cx-x, cy+y))
            points.append((cx+x, cy-y)); points.append((cx-x, cy-y))
            d = d + a2*(2*y - 3)
            y = y - 1

    def _draw_ellipse_filled(self, rx, ry, w, h):
        """Scan-line fill for ellipse."""
        cx = rx + w // 2; cy = ry + h // 2
        a  = w // 2;      b  = h // 2
        if a == 0 or b == 0:
            return None
        for dy in range(-b, b+1):
            t  = dy * dy / (b * b) if b != 0 else 1
            if t > 1: continue
            dx = int(a * (1 - t) ** 0.5)
            self._ren.draw_line(cx - dx, cy + dy, cx + dx, cy + dy)

    def _draw_triangle_filled(self, x0, y0, x1, y1, x2, y2):
        """Scan-line triangle fill."""
        pts = sorted([(x0,y0),(x1,y1),(x2,y2)], key=lambda p: p[1])
        (ax, ay), (bx, by), (cx, cy) = pts[0], pts[1], pts[2]
        if cy == ay: return None

        def lerp(t, a, b):
            return a + t * (b - a)

        for y in range(ay, cy + 1):
            if y < by:
                t1 = (y - ay) / (by - ay) if by != ay else 0
                t2 = (y - ay) / (cy - ay) if cy != ay else 0
                x_left  = int(lerp(t1, ax, bx))
                x_right = int(lerp(t2, ax, cx))
            else:
                t1 = (y - by) / (cy - by) if cy != by else 1
                t2 = (y - ay) / (cy - ay) if cy != ay else 1
                x_left  = int(lerp(t1, bx, cx))
                x_right = int(lerp(t2, ax, cx))
            if x_left > x_right:
                x_left, x_right = x_right, x_left
            self._ren.draw_line(x_left, y, x_right, y)

# ==============================================================================
#  SDL2 Input Bridge (standalone, works without NuklearSDLBridge in sdl wrapper)
# ==============================================================================

class NuklearSDLInput:
    """
    Feeds SDL2 events into a Nuklear Context.
    Import your sdl module and pass events here.

    Usage:
        nk_input = NuklearSDLInput(ctx)
        ev = sdl.Event()
        # per frame:
        nk_input.begin()
        while ev.poll():
            nk_input.feed(ev)
        nk_input.end()
    """
    def __init__(self, ctx):
        self._ctx      = ctx
        self._scroll_x = 0.0
        self._scroll_y = 0.0

    def begin(self):
        self._ctx.input_begin()
        self._scroll_x = 0.0
        self._scroll_y = 0.0

    def end(self):
        self._ctx.input_scroll(self._scroll_x, self._scroll_y)
        self._ctx.input_end()

    def feed(self, event):
        """event: an sdl.Event instance."""
        t = event.type

        # Import SDL constants inline to avoid circular import
        _KEYDOWN        = 0x300
        _KEYUP          = 0x301
        _TEXTINPUT      = 0x303
        _MOUSEMOTION    = 0x400
        _MOUSEBUTTONDOWN = 0x401
        _MOUSEBUTTONUP  = 0x402
        _MOUSEWHEEL     = 0x403
        _FINGERDOWN     = 0x700
        _FINGERUP       = 0x701
        _FINGERMOTION   = 0x702

        if t == _KEYDOWN or t == _KEYUP:
            down = (t == _KEYDOWN)
            sc   = event.scancode
            mod  = event.mod
            ctrl = (mod & 0x00C0) != 0   # KMOD_CTRL
            # Map SDL scancodes → NK_KEY
            _map = {
                225: NK_KEY_SHIFT,   229: NK_KEY_SHIFT,   # LSHIFT, RSHIFT
                76:  NK_KEY_DEL,                           # DELETE
                40:  NK_KEY_ENTER,   88: NK_KEY_ENTER,    # RETURN, KP_ENTER
                43:  NK_KEY_TAB,                           # TAB
                42:  NK_KEY_BACKSPACE,                     # BACKSPACE
                74:  NK_KEY_TEXT_START,                    # HOME
                77:  NK_KEY_TEXT_END,                      # END
                75:  NK_KEY_SCROLL_UP,                     # PAGEUP
                78:  NK_KEY_SCROLL_DOWN,                   # PAGEDOWN
                80:  NK_KEY_LEFT,                          # LEFT
                79:  NK_KEY_RIGHT,                         # RIGHT
                82:  NK_KEY_UP,                            # UP
                81:  NK_KEY_DOWN,                          # DOWN
            }
            if sc == 27 and ctrl:    # C → COPY
                self._ctx.input_key(NK_KEY_COPY, down)
            elif sc == 25 and ctrl:  # V → PASTE
                self._ctx.input_key(NK_KEY_PASTE, down)
            elif sc == 27 and ctrl:  # X → CUT
                self._ctx.input_key(NK_KEY_CUT, down)
            elif sc == 29 and ctrl:  # Z → UNDO
                self._ctx.input_key(NK_KEY_TEXT_UNDO, down)
            elif sc == 28 and ctrl:  # Y → REDO
                self._ctx.input_key(NK_KEY_TEXT_REDO, down)
            elif sc == 4 and ctrl:   # A → SELECT_ALL
                self._ctx.input_key(NK_KEY_TEXT_SELECT_ALL, down)
            elif sc == 80 and ctrl:  # LEFT+CTRL → WORD_LEFT
                self._ctx.input_key(NK_KEY_TEXT_WORD_LEFT, down)
            elif sc == 79 and ctrl:  # RIGHT+CTRL → WORD_RIGHT
                self._ctx.input_key(NK_KEY_TEXT_WORD_RIGHT, down)
            elif sc in _map:
                self._ctx.input_key(_map[sc], down)

        elif t == _TEXTINPUT:
            for ch in event.text:
                self._ctx.input_char(ch)

        elif t == _MOUSEMOTION:
            self._ctx.input_motion(event.x, event.y)

        elif t == _MOUSEBUTTONDOWN or t == _MOUSEBUTTONUP:
            down = (t == _MOUSEBUTTONDOWN)
            btn  = event.button
            x    = event.x
            y    = event.y
            if btn == 1:   # LEFT
                if event.clicks == 2:
                    self._ctx.input_button(NK_BUTTON_DOUBLE, x, y, down)
                self._ctx.input_button(NK_BUTTON_LEFT, x, y, down)
            elif btn == 2: # MIDDLE
                self._ctx.input_button(NK_BUTTON_MIDDLE, x, y, down)
            elif btn == 3: # RIGHT
                self._ctx.input_button(NK_BUTTON_RIGHT, x, y, down)

        elif t == _MOUSEWHEEL:
            self._scroll_x = self._scroll_x + event.wheel_x
            self._scroll_y = self._scroll_y + event.wheel_y

        elif t == _FINGERDOWN or t == _FINGERMOTION:
            self._ctx.input_motion(int(event.finger_x * 1920), int(event.finger_y * 1080))
            self._ctx.input_button(NK_BUTTON_LEFT, int(event.finger_x * 1920), int(event.finger_y * 1080), True)
        elif t == _FINGERUP:
            self._ctx.input_button(NK_BUTTON_LEFT, int(event.finger_x * 1920), int(event.finger_y * 1080), False)

# ==============================================================================
#  Theme presets
# ==============================================================================

def theme_dark():
    """Return a dark color table dict for ctx.style_set_colors()."""
    return {
        NK_COLOR_TEXT:                    Color(210, 210, 210),
        NK_COLOR_WINDOW:                  Color( 57,  67,  71),
        NK_COLOR_HEADER:                  Color( 51,  51,  56, 220),
        NK_COLOR_BORDER:                  Color( 46,  46,  46),
        NK_COLOR_BUTTON:                  Color( 48,  83, 111),
        NK_COLOR_BUTTON_HOVER:            Color( 58,  93, 121),
        NK_COLOR_BUTTON_ACTIVE:           Color( 63,  98, 126),
        NK_COLOR_TOGGLE:                  Color( 50,  58,  61),
        NK_COLOR_TOGGLE_HOVER:            Color( 45,  53,  56),
        NK_COLOR_TOGGLE_CURSOR:           Color( 48,  83, 111),
        NK_COLOR_SELECT:                  Color( 57,  67,  61),
        NK_COLOR_SELECT_ACTIVE:           Color( 48,  83, 111),
        NK_COLOR_SLIDER:                  Color( 50,  58,  61),
        NK_COLOR_SLIDER_CURSOR:           Color( 48,  83, 111),
        NK_COLOR_SLIDER_CURSOR_HOVER:     Color( 53,  88, 116),
        NK_COLOR_SLIDER_CURSOR_ACTIVE:    Color( 58,  93, 121),
        NK_COLOR_PROPERTY:                Color( 50,  58,  61),
        NK_COLOR_EDIT:                    Color( 50,  58,  61),
        NK_COLOR_EDIT_CURSOR:             Color(210, 210, 210),
        NK_COLOR_COMBO:                   Color( 50,  58,  61),
        NK_COLOR_CHART:                   Color( 50,  58,  61),
        NK_COLOR_CHART_COLOR:             Color( 48,  83, 111),
        NK_COLOR_CHART_COLOR_HIGHLIGHT:   Color(255,   0,   0),
        NK_COLOR_SCROLLBAR:               Color( 40,  50,  55),
        NK_COLOR_SCROLLBAR_CURSOR:        Color( 48,  83, 111),
        NK_COLOR_SCROLLBAR_CURSOR_HOVER:  Color( 53,  88, 116),
        NK_COLOR_SCROLLBAR_CURSOR_ACTIVE: Color( 58,  93, 121),
        NK_COLOR_TAB_HEADER:              Color( 48,  83, 111),
    }

def theme_light():
    """Return a light color table dict."""
    return {
        NK_COLOR_TEXT:                    Color( 20,  20,  20),
        NK_COLOR_WINDOW:                  Color(245, 245, 245),
        NK_COLOR_HEADER:                  Color(210, 210, 210, 220),
        NK_COLOR_BORDER:                  Color(175, 175, 175),
        NK_COLOR_BUTTON:                  Color(185, 185, 185),
        NK_COLOR_BUTTON_HOVER:            Color(170, 170, 170),
        NK_COLOR_BUTTON_ACTIVE:           Color(160, 160, 160),
        NK_COLOR_TOGGLE:                  Color(150, 150, 150),
        NK_COLOR_TOGGLE_HOVER:            Color(120, 120, 120),
        NK_COLOR_TOGGLE_CURSOR:           Color( 45, 100, 180),
        NK_COLOR_SELECT:                  Color(190, 190, 190),
        NK_COLOR_SELECT_ACTIVE:           Color( 45, 100, 180),
        NK_COLOR_SLIDER:                  Color(190, 190, 190),
        NK_COLOR_SLIDER_CURSOR:           Color( 45, 100, 180),
        NK_COLOR_SLIDER_CURSOR_HOVER:     Color( 55, 110, 190),
        NK_COLOR_SLIDER_CURSOR_ACTIVE:    Color( 35,  90, 170),
        NK_COLOR_PROPERTY:                Color(190, 190, 190),
        NK_COLOR_EDIT:                    Color(240, 240, 240),
        NK_COLOR_EDIT_CURSOR:             Color( 20,  20,  20),
        NK_COLOR_COMBO:                   Color(190, 190, 190),
        NK_COLOR_CHART:                   Color(190, 190, 190),
        NK_COLOR_CHART_COLOR:             Color( 45, 100, 180),
        NK_COLOR_CHART_COLOR_HIGHLIGHT:   Color(255,   0,   0),
        NK_COLOR_SCROLLBAR:               Color(205, 205, 205),
        NK_COLOR_SCROLLBAR_CURSOR:        Color( 45, 100, 180),
        NK_COLOR_SCROLLBAR_CURSOR_HOVER:  Color( 55, 110, 190),
        NK_COLOR_SCROLLBAR_CURSOR_ACTIVE: Color( 35,  90, 170),
        NK_COLOR_TAB_HEADER:              Color( 45, 100, 180),
    }

def theme_nord():
    """Nord palette."""
    return {
        NK_COLOR_TEXT:                    Color(236, 239, 244),
        NK_COLOR_WINDOW:                  Color( 46,  52,  64),
        NK_COLOR_HEADER:                  Color( 59,  66,  82, 230),
        NK_COLOR_BORDER:                  Color( 76,  86, 106),
        NK_COLOR_BUTTON:                  Color( 94, 129, 172),
        NK_COLOR_BUTTON_HOVER:            Color(129, 161, 193),
        NK_COLOR_BUTTON_ACTIVE:           Color( 67,  76,  94),
        NK_COLOR_TOGGLE:                  Color( 67,  76,  94),
        NK_COLOR_TOGGLE_HOVER:            Color( 76,  86, 106),
        NK_COLOR_TOGGLE_CURSOR:           Color( 94, 129, 172),
        NK_COLOR_SELECT:                  Color( 67,  76,  94),
        NK_COLOR_SELECT_ACTIVE:           Color( 94, 129, 172),
        NK_COLOR_SLIDER:                  Color( 67,  76,  94),
        NK_COLOR_SLIDER_CURSOR:           Color( 94, 129, 172),
        NK_COLOR_SLIDER_CURSOR_HOVER:     Color(129, 161, 193),
        NK_COLOR_SLIDER_CURSOR_ACTIVE:    Color( 67,  76,  94),
        NK_COLOR_PROPERTY:                Color( 67,  76,  94),
        NK_COLOR_EDIT:                    Color( 59,  66,  82),
        NK_COLOR_EDIT_CURSOR:             Color(236, 239, 244),
        NK_COLOR_COMBO:                   Color( 67,  76,  94),
        NK_COLOR_CHART:                   Color( 67,  76,  94),
        NK_COLOR_CHART_COLOR:             Color( 94, 129, 172),
        NK_COLOR_CHART_COLOR_HIGHLIGHT:   Color(191,  97, 106),
        NK_COLOR_SCROLLBAR:               Color( 59,  66,  82),
        NK_COLOR_SCROLLBAR_CURSOR:        Color( 94, 129, 172),
        NK_COLOR_SCROLLBAR_CURSOR_HOVER:  Color(129, 161, 193),
        NK_COLOR_SCROLLBAR_CURSOR_ACTIVE: Color( 67,  76,  94),
        NK_COLOR_TAB_HEADER:              Color( 94, 129, 172),
    }

# ==============================================================================
#  Full SDL2 + Nuklear App helper
# ==============================================================================

class NuklearApp:
    """
    One-call setup for a Nuklear + SDL2 application.

    Usage:
        app = NuklearApp("My App", 1280, 720, theme=theme_dark())
        app.start()

        ev   = app.sdl.Event()
        running = True
        while running:
            app.frame_begin(ev)
            while ev.poll():
                if ev.type == app.sdl.SDL_QUIT:
                    running = False
                app.feed(ev)
            app.frame_input_end()

            if app.ctx.begin("Demo", 50, 50, 300, 200,
                             NK_WINDOW_BORDER | NK_WINDOW_TITLE | NK_WINDOW_MOVABLE):
                app.ctx.layout_row_dynamic(30, 1)
                if app.ctx.button_label("Quit"):
                    running = False
                app.ctx.label("Hello Nuklear!", NK_TEXT_LEFT)
            app.ctx.end()

            app.frame_end()

        app.destroy()
    """
    def __init__(self, title, width, height, theme=None,font_size=14.0, vsync=True):
        import sdl as _sdl
        self.sdl = _sdl

        _sdl.init(_sdl.SDL_INIT_EVERYTHING)

        self.window   = _sdl.Window(title, width, height)
        self.renderer = _sdl.Renderer(
            self.window,
            flags=_sdl.SDL_RENDERER_ACCELERATED |
                  (_sdl.SDL_RENDERER_PRESENTVSYNC if vsync else 0))

        self.atlas = FontAtlas()
        self.atlas.begin()
        self._default_font = self.atlas.add_default(font_size)
        raw_pixels, fw, fh = self.atlas.bake()

        # Upload font atlas to SDL2 texture
        self._font_tex = _sdl.Texture(
            self.renderer, _sdl.SDL_PIXELFORMAT_RGBA8888,
            _sdl.SDL_TEXTUREACCESS_STATIC, fw, fh)
        self._font_tex.set_blend_mode(_sdl.SDL_BLENDMODE_BLEND)
        self._font_tex.update(None, raw_pixels, fw * 4)

        self.atlas.end(self._font_tex.ptr.Address)

        self.ctx = Context()
        self.ctx.init(self._default_font.user_font_ptr)

        if theme:
            self.ctx.style_set_colors(theme)

        self.nk_input    = NuklearSDLInput(self.ctx)
        self.nk_renderer = NuklearSDLRenderer(self.renderer)

    def start(self):
        """No-op placeholder — setup is done in __init__."""
        pass

    def frame_begin(self, ev=None):
        """Call at the start of each frame before polling events."""
        self.nk_input.begin()

    def feed(self, event):
        """Feed one SDL2 event to Nuklear."""
        self.nk_input.feed(event)

    def frame_input_end(self):
        """Call after all events have been fed."""
        self.nk_input.end()

    def frame_end(self):
        """Render Nuklear output and present the SDL2 frame."""
        self.renderer.set_draw_color(30, 30, 30, 255)
        self.renderer.clear()
        self.nk_renderer.render(self.ctx)
        self.renderer.present()
        self.ctx.clear()

    def destroy(self):
        self.ctx.free()
        self.atlas.clear()
        self._font_tex.destroy()
        self.renderer.destroy()
        self.window.destroy()
        self.sdl.quit()