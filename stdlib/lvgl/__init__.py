"""
Full LVGL v9 wrapper for Pylearn/Swalang.
Covers: Core, Display, Input, Style, Animation, Timer, Layout,
        and all major widgets.
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
        "bin/x86_64-linux/lvgl/liblvgl.so",
        "liblvgl.so",
    ])
elif sys.platform == "windows":
    _lib = _try_load([
        "bin/x86_64-windows-gnu/lvgl/liblvgl.dll",
        "liblvgl.dll",
    ])
else:
    _lib = _try_load(["liblvgl.dylib"])

if _lib is None:
    raise ffi.FFIError("Could not load LVGL shared library.")

# ==============================================================================
#  Constants
# ==============================================================================

# --- Alignment ---
LV_ALIGN_DEFAULT       = 0
LV_ALIGN_TOP_LEFT      = 1
LV_ALIGN_TOP_MID       = 2
LV_ALIGN_TOP_RIGHT     = 3
LV_ALIGN_BOTTOM_LEFT   = 4
LV_ALIGN_BOTTOM_MID    = 5
LV_ALIGN_BOTTOM_RIGHT  = 6
LV_ALIGN_LEFT_MID      = 7
LV_ALIGN_RIGHT_MID     = 8
LV_ALIGN_CENTER        = 9
LV_ALIGN_OUT_TOP_LEFT      = 10
LV_ALIGN_OUT_TOP_MID       = 11
LV_ALIGN_OUT_TOP_RIGHT     = 12
LV_ALIGN_OUT_BOTTOM_LEFT   = 13
LV_ALIGN_OUT_BOTTOM_MID    = 14
LV_ALIGN_OUT_BOTTOM_RIGHT  = 15
LV_ALIGN_OUT_LEFT_TOP      = 16
LV_ALIGN_OUT_LEFT_MID      = 17
LV_ALIGN_OUT_LEFT_BOTTOM   = 18
LV_ALIGN_OUT_RIGHT_TOP     = 19
LV_ALIGN_OUT_RIGHT_MID     = 20
LV_ALIGN_OUT_RIGHT_BOTTOM  = 21

# --- Object parts ---
LV_PART_MAIN        = 0x000000
LV_PART_SCROLLBAR   = 0x010000
LV_PART_INDICATOR   = 0x020000
LV_PART_KNOB        = 0x030000
LV_PART_SELECTED    = 0x040000
LV_PART_ITEMS       = 0x050000
LV_PART_CURSOR      = 0x080000
LV_PART_ANY         = 0x0F0000

# --- Object states ---
LV_STATE_DEFAULT  = 0x0000
LV_STATE_CHECKED  = 0x0001
LV_STATE_FOCUSED  = 0x0002
LV_STATE_FOCUS_KEY= 0x0004
LV_STATE_EDITED   = 0x0008
LV_STATE_HOVERED  = 0x0010
LV_STATE_PRESSED  = 0x0020
LV_STATE_SCROLLED = 0x0040
LV_STATE_DISABLED = 0x0080
LV_STATE_ANY      = 0xFFFF

# --- Event codes ---
LV_EVENT_ALL              = 0
LV_EVENT_PRESSED          = 1
LV_EVENT_PRESSING         = 2
LV_EVENT_PRESS_LOST       = 3
LV_EVENT_SHORT_CLICKED    = 4
LV_EVENT_LONG_PRESSED     = 5
LV_EVENT_LONG_PRESSED_REPEAT = 6
LV_EVENT_CLICKED          = 7
LV_EVENT_RELEASED         = 8
LV_EVENT_SCROLL_BEGIN     = 9
LV_EVENT_SCROLL_END       = 10
LV_EVENT_SCROLL           = 11
LV_EVENT_GESTURE          = 12
LV_EVENT_KEY              = 13
LV_EVENT_FOCUSED          = 14
LV_EVENT_DEFOCUSED        = 15
LV_EVENT_LEAVE            = 16
LV_EVENT_HIT_TEST         = 17
LV_EVENT_COVER_CHECK      = 18
LV_EVENT_REFR_EXT_DRAW_SIZE = 19
LV_EVENT_DRAW_MAIN_BEGIN  = 20
LV_EVENT_DRAW_MAIN        = 21
LV_EVENT_DRAW_MAIN_END    = 22
LV_EVENT_DRAW_POST_BEGIN  = 23
LV_EVENT_DRAW_POST        = 24
LV_EVENT_DRAW_POST_END    = 25
LV_EVENT_VALUE_CHANGED    = 28
LV_EVENT_INSERT           = 29
LV_EVENT_REFRESH          = 30
LV_EVENT_READY            = 31
LV_EVENT_CANCEL           = 32
LV_EVENT_DELETE           = 33
LV_EVENT_CHILD_CHANGED    = 34
LV_EVENT_CHILD_CREATED    = 35
LV_EVENT_CHILD_DELETED    = 36
LV_EVENT_SCREEN_UNLOAD_START = 37
LV_EVENT_SCREEN_LOAD_START   = 38
LV_EVENT_SCREEN_LOADED       = 39
LV_EVENT_SCREEN_UNLOADED     = 40
LV_EVENT_SIZE_CHANGED        = 41
LV_EVENT_STYLE_CHANGED       = 42
LV_EVENT_LAYOUT_CHANGED      = 43
LV_EVENT_GET_SELF_SIZE       = 44

# --- Input device types ---
LV_INDEV_TYPE_NONE    = 0
LV_INDEV_TYPE_POINTER = 1
LV_INDEV_TYPE_KEYPAD  = 2
LV_INDEV_TYPE_BUTTON  = 3
LV_INDEV_TYPE_ENCODER = 4

LV_INDEV_STATE_RELEASED = 0
LV_INDEV_STATE_PRESSED  = 1

# --- Display render modes ---
LV_DISPLAY_RENDER_MODE_PARTIAL = 0
LV_DISPLAY_RENDER_MODE_DIRECT  = 1
LV_DISPLAY_RENDER_MODE_FULL    = 2

# --- Animation ---
LV_ANIM_OFF = 0
LV_ANIM_ON  = 1
LV_ANIM_REPEAT_INFINITE = 0xFFFF

# --- Label long modes ---
LV_LABEL_LONG_WRAP            = 0
LV_LABEL_LONG_DOT             = 1
LV_LABEL_LONG_SCROLL          = 2
LV_LABEL_LONG_SCROLL_CIRCULAR = 3
LV_LABEL_LONG_CLIP            = 4

# --- Arc modes ---
LV_ARC_MODE_NORMAL     = 0
LV_ARC_MODE_SYMMETRICAL = 1
LV_ARC_MODE_REVERSE    = 2

# --- Bar modes ---
LV_BAR_MODE_NORMAL    = 0
LV_BAR_MODE_SYMMETRICAL = 1
LV_BAR_MODE_RANGE     = 2

# --- Slider modes ---
LV_SLIDER_MODE_NORMAL    = 0
LV_SLIDER_MODE_SYMMETRICAL = 1
LV_SLIDER_MODE_RANGE     = 2

# --- Roller modes ---
LV_ROLLER_MODE_NORMAL   = 0
LV_ROLLER_MODE_INFINITE = 1

# --- Dropdown direction ---
LV_DIR_NONE   = 0x00
LV_DIR_LEFT   = 0x01
LV_DIR_RIGHT  = 0x02
LV_DIR_TOP    = 0x04
LV_DIR_BOTTOM = 0x08
LV_DIR_HOR    = 0x03
LV_DIR_VER    = 0x0C
LV_DIR_ALL    = 0x0F

# --- Keyboard modes ---
LV_KEYBOARD_MODE_TEXT_LOWER   = 0
LV_KEYBOARD_MODE_TEXT_UPPER   = 1
LV_KEYBOARD_MODE_SPECIAL      = 2
LV_KEYBOARD_MODE_NUMBER       = 3
LV_KEYBOARD_MODE_USER_1       = 4
LV_KEYBOARD_MODE_USER_2       = 5
LV_KEYBOARD_MODE_USER_3       = 6
LV_KEYBOARD_MODE_USER_4       = 7

# --- Flex layout ---
LV_FLEX_FLOW_ROW             = 0
LV_FLEX_FLOW_COLUMN          = 1
LV_FLEX_FLOW_ROW_WRAP        = 2
LV_FLEX_FLOW_ROW_REVERSE     = 3
LV_FLEX_FLOW_ROW_WRAP_REVERSE = 4
LV_FLEX_FLOW_COLUMN_WRAP     = 5
LV_FLEX_FLOW_COLUMN_REVERSE  = 6
LV_FLEX_FLOW_COLUMN_WRAP_REVERSE = 7

LV_FLEX_ALIGN_START         = 0
LV_FLEX_ALIGN_END           = 1
LV_FLEX_ALIGN_CENTER        = 2
LV_FLEX_ALIGN_SPACE_EVENLY  = 3
LV_FLEX_ALIGN_SPACE_AROUND  = 4
LV_FLEX_ALIGN_SPACE_BETWEEN = 5

# --- Grid ---
LV_GRID_ALIGN_START   = 0
LV_GRID_ALIGN_CENTER  = 1
LV_GRID_ALIGN_END     = 2
LV_GRID_ALIGN_STRETCH = 3
LV_GRID_ALIGN_SPACE_EVENLY  = 4
LV_GRID_ALIGN_SPACE_AROUND  = 5
LV_GRID_ALIGN_SPACE_BETWEEN = 6
LV_GRID_FR            = 0x8000  # fractional unit flag

# --- Scroll ---
LV_SCROLL_SNAP_NONE   = 0
LV_SCROLL_SNAP_START  = 1
LV_SCROLL_SNAP_END    = 2
LV_SCROLL_SNAP_CENTER = 3

LV_SCROLLBAR_MODE_OFF    = 0
LV_SCROLLBAR_MODE_ON     = 1
LV_SCROLLBAR_MODE_ACTIVE = 2
LV_SCROLLBAR_MODE_AUTO   = 3

# --- Object flags ---
LV_OBJ_FLAG_HIDDEN            = 0x0001
LV_OBJ_FLAG_CLICKABLE         = 0x0002
LV_OBJ_FLAG_CLICK_FOCUSABLE   = 0x0004
LV_OBJ_FLAG_CHECKABLE         = 0x0008
LV_OBJ_FLAG_SCROLLABLE        = 0x0010
LV_OBJ_FLAG_SCROLL_ELASTIC    = 0x0020
LV_OBJ_FLAG_SCROLL_MOMENTUM   = 0x0040
LV_OBJ_FLAG_SCROLL_ONE        = 0x0080
LV_OBJ_FLAG_SCROLL_CHAIN_HOR  = 0x0100
LV_OBJ_FLAG_SCROLL_CHAIN_VER  = 0x0200
LV_OBJ_FLAG_SCROLL_ON_FOCUS   = 0x0400
LV_OBJ_FLAG_SCROLL_WITH_ARROW = 0x0800
LV_OBJ_FLAG_SNAPPABLE         = 0x1000
LV_OBJ_FLAG_PRESS_LOCK        = 0x2000
LV_OBJ_FLAG_EVENT_BUBBLE      = 0x4000
LV_OBJ_FLAG_GESTURE_BUBBLE    = 0x8000
LV_OBJ_FLAG_ADV_HITTEST       = 0x10000
LV_OBJ_FLAG_IGNORE_LAYOUT     = 0x20000
LV_OBJ_FLAG_FLOATING          = 0x40000
LV_OBJ_FLAG_OVERFLOW_VISIBLE  = 0x200000
LV_OBJ_FLAG_FLEX_IN_NEW_TRACK = 0x400000
LV_OBJ_FLAG_USER_1            = 0x1000000
LV_OBJ_FLAG_USER_2            = 0x2000000
LV_OBJ_FLAG_USER_3            = 0x4000000
LV_OBJ_FLAG_USER_4            = 0x8000000

# --- Border sides ---
LV_BORDER_SIDE_NONE   = 0x00
LV_BORDER_SIDE_BOTTOM = 0x01
LV_BORDER_SIDE_TOP    = 0x02
LV_BORDER_SIDE_LEFT   = 0x04
LV_BORDER_SIDE_RIGHT  = 0x08
LV_BORDER_SIDE_FULL   = 0x0F
LV_BORDER_SIDE_INTERNAL = 0x10

# --- Text decorations ---
LV_TEXT_DECOR_NONE         = 0x00
LV_TEXT_DECOR_UNDERLINE    = 0x01
LV_TEXT_DECOR_STRIKETHROUGH = 0x02

# --- Text alignment ---
LV_TEXT_ALIGN_AUTO   = 0
LV_TEXT_ALIGN_LEFT   = 1
LV_TEXT_ALIGN_CENTER = 2
LV_TEXT_ALIGN_RIGHT  = 3

# --- Gradient directions ---
LV_GRAD_DIR_NONE = 0
LV_GRAD_DIR_VER  = 1
LV_GRAD_DIR_HOR  = 2

# --- Blend modes ---
LV_BLEND_MODE_NORMAL    = 0
LV_BLEND_MODE_ADDITIVE  = 1
LV_BLEND_MODE_SUBTRACTIVE = 2
LV_BLEND_MODE_MULTIPLY  = 3

# --- Special sizes ---
LV_SIZE_CONTENT = 0x1FFFE  # size to wrap content
LV_PCT_0   = 0x80000000    # 0%  — see lv_pct() below
LV_PCT_100 = 0x80000064    # 100%

# ==============================================================================
#  Color utilities  (pure Python — no FFI call needed for 32-bit builds)
# ==============================================================================

def color_make(r, g, b):
    """
    Build an lv_color_t as a uint32 for 32-bit color depth (ARGB8888).
    Layout in memory: [blue, green, red, alpha] (little-endian struct).
    """
    return (0xFF << 24) | (r << 16) | (g << 8) | b

def color_hex(h):
    """Build lv_color_t from 0xRRGGBB hex integer."""
    r = (h >> 16) & 0xFF
    g = (h >> 8)  & 0xFF
    b =  h        & 0xFF
    return color_make(r, g, b)

def color_hex_str(s):
    """Build lv_color_t from a hex string like '#FF8800' or 'FF8800'."""
    s = s.lstrip("#")
    return color_hex(int(s, 16))

def lv_pct(v):
    """Convert an integer percentage 0–100 to LVGL's packed percentage value."""
    return 0x80000000 | (v & 0x7FFFFFFF)

# Common colours
COLOR_WHITE   = color_make(0xFF, 0xFF, 0xFF)
COLOR_BLACK   = color_make(0x00, 0x00, 0x00)
COLOR_RED     = color_make(0xFF, 0x00, 0x00)
COLOR_GREEN   = color_make(0x00, 0xFF, 0x00)
COLOR_BLUE    = color_make(0x00, 0x00, 0xFF)
COLOR_YELLOW  = color_make(0xFF, 0xFF, 0x00)
COLOR_CYAN    = color_make(0x00, 0xFF, 0xFF)
COLOR_MAGENTA = color_make(0xFF, 0x00, 0xFF)
COLOR_GREY    = color_make(0x80, 0x80, 0x80)
COLOR_SILVER  = color_make(0xC0, 0xC0, 0xC0)
COLOR_ORANGE  = color_make(0xFF, 0xA5, 0x00)
COLOR_TRANSPARENT = 0x00000000

# Opaque context sizes
_SZ_STYLE  = 128
_SZ_ANIM   = 256
_SZ_POINT  = 16   # lv_point_t = int32 x, int32 y

# ==============================================================================
#  C Function Bindings
# ==============================================================================

# --- Core ---
_lv_init           = _lib.lv_init([], None)
_lv_deinit         = _lib.lv_deinit([], None)
_lv_tick_inc       = _lib.lv_tick_inc([ffi.c_uint32], None)
_lv_tick_get       = _lib.lv_tick_get([], ffi.c_uint32)
_lv_timer_handler  = _lib.lv_timer_handler([], ffi.c_uint32)
_lv_refr_now       = _lib.lv_refr_now([ffi.c_void_p], None)

# --- Display ---
_lv_display_create       = _lib.lv_display_create([ffi.c_int32, ffi.c_int32], ffi.c_void_p)
_lv_display_delete       = _lib.lv_display_delete([ffi.c_void_p], None)
_lv_display_set_buffers  = _lib.lv_display_set_buffers(
    [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_uint32, ffi.c_int32], None)
_lv_display_set_flush_cb = _lib.lv_display_set_flush_cb([ffi.c_void_p, ffi.c_void_p], None)
_lv_display_flush_ready  = _lib.lv_display_flush_ready([ffi.c_void_p], None)
_lv_display_set_resolution = _lib.lv_display_set_resolution(
    [ffi.c_void_p, ffi.c_int32, ffi.c_int32], None)
_lv_display_get_horizontal_resolution = _lib.lv_display_get_horizontal_resolution(
    [ffi.c_void_p], ffi.c_int32)
_lv_display_get_vertical_resolution   = _lib.lv_display_get_vertical_resolution(
    [ffi.c_void_p], ffi.c_int32)
_lv_display_set_rotation = _lib.lv_display_set_rotation([ffi.c_void_p, ffi.c_int32], None)
_lv_display_get_default  = _lib.lv_display_get_default([], ffi.c_void_p)
_lv_display_set_default  = _lib.lv_display_set_default([ffi.c_void_p], None)
_lv_display_get_screen_active = _lib.lv_display_get_screen_active([ffi.c_void_p], ffi.c_void_p)

# --- Input device ---
_lv_indev_create      = _lib.lv_indev_create([], ffi.c_void_p)
_lv_indev_delete      = _lib.lv_indev_delete([ffi.c_void_p], None)
_lv_indev_set_type    = _lib.lv_indev_set_type([ffi.c_void_p, ffi.c_int32], None)
_lv_indev_set_read_cb = _lib.lv_indev_set_read_cb([ffi.c_void_p, ffi.c_void_p], None)
_lv_indev_set_display = _lib.lv_indev_set_display([ffi.c_void_p, ffi.c_void_p], None)
_lv_indev_reset       = _lib.lv_indev_reset([ffi.c_void_p, ffi.c_void_p], None)
_lv_indev_get_active  = _lib.lv_indev_get_active([], ffi.c_void_p)

# --- Timer ---
_lv_timer_create       = _lib.lv_timer_create(
    [ffi.c_void_p, ffi.c_uint32, ffi.c_void_p], ffi.c_void_p)
_lv_timer_delete       = _lib.lv_timer_delete([ffi.c_void_p], None)
_lv_timer_set_period   = _lib.lv_timer_set_period([ffi.c_void_p, ffi.c_uint32], None)
_lv_timer_reset        = _lib.lv_timer_reset([ffi.c_void_p], None)
_lv_timer_pause        = _lib.lv_timer_pause([ffi.c_void_p], None)
_lv_timer_resume       = _lib.lv_timer_resume([ffi.c_void_p], None)
_lv_timer_set_repeat_count = _lib.lv_timer_set_repeat_count(
    [ffi.c_void_p, ffi.c_int32], None)
_lv_timer_get_user_data = _lib.lv_timer_get_user_data([ffi.c_void_p], ffi.c_void_p)

# --- Style ---
_lv_style_init    = _lib.lv_style_init([ffi.c_void_p], None)
_lv_style_reset   = _lib.lv_style_reset([ffi.c_void_p], None)

# Background
_lv_style_set_bg_color      = _lib.lv_style_set_bg_color(
    [ffi.c_void_p, ffi.c_uint32], None)
_lv_style_set_bg_opa        = _lib.lv_style_set_bg_opa(
    [ffi.c_void_p, ffi.c_uint8], None)
_lv_style_set_bg_grad_color = _lib.lv_style_set_bg_grad_color(
    [ffi.c_void_p, ffi.c_uint32], None)
_lv_style_set_bg_grad_dir   = _lib.lv_style_set_bg_grad_dir(
    [ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_bg_main_stop  = _lib.lv_style_set_bg_main_stop(
    [ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_bg_grad_stop  = _lib.lv_style_set_bg_grad_stop(
    [ffi.c_void_p, ffi.c_int32], None)

# Border
_lv_style_set_border_color  = _lib.lv_style_set_border_color(
    [ffi.c_void_p, ffi.c_uint32], None)
_lv_style_set_border_width  = _lib.lv_style_set_border_width(
    [ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_border_opa    = _lib.lv_style_set_border_opa(
    [ffi.c_void_p, ffi.c_uint8], None)
_lv_style_set_border_side   = _lib.lv_style_set_border_side(
    [ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_border_post   = _lib.lv_style_set_border_post(
    [ffi.c_void_p, ffi.c_int32], None)

# Outline
_lv_style_set_outline_color = _lib.lv_style_set_outline_color(
    [ffi.c_void_p, ffi.c_uint32], None)
_lv_style_set_outline_width = _lib.lv_style_set_outline_width(
    [ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_outline_pad   = _lib.lv_style_set_outline_pad(
    [ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_outline_opa   = _lib.lv_style_set_outline_opa(
    [ffi.c_void_p, ffi.c_uint8], None)

# Shadow
_lv_style_set_shadow_color  = _lib.lv_style_set_shadow_color(
    [ffi.c_void_p, ffi.c_uint32], None)
_lv_style_set_shadow_width  = _lib.lv_style_set_shadow_width(
    [ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_shadow_opa    = _lib.lv_style_set_shadow_opa(
    [ffi.c_void_p, ffi.c_uint8], None)
_lv_style_set_shadow_offset_x = _lib.lv_style_set_shadow_offset_x(
    [ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_shadow_offset_y = _lib.lv_style_set_shadow_offset_y(
    [ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_shadow_spread = _lib.lv_style_set_shadow_spread(
    [ffi.c_void_p, ffi.c_int32], None)

# Text
_lv_style_set_text_color    = _lib.lv_style_set_text_color(
    [ffi.c_void_p, ffi.c_uint32], None)
_lv_style_set_text_opa      = _lib.lv_style_set_text_opa(
    [ffi.c_void_p, ffi.c_uint8], None)
_lv_style_set_text_font     = _lib.lv_style_set_text_font(
    [ffi.c_void_p, ffi.c_void_p], None)
_lv_style_set_text_letter_space = _lib.lv_style_set_text_letter_space(
    [ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_text_line_space   = _lib.lv_style_set_text_line_space(
    [ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_text_decor    = _lib.lv_style_set_text_decor(
    [ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_text_align    = _lib.lv_style_set_text_align(
    [ffi.c_void_p, ffi.c_int32], None)

# Padding
_lv_style_set_pad_top    = _lib.lv_style_set_pad_top(   [ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_pad_bottom = _lib.lv_style_set_pad_bottom([ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_pad_left   = _lib.lv_style_set_pad_left(  [ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_pad_right  = _lib.lv_style_set_pad_right( [ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_pad_row    = _lib.lv_style_set_pad_row(   [ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_pad_column = _lib.lv_style_set_pad_column([ffi.c_void_p, ffi.c_int32], None)

# Margin
_lv_style_set_margin_top    = _lib.lv_style_set_margin_top(   [ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_margin_bottom = _lib.lv_style_set_margin_bottom([ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_margin_left   = _lib.lv_style_set_margin_left(  [ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_margin_right  = _lib.lv_style_set_margin_right( [ffi.c_void_p, ffi.c_int32], None)

# Layout / size
_lv_style_set_width        = _lib.lv_style_set_width(       [ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_min_width    = _lib.lv_style_set_min_width(   [ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_max_width    = _lib.lv_style_set_max_width(   [ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_height       = _lib.lv_style_set_height(      [ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_min_height   = _lib.lv_style_set_min_height(  [ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_max_height   = _lib.lv_style_set_max_height(  [ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_radius       = _lib.lv_style_set_radius(      [ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_clip_corner  = _lib.lv_style_set_clip_corner( [ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_opa          = _lib.lv_style_set_opa(         [ffi.c_void_p, ffi.c_uint8], None)
_lv_style_set_color_filter_opa = _lib.lv_style_set_color_filter_opa(
    [ffi.c_void_p, ffi.c_uint8], None)
_lv_style_set_blend_mode   = _lib.lv_style_set_blend_mode(  [ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_transform_width  = _lib.lv_style_set_transform_width( [ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_transform_height = _lib.lv_style_set_transform_height([ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_translate_x  = _lib.lv_style_set_translate_x( [ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_translate_y  = _lib.lv_style_set_translate_y( [ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_transform_scale_x = _lib.lv_style_set_transform_scale_x([ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_transform_scale_y = _lib.lv_style_set_transform_scale_y([ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_transform_rotation = _lib.lv_style_set_transform_rotation([ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_transform_pivot_x  = _lib.lv_style_set_transform_pivot_x( [ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_transform_pivot_y  = _lib.lv_style_set_transform_pivot_y( [ffi.c_void_p, ffi.c_int32], None)

# Line
_lv_style_set_line_color  = _lib.lv_style_set_line_color( [ffi.c_void_p, ffi.c_uint32], None)
_lv_style_set_line_width  = _lib.lv_style_set_line_width( [ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_line_dash_width = _lib.lv_style_set_line_dash_width([ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_line_dash_gap   = _lib.lv_style_set_line_dash_gap(  [ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_line_rounded    = _lib.lv_style_set_line_rounded(    [ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_line_opa        = _lib.lv_style_set_line_opa(        [ffi.c_void_p, ffi.c_uint8], None)

# Arc
_lv_style_set_arc_color  = _lib.lv_style_set_arc_color( [ffi.c_void_p, ffi.c_uint32], None)
_lv_style_set_arc_width  = _lib.lv_style_set_arc_width( [ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_arc_rounded = _lib.lv_style_set_arc_rounded([ffi.c_void_p, ffi.c_int32], None)
_lv_style_set_arc_opa    = _lib.lv_style_set_arc_opa(   [ffi.c_void_p, ffi.c_uint8], None)

# Image
_lv_style_set_image_opa        = _lib.lv_style_set_image_opa(       [ffi.c_void_p, ffi.c_uint8], None)
_lv_style_set_image_recolor    = _lib.lv_style_set_image_recolor(   [ffi.c_void_p, ffi.c_uint32], None)
_lv_style_set_image_recolor_opa = _lib.lv_style_set_image_recolor_opa([ffi.c_void_p, ffi.c_uint8], None)

# Layout
_lv_style_set_layout         = _lib.lv_style_set_layout([ffi.c_void_p, ffi.c_int32], None)

# Scrollbar
_lv_style_set_scrollbar_mode = _lib.lv_style_set_scrollbar_mode(
    [ffi.c_void_p, ffi.c_int32], None)

# --- Object base ---
_lv_obj_create         = _lib.lv_obj_create([ffi.c_void_p], ffi.c_void_p)
_lv_obj_delete         = _lib.lv_obj_delete([ffi.c_void_p], None)
_lv_obj_clean          = _lib.lv_obj_clean([ffi.c_void_p], None)
_lv_obj_set_pos        = _lib.lv_obj_set_pos(   [ffi.c_void_p, ffi.c_int32, ffi.c_int32], None)
_lv_obj_set_x          = _lib.lv_obj_set_x(     [ffi.c_void_p, ffi.c_int32], None)
_lv_obj_set_y          = _lib.lv_obj_set_y(     [ffi.c_void_p, ffi.c_int32], None)
_lv_obj_set_size       = _lib.lv_obj_set_size(  [ffi.c_void_p, ffi.c_int32, ffi.c_int32], None)
_lv_obj_set_width      = _lib.lv_obj_set_width( [ffi.c_void_p, ffi.c_int32], None)
_lv_obj_set_height     = _lib.lv_obj_set_height([ffi.c_void_p, ffi.c_int32], None)
_lv_obj_set_align      = _lib.lv_obj_set_align( [ffi.c_void_p, ffi.c_int32], None)
_lv_obj_align          = _lib.lv_obj_align(
    [ffi.c_void_p, ffi.c_int32, ffi.c_int32, ffi.c_int32], None)
_lv_obj_align_to       = _lib.lv_obj_align_to(
    [ffi.c_void_p, ffi.c_void_p, ffi.c_int32, ffi.c_int32, ffi.c_int32], None)
_lv_obj_center         = _lib.lv_obj_center([ffi.c_void_p], None)
_lv_obj_get_x          = _lib.lv_obj_get_x(     [ffi.c_void_p], ffi.c_int32)
_lv_obj_get_y          = _lib.lv_obj_get_y(     [ffi.c_void_p], ffi.c_int32)
_lv_obj_get_width      = _lib.lv_obj_get_width( [ffi.c_void_p], ffi.c_int32)
_lv_obj_get_height     = _lib.lv_obj_get_height([ffi.c_void_p], ffi.c_int32)
_lv_obj_get_content_width  = _lib.lv_obj_get_content_width( [ffi.c_void_p], ffi.c_int32)
_lv_obj_get_content_height = _lib.lv_obj_get_content_height([ffi.c_void_p], ffi.c_int32)
_lv_obj_get_parent     = _lib.lv_obj_get_parent([ffi.c_void_p], ffi.c_void_p)
_lv_obj_get_child      = _lib.lv_obj_get_child( [ffi.c_void_p, ffi.c_int32], ffi.c_void_p)
_lv_obj_get_child_count = _lib.lv_obj_get_child_count([ffi.c_void_p], ffi.c_uint32)
_lv_obj_set_parent     = _lib.lv_obj_set_parent([ffi.c_void_p, ffi.c_void_p], None)
_lv_obj_add_style      = _lib.lv_obj_add_style(
    [ffi.c_void_p, ffi.c_void_p, ffi.c_int32], None)
_lv_obj_remove_style   = _lib.lv_obj_remove_style(
    [ffi.c_void_p, ffi.c_void_p, ffi.c_int32], None)
_lv_obj_remove_style_all = _lib.lv_obj_remove_style_all([ffi.c_void_p], None)
_lv_obj_add_event_cb   = _lib.lv_obj_add_event_cb(
    [ffi.c_void_p, ffi.c_void_p, ffi.c_int32, ffi.c_void_p], ffi.c_void_p)
_lv_obj_remove_event_cb = _lib.lv_obj_remove_event_cb(
    [ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_lv_obj_send_event     = _lib.lv_obj_send_event(
    [ffi.c_void_p, ffi.c_int32, ffi.c_void_p], ffi.c_int32)
_lv_obj_add_flag       = _lib.lv_obj_add_flag(   [ffi.c_void_p, ffi.c_uint32], None)
_lv_obj_remove_flag    = _lib.lv_obj_remove_flag([ffi.c_void_p, ffi.c_uint32], None)
_lv_obj_has_flag       = _lib.lv_obj_has_flag(   [ffi.c_void_p, ffi.c_uint32], ffi.c_int32)
_lv_obj_add_state      = _lib.lv_obj_add_state(   [ffi.c_void_p, ffi.c_int32], None)
_lv_obj_remove_state   = _lib.lv_obj_remove_state([ffi.c_void_p, ffi.c_int32], None)
_lv_obj_has_state      = _lib.lv_obj_has_state(   [ffi.c_void_p, ffi.c_int32], ffi.c_int32)
_lv_obj_set_user_data  = _lib.lv_obj_set_user_data([ffi.c_void_p, ffi.c_void_p], None)
_lv_obj_get_user_data  = _lib.lv_obj_get_user_data([ffi.c_void_p], ffi.c_void_p)
_lv_obj_scroll_to      = _lib.lv_obj_scroll_to(
    [ffi.c_void_p, ffi.c_int32, ffi.c_int32, ffi.c_int32], None)
_lv_obj_scroll_to_view = _lib.lv_obj_scroll_to_view([ffi.c_void_p, ffi.c_int32], None)
_lv_obj_set_scrollbar_mode = _lib.lv_obj_set_scrollbar_mode([ffi.c_void_p, ffi.c_int32], None)
_lv_obj_set_scroll_snap_x  = _lib.lv_obj_set_scroll_snap_x([ffi.c_void_p, ffi.c_int32], None)
_lv_obj_set_scroll_snap_y  = _lib.lv_obj_set_scroll_snap_y([ffi.c_void_p, ffi.c_int32], None)
_lv_obj_invalidate     = _lib.lv_obj_invalidate([ffi.c_void_p], None)
_lv_obj_move_to_index  = _lib.lv_obj_move_to_index([ffi.c_void_p, ffi.c_int32], None)
_lv_obj_move_foreground = _lib.lv_obj_move_foreground([ffi.c_void_p], None)
_lv_obj_move_background = _lib.lv_obj_move_background([ffi.c_void_p], None)
_lv_obj_is_visible     = _lib.lv_obj_is_visible([ffi.c_void_p], ffi.c_int32)

# --- Screen ---
_lv_screen_active  = _lib.lv_screen_active([], ffi.c_void_p)
_lv_screen_load    = _lib.lv_screen_load([ffi.c_void_p], None)
_lv_screen_load_anim = _lib.lv_screen_load_anim(
    [ffi.c_void_p, ffi.c_int32, ffi.c_uint32, ffi.c_uint32, ffi.c_int32], None)

# Screen load anim types
SCR_LOAD_ANIM_NONE        = 0
SCR_LOAD_ANIM_OVER_LEFT   = 1
SCR_LOAD_ANIM_OVER_RIGHT  = 2
SCR_LOAD_ANIM_OVER_TOP    = 3
SCR_LOAD_ANIM_OVER_BOTTOM = 4
SCR_LOAD_ANIM_MOVE_LEFT   = 5
SCR_LOAD_ANIM_MOVE_RIGHT  = 6
SCR_LOAD_ANIM_MOVE_TOP    = 7
SCR_LOAD_ANIM_MOVE_BOTTOM = 8
SCR_LOAD_ANIM_FADE_IN     = 9
SCR_LOAD_ANIM_FADE_OUT    = 10
SCR_LOAD_ANIM_OUT_LEFT    = 11
SCR_LOAD_ANIM_OUT_RIGHT   = 12
SCR_LOAD_ANIM_OUT_TOP     = 13
SCR_LOAD_ANIM_OUT_BOTTOM  = 14

# --- Flex layout ---
_lv_obj_set_flex_flow   = _lib.lv_obj_set_flex_flow([ffi.c_void_p, ffi.c_int32], None)
_lv_obj_set_flex_align  = _lib.lv_obj_set_flex_align(
    [ffi.c_void_p, ffi.c_int32, ffi.c_int32, ffi.c_int32], None)
_lv_obj_set_flex_grow   = _lib.lv_obj_set_flex_grow([ffi.c_void_p, ffi.c_uint8], None)

# --- Animation ---
_lv_anim_init          = _lib.lv_anim_init([ffi.c_void_p], None)
_lv_anim_set_var       = _lib.lv_anim_set_var([ffi.c_void_p, ffi.c_void_p], None)
_lv_anim_set_exec_cb   = _lib.lv_anim_set_exec_cb([ffi.c_void_p, ffi.c_void_p], None)
_lv_anim_set_duration  = _lib.lv_anim_set_duration([ffi.c_void_p, ffi.c_uint32], None)
_lv_anim_set_values    = _lib.lv_anim_set_values(
    [ffi.c_void_p, ffi.c_int32, ffi.c_int32], None)
_lv_anim_set_path_cb   = _lib.lv_anim_set_path_cb([ffi.c_void_p, ffi.c_void_p], None)
_lv_anim_set_delay     = _lib.lv_anim_set_delay([ffi.c_void_p, ffi.c_uint32], None)
_lv_anim_set_repeat_count  = _lib.lv_anim_set_repeat_count([ffi.c_void_p, ffi.c_uint16], None)
_lv_anim_set_repeat_delay  = _lib.lv_anim_set_repeat_delay([ffi.c_void_p, ffi.c_uint32], None)
_lv_anim_set_playback_duration = _lib.lv_anim_set_playback_duration(
    [ffi.c_void_p, ffi.c_uint32], None)
_lv_anim_set_playback_delay    = _lib.lv_anim_set_playback_delay(
    [ffi.c_void_p, ffi.c_uint32], None)
_lv_anim_set_completed_cb = _lib.lv_anim_set_completed_cb([ffi.c_void_p, ffi.c_void_p], None)
_lv_anim_set_deleted_cb   = _lib.lv_anim_set_deleted_cb([ffi.c_void_p, ffi.c_void_p], None)
_lv_anim_start            = _lib.lv_anim_start([ffi.c_void_p], ffi.c_void_p)
_lv_anim_delete           = _lib.lv_anim_delete([ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_lv_anim_count_running     = _lib.lv_anim_count_running([], ffi.c_uint16)

# Animation path callbacks (function pointers — obtained via lv_ global symbols)
_lv_anim_path_linear       = _lib.lv_anim_path_linear(      [ffi.c_void_p], ffi.c_int32)
_lv_anim_path_ease_in      = _lib.lv_anim_path_ease_in(     [ffi.c_void_p], ffi.c_int32)
_lv_anim_path_ease_out     = _lib.lv_anim_path_ease_out(    [ffi.c_void_p], ffi.c_int32)
_lv_anim_path_ease_in_out  = _lib.lv_anim_path_ease_in_out( [ffi.c_void_p], ffi.c_int32)
_lv_anim_path_overshoot    = _lib.lv_anim_path_overshoot(   [ffi.c_void_p], ffi.c_int32)
_lv_anim_path_bounce       = _lib.lv_anim_path_bounce(      [ffi.c_void_p], ffi.c_int32)
_lv_anim_path_step         = _lib.lv_anim_path_step(        [ffi.c_void_p], ffi.c_int32)

# Path name → function pointer mapping (for Anim.set_path)
ANIM_PATH_LINEAR      = "linear"
ANIM_PATH_EASE_IN     = "ease_in"
ANIM_PATH_EASE_OUT    = "ease_out"
ANIM_PATH_EASE_IN_OUT = "ease_in_out"
ANIM_PATH_OVERSHOOT   = "overshoot"
ANIM_PATH_BOUNCE      = "bounce"
ANIM_PATH_STEP        = "step"

# --- Event ---
_lv_event_get_code        = _lib.lv_event_get_code([ffi.c_void_p], ffi.c_int32)
_lv_event_get_target      = _lib.lv_event_get_target([ffi.c_void_p], ffi.c_void_p)
_lv_event_get_current_target = _lib.lv_event_get_current_target([ffi.c_void_p], ffi.c_void_p)
_lv_event_get_user_data   = _lib.lv_event_get_user_data([ffi.c_void_p], ffi.c_void_p)
_lv_event_get_param       = _lib.lv_event_get_param([ffi.c_void_p], ffi.c_void_p)
_lv_event_stop_bubbling   = _lib.lv_event_stop_bubbling([ffi.c_void_p], None)
_lv_event_stop_processing = _lib.lv_event_stop_processing([ffi.c_void_p], None)

# --- Widgets: Button ---
_lv_button_create = _lib.lv_button_create([ffi.c_void_p], ffi.c_void_p)

# --- Widgets: Label ---
_lv_label_create        = _lib.lv_label_create([ffi.c_void_p], ffi.c_void_p)
_lv_label_set_text      = _lib.lv_label_set_text([ffi.c_void_p, ffi.c_char_p], None)
_lv_label_set_long_mode = _lib.lv_label_set_long_mode([ffi.c_void_p, ffi.c_int32], None)
_lv_label_get_text      = _lib.lv_label_get_text([ffi.c_void_p], ffi.c_char_p)
_lv_label_set_recolor   = _lib.lv_label_set_recolor([ffi.c_void_p, ffi.c_int32], None)
_lv_label_cut_text      = _lib.lv_label_cut_text([ffi.c_void_p, ffi.c_uint32, ffi.c_uint32], None)
_lv_label_ins_text      = _lib.lv_label_ins_text([ffi.c_void_p, ffi.c_uint32, ffi.c_char_p], None)

# --- Widgets: Slider ---
_lv_slider_create       = _lib.lv_slider_create([ffi.c_void_p], ffi.c_void_p)
_lv_slider_set_value    = _lib.lv_slider_set_value([ffi.c_void_p, ffi.c_int32, ffi.c_int32], None)
_lv_slider_set_range    = _lib.lv_slider_set_range([ffi.c_void_p, ffi.c_int32, ffi.c_int32], None)
_lv_slider_set_mode     = _lib.lv_slider_set_mode([ffi.c_void_p, ffi.c_int32], None)
_lv_slider_get_value    = _lib.lv_slider_get_value([ffi.c_void_p], ffi.c_int32)
_lv_slider_get_min_value = _lib.lv_slider_get_min_value([ffi.c_void_p], ffi.c_int32)
_lv_slider_get_max_value = _lib.lv_slider_get_max_value([ffi.c_void_p], ffi.c_int32)
_lv_slider_is_dragged   = _lib.lv_slider_is_dragged([ffi.c_void_p], ffi.c_int32)

# --- Widgets: Checkbox ---
_lv_checkbox_create   = _lib.lv_checkbox_create([ffi.c_void_p], ffi.c_void_p)
_lv_checkbox_set_text = _lib.lv_checkbox_set_text([ffi.c_void_p, ffi.c_char_p], None)
_lv_checkbox_get_text = _lib.lv_checkbox_get_text([ffi.c_void_p], ffi.c_char_p)

# --- Widgets: Switch ---
_lv_switch_create = _lib.lv_switch_create([ffi.c_void_p], ffi.c_void_p)

# --- Widgets: Arc ---
_lv_arc_create      = _lib.lv_arc_create([ffi.c_void_p], ffi.c_void_p)
_lv_arc_set_value   = _lib.lv_arc_set_value([ffi.c_void_p, ffi.c_int16], None)
_lv_arc_set_range   = _lib.lv_arc_set_range([ffi.c_void_p, ffi.c_int16, ffi.c_int16], None)
_lv_arc_set_angles  = _lib.lv_arc_set_angles([ffi.c_void_p, ffi.c_uint16, ffi.c_uint16], None)
_lv_arc_set_bg_angles = _lib.lv_arc_set_bg_angles([ffi.c_void_p, ffi.c_uint16, ffi.c_uint16], None)
_lv_arc_set_rotation = _lib.lv_arc_set_rotation([ffi.c_void_p, ffi.c_uint16], None)
_lv_arc_set_mode    = _lib.lv_arc_set_mode([ffi.c_void_p, ffi.c_int32], None)
_lv_arc_get_value   = _lib.lv_arc_get_value([ffi.c_void_p], ffi.c_int16)
_lv_arc_get_min_value = _lib.lv_arc_get_min_value([ffi.c_void_p], ffi.c_int16)
_lv_arc_get_max_value = _lib.lv_arc_get_max_value([ffi.c_void_p], ffi.c_int16)
_lv_arc_set_change_rate = _lib.lv_arc_set_change_rate([ffi.c_void_p, ffi.c_uint16], None)

# --- Widgets: Bar ---
_lv_bar_create      = _lib.lv_bar_create([ffi.c_void_p], ffi.c_void_p)
_lv_bar_set_value   = _lib.lv_bar_set_value([ffi.c_void_p, ffi.c_int32, ffi.c_int32], None)
_lv_bar_set_start_value = _lib.lv_bar_set_start_value([ffi.c_void_p, ffi.c_int32, ffi.c_int32], None)
_lv_bar_set_range   = _lib.lv_bar_set_range([ffi.c_void_p, ffi.c_int32, ffi.c_int32], None)
_lv_bar_set_mode    = _lib.lv_bar_set_mode([ffi.c_void_p, ffi.c_int32], None)
_lv_bar_get_value   = _lib.lv_bar_get_value([ffi.c_void_p], ffi.c_int32)
_lv_bar_get_start_value = _lib.lv_bar_get_start_value([ffi.c_void_p], ffi.c_int32)
_lv_bar_get_min_value = _lib.lv_bar_get_min_value([ffi.c_void_p], ffi.c_int32)
_lv_bar_get_max_value = _lib.lv_bar_get_max_value([ffi.c_void_p], ffi.c_int32)

# --- Widgets: Dropdown ---
_lv_dropdown_create        = _lib.lv_dropdown_create([ffi.c_void_p], ffi.c_void_p)
_lv_dropdown_set_options   = _lib.lv_dropdown_set_options([ffi.c_void_p, ffi.c_char_p], None)
_lv_dropdown_add_option    = _lib.lv_dropdown_add_option([ffi.c_void_p, ffi.c_char_p, ffi.c_uint32], None)
_lv_dropdown_set_selected  = _lib.lv_dropdown_set_selected([ffi.c_void_p, ffi.c_uint16], None)
_lv_dropdown_get_selected  = _lib.lv_dropdown_get_selected([ffi.c_void_p], ffi.c_uint16)
_lv_dropdown_get_selected_str = _lib.lv_dropdown_get_selected_str(
    [ffi.c_void_p, ffi.c_char_p, ffi.c_uint32], None)
_lv_dropdown_set_text      = _lib.lv_dropdown_set_text([ffi.c_void_p, ffi.c_char_p], None)
_lv_dropdown_set_dir       = _lib.lv_dropdown_set_dir([ffi.c_void_p, ffi.c_int32], None)
_lv_dropdown_open          = _lib.lv_dropdown_open([ffi.c_void_p], None)
_lv_dropdown_close         = _lib.lv_dropdown_close([ffi.c_void_p], None)
_lv_dropdown_is_open       = _lib.lv_dropdown_is_open([ffi.c_void_p], ffi.c_int32)
_lv_dropdown_get_option_count = _lib.lv_dropdown_get_option_count([ffi.c_void_p], ffi.c_uint32)

# --- Widgets: Roller ---
_lv_roller_create          = _lib.lv_roller_create([ffi.c_void_p], ffi.c_void_p)
_lv_roller_set_options     = _lib.lv_roller_set_options(
    [ffi.c_void_p, ffi.c_char_p, ffi.c_int32], None)
_lv_roller_set_selected    = _lib.lv_roller_set_selected(
    [ffi.c_void_p, ffi.c_uint16, ffi.c_int32], None)
_lv_roller_get_selected    = _lib.lv_roller_get_selected([ffi.c_void_p], ffi.c_uint16)
_lv_roller_get_selected_str = _lib.lv_roller_get_selected_str(
    [ffi.c_void_p, ffi.c_char_p, ffi.c_uint32], None)
_lv_roller_set_visible_row_count = _lib.lv_roller_set_visible_row_count(
    [ffi.c_void_p, ffi.c_uint8], None)
_lv_roller_get_option_count = _lib.lv_roller_get_option_count([ffi.c_void_p], ffi.c_uint16)

# --- Widgets: Textarea ---
_lv_textarea_create           = _lib.lv_textarea_create([ffi.c_void_p], ffi.c_void_p)
_lv_textarea_set_text         = _lib.lv_textarea_set_text([ffi.c_void_p, ffi.c_char_p], None)
_lv_textarea_get_text         = _lib.lv_textarea_get_text([ffi.c_void_p], ffi.c_char_p)
_lv_textarea_add_char         = _lib.lv_textarea_add_char([ffi.c_void_p, ffi.c_uint32], None)
_lv_textarea_add_text         = _lib.lv_textarea_add_text([ffi.c_void_p, ffi.c_char_p], None)
_lv_textarea_delete_char      = _lib.lv_textarea_delete_char([ffi.c_void_p], None)
_lv_textarea_delete_char_forward = _lib.lv_textarea_delete_char_forward([ffi.c_void_p], None)
_lv_textarea_set_placeholder_text = _lib.lv_textarea_set_placeholder_text(
    [ffi.c_void_p, ffi.c_char_p], None)
_lv_textarea_set_one_line     = _lib.lv_textarea_set_one_line([ffi.c_void_p, ffi.c_int32], None)
_lv_textarea_set_password_mode = _lib.lv_textarea_set_password_mode([ffi.c_void_p, ffi.c_int32], None)
_lv_textarea_set_max_length   = _lib.lv_textarea_set_max_length([ffi.c_void_p, ffi.c_uint32], None)
_lv_textarea_set_accepted_chars = _lib.lv_textarea_set_accepted_chars(
    [ffi.c_void_p, ffi.c_char_p], None)
_lv_textarea_set_cursor_pos   = _lib.lv_textarea_set_cursor_pos([ffi.c_void_p, ffi.c_int32], None)
_lv_textarea_get_cursor_pos   = _lib.lv_textarea_get_cursor_pos([ffi.c_void_p], ffi.c_uint32)
_lv_textarea_set_password_show_time = _lib.lv_textarea_set_password_show_time(
    [ffi.c_void_p, ffi.c_uint16], None)

# --- Widgets: Keyboard ---
_lv_keyboard_create      = _lib.lv_keyboard_create([ffi.c_void_p], ffi.c_void_p)
_lv_keyboard_set_textarea = _lib.lv_keyboard_set_textarea([ffi.c_void_p, ffi.c_void_p], None)
_lv_keyboard_set_mode    = _lib.lv_keyboard_set_mode([ffi.c_void_p, ffi.c_int32], None)
_lv_keyboard_get_textarea = _lib.lv_keyboard_get_textarea([ffi.c_void_p], ffi.c_void_p)

# --- Widgets: Spinner ---
_lv_spinner_create     = _lib.lv_spinner_create([ffi.c_void_p], ffi.c_void_p)
_lv_spinner_set_anim_params = _lib.lv_spinner_set_anim_params(
    [ffi.c_void_p, ffi.c_uint32, ffi.c_uint32], None)

# --- Widgets: Image ---
_lv_image_create      = _lib.lv_image_create([ffi.c_void_p], ffi.c_void_p)
_lv_image_set_src     = _lib.lv_image_set_src([ffi.c_void_p, ffi.c_void_p], None)
_lv_image_set_angle   = _lib.lv_image_set_angle([ffi.c_void_p, ffi.c_int16], None)
_lv_image_set_zoom    = _lib.lv_image_set_zoom([ffi.c_void_p, ffi.c_uint16], None)
_lv_image_set_pivot   = _lib.lv_image_set_pivot([ffi.c_void_p, ffi.c_int32, ffi.c_int32], None)
_lv_image_set_antialias = _lib.lv_image_set_antialias([ffi.c_void_p, ffi.c_int32], None)
_lv_image_set_offset_x  = _lib.lv_image_set_offset_x([ffi.c_void_p, ffi.c_int32], None)
_lv_image_set_offset_y  = _lib.lv_image_set_offset_y([ffi.c_void_p, ffi.c_int32], None)
_lv_image_get_src       = _lib.lv_image_get_src([ffi.c_void_p], ffi.c_void_p)

# --- Widgets: Line ---
_lv_line_create       = _lib.lv_line_create([ffi.c_void_p], ffi.c_void_p)
_lv_line_set_points   = _lib.lv_line_set_points(
    [ffi.c_void_p, ffi.c_void_p, ffi.c_uint16], None)
_lv_line_set_y_invert = _lib.lv_line_set_y_invert([ffi.c_void_p, ffi.c_int32], None)

# --- Widgets: Table ---
_lv_table_create        = _lib.lv_table_create([ffi.c_void_p], ffi.c_void_p)
_lv_table_set_cell_value = _lib.lv_table_set_cell_value(
    [ffi.c_void_p, ffi.c_uint16, ffi.c_uint16, ffi.c_char_p], None)
_lv_table_get_cell_value = _lib.lv_table_get_cell_value(
    [ffi.c_void_p, ffi.c_uint16, ffi.c_uint16], ffi.c_char_p)
_lv_table_set_column_count = _lib.lv_table_set_column_count([ffi.c_void_p, ffi.c_uint16], None)
_lv_table_get_column_count = _lib.lv_table_get_column_count([ffi.c_void_p], ffi.c_uint16)
_lv_table_set_row_count    = _lib.lv_table_set_row_count([ffi.c_void_p, ffi.c_uint16], None)
_lv_table_get_row_count    = _lib.lv_table_get_row_count([ffi.c_void_p], ffi.c_uint16)
_lv_table_set_column_width = _lib.lv_table_set_column_width(
    [ffi.c_void_p, ffi.c_uint16, ffi.c_int32], None)
_lv_table_get_selected_cell = _lib.lv_table_get_selected_cell(
    [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], None)

# --- Widgets: List ---
_lv_list_create     = _lib.lv_list_create([ffi.c_void_p], ffi.c_void_p)
_lv_list_add_text   = _lib.lv_list_add_text([ffi.c_void_p, ffi.c_char_p], ffi.c_void_p)
_lv_list_add_button = _lib.lv_list_add_button(
    [ffi.c_void_p, ffi.c_void_p, ffi.c_char_p], ffi.c_void_p)
_lv_list_get_button_text = _lib.lv_list_get_button_text(
    [ffi.c_void_p, ffi.c_void_p], ffi.c_char_p)

# --- Widgets: Message box ---
_lv_msgbox_create         = _lib.lv_msgbox_create([ffi.c_void_p], ffi.c_void_p)
_lv_msgbox_add_title      = _lib.lv_msgbox_add_title([ffi.c_void_p, ffi.c_char_p], ffi.c_void_p)
_lv_msgbox_add_text       = _lib.lv_msgbox_add_text( [ffi.c_void_p, ffi.c_char_p], ffi.c_void_p)
_lv_msgbox_add_close_button = _lib.lv_msgbox_add_close_button([ffi.c_void_p], ffi.c_void_p)
_lv_msgbox_add_footer_button = _lib.lv_msgbox_add_footer_button(
    [ffi.c_void_p, ffi.c_char_p], ffi.c_void_p)
_lv_msgbox_close          = _lib.lv_msgbox_close([ffi.c_void_p], None)

# ==============================================================================
#  Style class
# ==============================================================================

class Style:
    """
    Wraps lv_style_t.  Allocate, set properties, apply to one or more objects.

    Usage:
        s = Style()
        s.bg_color(color_hex(0x2196F3))
        s.text_color(COLOR_WHITE)
        s.radius(8)
        s.pad_all(10)
        btn.add_style(s)           # applies to LV_PART_MAIN | LV_STATE_DEFAULT
        btn.add_style(s, LV_PART_MAIN | LV_STATE_PRESSED)
        s.free()
    """
    def __init__(self):
        self._buf   = ffi.malloc(_SZ_STYLE)
        self._freed = False
        _lv_style_init(self._buf)

    # --- background ---
    def bg_color(self, color):              _lv_style_set_bg_color(self._buf, color)
    def bg_opa(self, opa):                  _lv_style_set_bg_opa(self._buf, opa)
    def bg_grad_color(self, color):         _lv_style_set_bg_grad_color(self._buf, color)
    def bg_grad_dir(self, dir_):            _lv_style_set_bg_grad_dir(self._buf, dir_)
    def bg_main_stop(self, stop):           _lv_style_set_bg_main_stop(self._buf, stop)
    def bg_grad_stop(self, stop):           _lv_style_set_bg_grad_stop(self._buf, stop)

    # --- border ---
    def border_color(self, color):          _lv_style_set_border_color(self._buf, color)
    def border_width(self, w):              _lv_style_set_border_width(self._buf, w)
    def border_opa(self, opa):              _lv_style_set_border_opa(self._buf, opa)
    def border_side(self, side):            _lv_style_set_border_side(self._buf, side)

    # --- outline ---
    def outline_color(self, color):         _lv_style_set_outline_color(self._buf, color)
    def outline_width(self, w):             _lv_style_set_outline_width(self._buf, w)
    def outline_pad(self, p):               _lv_style_set_outline_pad(self._buf, p)
    def outline_opa(self, opa):             _lv_style_set_outline_opa(self._buf, opa)

    # --- shadow ---
    def shadow_color(self, color):          _lv_style_set_shadow_color(self._buf, color)
    def shadow_width(self, w):              _lv_style_set_shadow_width(self._buf, w)
    def shadow_opa(self, opa):              _lv_style_set_shadow_opa(self._buf, opa)
    def shadow_offset_x(self, x):          _lv_style_set_shadow_offset_x(self._buf, x)
    def shadow_offset_y(self, y):          _lv_style_set_shadow_offset_y(self._buf, y)
    def shadow_spread(self, s):             _lv_style_set_shadow_spread(self._buf, s)

    # --- text ---
    def text_color(self, color):            _lv_style_set_text_color(self._buf, color)
    def text_opa(self, opa):                _lv_style_set_text_opa(self._buf, opa)
    def text_font(self, font_ptr):          _lv_style_set_text_font(self._buf, font_ptr)
    def text_letter_space(self, s):         _lv_style_set_text_letter_space(self._buf, s)
    def text_line_space(self, s):           _lv_style_set_text_line_space(self._buf, s)
    def text_decor(self, d):                _lv_style_set_text_decor(self._buf, d)
    def text_align(self, a):                _lv_style_set_text_align(self._buf, a)

    # --- padding ---
    def pad_top(self, v):                   _lv_style_set_pad_top(self._buf, v)
    def pad_bottom(self, v):                _lv_style_set_pad_bottom(self._buf, v)
    def pad_left(self, v):                  _lv_style_set_pad_left(self._buf, v)
    def pad_right(self, v):                 _lv_style_set_pad_right(self._buf, v)
    def pad_row(self, v):                   _lv_style_set_pad_row(self._buf, v)
    def pad_column(self, v):                _lv_style_set_pad_column(self._buf, v)
    def pad_hor(self, v):
        _lv_style_set_pad_left(self._buf, v)
        _lv_style_set_pad_right(self._buf, v)
    def pad_ver(self, v):
        _lv_style_set_pad_top(self._buf, v)
        _lv_style_set_pad_bottom(self._buf, v)
    def pad_all(self, v):
        self.pad_hor(v)
        self.pad_ver(v)
    def pad_gap(self, v):
        self.pad_row(v)
        self.pad_column(v)

    # --- margin ---
    def margin_top(self, v):                _lv_style_set_margin_top(self._buf, v)
    def margin_bottom(self, v):             _lv_style_set_margin_bottom(self._buf, v)
    def margin_left(self, v):               _lv_style_set_margin_left(self._buf, v)
    def margin_right(self, v):              _lv_style_set_margin_right(self._buf, v)

    # --- size and shape ---
    def width(self, v):                     _lv_style_set_width(self._buf, v)
    def min_width(self, v):                 _lv_style_set_min_width(self._buf, v)
    def max_width(self, v):                 _lv_style_set_max_width(self._buf, v)
    def height(self, v):                    _lv_style_set_height(self._buf, v)
    def min_height(self, v):               _lv_style_set_min_height(self._buf, v)
    def max_height(self, v):               _lv_style_set_max_height(self._buf, v)
    def radius(self, v):                    _lv_style_set_radius(self._buf, v)
    def clip_corner(self, v):               _lv_style_set_clip_corner(self._buf, v)
    def opa(self, v):                       _lv_style_set_opa(self._buf, v)
    def blend_mode(self, v):                _lv_style_set_blend_mode(self._buf, v)

    # --- transform ---
    def transform_width(self, v):           _lv_style_set_transform_width(self._buf, v)
    def transform_height(self, v):          _lv_style_set_transform_height(self._buf, v)
    def translate_x(self, v):              _lv_style_set_translate_x(self._buf, v)
    def translate_y(self, v):              _lv_style_set_translate_y(self._buf, v)
    def transform_scale_x(self, v):        _lv_style_set_transform_scale_x(self._buf, v)
    def transform_scale_y(self, v):        _lv_style_set_transform_scale_y(self._buf, v)
    def transform_rotation(self, v):        _lv_style_set_transform_rotation(self._buf, v)
    def transform_pivot_x(self, v):        _lv_style_set_transform_pivot_x(self._buf, v)
    def transform_pivot_y(self, v):        _lv_style_set_transform_pivot_y(self._buf, v)

    # --- line ---
    def line_color(self, color):            _lv_style_set_line_color(self._buf, color)
    def line_width(self, w):               _lv_style_set_line_width(self._buf, w)
    def line_dash_width(self, w):           _lv_style_set_line_dash_width(self._buf, w)
    def line_dash_gap(self, g):             _lv_style_set_line_dash_gap(self._buf, g)
    def line_rounded(self, v):              _lv_style_set_line_rounded(self._buf, v)
    def line_opa(self, v):                  _lv_style_set_line_opa(self._buf, v)

    # --- arc ---
    def arc_color(self, color):             _lv_style_set_arc_color(self._buf, color)
    def arc_width(self, w):                 _lv_style_set_arc_width(self._buf, w)
    def arc_rounded(self, v):               _lv_style_set_arc_rounded(self._buf, v)
    def arc_opa(self, v):                   _lv_style_set_arc_opa(self._buf, v)

    # --- image ---
    def image_opa(self, v):                 _lv_style_set_image_opa(self._buf, v)
    def image_recolor(self, color):         _lv_style_set_image_recolor(self._buf, color)
    def image_recolor_opa(self, v):         _lv_style_set_image_recolor_opa(self._buf, v)

    # --- layout ---
    def layout(self, v):                    _lv_style_set_layout(self._buf, v)
    def scrollbar_mode(self, v):            _lv_style_set_scrollbar_mode(self._buf, v)

    def reset(self):
        _lv_style_reset(self._buf)

    def free(self):
        if not self._freed and self._buf:
            _lv_style_reset(self._buf)
            ffi.free(self._buf)
            self._buf   = None
            self._freed = True

# ==============================================================================
#  Animation class
# ==============================================================================

_ANIM_PATH_MAP = {
    ANIM_PATH_LINEAR:      _lv_anim_path_linear,
    ANIM_PATH_EASE_IN:     _lv_anim_path_ease_in,
    ANIM_PATH_EASE_OUT:    _lv_anim_path_ease_out,
    ANIM_PATH_EASE_IN_OUT: _lv_anim_path_ease_in_out,
    ANIM_PATH_OVERSHOOT:   _lv_anim_path_overshoot,
    ANIM_PATH_BOUNCE:      _lv_anim_path_bounce,
    ANIM_PATH_STEP:        _lv_anim_path_step,
}

class Anim:
    """
    Wraps lv_anim_t.

    Usage:
        a = Anim()
        a.set_var(obj.ptr)
        a.set_exec_cb(my_exec_cb)       # ffi.callback wrapping (obj_ptr, int32_t value)
        a.set_values(0, 100)
        a.set_duration(500)
        a.set_path(ANIM_PATH_EASE_IN_OUT)
        a.set_repeat_count(LV_ANIM_REPEAT_INFINITE)
        a.start()
    """
    def __init__(self):
        self._buf = ffi.malloc(_SZ_ANIM)
        self._exec_cb_ref = None
        self._done_cb_ref = None
        _lv_anim_init(self._buf)

    def set_var(self, obj_ptr):
        _lv_anim_set_var(self._buf, obj_ptr)

    def set_exec_cb(self, cb):
        """cb = ffi.callback(fn, None, [ffi.c_void_p, ffi.c_int32])"""
        self._exec_cb_ref = cb
        _lv_anim_set_exec_cb(self._buf, cb)

    def set_values(self, start, end):
        _lv_anim_set_values(self._buf, start, end)

    def set_duration(self, ms):
        _lv_anim_set_duration(self._buf, ms)

    def set_path(self, name):
        fn = _ANIM_PATH_MAP.get(name)
        if fn is None:
            raise ValueError(format_str("Unknown anim path: {name}"))
        _lv_anim_set_path_cb(self._buf, fn)

    def set_delay(self, ms):
        _lv_anim_set_delay(self._buf, ms)

    def set_repeat_count(self, cnt):
        _lv_anim_set_repeat_count(self._buf, cnt)

    def set_repeat_delay(self, ms):
        _lv_anim_set_repeat_delay(self._buf, ms)

    def set_playback_duration(self, ms):
        _lv_anim_set_playback_duration(self._buf, ms)

    def set_playback_delay(self, ms):
        _lv_anim_set_playback_delay(self._buf, ms)

    def set_completed_cb(self, cb):
        """cb = ffi.callback(fn, None, [ffi.c_void_p])"""
        self._done_cb_ref = cb
        _lv_anim_set_completed_cb(self._buf, cb)

    def start(self):
        """Start the animation. Returns C pointer to running lv_anim_t."""
        return _lv_anim_start(self._buf)

    @staticmethod
    def delete(obj_ptr, exec_cb=None):
        """Stop animation(s) running on obj_ptr."""
        _lv_anim_delete(obj_ptr, exec_cb)

    @staticmethod
    def count_running():
        return _lv_anim_count_running()

    def free(self):
        if self._buf:
            ffi.free(self._buf)
            self._buf = None

# ==============================================================================
#  Timer class
# ==============================================================================

class Timer:
    """
    Wraps lv_timer_t.

    Usage:
        def on_tick(timer_ptr):
            print("tick!")
        cb  = ffi.callback(on_tick, None, [ffi.c_void_p])
        t   = Timer(cb, period_ms=500)
        # later:
        t.set_period(1000)
        t.pause()
        t.resume()
        t.delete()
    """
    def __init__(self, cb, period_ms=1000, repeat_count=-1):
        self._cb  = cb          # keep alive
        self._ptr = _lv_timer_create(cb, period_ms, ffi.c_void_p(0))
        if repeat_count != -1:
            _lv_timer_set_repeat_count(self._ptr, repeat_count)

    def set_period(self, ms):
        _lv_timer_set_period(self._ptr, ms)

    def reset(self):
        _lv_timer_reset(self._ptr)

    def pause(self):
        _lv_timer_pause(self._ptr)

    def resume(self):
        _lv_timer_resume(self._ptr)

    def set_repeat_count(self, cnt):
        _lv_timer_set_repeat_count(self._ptr, cnt)

    def delete(self):
        if self._ptr:
            _lv_timer_delete(self._ptr)
            self._ptr = None

# ==============================================================================
#  Object base (Obj)
# ==============================================================================

class Obj:
    """
    Base class wrapping lv_obj_t.  All widget classes extend this.
    Can also be used directly to create a plain container.

    Usage:
        scr  = screen_active()
        cont = Obj(scr.ptr)
        cont.set_size(200, 100)
        cont.center()
    """
    def __init__(self, parent_ptr=None):
        if parent_ptr is None:
            parent_ptr = _lv_screen_active()
        self.ptr        = _lv_obj_create(parent_ptr)
        self._event_cbs = []
        self._styles    = []

    # --- geometry ---
    def set_pos(self, x, y):                _lv_obj_set_pos(self.ptr, x, y)
    def set_x(self, x):                     _lv_obj_set_x(self.ptr, x)
    def set_y(self, y):                     _lv_obj_set_y(self.ptr, y)
    def set_size(self, w, h):               _lv_obj_set_size(self.ptr, w, h)
    def set_width(self, w):                 _lv_obj_set_width(self.ptr, w)
    def set_height(self, h):               _lv_obj_set_height(self.ptr, h)
    def set_align(self, align):             _lv_obj_set_align(self.ptr, align)
    def align(self, align, x_ofs=0, y_ofs=0):
        _lv_obj_align(self.ptr, align, x_ofs, y_ofs)
    def align_to(self, base_ptr, align, x_ofs=0, y_ofs=0):
        _lv_obj_align_to(self.ptr, base_ptr, align, x_ofs, y_ofs)
    def center(self):                       _lv_obj_center(self.ptr)
    def get_x(self):                        return _lv_obj_get_x(self.ptr)
    def get_y(self):                        return _lv_obj_get_y(self.ptr)
    def get_width(self):                    return _lv_obj_get_width(self.ptr)
    def get_height(self):                   return _lv_obj_get_height(self.ptr)
    def get_content_width(self):            return _lv_obj_get_content_width(self.ptr)
    def get_content_height(self):           return _lv_obj_get_content_height(self.ptr)

    # --- hierarchy ---
    def get_parent(self):                   return _lv_obj_get_parent(self.ptr)
    def get_child(self, index):             return _lv_obj_get_child(self.ptr, index)
    def get_child_count(self):              return _lv_obj_get_child_count(self.ptr)
    def set_parent(self, parent_ptr):       _lv_obj_set_parent(self.ptr, parent_ptr)
    def move_foreground(self):              _lv_obj_move_foreground(self.ptr)
    def move_background(self):              _lv_obj_move_background(self.ptr)
    def move_to_index(self, index):         _lv_obj_move_to_index(self.ptr, index)

    # --- style ---
    def add_style(self, style, selector=0):
        """style can be a Style instance or a raw C pointer."""
        ptr = style._buf if isinstance(style, Style) else style
        self._styles.append(style)
        _lv_obj_add_style(self.ptr, ptr, selector)

    def remove_style(self, style, selector=0):
        ptr = style._buf if isinstance(style, Style) else style
        _lv_obj_remove_style(self.ptr, ptr, selector)

    def remove_style_all(self):
        _lv_obj_remove_style_all(self.ptr)

    # --- flags ---
    def add_flag(self, flag):               _lv_obj_add_flag(self.ptr, flag)
    def remove_flag(self, flag):            _lv_obj_remove_flag(self.ptr, flag)
    def has_flag(self, flag):               return _lv_obj_has_flag(self.ptr, flag) != 0
    def show(self):                         self.remove_flag(LV_OBJ_FLAG_HIDDEN)
    def hide(self):                         self.add_flag(LV_OBJ_FLAG_HIDDEN)
    def is_visible(self):                   return _lv_obj_is_visible(self.ptr) != 0

    # --- state ---
    def add_state(self, state):             _lv_obj_add_state(self.ptr, state)
    def remove_state(self, state):          _lv_obj_remove_state(self.ptr, state)
    def has_state(self, state):             return _lv_obj_has_state(self.ptr, state) != 0

    # --- events ---
    def on(self, event_code, callback, user_data=None):
        """
        Register an event callback.
        callback receives a single event_ptr argument.
        Returns the C event descriptor (can be passed to off()).
        """
        cb  = ffi.callback(callback, None, [ffi.c_void_p])
        ud  = user_data if user_data is not None else ffi.c_void_p(0)
        desc = _lv_obj_add_event_cb(self.ptr, cb, event_code, ud)
        self._event_cbs.append(cb)   # keep alive
        return desc

    def off(self, callback):
        _lv_obj_remove_event_cb(self.ptr, callback)

    def send_event(self, event_code, param=None):
        p = param if param is not None else ffi.c_void_p(0)
        _lv_obj_send_event(self.ptr, event_code, p)

    # --- user data ---
    def set_user_data(self, ptr):           _lv_obj_set_user_data(self.ptr, ptr)
    def get_user_data(self):                return _lv_obj_get_user_data(self.ptr)

    # --- scroll ---
    def scroll_to(self, x, y, anim=LV_ANIM_ON):
        _lv_obj_scroll_to(self.ptr, x, y, anim)
    def scroll_to_view(self, anim=LV_ANIM_ON):
        _lv_obj_scroll_to_view(self.ptr, anim)
    def set_scrollbar_mode(self, mode):
        _lv_obj_set_scrollbar_mode(self.ptr, mode)
    def set_scroll_snap_x(self, snap):      _lv_obj_set_scroll_snap_x(self.ptr, snap)
    def set_scroll_snap_y(self, snap):      _lv_obj_set_scroll_snap_y(self.ptr, snap)

    # --- layout ---
    def set_flex_flow(self, flow):          _lv_obj_set_flex_flow(self.ptr, flow)
    def set_flex_align(self, main, cross, track):
        _lv_obj_set_flex_align(self.ptr, main, cross, track)
    def set_flex_grow(self, grow):          _lv_obj_set_flex_grow(self.ptr, grow)

    # --- misc ---
    def invalidate(self):                   _lv_obj_invalidate(self.ptr)
    def clean(self):                        _lv_obj_clean(self.ptr)

    def delete(self):
        if self.ptr:
            _lv_obj_delete(self.ptr)
            self.ptr = None

# ==============================================================================
#  Widget classes
# ==============================================================================

class Button(Obj):
    """
    Usage:
        btn = Button(parent_ptr)
        btn.set_size(120, 50)
        btn.center()
        btn.on(LV_EVENT_CLICKED, lambda e: print("clicked!"))
    """
    def __init__(self, parent_ptr):
        self.ptr        = _lv_button_create(parent_ptr)
        self._event_cbs = []
        self._styles    = []


class Label(Obj):
    """
    Usage:
        lbl = Label(parent_ptr)
        lbl.set_text("Hello LVGL!")
        lbl.center()
    """
    def __init__(self, parent_ptr, text=""):
        self.ptr        = _lv_label_create(parent_ptr)
        self._event_cbs = []
        self._styles    = []
        if text:
            self.set_text(text)

    def set_text(self, text):
        if isinstance(text, str):
            text = text.encode("utf-8")
        _lv_label_set_text(self.ptr, text)

    def get_text(self):
        raw = _lv_label_get_text(self.ptr)
        return raw.decode("utf-8") if raw else ""

    def set_long_mode(self, mode):
        _lv_label_set_long_mode(self.ptr, mode)

    def set_recolor(self, enabled):
        _lv_label_set_recolor(self.ptr, 1 if enabled else 0)

    def ins_text(self, pos, text):
        if isinstance(text, str):
            text = text.encode("utf-8")
        _lv_label_ins_text(self.ptr, pos, text)

    def cut_text(self, pos, count):
        _lv_label_cut_text(self.ptr, pos, count)


class Slider(Obj):
    """
    Usage:
        s = Slider(parent_ptr)
        s.set_range(0, 100)
        s.set_value(50, LV_ANIM_OFF)
        s.on(LV_EVENT_VALUE_CHANGED, lambda e: print(s.get_value()))
    """
    def __init__(self, parent_ptr):
        self.ptr        = _lv_slider_create(parent_ptr)
        self._event_cbs = []
        self._styles    = []

    def set_value(self, val, anim=LV_ANIM_OFF):
        _lv_slider_set_value(self.ptr, val, anim)

    def set_range(self, min_, max_):
        _lv_slider_set_range(self.ptr, min_, max_)

    def set_mode(self, mode):
        _lv_slider_set_mode(self.ptr, mode)

    def get_value(self):        return _lv_slider_get_value(self.ptr)
    def get_min_value(self):    return _lv_slider_get_min_value(self.ptr)
    def get_max_value(self):    return _lv_slider_get_max_value(self.ptr)
    def is_dragged(self):       return _lv_slider_is_dragged(self.ptr) != 0


class Checkbox(Obj):
    """
    Usage:
        cb = Checkbox(parent_ptr, "Accept terms")
        cb.on(LV_EVENT_VALUE_CHANGED, lambda e: print(cb.is_checked()))
    """
    def __init__(self, parent_ptr, text=""):
        self.ptr        = _lv_checkbox_create(parent_ptr)
        self._event_cbs = []
        self._styles    = []
        if text:
            self.set_text(text)

    def set_text(self, text):
        if isinstance(text, str):
            text = text.encode("utf-8")
        _lv_checkbox_set_text(self.ptr, text)

    def get_text(self):
        raw = _lv_checkbox_get_text(self.ptr)
        return raw.decode("utf-8") if raw else ""

    def check(self):        self.add_state(LV_STATE_CHECKED)
    def uncheck(self):      self.remove_state(LV_STATE_CHECKED)
    def is_checked(self):   return self.has_state(LV_STATE_CHECKED)


class Switch(Obj):
    """
    Usage:
        sw = Switch(parent_ptr)
        sw.on(LV_EVENT_VALUE_CHANGED, lambda e: print(sw.is_on()))
    """
    def __init__(self, parent_ptr):
        self.ptr        = _lv_switch_create(parent_ptr)
        self._event_cbs = []
        self._styles    = []

    def turn_on(self, anim=LV_ANIM_ON):
        self.add_state(LV_STATE_CHECKED)

    def turn_off(self, anim=LV_ANIM_ON):
        self.remove_state(LV_STATE_CHECKED)

    def toggle(self):
        if self.is_on():
            self.turn_off()
        else:
            self.turn_on()

    def is_on(self):
        return self.has_state(LV_STATE_CHECKED)


class Arc(Obj):
    """
    Usage:
        a = Arc(parent_ptr)
        a.set_range(0, 360)
        a.set_value(90)
    """
    def __init__(self, parent_ptr):
        self.ptr        = _lv_arc_create(parent_ptr)
        self._event_cbs = []
        self._styles    = []

    def set_value(self, val):               _lv_arc_set_value(self.ptr, val)
    def set_range(self, min_, max_):        _lv_arc_set_range(self.ptr, min_, max_)
    def set_angles(self, start, end):       _lv_arc_set_angles(self.ptr, start, end)
    def set_bg_angles(self, start, end):    _lv_arc_set_bg_angles(self.ptr, start, end)
    def set_rotation(self, rot):            _lv_arc_set_rotation(self.ptr, rot)
    def set_mode(self, mode):               _lv_arc_set_mode(self.ptr, mode)
    def set_change_rate(self, rate):        _lv_arc_set_change_rate(self.ptr, rate)
    def get_value(self):                    return _lv_arc_get_value(self.ptr)
    def get_min_value(self):                return _lv_arc_get_min_value(self.ptr)
    def get_max_value(self):                return _lv_arc_get_max_value(self.ptr)


class Bar(Obj):
    """
    Usage:
        b = Bar(parent_ptr)
        b.set_range(0, 100)
        b.set_value(70, LV_ANIM_OFF)
    """
    def __init__(self, parent_ptr):
        self.ptr        = _lv_bar_create(parent_ptr)
        self._event_cbs = []
        self._styles    = []

    def set_value(self, val, anim=LV_ANIM_OFF):
        _lv_bar_set_value(self.ptr, val, anim)

    def set_start_value(self, val, anim=LV_ANIM_OFF):
        _lv_bar_set_start_value(self.ptr, val, anim)

    def set_range(self, min_, max_):        _lv_bar_set_range(self.ptr, min_, max_)
    def set_mode(self, mode):               _lv_bar_set_mode(self.ptr, mode)
    def get_value(self):                    return _lv_bar_get_value(self.ptr)
    def get_start_value(self):              return _lv_bar_get_start_value(self.ptr)
    def get_min_value(self):                return _lv_bar_get_min_value(self.ptr)
    def get_max_value(self):                return _lv_bar_get_max_value(self.ptr)


class Dropdown(Obj):
    """
    Usage:
        dd = Dropdown(parent_ptr)
        dd.set_options("Apple\nBanana\nCherry")
        dd.on(LV_EVENT_VALUE_CHANGED, lambda e: print(dd.get_selected_str()))
    """
    def __init__(self, parent_ptr, options=""):
        self.ptr        = _lv_dropdown_create(parent_ptr)
        self._event_cbs = []
        self._styles    = []
        if options:
            self.set_options(options)

    def set_options(self, opts):
        if isinstance(opts, str):
            opts = opts.encode("utf-8")
        _lv_dropdown_set_options(self.ptr, opts)

    def add_option(self, text, pos=0xFFFFFFFF):
        if isinstance(text, str):
            text = text.encode("utf-8")
        _lv_dropdown_add_option(self.ptr, text, pos)

    def set_selected(self, idx):            _lv_dropdown_set_selected(self.ptr, idx)
    def get_selected(self):                 return _lv_dropdown_get_selected(self.ptr)
    def get_option_count(self):             return _lv_dropdown_get_option_count(self.ptr)
    def set_dir(self, dir_):               _lv_dropdown_set_dir(self.ptr, dir_)
    def open(self):                         _lv_dropdown_open(self.ptr)
    def close(self):                        _lv_dropdown_close(self.ptr)
    def is_open(self):                      return _lv_dropdown_is_open(self.ptr) != 0

    def get_selected_str(self, max_len=64):
        buf = ffi.malloc(max_len)
        try:
            _lv_dropdown_get_selected_str(self.ptr, buf, max_len)
            raw = ffi.buffer_to_bytes(buf, max_len)
            text = ""
            for b in raw:
                if b == 0:
                    break
                text = text + chr(b)
            return text
        finally:
            ffi.free(buf)


class Roller(Obj):
    """
    Usage:
        r = Roller(parent_ptr, "Mon\nTue\nWed\nThu\nFri")
        r.set_visible_row_count(3)
        r.on(LV_EVENT_VALUE_CHANGED, lambda e: print(r.get_selected_str()))
    """
    def __init__(self, parent_ptr, options="", mode=LV_ROLLER_MODE_NORMAL):
        self.ptr        = _lv_roller_create(parent_ptr)
        self._event_cbs = []
        self._styles    = []
        if options:
            self.set_options(options, mode)

    def set_options(self, opts, mode=LV_ROLLER_MODE_NORMAL):
        if isinstance(opts, str):
            opts = opts.encode("utf-8")
        _lv_roller_set_options(self.ptr, opts, mode)

    def set_selected(self, idx, anim=LV_ANIM_ON):
        _lv_roller_set_selected(self.ptr, idx, anim)

    def set_visible_row_count(self, n):     _lv_roller_set_visible_row_count(self.ptr, n)
    def get_selected(self):                 return _lv_roller_get_selected(self.ptr)
    def get_option_count(self):             return _lv_roller_get_option_count(self.ptr)

    def get_selected_str(self, max_len=64):
        buf = ffi.malloc(max_len)
        try:
            _lv_roller_get_selected_str(self.ptr, buf, max_len)
            raw = ffi.buffer_to_bytes(buf, max_len)
            text = ""
            for b in raw:
                if b == 0:
                    break
                text = text + chr(b)
            return text
        finally:
            ffi.free(buf)


class Textarea(Obj):
    """
    Usage:
        ta = Textarea(parent_ptr)
        ta.set_one_line(True)
        ta.set_placeholder_text("Type here...")
        ta.on(LV_EVENT_VALUE_CHANGED, lambda e: print(ta.get_text()))
    """
    def __init__(self, parent_ptr, text=""):
        self.ptr        = _lv_textarea_create(parent_ptr)
        self._event_cbs = []
        self._styles    = []
        if text:
            self.set_text(text)

    def set_text(self, text):
        if isinstance(text, str):
            text = text.encode("utf-8")
        _lv_textarea_set_text(self.ptr, text)

    def get_text(self):
        raw = _lv_textarea_get_text(self.ptr)
        return raw.decode("utf-8") if raw else ""

    def add_char(self, c):
        _lv_textarea_add_char(self.ptr, ord(c) if isinstance(c, str) else c)

    def add_text(self, text):
        if isinstance(text, str):
            text = text.encode("utf-8")
        _lv_textarea_add_text(self.ptr, text)

    def delete_char(self):                  _lv_textarea_delete_char(self.ptr)
    def delete_char_forward(self):          _lv_textarea_delete_char_forward(self.ptr)

    def set_placeholder_text(self, text):
        if isinstance(text, str):
            text = text.encode("utf-8")
        _lv_textarea_set_placeholder_text(self.ptr, text)

    def set_one_line(self, en):
        _lv_textarea_set_one_line(self.ptr, 1 if en else 0)

    def set_password_mode(self, en):
        _lv_textarea_set_password_mode(self.ptr, 1 if en else 0)

    def set_max_length(self, n):            _lv_textarea_set_max_length(self.ptr, n)

    def set_accepted_chars(self, chars):
        if isinstance(chars, str):
            chars = chars.encode("utf-8")
        _lv_textarea_set_accepted_chars(self.ptr, chars)

    def set_cursor_pos(self, pos):          _lv_textarea_set_cursor_pos(self.ptr, pos)
    def get_cursor_pos(self):               return _lv_textarea_get_cursor_pos(self.ptr)
    def set_password_show_time(self, ms):   _lv_textarea_set_password_show_time(self.ptr, ms)


class Keyboard(Obj):
    """
    Usage:
        ta  = Textarea(scr.ptr)
        kb  = Keyboard(scr.ptr)
        kb.set_textarea(ta)
    """
    def __init__(self, parent_ptr):
        self.ptr        = _lv_keyboard_create(parent_ptr)
        self._event_cbs = []
        self._styles    = []

    def set_textarea(self, ta):
        ptr = ta.ptr if isinstance(ta, Obj) else ta
        _lv_keyboard_set_textarea(self.ptr, ptr)

    def set_mode(self, mode):               _lv_keyboard_set_mode(self.ptr, mode)
    def get_textarea(self):                 return _lv_keyboard_get_textarea(self.ptr)


class Spinner(Obj):
    """
    Usage:
        sp = Spinner(parent_ptr)
        sp.set_anim_params(arc_length=60, speed=1000)
        sp.set_size(80, 80)
        sp.center()
    """
    def __init__(self, parent_ptr):
        self.ptr        = _lv_spinner_create(parent_ptr)
        self._event_cbs = []
        self._styles    = []

    def set_anim_params(self, arc_length=60, speed=1000):
        _lv_spinner_set_anim_params(self.ptr, arc_length, speed)


class Image(Obj):
    """
    Usage:
        img = Image(parent_ptr)
        img.set_src(b"path/to/image.png")   # file path or lv_image_dsc_t ptr
        img.set_zoom(384)                    # 256=100%, 384=150%
    """
    def __init__(self, parent_ptr):
        self.ptr        = _lv_image_create(parent_ptr)
        self._event_cbs = []
        self._styles    = []

    def set_src(self, src):
        if isinstance(src, (str, bytes)):
            if isinstance(src, str):
                src = src.encode("utf-8")
            _lv_image_set_src(self.ptr, src)
        else:
            _lv_image_set_src(self.ptr, src)  # raw C ptr (lv_image_dsc_t*)

    def set_angle(self, angle):             _lv_image_set_angle(self.ptr, angle)
    def set_zoom(self, zoom):               _lv_image_set_zoom(self.ptr, zoom)
    def set_pivot(self, x, y):              _lv_image_set_pivot(self.ptr, x, y)
    def set_antialias(self, en):            _lv_image_set_antialias(self.ptr, 1 if en else 0)
    def set_offset_x(self, x):             _lv_image_set_offset_x(self.ptr, x)
    def set_offset_y(self, y):             _lv_image_set_offset_y(self.ptr, y)
    def get_src(self):                      return _lv_image_get_src(self.ptr)


class Line(Obj):
    """
    Usage:
        pts  = make_points([(0,0),(100,50),(200,0)])
        line = Line(parent_ptr)
        line.set_points(pts, 3)
    """
    def __init__(self, parent_ptr):
        self.ptr        = _lv_line_create(parent_ptr)
        self._event_cbs = []
        self._styles    = []
        self._points_buf = None

    def set_points(self, points_ptr, count):
        """points_ptr: raw C buffer of lv_point_t structs (int32 x, int32 y each)."""
        _lv_line_set_points(self.ptr, points_ptr, count)

    def set_points_list(self, points):
        """
        Convenience: accepts a list of (x, y) tuples.
        Allocates and keeps the buffer alive on the Line instance.
        """
        if self._points_buf:
            ffi.free(self._points_buf)
        n   = len(points)
        buf = ffi.malloc(n * 8)  # 2 × int32 per point
        for i in range(n):
            ffi.write_memory_with_offset(buf, i * 8,     ffi.c_int32, points[i][0])
            ffi.write_memory_with_offset(buf, i * 8 + 4, ffi.c_int32, points[i][1])
        self._points_buf = buf
        _lv_line_set_points(self.ptr, buf, n)

    def set_y_invert(self, en):
        _lv_line_set_y_invert(self.ptr, 1 if en else 0)

    def delete(self):
        if self._points_buf:
            ffi.free(self._points_buf)
            self._points_buf = None
        super().delete()


class Table(Obj):
    """
    Usage:
        t = Table(parent_ptr)
        t.set_column_count(3)
        t.set_row_count(4)
        t.set_cell_value(0, 0, "Name")
        t.set_column_width(0, 100)
        t.on(LV_EVENT_VALUE_CHANGED, lambda e: print(t.get_selected_cell()))
    """
    def __init__(self, parent_ptr):
        self.ptr        = _lv_table_create(parent_ptr)
        self._event_cbs = []
        self._styles    = []

    def set_cell_value(self, row, col, text):
        if isinstance(text, str):
            text = text.encode("utf-8")
        _lv_table_set_cell_value(self.ptr, row, col, text)

    def get_cell_value(self, row, col):
        raw = _lv_table_get_cell_value(self.ptr, row, col)
        return raw.decode("utf-8") if raw else ""

    def set_column_count(self, n):          _lv_table_set_column_count(self.ptr, n)
    def get_column_count(self):             return _lv_table_get_column_count(self.ptr)
    def set_row_count(self, n):             _lv_table_set_row_count(self.ptr, n)
    def get_row_count(self):                return _lv_table_get_row_count(self.ptr)
    def set_column_width(self, col, w):     _lv_table_set_column_width(self.ptr, col, w)

    def get_selected_cell(self):
        """Returns (row, col) tuple of the currently selected cell."""
        row_buf = ffi.malloc(4)
        col_buf = ffi.malloc(4)
        try:
            _lv_table_get_selected_cell(self.ptr, row_buf, col_buf)
            row_b = ffi.buffer_to_bytes(row_buf, 4)
            col_b = ffi.buffer_to_bytes(col_buf, 4)
            row = row_b[0] | (row_b[1] << 8) | (row_b[2] << 16) | (row_b[3] << 24)
            col = col_b[0] | (col_b[1] << 8) | (col_b[2] << 16) | (col_b[3] << 24)
            return (row, col)
        finally:
            ffi.free(row_buf)
            ffi.free(col_buf)


class List(Obj):
    """
    Usage:
        lst = List(parent_ptr)
        lst.set_size(200, 300)
        lst.add_text("Fruits")
        b1 = lst.add_button(None, "Apple")
        b2 = lst.add_button(None, "Banana")
        b1.on(LV_EVENT_CLICKED, lambda e: print("Apple!"))
    """
    def __init__(self, parent_ptr):
        self.ptr        = _lv_list_create(parent_ptr)
        self._event_cbs = []
        self._styles    = []

    def add_text(self, text):
        if isinstance(text, str):
            text = text.encode("utf-8")
        ptr = _lv_list_add_text(self.ptr, text)
        obj = Label.__new__(Label)
        obj.ptr        = ptr
        obj._event_cbs = []
        obj._styles    = []
        return obj

    def add_button(self, icon_ptr, text):
        if isinstance(text, str):
            text = text.encode("utf-8")
        icon = icon_ptr if icon_ptr is not None else ffi.c_void_p(0)
        ptr  = _lv_list_add_button(self.ptr, icon, text)
        obj  = Button.__new__(Button)
        obj.ptr        = ptr
        obj._event_cbs = []
        obj._styles    = []
        return obj

    def get_button_text(self, btn):
        ptr = btn.ptr if isinstance(btn, Obj) else btn
        raw = _lv_list_get_button_text(self.ptr, ptr)
        return raw.decode("utf-8") if raw else ""


class Msgbox(Obj):
    """
    Usage:
        mb = Msgbox(scr.ptr)
        mb.add_title("Warning")
        mb.add_text("Are you sure you want to delete this file?")
        ok  = mb.add_footer_button("OK")
        no  = mb.add_footer_button("Cancel")
        ok.on(LV_EVENT_CLICKED, lambda e: mb.close())
    """
    def __init__(self, parent_ptr):
        self.ptr        = _lv_msgbox_create(parent_ptr)
        self._event_cbs = []
        self._styles    = []

    def add_title(self, text):
        if isinstance(text, str):
            text = text.encode("utf-8")
        ptr = _lv_msgbox_add_title(self.ptr, text)
        obj = Label.__new__(Label)
        obj.ptr        = ptr
        obj._event_cbs = []
        obj._styles    = []
        return obj

    def add_text(self, text):
        if isinstance(text, str):
            text = text.encode("utf-8")
        ptr = _lv_msgbox_add_text(self.ptr, text)
        obj = Label.__new__(Label)
        obj.ptr        = ptr
        obj._event_cbs = []
        obj._styles    = []
        return obj

    def add_close_button(self):
        ptr = _lv_msgbox_add_close_button(self.ptr)
        obj = Button.__new__(Button)
        obj.ptr        = ptr
        obj._event_cbs = []
        obj._styles    = []
        return obj

    def add_footer_button(self, text):
        if isinstance(text, str):
            text = text.encode("utf-8")
        ptr = _lv_msgbox_add_footer_button(self.ptr, text)
        obj = Button.__new__(Button)
        obj.ptr        = ptr
        obj._event_cbs = []
        obj._styles    = []
        return obj

    def close(self):
        _lv_msgbox_close(self.ptr)
        self.ptr = None

# ==============================================================================
#  Display and Input
# ==============================================================================

class Display:
    """
    Wraps lv_display_t.

    Usage:
        disp = Display(800, 480)
        buf  = ffi.malloc(800 * 480 * 4)
        disp.set_buffers(buf, None, 800*480*4, LV_DISPLAY_RENDER_MODE_FULL)

        def my_flush(disp_ptr, area_ptr, px_map):
            # blit px_map to hardware
            Display.flush_ready_ptr(disp_ptr)

        disp.set_flush_cb(my_flush)
    """
    def __init__(self, w, h):
        self.ptr  = _lv_display_create(w, h)
        self.w    = w
        self.h    = h
        self._flush_cb = None

    def set_buffers(self, buf1, buf2, size, mode=LV_DISPLAY_RENDER_MODE_PARTIAL):
        b2 = buf2 if buf2 is not None else ffi.c_void_p(0)
        _lv_display_set_buffers(self.ptr, buf1, b2, size, mode)

    def set_flush_cb(self, fn):
        """fn(disp_ptr, area_ptr, px_map_ptr)"""
        cb = ffi.callback(fn, None, [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p])
        self._flush_cb = cb
        _lv_display_set_flush_cb(self.ptr, cb)

    @staticmethod
    def flush_ready_ptr(disp_ptr):
        """Call from inside flush callback to signal completion."""
        _lv_display_flush_ready(disp_ptr)

    def flush_ready(self):
        _lv_display_flush_ready(self.ptr)

    def set_resolution(self, w, h):
        _lv_display_set_resolution(self.ptr, w, h)

    def get_width(self):
        return _lv_display_get_horizontal_resolution(self.ptr)

    def get_height(self):
        return _lv_display_get_vertical_resolution(self.ptr)

    def set_rotation(self, rot):
        _lv_display_set_rotation(self.ptr, rot)

    def set_default(self):
        _lv_display_set_default(self.ptr)

    def get_screen_active(self):
        ptr = _lv_display_get_screen_active(self.ptr)
        obj = Obj.__new__(Obj)
        obj.ptr        = ptr
        obj._event_cbs = []
        obj._styles    = []
        return obj

    def delete(self):
        if self.ptr:
            _lv_display_delete(self.ptr)
            self.ptr = None


class InputDevice:
    """
    Wraps lv_indev_t.

    Usage:
        indev = InputDevice(LV_INDEV_TYPE_POINTER)
        indev.set_display(disp)

        def read_cb(indev_ptr, data_ptr):
            x, y = get_touch()
            InputDevice.write_pointer(data_ptr, x, y, LV_INDEV_STATE_PRESSED)

        indev.set_read_cb(read_cb)
    """
    def __init__(self, type_=LV_INDEV_TYPE_POINTER):
        self.ptr     = _lv_indev_create()
        self._read_cb = None
        _lv_indev_set_type(self.ptr, type_)

    def set_read_cb(self, fn):
        """fn(indev_ptr, data_ptr)"""
        cb = ffi.callback(fn, None, [ffi.c_void_p, ffi.c_void_p])
        self._read_cb = cb
        _lv_indev_set_read_cb(self.ptr, cb)

    def set_display(self, disp):
        ptr = disp.ptr if isinstance(disp, Display) else disp
        _lv_indev_set_display(self.ptr, ptr)

    def reset(self, obj_ptr=None):
        p = obj_ptr if obj_ptr is not None else ffi.c_void_p(0)
        _lv_indev_reset(self.ptr, p)

    @staticmethod
    def write_pointer(data_ptr, x, y, state):
        """
        Write x, y, state into lv_indev_data_t.
        Layout (LVGL v9, 64-bit): int32 x @ 0, int32 y @ 4, uint32 state @ 8
        """
        ffi.write_memory_with_offset(data_ptr,  0, ffi.c_int32,  x)
        ffi.write_memory_with_offset(data_ptr,  4, ffi.c_int32,  y)
        ffi.write_memory_with_offset(data_ptr,  8, ffi.c_uint32, state)

    def delete(self):
        if self.ptr:
            _lv_indev_delete(self.ptr)
            self.ptr = None

# ==============================================================================
#  Event helpers
# ==============================================================================

def event_get_code(event_ptr):
    return _lv_event_get_code(event_ptr)

def event_get_target(event_ptr):
    return _lv_event_get_target(event_ptr)

def event_get_current_target(event_ptr):
    return _lv_event_get_current_target(event_ptr)

def event_get_user_data(event_ptr):
    return _lv_event_get_user_data(event_ptr)

def event_get_param(event_ptr):
    return _lv_event_get_param(event_ptr)

def event_stop_bubbling(event_ptr):
    _lv_event_stop_bubbling(event_ptr)

def event_stop_processing(event_ptr):
    _lv_event_stop_processing(event_ptr)

# ==============================================================================
#  Screen helpers
# ==============================================================================

def screen_active():
    """Return the active screen as an Obj wrapper."""
    ptr = _lv_screen_active()
    obj = Obj.__new__(Obj)
    obj.ptr        = ptr
    obj._event_cbs = []
    obj._styles    = []
    return obj

def screen_load(scr):
    ptr = scr.ptr if isinstance(scr, Obj) else scr
    _lv_screen_load(ptr)

def screen_load_anim(scr, anim_type, time_ms, delay_ms, auto_del=False):
    ptr = scr.ptr if isinstance(scr, Obj) else scr
    _lv_screen_load_anim(ptr, anim_type, time_ms, delay_ms, 1 if auto_del else 0)

def new_screen():
    """Create a blank new screen object."""
    obj = Obj.__new__(Obj)
    obj.ptr        = _lv_obj_create(ffi.c_void_p(0))
    obj._event_cbs = []
    obj._styles    = []
    return obj

# ==============================================================================
#  Core helpers
# ==============================================================================

def init():
    """Initialize the LVGL library. Call once at startup."""
    _lv_init()

def deinit():
    """De-initialize LVGL."""
    _lv_deinit()

def tick_inc(ms):
    """Inform LVGL that `ms` milliseconds have elapsed. Call regularly."""
    _lv_tick_inc(ms)

def tick_get():
    """Return the current LVGL tick count in milliseconds."""
    return _lv_tick_get()

def timer_handler():
    """
    Process all pending LVGL tasks (timers, redraws, input).
    Call as frequently as possible from the main loop.
    Returns the time until the next call is needed (ms).
    """
    return _lv_timer_handler()

def refr_now(disp=None):
    """Force an immediate display refresh."""
    ptr = disp.ptr if isinstance(disp, Display) else (disp if disp else ffi.c_void_p(0))
    _lv_refr_now(ptr)

def get_default_display():
    """Return the default Display object."""
    ptr  = _lv_display_get_default()
    disp = Display.__new__(Display)
    disp.ptr       = ptr
    disp._flush_cb = None
    return disp