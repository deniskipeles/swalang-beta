"""
Full SDL2 wrapper for Swalang/Pylearn.

Subsystems covered:
  Core · Window · Renderer · Surface · Texture · Events · Keyboard ·
  Mouse · Touch · Audio · Timer · Clipboard · Cursor · OpenGL ·
  Display info · LVGL backend · Nuklear backend hooks
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
        "bin/x86_64-linux/sdl2/libSDL2-2.0.so",
        "bin/x86_64-linux/sdl2/libSDL2.so",
        "libSDL2-2.0.so.0",
        "libSDL2.so",
    ])
elif sys.platform == "windows":
    _lib = _try_load([
        "bin/x86_64-windows-gnu/sdl2/SDL2.dll",
        "SDL2.dll",
    ])
else:
    _lib = _try_load(["libSDL2.dylib", "libSDL2-2.0.dylib"])

if _lib is None:
    raise ffi.FFIError("Could not load SDL2 shared library.")

# ==============================================================================
#  Constants
# ==============================================================================

# --- Init flags ---
SDL_INIT_TIMER          = 0x00000001
SDL_INIT_AUDIO          = 0x00000010
SDL_INIT_VIDEO          = 0x00000020
SDL_INIT_JOYSTICK       = 0x00000200
SDL_INIT_HAPTIC         = 0x00001000
SDL_INIT_GAMECONTROLLER = 0x00002000
SDL_INIT_EVENTS         = 0x00004000
SDL_INIT_SENSOR         = 0x00008000
SDL_INIT_EVERYTHING     = 0x0000FFFF

# --- Window flags ---
SDL_WINDOW_FULLSCREEN         = 0x00000001
SDL_WINDOW_OPENGL             = 0x00000002
SDL_WINDOW_SHOWN              = 0x00000004
SDL_WINDOW_HIDDEN             = 0x00000008
SDL_WINDOW_BORDERLESS         = 0x00000010
SDL_WINDOW_RESIZABLE          = 0x00000020
SDL_WINDOW_MINIMIZED          = 0x00000040
SDL_WINDOW_MAXIMIZED          = 0x00000080
SDL_WINDOW_MOUSE_GRABBED      = 0x00000100
SDL_WINDOW_INPUT_FOCUS        = 0x00000200
SDL_WINDOW_MOUSE_FOCUS        = 0x00000400
SDL_WINDOW_FULLSCREEN_DESKTOP = 0x00001001
SDL_WINDOW_ALLOW_HIGHDPI      = 0x00002000
SDL_WINDOW_MOUSE_CAPTURE      = 0x00004000
SDL_WINDOW_ALWAYS_ON_TOP      = 0x00008000
SDL_WINDOW_SKIP_TASKBAR       = 0x00010000
SDL_WINDOW_UTILITY            = 0x00020000
SDL_WINDOW_TOOLTIP            = 0x00040000
SDL_WINDOW_POPUP_MENU         = 0x00080000
SDL_WINDOW_VULKAN             = 0x10000000

SDL_WINDOWPOS_UNDEFINED = 0x1FFF0000
SDL_WINDOWPOS_CENTERED  = 0x2FFF0000

# --- Window events ---
SDL_WINDOWEVENT_NONE         = 0
SDL_WINDOWEVENT_SHOWN        = 1
SDL_WINDOWEVENT_HIDDEN       = 2
SDL_WINDOWEVENT_EXPOSED      = 3
SDL_WINDOWEVENT_MOVED        = 4
SDL_WINDOWEVENT_RESIZED      = 5
SDL_WINDOWEVENT_SIZE_CHANGED = 6
SDL_WINDOWEVENT_MINIMIZED    = 7
SDL_WINDOWEVENT_MAXIMIZED    = 8
SDL_WINDOWEVENT_RESTORED     = 9
SDL_WINDOWEVENT_ENTER        = 10
SDL_WINDOWEVENT_LEAVE        = 11
SDL_WINDOWEVENT_FOCUS_GAINED = 12
SDL_WINDOWEVENT_FOCUS_LOST   = 13
SDL_WINDOWEVENT_CLOSE        = 14
SDL_WINDOWEVENT_TAKE_FOCUS   = 15
SDL_WINDOWEVENT_HIT_TEST     = 16

# --- Renderer flags ---
SDL_RENDERER_SOFTWARE      = 0x00000001
SDL_RENDERER_ACCELERATED   = 0x00000002
SDL_RENDERER_PRESENTVSYNC  = 0x00000004
SDL_RENDERER_TARGETTEXTURE = 0x00000008

# --- Texture access ---
SDL_TEXTUREACCESS_STATIC    = 0
SDL_TEXTUREACCESS_STREAMING = 1
SDL_TEXTUREACCESS_TARGET    = 2

# --- Pixel formats ---
SDL_PIXELFORMAT_UNKNOWN  = 0
SDL_PIXELFORMAT_RGB332   = 0x14110801
SDL_PIXELFORMAT_RGB444   = 0x15120C02
SDL_PIXELFORMAT_RGB555   = 0x15130F02
SDL_PIXELFORMAT_BGR555   = 0x15530F02
SDL_PIXELFORMAT_ARGB4444 = 0x15321002
SDL_PIXELFORMAT_RGBA4444 = 0x15421002
SDL_PIXELFORMAT_ABGR4444 = 0x15721002
SDL_PIXELFORMAT_BGRA4444 = 0x15821002
SDL_PIXELFORMAT_ARGB1555 = 0x15331002
SDL_PIXELFORMAT_RGBA5551 = 0x15441002
SDL_PIXELFORMAT_ABGR1555 = 0x15731002
SDL_PIXELFORMAT_BGRA5551 = 0x15841002
SDL_PIXELFORMAT_RGB565   = 0x15151002
SDL_PIXELFORMAT_BGR565   = 0x15551002
SDL_PIXELFORMAT_RGB24    = 0x17101803
SDL_PIXELFORMAT_BGR24    = 0x17401803
SDL_PIXELFORMAT_RGB888   = 0x16161804
SDL_PIXELFORMAT_RGBX8888 = 0x16261804
SDL_PIXELFORMAT_BGR888   = 0x16561804
SDL_PIXELFORMAT_BGRX8888 = 0x16661804
SDL_PIXELFORMAT_ARGB8888 = 0x16362004
SDL_PIXELFORMAT_RGBA8888 = 0x16462004
SDL_PIXELFORMAT_ABGR8888 = 0x16762004
SDL_PIXELFORMAT_BGRA8888 = 0x16862004

# --- Blend modes ---
SDL_BLENDMODE_NONE    = 0x00000000
SDL_BLENDMODE_BLEND   = 0x00000001
SDL_BLENDMODE_ADD     = 0x00000002
SDL_BLENDMODE_MOD     = 0x00000004
SDL_BLENDMODE_MUL     = 0x00000008

# --- Flip ---
SDL_FLIP_NONE       = 0
SDL_FLIP_HORIZONTAL = 1
SDL_FLIP_VERTICAL   = 2

# --- Event types ---
SDL_QUIT           = 0x100
SDL_DISPLAYEVENT   = 0x150
SDL_WINDOWEVENT    = 0x200
SDL_SYSWMEVENT     = 0x201
SDL_KEYDOWN        = 0x300
SDL_KEYUP          = 0x301
SDL_TEXTEDITING    = 0x302
SDL_TEXTINPUT      = 0x303
SDL_KEYMAPCHANGED  = 0x304
SDL_MOUSEMOTION    = 0x400
SDL_MOUSEBUTTONDOWN = 0x401
SDL_MOUSEBUTTONUP  = 0x402
SDL_MOUSEWHEEL     = 0x403
SDL_JOYAXISMOTION  = 0x600
SDL_JOYBALLMOTION  = 0x601
SDL_JOYHATMOTION   = 0x602
SDL_JOYBUTTONDOWN  = 0x603
SDL_JOYBUTTONUP    = 0x604
SDL_JOYDEVICEADDED   = 0x605
SDL_JOYDEVICEREMOVED = 0x606
SDL_CONTROLLERAXISMOTION     = 0x650
SDL_CONTROLLERBUTTONDOWN     = 0x651
SDL_CONTROLLERBUTTONUP       = 0x652
SDL_CONTROLLERDEVICEADDED    = 0x653
SDL_CONTROLLERDEVICEREMOVED  = 0x654
SDL_CONTROLLERDEVICEREMAPPED = 0x655
SDL_FINGERDOWN   = 0x700
SDL_FINGERUP     = 0x701
SDL_FINGERMOTION = 0x702
SDL_DOLLARGESTURE = 0x800
SDL_MULTIGESTURE  = 0x802
SDL_CLIPBOARDUPDATE = 0x900
SDL_DROPFILE      = 0x1000
SDL_DROPTEXT      = 0x1001
SDL_DROPBEGIN     = 0x1002
SDL_DROPCOMPLETE  = 0x1003
SDL_AUDIODEVICEADDED   = 0x1100
SDL_AUDIODEVICEREMOVED = 0x1101
SDL_SENSORUPDATE  = 0x1200
SDL_USEREVENT     = 0x8000
SDL_LASTEVENT     = 0xFFFF

# --- Mouse buttons ---
SDL_BUTTON_LEFT   = 1
SDL_BUTTON_MIDDLE = 2
SDL_BUTTON_RIGHT  = 3
SDL_BUTTON_X1     = 4
SDL_BUTTON_X2     = 5

SDL_BUTTON_LMASK  = 1 << (SDL_BUTTON_LEFT   - 1)
SDL_BUTTON_MMASK  = 1 << (SDL_BUTTON_MIDDLE - 1)
SDL_BUTTON_RMASK  = 1 << (SDL_BUTTON_RIGHT  - 1)
SDL_BUTTON_X1MASK = 1 << (SDL_BUTTON_X1     - 1)
SDL_BUTTON_X2MASK = 1 << (SDL_BUTTON_X2     - 1)

SDL_MOUSEWHEEL_NORMAL  = 0
SDL_MOUSEWHEEL_FLIPPED = 1

# --- Keyboard scancodes (most common) ---
SDL_SCANCODE_UNKNOWN = 0
SDL_SCANCODE_A = 4;  SDL_SCANCODE_B = 5;  SDL_SCANCODE_C = 6
SDL_SCANCODE_D = 7;  SDL_SCANCODE_E = 8;  SDL_SCANCODE_F = 9
SDL_SCANCODE_G = 10; SDL_SCANCODE_H = 11; SDL_SCANCODE_I = 12
SDL_SCANCODE_J = 13; SDL_SCANCODE_K = 14; SDL_SCANCODE_L = 15
SDL_SCANCODE_M = 16; SDL_SCANCODE_N = 17; SDL_SCANCODE_O = 18
SDL_SCANCODE_P = 19; SDL_SCANCODE_Q = 20; SDL_SCANCODE_R = 21
SDL_SCANCODE_S = 22; SDL_SCANCODE_T = 23; SDL_SCANCODE_U = 24
SDL_SCANCODE_V = 25; SDL_SCANCODE_W = 26; SDL_SCANCODE_X = 27
SDL_SCANCODE_Y = 28; SDL_SCANCODE_Z = 29
SDL_SCANCODE_1 = 30; SDL_SCANCODE_2 = 31; SDL_SCANCODE_3 = 32
SDL_SCANCODE_4 = 33; SDL_SCANCODE_5 = 34; SDL_SCANCODE_6 = 35
SDL_SCANCODE_7 = 36; SDL_SCANCODE_8 = 37; SDL_SCANCODE_9 = 38
SDL_SCANCODE_0 = 39
SDL_SCANCODE_RETURN    = 40
SDL_SCANCODE_ESCAPE    = 41
SDL_SCANCODE_BACKSPACE = 42
SDL_SCANCODE_TAB       = 43
SDL_SCANCODE_SPACE     = 44
SDL_SCANCODE_MINUS     = 45
SDL_SCANCODE_EQUALS    = 46
SDL_SCANCODE_LEFTBRACKET  = 47
SDL_SCANCODE_RIGHTBRACKET = 48
SDL_SCANCODE_BACKSLASH = 49
SDL_SCANCODE_SEMICOLON = 51
SDL_SCANCODE_APOSTROPHE = 52
SDL_SCANCODE_GRAVE     = 53
SDL_SCANCODE_COMMA     = 54
SDL_SCANCODE_PERIOD    = 55
SDL_SCANCODE_SLASH     = 56
SDL_SCANCODE_CAPSLOCK  = 57
SDL_SCANCODE_F1  = 58; SDL_SCANCODE_F2  = 59; SDL_SCANCODE_F3  = 60
SDL_SCANCODE_F4  = 61; SDL_SCANCODE_F5  = 62; SDL_SCANCODE_F6  = 63
SDL_SCANCODE_F7  = 64; SDL_SCANCODE_F8  = 65; SDL_SCANCODE_F9  = 66
SDL_SCANCODE_F10 = 67; SDL_SCANCODE_F11 = 68; SDL_SCANCODE_F12 = 69
SDL_SCANCODE_PRINTSCREEN = 70
SDL_SCANCODE_SCROLLLOCK  = 71
SDL_SCANCODE_PAUSE       = 72
SDL_SCANCODE_INSERT      = 73
SDL_SCANCODE_HOME        = 74
SDL_SCANCODE_PAGEUP      = 75
SDL_SCANCODE_DELETE      = 76
SDL_SCANCODE_END         = 77
SDL_SCANCODE_PAGEDOWN    = 78
SDL_SCANCODE_RIGHT       = 79
SDL_SCANCODE_LEFT        = 80
SDL_SCANCODE_DOWN        = 81
SDL_SCANCODE_UP          = 82
SDL_SCANCODE_KP_DIVIDE   = 84
SDL_SCANCODE_KP_MULTIPLY = 85
SDL_SCANCODE_KP_MINUS    = 86
SDL_SCANCODE_KP_PLUS     = 87
SDL_SCANCODE_KP_ENTER    = 88
SDL_SCANCODE_KP_1 = 89; SDL_SCANCODE_KP_2 = 90; SDL_SCANCODE_KP_3 = 91
SDL_SCANCODE_KP_4 = 92; SDL_SCANCODE_KP_5 = 93; SDL_SCANCODE_KP_6 = 94
SDL_SCANCODE_KP_7 = 95; SDL_SCANCODE_KP_8 = 96; SDL_SCANCODE_KP_9 = 97
SDL_SCANCODE_KP_0 = 98
SDL_SCANCODE_LCTRL  = 224; SDL_SCANCODE_LSHIFT = 225
SDL_SCANCODE_LALT   = 226; SDL_SCANCODE_LGUI   = 227
SDL_SCANCODE_RCTRL  = 228; SDL_SCANCODE_RSHIFT = 229
SDL_SCANCODE_RALT   = 230; SDL_SCANCODE_RGUI   = 231

# --- Key modifiers ---
KMOD_NONE   = 0x0000
KMOD_LSHIFT = 0x0001; KMOD_RSHIFT = 0x0002
KMOD_LCTRL  = 0x0040; KMOD_RCTRL  = 0x0080
KMOD_LALT   = 0x0100; KMOD_RALT   = 0x0200
KMOD_LGUI   = 0x0400; KMOD_RGUI   = 0x0800
KMOD_NUM    = 0x1000; KMOD_CAPS   = 0x2000
KMOD_SCROLL = 0x8000
KMOD_SHIFT  = KMOD_LSHIFT | KMOD_RSHIFT
KMOD_CTRL   = KMOD_LCTRL  | KMOD_RCTRL
KMOD_ALT    = KMOD_LALT   | KMOD_RALT
KMOD_GUI    = KMOD_LGUI   | KMOD_RGUI

# --- Audio formats ---
AUDIO_U8     = 0x0008
AUDIO_S8     = 0x8008
AUDIO_U16LSB = 0x0010
AUDIO_S16LSB = 0x8010
AUDIO_U16MSB = 0x1010
AUDIO_S16MSB = 0x9010
AUDIO_S32LSB = 0x8020
AUDIO_S32MSB = 0x9020
AUDIO_F32LSB = 0x8120
AUDIO_F32MSB = 0x9120
AUDIO_S16    = AUDIO_S16LSB
AUDIO_S32    = AUDIO_S32LSB
AUDIO_F32    = AUDIO_F32LSB

# --- GL attributes ---
SDL_GL_RED_SIZE           = 0
SDL_GL_GREEN_SIZE         = 1
SDL_GL_BLUE_SIZE          = 2
SDL_GL_ALPHA_SIZE         = 3
SDL_GL_BUFFER_SIZE        = 4
SDL_GL_DOUBLEBUFFER       = 5
SDL_GL_DEPTH_SIZE         = 6
SDL_GL_STENCIL_SIZE       = 7
SDL_GL_CONTEXT_MAJOR_VERSION = 17
SDL_GL_CONTEXT_MINOR_VERSION = 18
SDL_GL_CONTEXT_PROFILE_MASK = 21
SDL_GL_CONTEXT_PROFILE_CORE          = 0x0001
SDL_GL_CONTEXT_PROFILE_COMPATIBILITY = 0x0002
SDL_GL_CONTEXT_PROFILE_ES            = 0x0004

# --- Hat positions (joystick) ---
SDL_HAT_CENTERED  = 0x00
SDL_HAT_UP        = 0x01
SDL_HAT_RIGHT     = 0x02
SDL_HAT_DOWN      = 0x04
SDL_HAT_LEFT      = 0x08
SDL_HAT_RIGHTUP   = SDL_HAT_RIGHT | SDL_HAT_UP
SDL_HAT_RIGHTDOWN = SDL_HAT_RIGHT | SDL_HAT_DOWN
SDL_HAT_LEFTUP    = SDL_HAT_LEFT  | SDL_HAT_UP
SDL_HAT_LEFTDOWN  = SDL_HAT_LEFT  | SDL_HAT_DOWN

# --- Hit test ---
SDL_HITTEST_NORMAL          = 0
SDL_HITTEST_DRAGGABLE       = 1
SDL_HITTEST_RESIZE_TOPLEFT  = 2
SDL_HITTEST_RESIZE_TOP      = 3
SDL_HITTEST_RESIZE_TOPRIGHT = 4
SDL_HITTEST_RESIZE_RIGHT    = 5
SDL_HITTEST_RESIZE_BOTTOMRIGHT = 6
SDL_HITTEST_RESIZE_BOTTOM   = 7
SDL_HITTEST_RESIZE_BOTTOMLEFT = 8
SDL_HITTEST_RESIZE_LEFT     = 9

# Struct / buffer sizes
_SZ_EVENT       = 56
_SZ_AUDIOSPEC   = 40
_SZ_DISPLAYMODE = 28
_SZ_RECT        = 16
_SZ_FRECT       = 16
_SZ_POINT       = 8
_SZ_FPOINT      = 8
_SZ_COLOR       = 4
_SZ_VERTEX      = 20   # SDL_Vertex: FPoint pos, Color color, FPoint tex_coord

# ==============================================================================
#  SDL_Rect / SDL_FRect helpers (pure Python — no FFI needed)
# ==============================================================================

def make_rect(x, y, w, h):
    """Allocate and fill an SDL_Rect buffer. Caller must ffi.free() it."""
    buf = ffi.malloc(_SZ_RECT)
    ffi.write_memory_with_offset(buf,  0, ffi.c_int32, x)
    ffi.write_memory_with_offset(buf,  4, ffi.c_int32, y)
    ffi.write_memory_with_offset(buf,  8, ffi.c_int32, w)
    ffi.write_memory_with_offset(buf, 12, ffi.c_int32, h)
    return buf

def make_frect(x, y, w, h):
    """Allocate and fill an SDL_FRect buffer. Caller must ffi.free() it."""
    buf = ffi.malloc(_SZ_FRECT)
    ffi.write_memory_with_offset(buf,  0, ffi.c_float, x)
    ffi.write_memory_with_offset(buf,  4, ffi.c_float, y)
    ffi.write_memory_with_offset(buf,  8, ffi.c_float, w)
    ffi.write_memory_with_offset(buf, 12, ffi.c_float, h)
    return buf

def make_point(x, y):
    buf = ffi.malloc(_SZ_POINT)
    ffi.write_memory_with_offset(buf, 0, ffi.c_int32, x)
    ffi.write_memory_with_offset(buf, 4, ffi.c_int32, y)
    return buf

def _read_i32(buf, off):
    return ffi.read_memory_with_offset(buf, off, ffi.c_int32)

def _read_u32(buf, off):
    return ffi.read_memory_with_offset(buf, off, ffi.c_uint32)

def _read_u16(buf, off):
    return ffi.read_memory_with_offset(buf, off, ffi.c_uint16)

def _read_u8(buf, off):
    return ffi.read_memory_with_offset(buf, off, ffi.c_uint8)

def _read_f32(buf, off):
    return ffi.read_memory_with_offset(buf, off, ffi.c_float)

# ==============================================================================
#  C Function Bindings
# ==============================================================================

# --- Core ---
_SDL_Init          = _lib.SDL_Init([ffi.c_uint32], ffi.c_int32)
_SDL_InitSubSystem = _lib.SDL_InitSubSystem([ffi.c_uint32], ffi.c_int32)
_SDL_QuitSubSystem = _lib.SDL_QuitSubSystem([ffi.c_uint32], None)
_SDL_Quit          = _lib.SDL_Quit([], None)
_SDL_WasInit       = _lib.SDL_WasInit([ffi.c_uint32], ffi.c_uint32)
_SDL_GetError      = _lib.SDL_GetError([], ffi.c_char_p)
_SDL_ClearError    = _lib.SDL_ClearError([], None)
_SDL_SetHint       = _lib.SDL_SetHint([ffi.c_char_p, ffi.c_char_p], ffi.c_int32)
_SDL_GetHint       = _lib.SDL_GetHint([ffi.c_char_p], ffi.c_char_p)
_SDL_GetVersion    = _lib.SDL_GetVersion([ffi.c_void_p], None)
_SDL_GetRevision   = _lib.SDL_GetRevision([], ffi.c_char_p)

# --- Timer ---
_SDL_GetTicks      = _lib.SDL_GetTicks([], ffi.c_uint32)
_SDL_GetTicks64    = _lib.SDL_GetTicks64([], ffi.c_uint64)
_SDL_GetPerformanceCounter   = _lib.SDL_GetPerformanceCounter([], ffi.c_uint64)
_SDL_GetPerformanceFrequency = _lib.SDL_GetPerformanceFrequency([], ffi.c_uint64)
_SDL_Delay         = _lib.SDL_Delay([ffi.c_uint32], None)
_SDL_AddTimer      = _lib.SDL_AddTimer([ffi.c_uint32, ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_SDL_RemoveTimer   = _lib.SDL_RemoveTimer([ffi.c_int32], ffi.c_int32)

# --- Display ---
_SDL_GetNumVideoDisplays      = _lib.SDL_GetNumVideoDisplays([], ffi.c_int32)
_SDL_GetDisplayName           = _lib.SDL_GetDisplayName([ffi.c_int32], ffi.c_char_p)
_SDL_GetCurrentDisplayMode    = _lib.SDL_GetCurrentDisplayMode(
    [ffi.c_int32, ffi.c_void_p], ffi.c_int32)
_SDL_GetDesktopDisplayMode    = _lib.SDL_GetDesktopDisplayMode(
    [ffi.c_int32, ffi.c_void_p], ffi.c_int32)
_SDL_GetDisplayBounds         = _lib.SDL_GetDisplayBounds(
    [ffi.c_int32, ffi.c_void_p], ffi.c_int32)
_SDL_GetNumDisplayModes       = _lib.SDL_GetNumDisplayModes([ffi.c_int32], ffi.c_int32)
_SDL_GetDisplayMode           = _lib.SDL_GetDisplayMode(
    [ffi.c_int32, ffi.c_int32, ffi.c_void_p], ffi.c_int32)
_SDL_GetDisplayDPI            = _lib.SDL_GetDisplayDPI(
    [ffi.c_int32, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], ffi.c_int32)

# --- Window ---
_SDL_CreateWindow    = _lib.SDL_CreateWindow(
    [ffi.c_char_p, ffi.c_int32, ffi.c_int32, ffi.c_int32, ffi.c_int32, ffi.c_uint32],
    ffi.c_void_p)
_SDL_CreateWindowFrom = _lib.SDL_CreateWindowFrom([ffi.c_void_p], ffi.c_void_p)
_SDL_DestroyWindow   = _lib.SDL_DestroyWindow([ffi.c_void_p], None)
_SDL_GetWindowID     = _lib.SDL_GetWindowID([ffi.c_void_p], ffi.c_uint32)
_SDL_GetWindowFromID = _lib.SDL_GetWindowFromID([ffi.c_uint32], ffi.c_void_p)
_SDL_GetWindowTitle  = _lib.SDL_GetWindowTitle([ffi.c_void_p], ffi.c_char_p)
_SDL_SetWindowTitle  = _lib.SDL_SetWindowTitle([ffi.c_void_p, ffi.c_char_p], None)
_SDL_GetWindowSize   = _lib.SDL_GetWindowSize([ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], None)
_SDL_SetWindowSize   = _lib.SDL_SetWindowSize([ffi.c_void_p, ffi.c_int32, ffi.c_int32], None)
_SDL_GetWindowPosition  = _lib.SDL_GetWindowPosition(
    [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], None)
_SDL_SetWindowPosition  = _lib.SDL_SetWindowPosition(
    [ffi.c_void_p, ffi.c_int32, ffi.c_int32], None)
_SDL_GetWindowFlags     = _lib.SDL_GetWindowFlags([ffi.c_void_p], ffi.c_uint32)
_SDL_SetWindowFullscreen = _lib.SDL_SetWindowFullscreen(
    [ffi.c_void_p, ffi.c_uint32], ffi.c_int32)
_SDL_ShowWindow      = _lib.SDL_ShowWindow([ffi.c_void_p], None)
_SDL_HideWindow      = _lib.SDL_HideWindow([ffi.c_void_p], None)
_SDL_RaiseWindow     = _lib.SDL_RaiseWindow([ffi.c_void_p], None)
_SDL_MinimizeWindow  = _lib.SDL_MinimizeWindow([ffi.c_void_p], None)
_SDL_MaximizeWindow  = _lib.SDL_MaximizeWindow([ffi.c_void_p], None)
_SDL_RestoreWindow   = _lib.SDL_RestoreWindow([ffi.c_void_p], None)
_SDL_SetWindowBordered  = _lib.SDL_SetWindowBordered([ffi.c_void_p, ffi.c_int32], None)
_SDL_SetWindowResizable = _lib.SDL_SetWindowResizable([ffi.c_void_p, ffi.c_int32], None)
_SDL_SetWindowOpacity   = _lib.SDL_SetWindowOpacity([ffi.c_void_p, ffi.c_float], ffi.c_int32)
_SDL_GetWindowOpacity   = _lib.SDL_GetWindowOpacity(
    [ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_SDL_SetWindowMinimumSize = _lib.SDL_SetWindowMinimumSize(
    [ffi.c_void_p, ffi.c_int32, ffi.c_int32], None)
_SDL_SetWindowMaximumSize = _lib.SDL_SetWindowMaximumSize(
    [ffi.c_void_p, ffi.c_int32, ffi.c_int32], None)
_SDL_GetWindowDisplayIndex = _lib.SDL_GetWindowDisplayIndex([ffi.c_void_p], ffi.c_int32)
_SDL_SetWindowDisplayMode  = _lib.SDL_SetWindowDisplayMode(
    [ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_SDL_GetWindowSurface   = _lib.SDL_GetWindowSurface([ffi.c_void_p], ffi.c_void_p)
_SDL_UpdateWindowSurface = _lib.SDL_UpdateWindowSurface([ffi.c_void_p], ffi.c_int32)
_SDL_SetWindowIcon      = _lib.SDL_SetWindowIcon([ffi.c_void_p, ffi.c_void_p], None)
_SDL_WarpMouseInWindow  = _lib.SDL_WarpMouseInWindow(
    [ffi.c_void_p, ffi.c_int32, ffi.c_int32], None)
_SDL_SetWindowGrab      = _lib.SDL_SetWindowGrab([ffi.c_void_p, ffi.c_int32], None)
_SDL_GetWindowGrab      = _lib.SDL_GetWindowGrab([ffi.c_void_p], ffi.c_int32)
_SDL_FlashWindow        = _lib.SDL_FlashWindow([ffi.c_void_p, ffi.c_int32], ffi.c_int32)

SDL_FLASH_CANCEL         = 0
SDL_FLASH_BRIEFLY        = 1
SDL_FLASH_UNTIL_FOCUSED  = 2

# --- Renderer ---
_SDL_CreateRenderer         = _lib.SDL_CreateRenderer(
    [ffi.c_void_p, ffi.c_int32, ffi.c_uint32], ffi.c_void_p)
_SDL_CreateSoftwareRenderer = _lib.SDL_CreateSoftwareRenderer(
    [ffi.c_void_p], ffi.c_void_p)
_SDL_DestroyRenderer        = _lib.SDL_DestroyRenderer([ffi.c_void_p], None)
_SDL_GetRenderer            = _lib.SDL_GetRenderer([ffi.c_void_p], ffi.c_void_p)
_SDL_GetNumRenderDrivers    = _lib.SDL_GetNumRenderDrivers([], ffi.c_int32)
_SDL_RenderSetLogicalSize   = _lib.SDL_RenderSetLogicalSize(
    [ffi.c_void_p, ffi.c_int32, ffi.c_int32], ffi.c_int32)
_SDL_RenderGetLogicalSize   = _lib.SDL_RenderGetLogicalSize(
    [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], None)
_SDL_RenderSetScale         = _lib.SDL_RenderSetScale(
    [ffi.c_void_p, ffi.c_float, ffi.c_float], ffi.c_int32)
_SDL_RenderGetScale         = _lib.SDL_RenderGetScale(
    [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], None)
_SDL_RenderSetViewport      = _lib.SDL_RenderSetViewport(
    [ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_SDL_RenderGetViewport      = _lib.SDL_RenderGetViewport(
    [ffi.c_void_p, ffi.c_void_p], None)
_SDL_RenderSetClipRect      = _lib.SDL_RenderSetClipRect(
    [ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_SDL_RenderGetClipRect      = _lib.SDL_RenderGetClipRect(
    [ffi.c_void_p, ffi.c_void_p], None)
_SDL_RenderIsClipEnabled    = _lib.SDL_RenderIsClipEnabled([ffi.c_void_p], ffi.c_int32)
_SDL_GetRendererOutputSize  = _lib.SDL_GetRendererOutputSize(
    [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_SDL_SetRenderDrawColor     = _lib.SDL_SetRenderDrawColor(
    [ffi.c_void_p, ffi.c_uint8, ffi.c_uint8, ffi.c_uint8, ffi.c_uint8], ffi.c_int32)
_SDL_GetRenderDrawColor     = _lib.SDL_GetRenderDrawColor(
    [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_SDL_SetRenderDrawBlendMode = _lib.SDL_SetRenderDrawBlendMode(
    [ffi.c_void_p, ffi.c_int32], ffi.c_int32)
_SDL_GetRenderDrawBlendMode = _lib.SDL_GetRenderDrawBlendMode(
    [ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_SDL_RenderClear            = _lib.SDL_RenderClear([ffi.c_void_p], ffi.c_int32)
_SDL_RenderPresent          = _lib.SDL_RenderPresent([ffi.c_void_p], None)
_SDL_RenderDrawPoint        = _lib.SDL_RenderDrawPoint(
    [ffi.c_void_p, ffi.c_int32, ffi.c_int32], ffi.c_int32)
_SDL_RenderDrawPoints       = _lib.SDL_RenderDrawPoints(
    [ffi.c_void_p, ffi.c_void_p, ffi.c_int32], ffi.c_int32)
_SDL_RenderDrawLine         = _lib.SDL_RenderDrawLine(
    [ffi.c_void_p, ffi.c_int32, ffi.c_int32, ffi.c_int32, ffi.c_int32], ffi.c_int32)
_SDL_RenderDrawLines        = _lib.SDL_RenderDrawLines(
    [ffi.c_void_p, ffi.c_void_p, ffi.c_int32], ffi.c_int32)
_SDL_RenderDrawRect         = _lib.SDL_RenderDrawRect(
    [ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_SDL_RenderDrawRects        = _lib.SDL_RenderDrawRects(
    [ffi.c_void_p, ffi.c_void_p, ffi.c_int32], ffi.c_int32)
_SDL_RenderFillRect         = _lib.SDL_RenderFillRect(
    [ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_SDL_RenderFillRects        = _lib.SDL_RenderFillRects(
    [ffi.c_void_p, ffi.c_void_p, ffi.c_int32], ffi.c_int32)
_SDL_RenderCopy             = _lib.SDL_RenderCopy(
    [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_SDL_RenderCopyEx           = _lib.SDL_RenderCopyEx(
    [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p,
     ffi.c_double, ffi.c_void_p, ffi.c_int32], ffi.c_int32)
_SDL_RenderReadPixels       = _lib.SDL_RenderReadPixels(
    [ffi.c_void_p, ffi.c_void_p, ffi.c_uint32,
     ffi.c_void_p, ffi.c_int32], ffi.c_int32)
_SDL_SetRenderTarget        = _lib.SDL_SetRenderTarget(
    [ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_SDL_GetRenderTarget        = _lib.SDL_GetRenderTarget([ffi.c_void_p], ffi.c_void_p)
_SDL_RenderGeometry         = _lib.SDL_RenderGeometry(
    [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p,
     ffi.c_int32, ffi.c_void_p, ffi.c_int32], ffi.c_int32)
_SDL_RenderFlush            = _lib.SDL_RenderFlush([ffi.c_void_p], ffi.c_int32)

# --- Texture ---
_SDL_CreateTexture       = _lib.SDL_CreateTexture(
    [ffi.c_void_p, ffi.c_uint32, ffi.c_int32, ffi.c_int32, ffi.c_int32], ffi.c_void_p)
_SDL_CreateTextureFromSurface = _lib.SDL_CreateTextureFromSurface(
    [ffi.c_void_p, ffi.c_void_p], ffi.c_void_p)
_SDL_DestroyTexture      = _lib.SDL_DestroyTexture([ffi.c_void_p], None)
_SDL_QueryTexture        = _lib.SDL_QueryTexture(
    [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p,
     ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_SDL_UpdateTexture       = _lib.SDL_UpdateTexture(
    [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_int32], ffi.c_int32)
_SDL_UpdateYUVTexture    = _lib.SDL_UpdateYUVTexture(
    [ffi.c_void_p, ffi.c_void_p,
     ffi.c_void_p, ffi.c_int32,
     ffi.c_void_p, ffi.c_int32,
     ffi.c_void_p, ffi.c_int32], ffi.c_int32)
_SDL_LockTexture         = _lib.SDL_LockTexture(
    [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_SDL_UnlockTexture       = _lib.SDL_UnlockTexture([ffi.c_void_p], None)
_SDL_SetTextureBlendMode = _lib.SDL_SetTextureBlendMode(
    [ffi.c_void_p, ffi.c_int32], ffi.c_int32)
_SDL_GetTextureBlendMode = _lib.SDL_GetTextureBlendMode(
    [ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_SDL_SetTextureColorMod  = _lib.SDL_SetTextureColorMod(
    [ffi.c_void_p, ffi.c_uint8, ffi.c_uint8, ffi.c_uint8], ffi.c_int32)
_SDL_GetTextureColorMod  = _lib.SDL_GetTextureColorMod(
    [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_SDL_SetTextureAlphaMod  = _lib.SDL_SetTextureAlphaMod(
    [ffi.c_void_p, ffi.c_uint8], ffi.c_int32)
_SDL_GetTextureAlphaMod  = _lib.SDL_GetTextureAlphaMod(
    [ffi.c_void_p, ffi.c_void_p], ffi.c_int32)

# --- Surface ---
_SDL_CreateRGBSurface        = _lib.SDL_CreateRGBSurface(
    [ffi.c_uint32, ffi.c_int32, ffi.c_int32, ffi.c_int32,
     ffi.c_uint32, ffi.c_uint32, ffi.c_uint32, ffi.c_uint32], ffi.c_void_p)
_SDL_CreateRGBSurfaceFrom    = _lib.SDL_CreateRGBSurfaceFrom(
    [ffi.c_void_p, ffi.c_int32, ffi.c_int32, ffi.c_int32, ffi.c_int32,
     ffi.c_uint32, ffi.c_uint32, ffi.c_uint32, ffi.c_uint32], ffi.c_void_p)
_SDL_CreateRGBSurfaceWithFormat = _lib.SDL_CreateRGBSurfaceWithFormat(
    [ffi.c_uint32, ffi.c_int32, ffi.c_int32, ffi.c_int32, ffi.c_uint32], ffi.c_void_p)
_SDL_FreeSurface             = _lib.SDL_FreeSurface([ffi.c_void_p], None)
_SDL_FillRect                = _lib.SDL_FillRect(
    [ffi.c_void_p, ffi.c_void_p, ffi.c_uint32], ffi.c_int32)
_SDL_BlitSurface             = _lib.SDL_UpperBlit(
    [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_SDL_BlitScaled              = _lib.SDL_UpperBlitScaled(
    [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_SDL_ConvertSurface          = _lib.SDL_ConvertSurface(
    [ffi.c_void_p, ffi.c_void_p, ffi.c_uint32], ffi.c_void_p)
_SDL_ConvertSurfaceFormat    = _lib.SDL_ConvertSurfaceFormat(
    [ffi.c_void_p, ffi.c_uint32, ffi.c_uint32], ffi.c_void_p)
_SDL_SetSurfaceBlendMode     = _lib.SDL_SetSurfaceBlendMode(
    [ffi.c_void_p, ffi.c_int32], ffi.c_int32)
_SDL_SetColorKey             = _lib.SDL_SetColorKey(
    [ffi.c_void_p, ffi.c_int32, ffi.c_uint32], ffi.c_int32)
_SDL_SetSurfaceAlphaMod      = _lib.SDL_SetSurfaceAlphaMod(
    [ffi.c_void_p, ffi.c_uint8], ffi.c_int32)
_SDL_LockSurface             = _lib.SDL_LockSurface([ffi.c_void_p], ffi.c_int32)
_SDL_UnlockSurface           = _lib.SDL_UnlockSurface([ffi.c_void_p], None)
_SDL_MapRGB                  = _lib.SDL_MapRGB(
    [ffi.c_void_p, ffi.c_uint8, ffi.c_uint8, ffi.c_uint8], ffi.c_uint32)
_SDL_MapRGBA                 = _lib.SDL_MapRGBA(
    [ffi.c_void_p, ffi.c_uint8, ffi.c_uint8, ffi.c_uint8, ffi.c_uint8], ffi.c_uint32)
_SDL_GetRGB                  = _lib.SDL_GetRGB(
    [ffi.c_uint32, ffi.c_void_p,
     ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], None)
_SDL_SaveBMP_RW              = _lib.SDL_SaveBMP_RW(
    [ffi.c_void_p, ffi.c_void_p, ffi.c_int32], ffi.c_int32)

# --- Events ---
_SDL_PollEvent    = _lib.SDL_PollEvent([ffi.c_void_p], ffi.c_int32)
_SDL_WaitEvent    = _lib.SDL_WaitEvent([ffi.c_void_p], ffi.c_int32)
_SDL_WaitEventTimeout = _lib.SDL_WaitEventTimeout([ffi.c_void_p, ffi.c_int32], ffi.c_int32)
_SDL_PushEvent    = _lib.SDL_PushEvent([ffi.c_void_p], ffi.c_int32)
_SDL_PeepEvents   = _lib.SDL_PeepEvents(
    [ffi.c_void_p, ffi.c_int32, ffi.c_int32, ffi.c_uint32, ffi.c_uint32], ffi.c_int32)
_SDL_FlushEvent   = _lib.SDL_FlushEvent([ffi.c_uint32], None)
_SDL_FlushEvents  = _lib.SDL_FlushEvents([ffi.c_uint32, ffi.c_uint32], None)
_SDL_HasEvent     = _lib.SDL_HasEvent([ffi.c_uint32], ffi.c_int32)
_SDL_HasEvents    = _lib.SDL_HasEvents([ffi.c_uint32, ffi.c_uint32], ffi.c_int32)
_SDL_EventState   = _lib.SDL_EventState([ffi.c_uint32, ffi.c_int32], ffi.c_uint8)
_SDL_RegisterEvents = _lib.SDL_RegisterEvents([ffi.c_int32], ffi.c_uint32)

SDL_ADDEVENT   = 0
SDL_PEEKEVENT  = 1
SDL_GETEVENT   = 2
SDL_ENABLE     = 1
SDL_DISABLE    = 0
SDL_QUERY      = -1
SDL_IGNORE     = 0

# --- Keyboard ---
_SDL_GetKeyboardState = _lib.SDL_GetKeyboardState([ffi.c_void_p], ffi.c_void_p)
_SDL_GetModState      = _lib.SDL_GetModState([], ffi.c_int32)
_SDL_SetModState      = _lib.SDL_SetModState([ffi.c_int32], None)
_SDL_GetKeyName       = _lib.SDL_GetKeyName([ffi.c_int32], ffi.c_char_p)
_SDL_GetScancodeFromName = _lib.SDL_GetScancodeFromName([ffi.c_char_p], ffi.c_int32)
_SDL_StartTextInput   = _lib.SDL_StartTextInput([], None)
_SDL_StopTextInput    = _lib.SDL_StopTextInput([], None)
_SDL_IsTextInputActive = _lib.SDL_IsTextInputActive([], ffi.c_int32)
_SDL_SetTextInputRect  = _lib.SDL_SetTextInputRect([ffi.c_void_p], None)

# --- Mouse ---
_SDL_GetMouseState         = _lib.SDL_GetMouseState(
    [ffi.c_void_p, ffi.c_void_p], ffi.c_uint32)
_SDL_GetGlobalMouseState   = _lib.SDL_GetGlobalMouseState(
    [ffi.c_void_p, ffi.c_void_p], ffi.c_uint32)
_SDL_GetRelativeMouseState = _lib.SDL_GetRelativeMouseState(
    [ffi.c_void_p, ffi.c_void_p], ffi.c_uint32)
_SDL_SetRelativeMouseMode  = _lib.SDL_SetRelativeMouseMode([ffi.c_int32], ffi.c_int32)
_SDL_GetRelativeMouseMode  = _lib.SDL_GetRelativeMouseMode([], ffi.c_int32)
_SDL_ShowCursor            = _lib.SDL_ShowCursor([ffi.c_int32], ffi.c_int32)
_SDL_WarpMouseGlobal       = _lib.SDL_WarpMouseGlobal(
    [ffi.c_int32, ffi.c_int32], ffi.c_int32)
_SDL_CreateCursor          = _lib.SDL_CreateCursor(
    [ffi.c_void_p, ffi.c_void_p, ffi.c_int32, ffi.c_int32,
     ffi.c_int32, ffi.c_int32], ffi.c_void_p)
_SDL_CreateColorCursor     = _lib.SDL_CreateColorCursor(
    [ffi.c_void_p, ffi.c_int32, ffi.c_int32], ffi.c_void_p)
_SDL_CreateSystemCursor    = _lib.SDL_CreateSystemCursor([ffi.c_int32], ffi.c_void_p)
_SDL_SetCursor             = _lib.SDL_SetCursor([ffi.c_void_p], None)
_SDL_GetCursor             = _lib.SDL_GetCursor([], ffi.c_void_p)
_SDL_FreeCursor            = _lib.SDL_FreeCursor([ffi.c_void_p], None)

SDL_SYSTEM_CURSOR_ARROW     = 0
SDL_SYSTEM_CURSOR_IBEAM     = 1
SDL_SYSTEM_CURSOR_WAIT      = 2
SDL_SYSTEM_CURSOR_CROSSHAIR = 3
SDL_SYSTEM_CURSOR_WAITARROW = 4
SDL_SYSTEM_CURSOR_SIZENWSE  = 5
SDL_SYSTEM_CURSOR_SIZENESW  = 6
SDL_SYSTEM_CURSOR_SIZEWE    = 7
SDL_SYSTEM_CURSOR_SIZENS    = 8
SDL_SYSTEM_CURSOR_SIZEALL   = 9
SDL_SYSTEM_CURSOR_NO        = 10
SDL_SYSTEM_CURSOR_HAND      = 11

# --- Clipboard ---
_SDL_GetClipboardText  = _lib.SDL_GetClipboardText([], ffi.c_char_p)
_SDL_SetClipboardText  = _lib.SDL_SetClipboardText([ffi.c_char_p], ffi.c_int32)
_SDL_HasClipboardText  = _lib.SDL_HasClipboardText([], ffi.c_int32)

# --- Audio ---
_SDL_GetNumAudioDrivers  = _lib.SDL_GetNumAudioDrivers([], ffi.c_int32)
_SDL_GetAudioDriver      = _lib.SDL_GetAudioDriver([ffi.c_int32], ffi.c_char_p)
_SDL_GetCurrentAudioDriver = _lib.SDL_GetCurrentAudioDriver([], ffi.c_char_p)
_SDL_GetNumAudioDevices  = _lib.SDL_GetNumAudioDevices([ffi.c_int32], ffi.c_int32)
_SDL_GetAudioDeviceName  = _lib.SDL_GetAudioDeviceName(
    [ffi.c_int32, ffi.c_int32], ffi.c_char_p)
_SDL_OpenAudioDevice     = _lib.SDL_OpenAudioDevice(
    [ffi.c_char_p, ffi.c_int32, ffi.c_void_p, ffi.c_void_p, ffi.c_int32],
    ffi.c_uint32)
_SDL_CloseAudioDevice    = _lib.SDL_CloseAudioDevice([ffi.c_uint32], None)
_SDL_PauseAudioDevice    = _lib.SDL_PauseAudioDevice([ffi.c_uint32, ffi.c_int32], None)
_SDL_QueueAudio          = _lib.SDL_QueueAudio(
    [ffi.c_uint32, ffi.c_void_p, ffi.c_uint32], ffi.c_int32)
_SDL_DequeueAudio        = _lib.SDL_DequeueAudio(
    [ffi.c_uint32, ffi.c_void_p, ffi.c_uint32], ffi.c_uint32)
_SDL_GetQueuedAudioSize  = _lib.SDL_GetQueuedAudioSize([ffi.c_uint32], ffi.c_uint32)
_SDL_ClearQueuedAudio    = _lib.SDL_ClearQueuedAudio([ffi.c_uint32], None)
_SDL_GetAudioDeviceStatus = _lib.SDL_GetAudioDeviceStatus([ffi.c_uint32], ffi.c_int32)
_SDL_LockAudioDevice     = _lib.SDL_LockAudioDevice([ffi.c_uint32], None)
_SDL_UnlockAudioDevice   = _lib.SDL_UnlockAudioDevice([ffi.c_uint32], None)
_SDL_MixAudioFormat      = _lib.SDL_MixAudioFormat(
    [ffi.c_void_p, ffi.c_void_p, ffi.c_uint16, ffi.c_uint32, ffi.c_int32], None)

SDL_AUDIO_STOPPED = 0
SDL_AUDIO_PLAYING = 1
SDL_AUDIO_PAUSED  = 2
SDL_MIX_MAXVOLUME = 128

SDL_AUDIO_ALLOW_FREQUENCY_CHANGE  = 0x00000001
SDL_AUDIO_ALLOW_FORMAT_CHANGE     = 0x00000002
SDL_AUDIO_ALLOW_CHANNELS_CHANGE   = 0x00000004
SDL_AUDIO_ALLOW_SAMPLES_CHANGE    = 0x00000008
SDL_AUDIO_ALLOW_ANY_CHANGE        = 0x0000000F

# --- OpenGL ---
_SDL_GL_CreateContext   = _lib.SDL_GL_CreateContext([ffi.c_void_p], ffi.c_void_p)
_SDL_GL_DeleteContext   = _lib.SDL_GL_DeleteContext([ffi.c_void_p], None)
_SDL_GL_MakeCurrent     = _lib.SDL_GL_MakeCurrent(
    [ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_SDL_GL_GetCurrentContext = _lib.SDL_GL_GetCurrentContext([], ffi.c_void_p)
_SDL_GL_SwapWindow      = _lib.SDL_GL_SwapWindow([ffi.c_void_p], None)
_SDL_GL_SetSwapInterval = _lib.SDL_GL_SetSwapInterval([ffi.c_int32], ffi.c_int32)
_SDL_GL_GetSwapInterval = _lib.SDL_GL_GetSwapInterval([], ffi.c_int32)
_SDL_GL_SetAttribute    = _lib.SDL_GL_SetAttribute([ffi.c_int32, ffi.c_int32], ffi.c_int32)
_SDL_GL_GetAttribute    = _lib.SDL_GL_GetAttribute(
    [ffi.c_int32, ffi.c_void_p], ffi.c_int32)
_SDL_GL_GetProcAddress  = _lib.SDL_GL_GetProcAddress([ffi.c_char_p], ffi.c_void_p)
_SDL_GL_GetDrawableSize = _lib.SDL_GL_GetDrawableSize(
    [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], None)
_SDL_GL_LoadLibrary     = _lib.SDL_GL_LoadLibrary([ffi.c_char_p], ffi.c_int32)
_SDL_GL_UnloadLibrary   = _lib.SDL_GL_UnloadLibrary([], None)

# --- RWops (for BMP save) ---
_SDL_RWFromFile         = _lib.SDL_RWFromFile([ffi.c_char_p, ffi.c_char_p], ffi.c_void_p)
_SDL_RWclose            = _lib.SDL_RWclose([ffi.c_void_p], ffi.c_int32)

# ==============================================================================
#  Helpers
# ==============================================================================

def get_error():
    raw = _SDL_GetError()
    return raw if raw else ""

def clear_error():
    _SDL_ClearError()

def _check(ret, label="SDL"):
    if ret < 0:
        raise SDLError(format_str("{label} failed: {get_error()}"))

class SDLError(Exception):
    pass

def _enc(s):
    return s.encode("utf-8") if isinstance(s, str) else s

def _read_int_buf(buf):
    b = ffi.buffer_to_bytes(buf, 4)
    return b[0] | (b[1] << 8) | (b[2] << 16) | (b[3] << 24)

# ==============================================================================
#  Event class
# ==============================================================================

class Event:
    """
    Wraps a raw SDL_Event buffer (56 bytes) and exposes typed field access.

    Usage:
        ev = Event()
        while ev.poll():
            if ev.type == SDL_QUIT:
                running = False
            elif ev.type == SDL_KEYDOWN:
                print(ev.scancode, ev.sym)
            elif ev.type == SDL_MOUSEBUTTONDOWN:
                print(ev.x, ev.y, ev.button)
    """
    def __init__(self):
        self._buf = ffi.malloc(_SZ_EVENT)
        self.type = 0

    # --- Polling ---
    def poll(self):
        """Return True if an event was available and fill self."""
        got = _SDL_PollEvent(self._buf)
        if got:
            self.type = _read_u32(self._buf, 0)
        return got != 0

    def wait(self):
        """Block until an event is available."""
        _SDL_WaitEvent(self._buf)
        self.type = _read_u32(self._buf, 0)

    def wait_timeout(self, ms):
        """Block up to ms milliseconds. Returns True if event received."""
        got = _SDL_WaitEventTimeout(self._buf, ms)
        if got:
            self.type = _read_u32(self._buf, 0)
        return got != 0

    def push(self):
        """Push this event onto the queue."""
        _read_u32(self._buf, 0)   # ensure type is set
        _SDL_PushEvent(self._buf)

    # --- Common fields ---
    @property
    def timestamp(self):            return _read_u32(self._buf,  4)

    # --- Keyboard fields (SDL_KEYDOWN / SDL_KEYUP) ---
    @property
    def window_id(self):            return _read_u32(self._buf,  8)
    @property
    def state(self):                return _read_u8( self._buf, 12)
    @property
    def repeat(self):               return _read_u8( self._buf, 13)
    @property
    def scancode(self):             return _read_i32(self._buf, 16)
    @property
    def sym(self):                  return _read_i32(self._buf, 20)
    @property
    def mod(self):                  return _read_u16(self._buf, 24)

    # --- Mouse motion (SDL_MOUSEMOTION) ---
    @property
    def which(self):                return _read_u32(self._buf, 12)
    @property
    def mouse_state(self):          return _read_u32(self._buf, 16)
    @property
    def x(self):                    return _read_i32(self._buf, 20)
    @property
    def y(self):                    return _read_i32(self._buf, 24)
    @property
    def xrel(self):                 return _read_i32(self._buf, 28)
    @property
    def yrel(self):                 return _read_i32(self._buf, 32)

    # --- Mouse button (SDL_MOUSEBUTTONDOWN / SDL_MOUSEBUTTONUP) ---
    @property
    def button(self):               return _read_u8( self._buf, 16)
    @property
    def button_state(self):         return _read_u8( self._buf, 17)
    @property
    def clicks(self):               return _read_u8( self._buf, 18)

    # --- Mouse wheel (SDL_MOUSEWHEEL) ---
    @property
    def wheel_x(self):              return _read_i32(self._buf, 16)
    @property
    def wheel_y(self):              return _read_i32(self._buf, 20)
    @property
    def wheel_direction(self):      return _read_u32(self._buf, 24)
    @property
    def wheel_precise_x(self):      return _read_f32(self._buf, 28)
    @property
    def wheel_precise_y(self):      return _read_f32(self._buf, 32)

    # --- Window event (SDL_WINDOWEVENT) ---
    @property
    def window_event(self):         return _read_u8( self._buf, 12)
    @property
    def window_data1(self):         return _read_i32(self._buf, 16)
    @property
    def window_data2(self):         return _read_i32(self._buf, 20)

    # --- Text input (SDL_TEXTINPUT) ---
    @property
    def text(self):
        raw = ffi.buffer_to_bytes(self._buf, _SZ_EVENT)
        result = ""
        for i in range(12, 44):
            if raw[i] == 0:
                break
            result = result + chr(raw[i])
        return result

    # --- Touch / finger events ---
    @property
    def touch_id(self):
        lo = _read_u32(self._buf,  8)
        hi = _read_u32(self._buf, 12)
        return lo | (hi << 32)
    @property
    def finger_id(self):
        lo = _read_u32(self._buf, 16)
        hi = _read_u32(self._buf, 20)
        return lo | (hi << 32)
    @property
    def finger_x(self):             return _read_f32(self._buf, 24)
    @property
    def finger_y(self):             return _read_f32(self._buf, 28)
    @property
    def finger_dx(self):            return _read_f32(self._buf, 32)
    @property
    def finger_dy(self):            return _read_f32(self._buf, 36)
    @property
    def pressure(self):             return _read_f32(self._buf, 40)

    # --- User event data ---
    @property
    def user_code(self):            return _read_i32(self._buf, 12)
    @property
    def user_data1(self):           return _read_u32(self._buf, 16)
    @property
    def user_data2(self):           return _read_u32(self._buf, 20)

    # --- Helpers ---
    def set_type(self, event_type):
        ffi.write_memory_with_offset(self._buf, 0, ffi.c_uint32, event_type)
        self.type = event_type

    def mod_has(self, flag):
        return (self.mod & flag) != 0

    def free(self):
        if self._buf:
            ffi.free(self._buf)
            self._buf = None

# ==============================================================================
#  Window class
# ==============================================================================

class Window:
    """
    Usage:
        win = Window("My App", 1280, 720)
        win = Window("My App", 1280, 720, SDL_WINDOW_RESIZABLE | SDL_WINDOW_OPENGL)
    """
    def __init__(self, title, w, h, flags=SDL_WINDOW_SHOWN, x=SDL_WINDOWPOS_CENTERED, y=SDL_WINDOWPOS_CENTERED):
        self.ptr = _SDL_CreateWindow(_enc(title), x, y, w, h, flags)
        if self.ptr is None or self.ptr.Address == 0:
            raise SDLError(format_str("SDL_CreateWindow failed: {get_error()}"))
        self._title = title
        self.w = w
        self.h = h

    # --- Properties ---
    def get_id(self):               return _SDL_GetWindowID(self.ptr)
    def get_title(self):
        raw = _SDL_GetWindowTitle(self.ptr)
        return raw if raw else ""
    def set_title(self, title):
        self._title = title
        _SDL_SetWindowTitle(self.ptr, _enc(title))

    def get_size(self):
        wb = ffi.malloc(4); hb = ffi.malloc(4)
        try:
            _SDL_GetWindowSize(self.ptr, wb, hb)
            return (_read_int_buf(wb), _read_int_buf(hb))
        finally:
            ffi.free(wb); ffi.free(hb)

    def set_size(self, w, h):
        _SDL_SetWindowSize(self.ptr, w, h)

    def get_position(self):
        xb = ffi.malloc(4); yb = ffi.malloc(4)
        try:
            _SDL_GetWindowPosition(self.ptr, xb, yb)
            return (_read_int_buf(xb), _read_int_buf(yb))
        finally:
            ffi.free(xb); ffi.free(yb)

    def set_position(self, x, y):  _SDL_SetWindowPosition(self.ptr, x, y)
    def get_flags(self):            return _SDL_GetWindowFlags(self.ptr)
    def get_display_index(self):    return _SDL_GetWindowDisplayIndex(self.ptr)

    # --- Visibility / state ---
    def show(self):                 _SDL_ShowWindow(self.ptr)
    def hide(self):                 _SDL_HideWindow(self.ptr)
    def raise_window(self):         _SDL_RaiseWindow(self.ptr)
    def minimize(self):             _SDL_MinimizeWindow(self.ptr)
    def maximize(self):             _SDL_MaximizeWindow(self.ptr)
    def restore(self):              _SDL_RestoreWindow(self.ptr)
    def flash(self, op=SDL_FLASH_BRIEFLY):
        _SDL_FlashWindow(self.ptr, op)

    def set_fullscreen(self, flags=SDL_WINDOW_FULLSCREEN_DESKTOP):
        _check(_SDL_SetWindowFullscreen(self.ptr, flags), "SetWindowFullscreen")

    def set_bordered(self, bordered):
        _SDL_SetWindowBordered(self.ptr, 1 if bordered else 0)

    def set_resizable(self, resizable):
        _SDL_SetWindowResizable(self.ptr, 1 if resizable else 0)

    def set_opacity(self, alpha):
        _check(_SDL_SetWindowOpacity(self.ptr, alpha), "SetWindowOpacity")

    def get_opacity(self):
        buf = ffi.malloc(4)
        try:
            _SDL_GetWindowOpacity(self.ptr, buf)
            return _read_f32(buf, 0)
        finally:
            ffi.free(buf)

    def set_min_size(self, w, h):   _SDL_SetWindowMinimumSize(self.ptr, w, h)
    def set_max_size(self, w, h):   _SDL_SetWindowMaximumSize(self.ptr, w, h)
    def set_grab(self, grabbed):    _SDL_SetWindowGrab(self.ptr, 1 if grabbed else 0)
    def get_grab(self):             return _SDL_GetWindowGrab(self.ptr) != 0
    def set_icon(self, surface_ptr): _SDL_SetWindowIcon(self.ptr, surface_ptr)
    def warp_mouse(self, x, y):     _SDL_WarpMouseInWindow(self.ptr, x, y)

    def get_surface(self):
        ptr = _SDL_GetWindowSurface(self.ptr)
        s = Surface.__new__(Surface)
        s.ptr = ptr
        s._owned = False
        return s

    def update_surface(self):
        _check(_SDL_UpdateWindowSurface(self.ptr), "UpdateWindowSurface")

    # --- GL ---
    def gl_swap(self):              _SDL_GL_SwapWindow(self.ptr)

    def gl_get_drawable_size(self):
        wb = ffi.malloc(4); hb = ffi.malloc(4)
        try:
            _SDL_GL_GetDrawableSize(self.ptr, wb, hb)
            return (_read_int_buf(wb), _read_int_buf(hb))
        finally:
            ffi.free(wb); ffi.free(hb)

    def destroy(self):
        if self.ptr:
            _SDL_DestroyWindow(self.ptr)
            self.ptr = None

# ==============================================================================
#  Renderer class
# ==============================================================================

class Renderer:
    """
    Usage:
        r = Renderer(window)
        r = Renderer(window, flags=SDL_RENDERER_ACCELERATED|SDL_RENDERER_PRESENTVSYNC)

        r.set_draw_color(255, 0, 0, 255)
        r.clear()
        r.fill_rect(10, 10, 100, 50)
        r.present()
    """
    def __init__(self, window, index=-1, flags=0):
        win_ptr = window.ptr if isinstance(window, Window) else window
        self.ptr = _SDL_CreateRenderer(win_ptr, index, flags)
        if self.ptr is None or self.ptr.Address == 0:
            raise SDLError(format_str("SDL_CreateRenderer failed: {get_error()}"))

    # --- Draw color ---
    def set_draw_color(self, r, g, b, a=255):
        _SDL_SetRenderDrawColor(self.ptr, r, g, b, a)

    def get_draw_color(self):
        rb = ffi.malloc(1); gb = ffi.malloc(1)
        bb = ffi.malloc(1); ab = ffi.malloc(1)
        try:
            _SDL_GetRenderDrawColor(self.ptr, rb, gb, bb, ab)
            r = ffi.buffer_to_bytes(rb, 1)[0]
            g = ffi.buffer_to_bytes(gb, 1)[0]
            b = ffi.buffer_to_bytes(bb, 1)[0]
            a = ffi.buffer_to_bytes(ab, 1)[0]
            return (r, g, b, a)
        finally:
            ffi.free(rb); ffi.free(gb); ffi.free(bb); ffi.free(ab)

    def set_blend_mode(self, mode):
        _SDL_SetRenderDrawBlendMode(self.ptr, mode)

    # --- Frame ---
    def clear(self):                _SDL_RenderClear(self.ptr)
    def present(self):              _SDL_RenderPresent(self.ptr)
    def flush(self):                _SDL_RenderFlush(self.ptr)

    # --- Primitives ---
    def draw_point(self, x, y):
        _SDL_RenderDrawPoint(self.ptr, x, y)

    def draw_points(self, points):
        """points: list of (x, y) tuples."""
        n   = len(points)
        buf = ffi.malloc(n * _SZ_POINT)
        try:
            for i in range(n):
                ffi.write_memory_with_offset(buf, i*8,   ffi.c_int32, points[i][0])
                ffi.write_memory_with_offset(buf, i*8+4, ffi.c_int32, points[i][1])
            _SDL_RenderDrawPoints(self.ptr, buf, n)
        finally:
            ffi.free(buf)

    def draw_line(self, x1, y1, x2, y2):
        _SDL_RenderDrawLine(self.ptr, x1, y1, x2, y2)

    def draw_lines(self, points):
        """points: list of (x, y) tuples defining a polyline."""
        n   = len(points)
        buf = ffi.malloc(n * _SZ_POINT)
        try:
            for i in range(n):
                ffi.write_memory_with_offset(buf, i*8,   ffi.c_int32, points[i][0])
                ffi.write_memory_with_offset(buf, i*8+4, ffi.c_int32, points[i][1])
            _SDL_RenderDrawLines(self.ptr, buf, n)
        finally:
            ffi.free(buf)

    def draw_rect(self, x, y, w, h):
        buf = make_rect(x, y, w, h)
        try:
            _SDL_RenderDrawRect(self.ptr, buf)
        finally:
            ffi.free(buf)

    def fill_rect(self, x, y, w, h):
        buf = make_rect(x, y, w, h)
        try:
            _SDL_RenderFillRect(self.ptr, buf)
        finally:
            ffi.free(buf)

    def fill_screen(self):
        _SDL_RenderFillRect(self.ptr, ffi.c_void_p(0))

    def draw_rects(self, rects):
        """rects: list of (x,y,w,h) tuples."""
        n   = len(rects)
        buf = ffi.malloc(n * _SZ_RECT)
        try:
            for i in range(n):
                ffi.write_memory_with_offset(buf, i*16,    ffi.c_int32, rects[i][0])
                ffi.write_memory_with_offset(buf, i*16+4,  ffi.c_int32, rects[i][1])
                ffi.write_memory_with_offset(buf, i*16+8,  ffi.c_int32, rects[i][2])
                ffi.write_memory_with_offset(buf, i*16+12, ffi.c_int32, rects[i][3])
            _SDL_RenderDrawRects(self.ptr, buf, n)
        finally:
            ffi.free(buf)

    def fill_rects(self, rects):
        n   = len(rects)
        buf = ffi.malloc(n * _SZ_RECT)
        try:
            for i in range(n):
                ffi.write_memory_with_offset(buf, i*16,    ffi.c_int32, rects[i][0])
                ffi.write_memory_with_offset(buf, i*16+4,  ffi.c_int32, rects[i][1])
                ffi.write_memory_with_offset(buf, i*16+8,  ffi.c_int32, rects[i][2])
                ffi.write_memory_with_offset(buf, i*16+12, ffi.c_int32, rects[i][3])
            _SDL_RenderFillRects(self.ptr, buf, n)
        finally:
            ffi.free(buf)

    # --- Texture ---
    def copy(self, texture, src=None, dst=None):
        """src / dst: (x,y,w,h) tuples or None for full area."""
        tex_ptr = texture.ptr if isinstance(texture, Texture) else texture
        src_buf = make_rect(*src) if src else ffi.c_void_p(0)
        dst_buf = make_rect(*dst) if dst else ffi.c_void_p(0)
        try:
            _SDL_RenderCopy(self.ptr, tex_ptr, src_buf, dst_buf)
        finally:
            if src:
                ffi.free(src_buf)
            if dst:
                ffi.free(dst_buf)

    def copy_ex(self, texture, src, dst, angle, center=None, flip=SDL_FLIP_NONE):
        tex_ptr = texture.ptr if isinstance(texture, Texture) else texture
        src_buf = make_rect(*src) if src else ffi.c_void_p(0)
        dst_buf = make_rect(*dst) if dst else ffi.c_void_p(0)
        ctr_buf = make_point(*center) if center else ffi.c_void_p(0)
        try:
            _SDL_RenderCopyEx(self.ptr, tex_ptr, src_buf, dst_buf, angle, ctr_buf, flip)
        finally:
            if src:    ffi.free(src_buf)
            if dst:    ffi.free(dst_buf)
            if center: ffi.free(ctr_buf)

    def geometry(self, texture, vertices, indices=None):
        """
        Render triangles.  vertices: list of (x, y, r, g, b, a, u, v) tuples.
        SDL_Vertex layout: [float x, float y, uint8 r,g,b,a, float u, float v] = 20 bytes
        """
        tex_ptr = texture.ptr if isinstance(texture, Texture) else (texture if texture else ffi.c_void_p(0))
        n    = len(vertices)
        vbuf = ffi.malloc(n * _SZ_VERTEX)
        ibuf = None
        ni   = 0
        try:
            for i in range(n):
                v   = vertices[i]
                off = i * _SZ_VERTEX
                ffi.write_memory_with_offset(vbuf, off,    ffi.c_float,  v[0])
                ffi.write_memory_with_offset(vbuf, off+4,  ffi.c_float,  v[1])
                ffi.write_memory_with_offset(vbuf, off+8,  ffi.c_uint8,  v[2])
                ffi.write_memory_with_offset(vbuf, off+9,  ffi.c_uint8,  v[3])
                ffi.write_memory_with_offset(vbuf, off+10, ffi.c_uint8,  v[4])
                ffi.write_memory_with_offset(vbuf, off+11, ffi.c_uint8,  v[5])
                ffi.write_memory_with_offset(vbuf, off+12, ffi.c_float,  v[6])
                ffi.write_memory_with_offset(vbuf, off+16, ffi.c_float,  v[7])

            if indices:
                ni   = len(indices)
                ibuf = ffi.malloc(ni * 4)
                for i in range(ni):
                    ffi.write_memory_with_offset(ibuf, i*4, ffi.c_int32, indices[i])

            _SDL_RenderGeometry(self.ptr, tex_ptr, vbuf, n, ibuf if ibuf else ffi.c_void_p(0), ni)
        finally:
            ffi.free(vbuf)
            if ibuf:
                ffi.free(ibuf)

    # --- Target texture ---
    def set_target(self, texture=None):
        ptr = texture.ptr if isinstance(texture, Texture) else (texture if texture else ffi.c_void_p(0))
        _check(_SDL_SetRenderTarget(self.ptr, ptr), "SetRenderTarget")

    def get_target(self):           return _SDL_GetRenderTarget(self.ptr)

    # --- Viewport / clip ---
    def set_viewport(self, x, y, w, h):
        buf = make_rect(x, y, w, h)
        try:
            _SDL_RenderSetViewport(self.ptr, buf)
        finally:
            ffi.free(buf)

    def reset_viewport(self):
        _SDL_RenderSetViewport(self.ptr, ffi.c_void_p(0))

    def set_clip_rect(self, x, y, w, h):
        buf = make_rect(x, y, w, h)
        try:
            _SDL_RenderSetClipRect(self.ptr, buf)
        finally:
            ffi.free(buf)

    def disable_clip(self):
        _SDL_RenderSetClipRect(self.ptr, ffi.c_void_p(0))

    # --- Scale / logical size ---
    def set_logical_size(self, w, h):
        _check(_SDL_RenderSetLogicalSize(self.ptr, w, h), "SetLogicalSize")

    def set_scale(self, sx, sy):
        _check(_SDL_RenderSetScale(self.ptr, sx, sy), "SetScale")

    def get_output_size(self):
        wb = ffi.malloc(4); hb = ffi.malloc(4)
        try:
            _SDL_GetRendererOutputSize(self.ptr, wb, hb)
            return (_read_int_buf(wb), _read_int_buf(hb))
        finally:
            ffi.free(wb); ffi.free(hb)

    # --- Pixel readback ---
    def read_pixels(self, x, y, w, h, fmt=SDL_PIXELFORMAT_ARGB8888):
        pitch  = w * 4
        buf    = ffi.malloc(h * pitch)
        rect   = make_rect(x, y, w, h)
        try:
            _SDL_RenderReadPixels(self.ptr, rect, fmt, buf, pitch)
            return ffi.buffer_to_bytes(buf, h * pitch)
        finally:
            ffi.free(buf)
            ffi.free(rect)

    def destroy(self):
        if self.ptr:
            _SDL_DestroyRenderer(self.ptr)
            self.ptr = None

# ==============================================================================
#  Texture class
# ==============================================================================

class Texture:
    """
    Usage:
        tex = Texture(renderer, SDL_PIXELFORMAT_ARGB8888,
                      SDL_TEXTUREACCESS_STREAMING, 800, 600)
        tex.update(None, pixel_bytes, pitch=800*4)
        renderer.copy(tex)
        tex.destroy()
    """
    def __init__(self, renderer, fmt, access, w, h):
        rptr = renderer.ptr if isinstance(renderer, Renderer) else renderer
        self.ptr = _SDL_CreateTexture(rptr, fmt, access, w, h)
        if self.ptr is None or self.ptr.Address == 0:
            raise SDLError(format_str("SDL_CreateTexture failed: {get_error()}"))
        self.w = w
        self.h = h
        self.fmt = fmt

    @staticmethod
    def from_surface(renderer, surface):
        rptr = renderer.ptr if isinstance(renderer, Renderer) else renderer
        sptr = surface.ptr  if isinstance(surface,  Surface)  else surface
        ptr  = _SDL_CreateTextureFromSurface(rptr, sptr)
        if ptr is None or ptr.Address == 0:
            raise SDLError(format_str("CreateTextureFromSurface: {get_error()}"))
        t     = Texture.__new__(Texture)
        t.ptr = ptr
        t.fmt = 0
        # Query actual size
        wb = ffi.malloc(4); hb = ffi.malloc(4)
        try:
            _SDL_QueryTexture(ptr, ffi.c_void_p(0), ffi.c_void_p(0), wb, hb)
            t.w = _read_int_buf(wb)
            t.h = _read_int_buf(hb)
        finally:
            ffi.free(wb); ffi.free(hb)
        return t

    def update(self, rect, pixels, pitch):
        """pixels: bytes or a C pointer. rect: (x,y,w,h) or None."""
        rect_buf = make_rect(*rect) if rect else ffi.c_void_p(0)
        try:
            _SDL_UpdateTexture(self.ptr, rect_buf, pixels, pitch)
        finally:
            if rect:
                ffi.free(rect_buf)

    def lock(self, rect=None):
        """Returns (pixel_ptr, pitch). Must call unlock() when done."""
        rect_buf = make_rect(*rect) if rect else ffi.c_void_p(0)
        pxbuf    = ffi.malloc(8)   # void*
        ptbuf    = ffi.malloc(4)   # int
        try:
            _check(_SDL_LockTexture(self.ptr, rect_buf, pxbuf, ptbuf), "LockTexture")
            px    = ffi.read_memory_with_offset(pxbuf, 0, ffi.c_void_p)
            pitch = _read_int_buf(ptbuf)
            return (px, pitch)
        finally:
            ffi.free(pxbuf); ffi.free(ptbuf)
            if rect: ffi.free(rect_buf)

    def unlock(self):               _SDL_UnlockTexture(self.ptr)

    def set_blend_mode(self, mode): _SDL_SetTextureBlendMode(self.ptr, mode)
    def set_color_mod(self, r, g, b): _SDL_SetTextureColorMod(self.ptr, r, g, b)
    def set_alpha_mod(self, a):     _SDL_SetTextureAlphaMod(self.ptr, a)

    def query(self):
        fmtb = ffi.malloc(4); accb = ffi.malloc(4)
        wb   = ffi.malloc(4); hb   = ffi.malloc(4)
        try:
            _SDL_QueryTexture(self.ptr, fmtb, accb, wb, hb)
            return {
                "format": _read_int_buf(fmtb),
                "access": _read_int_buf(accb),
                "w":      _read_int_buf(wb),
                "h":      _read_int_buf(hb),
            }
        finally:
            ffi.free(fmtb); ffi.free(accb); ffi.free(wb); ffi.free(hb)

    def destroy(self):
        if self.ptr:
            _SDL_DestroyTexture(self.ptr)
            self.ptr = None

# ==============================================================================
#  Surface class
# ==============================================================================

class Surface:
    """
    Usage:
        s = Surface.create(800, 600)
        s = Surface.create_with_format(800, 600, SDL_PIXELFORMAT_ARGB8888)
        s.fill(None, 0xFF000000)
        s.save_bmp("screenshot.bmp")
        s.free()
    """
    def __init__(self, w, h, depth=32, rmask=0xFF0000, gmask=0x00FF00, bmask=0x0000FF, amask=0xFF000000):
        self.ptr    = _SDL_CreateRGBSurface(0, w, h, depth, rmask, gmask, bmask, amask)
        self._owned = True
        if self.ptr is None or self.ptr.Address == 0:
            raise SDLError(format_str("SDL_CreateRGBSurface: {get_error()}"))

    @staticmethod
    def create(w, h):
        return Surface(w, h)

    @staticmethod
    def create_with_format(w, h, fmt=SDL_PIXELFORMAT_ARGB8888):
        s       = Surface.__new__(Surface)
        s.ptr   = _SDL_CreateRGBSurfaceWithFormat(0, w, h, 32, fmt)
        s._owned = True
        if s.ptr is None or s.ptr.Address == 0:
            raise SDLError(format_str("CreateRGBSurfaceWithFormat: {get_error()}"))
        return s

    @staticmethod
    def from_ptr(ptr, owned=False):
        s        = Surface.__new__(Surface)
        s.ptr    = ptr
        s._owned = owned
        return s

    def fill(self, rect, color):
        rect_buf = make_rect(*rect) if rect else ffi.c_void_p(0)
        try:
            _SDL_FillRect(self.ptr, rect_buf, color)
        finally:
            if rect: ffi.free(rect_buf)

    def blit(self, src, src_rect=None, dst_rect=None):
        sptr = src.ptr if isinstance(src, Surface) else src
        sbuf = make_rect(*src_rect) if src_rect else ffi.c_void_p(0)
        dbuf = make_rect(*dst_rect) if dst_rect else ffi.c_void_p(0)
        try:
            _SDL_BlitSurface(sptr, sbuf, self.ptr, dbuf)
        finally:
            if src_rect: ffi.free(sbuf)
            if dst_rect: ffi.free(dbuf)

    def blit_scaled(self, src, src_rect=None, dst_rect=None):
        sptr = src.ptr if isinstance(src, Surface) else src
        sbuf = make_rect(*src_rect) if src_rect else ffi.c_void_p(0)
        dbuf = make_rect(*dst_rect) if dst_rect else ffi.c_void_p(0)
        try:
            _SDL_BlitScaled(sptr, sbuf, self.ptr, dbuf)
        finally:
            if src_rect: ffi.free(sbuf)
            if dst_rect: ffi.free(dbuf)

    def set_blend_mode(self, mode): _SDL_SetSurfaceBlendMode(self.ptr, mode)
    def set_alpha_mod(self, a):     _SDL_SetSurfaceAlphaMod(self.ptr, a)
    def set_color_key(self, flag, color): _SDL_SetColorKey(self.ptr, flag, color)

    def convert_format(self, fmt=SDL_PIXELFORMAT_ARGB8888):
        ptr = _SDL_ConvertSurfaceFormat(self.ptr, fmt, 0)
        return Surface.from_ptr(ptr, owned=True)

    def lock(self):                 _check(_SDL_LockSurface(self.ptr), "LockSurface")
    def unlock(self):               _SDL_UnlockSurface(self.ptr)

    def save_bmp(self, path):
        rw = _SDL_RWFromFile(_enc(path), b"wb")
        _SDL_SaveBMP_RW(self.ptr, rw, 1)   # 1 = close RW after write

    def free(self):
        if self._owned and self.ptr:
            _SDL_FreeSurface(self.ptr)
            self.ptr = None

# ==============================================================================
#  Audio class
# ==============================================================================

class AudioDevice:
    """
    Queue-based audio output (no callback).

    Usage:
        dev = AudioDevice(freq=44100, fmt=AUDIO_S16, channels=2, samples=1024)
        dev.queue(pcm_bytes)
        dev.play()
        # in loop:
        while dev.queued_size() > target:
            delay(1)
        dev.close()
    """
    def __init__(self, device_name=None, is_capture=False, freq=44100, fmt=AUDIO_S16, channels=2, samples=1024):
        spec = ffi.malloc(_SZ_AUDIOSPEC)
        try:
            ffi.write_memory_with_offset(spec,  0, ffi.c_int32,  freq)
            ffi.write_memory_with_offset(spec,  4, ffi.c_uint16, fmt)
            ffi.write_memory_with_offset(spec,  6, ffi.c_uint8,  channels)
            ffi.write_memory_with_offset(spec,  8, ffi.c_uint16, samples)
            # callback = NULL (queue mode)
            ffi.write_memory_with_offset(spec, 16, ffi.c_uint64, 0)

            name = _enc(device_name) if device_name else ffi.c_void_p(0)
            self._id = _SDL_OpenAudioDevice( name, 1 if is_capture else 0, spec, ffi.c_void_p(0), 0)
        finally:
            ffi.free(spec)

        if self._id == 0:
            raise SDLError(format_str("SDL_OpenAudioDevice: {get_error()}"))

    def play(self):             _SDL_PauseAudioDevice(self._id, 0)
    def pause(self):            _SDL_PauseAudioDevice(self._id, 1)
    def get_status(self):       return _SDL_GetAudioDeviceStatus(self._id)
    def is_playing(self):       return self.get_status() == SDL_AUDIO_PLAYING

    def queue(self, data):
        if isinstance(data, (bytes, bytearray)):
            _check(_SDL_QueueAudio(self._id, data, len(data)), "QueueAudio")
        else:
            raise TypeError("data must be bytes or bytearray")

    def dequeue(self, n):
        buf = ffi.malloc(n)
        try:
            got = _SDL_DequeueAudio(self._id, buf, n)
            return ffi.buffer_to_bytes(buf, got)
        finally:
            ffi.free(buf)

    def queued_size(self):      return _SDL_GetQueuedAudioSize(self._id)
    def clear(self):            _SDL_ClearQueuedAudio(self._id)
    def lock(self):             _SDL_LockAudioDevice(self._id)
    def unlock(self):           _SDL_UnlockAudioDevice(self._id)

    def close(self):
        if self._id:
            _SDL_CloseAudioDevice(self._id)
            self._id = 0

# ==============================================================================
#  OpenGL Context class
# ==============================================================================

class GLContext:
    """
    Usage:
        window = Window("GL", 800, 600, SDL_WINDOW_OPENGL)
        GL.set_attribute(SDL_GL_CONTEXT_MAJOR_VERSION, 3)
        GL.set_attribute(SDL_GL_CONTEXT_MINOR_VERSION, 3)
        GL.set_attribute(SDL_GL_CONTEXT_PROFILE_MASK, SDL_GL_CONTEXT_PROFILE_CORE)
        ctx = GLContext(window)
        ctx.make_current(window)
        window.gl_swap()
        ctx.delete()
    """
    def __init__(self, window):
        win_ptr   = window.ptr if isinstance(window, Window) else window
        self.ptr  = _SDL_GL_CreateContext(win_ptr)
        if self.ptr is None or self.ptr.Address == 0:
            raise SDLError(format_str("SDL_GL_CreateContext: {get_error()}"))

    def make_current(self, window):
        win_ptr = window.ptr if isinstance(window, Window) else window
        _check(_SDL_GL_MakeCurrent(win_ptr, self.ptr), "GL_MakeCurrent")

    def delete(self):
        if self.ptr:
            _SDL_GL_DeleteContext(self.ptr)
            self.ptr = None

class GL:
    """Static helpers for OpenGL context configuration."""
    @staticmethod
    def set_attribute(attr, value):
        _check(_SDL_GL_SetAttribute(attr, value), "GL_SetAttribute")

    @staticmethod
    def get_attribute(attr):
        buf = ffi.malloc(4)
        try:
            _SDL_GL_GetAttribute(attr, buf)
            return _read_int_buf(buf)
        finally:
            ffi.free(buf)

    @staticmethod
    def set_swap_interval(interval):
        _check(_SDL_GL_SetSwapInterval(interval), "GL_SetSwapInterval")

    @staticmethod
    def get_swap_interval():
        return _SDL_GL_GetSwapInterval()

    @staticmethod
    def get_proc_address(name):
        return _SDL_GL_GetProcAddress(_enc(name))

    @staticmethod
    def load_library(path=None):
        p = _enc(path) if path else ffi.c_void_p(0)
        _check(_SDL_GL_LoadLibrary(p), "GL_LoadLibrary")

    @staticmethod
    def unload_library():
        _SDL_GL_UnloadLibrary()

# ==============================================================================
#  Cursor class
# ==============================================================================

class Cursor:
    """
    Usage:
        c = Cursor.system(SDL_SYSTEM_CURSOR_HAND)
        c.set()
        c.free()
    """
    def __init__(self, ptr):
        self.ptr = ptr

    @staticmethod
    def system(cursor_id):
        ptr = _SDL_CreateSystemCursor(cursor_id)
        if ptr is None or ptr.Address == 0:
            raise SDLError(format_str("CreateSystemCursor: {get_error()}"))
        return Cursor(ptr)

    @staticmethod
    def from_surface(surface, hot_x, hot_y):
        sptr = surface.ptr if isinstance(surface, Surface) else surface
        ptr  = _SDL_CreateColorCursor(sptr, hot_x, hot_y)
        return Cursor(ptr)

    def set(self):
        _SDL_SetCursor(self.ptr)

    @staticmethod
    def get_current():
        return Cursor(_SDL_GetCursor())

    @staticmethod
    def show():                 _SDL_ShowCursor(1)

    @staticmethod
    def hide():                 _SDL_ShowCursor(0)

    @staticmethod
    def is_shown():             return _SDL_ShowCursor(-1) != 0

    def free(self):
        if self.ptr:
            _SDL_FreeCursor(self.ptr)
            self.ptr = None

# ==============================================================================
#  Keyboard helpers
# ==============================================================================

class Keyboard:
    """
    Usage:
        if Keyboard.is_pressed(SDL_SCANCODE_SPACE):
            jump()
        mod = Keyboard.get_mod()
        if mod & KMOD_CTRL:
            ...
    """
    _state_ptr = None

    @staticmethod
    def get_state():
        """Returns raw C pointer to the keyboard state array (256 uint8 values)."""
        return _SDL_GetKeyboardState(ffi.c_void_p(0))

    @staticmethod
    def is_pressed(scancode):
        """Returns True if the key with the given scancode is currently held."""
        state = _SDL_GetKeyboardState(ffi.c_void_p(0))
        val   = ffi.read_memory_with_offset(state, scancode, ffi.c_uint8)
        return val != 0

    @staticmethod
    def get_mod():              return _SDL_GetModState()
    @staticmethod
    def set_mod(mod):           _SDL_SetModState(mod)
    @staticmethod
    def mod_has(flag):          return (_SDL_GetModState() & flag) != 0
    @staticmethod
    def get_key_name(keycode):
        raw = _SDL_GetKeyName(keycode)
        return raw if raw else ""
    @staticmethod
    def start_text_input():     _SDL_StartTextInput()
    @staticmethod
    def stop_text_input():      _SDL_StopTextInput()
    @staticmethod
    def is_text_input_active(): return _SDL_IsTextInputActive() != 0
    @staticmethod
    def set_text_input_rect(x, y, w, h):
        buf = make_rect(x, y, w, h)
        try:
            _SDL_SetTextInputRect(buf)
        finally:
            ffi.free(buf)

# ==============================================================================
#  Mouse helpers
# ==============================================================================

class Mouse:
    """
    Usage:
        x, y, buttons = Mouse.get_state()
        if buttons & SDL_BUTTON_LMASK:
            print("Left button held at", x, y)
    """
    @staticmethod
    def get_state():
        xb = ffi.malloc(4); yb = ffi.malloc(4)
        try:
            mask = _SDL_GetMouseState(xb, yb)
            return (_read_int_buf(xb), _read_int_buf(yb), mask)
        finally:
            ffi.free(xb); ffi.free(yb)

    @staticmethod
    def get_global_state():
        xb = ffi.malloc(4); yb = ffi.malloc(4)
        try:
            mask = _SDL_GetGlobalMouseState(xb, yb)
            return (_read_int_buf(xb), _read_int_buf(yb), mask)
        finally:
            ffi.free(xb); ffi.free(yb)

    @staticmethod
    def get_relative_state():
        xb = ffi.malloc(4); yb = ffi.malloc(4)
        try:
            mask = _SDL_GetRelativeMouseState(xb, yb)
            return (_read_int_buf(xb), _read_int_buf(yb), mask)
        finally:
            ffi.free(xb); ffi.free(yb)

    @staticmethod
    def set_relative_mode(en):
        _check(_SDL_SetRelativeMouseMode(1 if en else 0), "SetRelativeMouseMode")

    @staticmethod
    def is_relative_mode():     return _SDL_GetRelativeMouseMode() != 0

    @staticmethod
    def warp_global(x, y):      _SDL_WarpMouseGlobal(x, y)

    @staticmethod
    def button_held(mask, button):
        return (mask & (1 << (button - 1))) != 0

# ==============================================================================
#  Clipboard helpers
# ==============================================================================

class Clipboard:
    @staticmethod
    def get():
        raw = _SDL_GetClipboardText()
        return raw if raw else ""

    @staticmethod
    def set(text):
        _check(_SDL_SetClipboardText(_enc(text)), "SetClipboardText")

    @staticmethod
    def has_text():             return _SDL_HasClipboardText() != 0

# ==============================================================================
#  Display info helpers
# ==============================================================================

class DisplayInfo:
    @staticmethod
    def get_num_displays():     return _SDL_GetNumVideoDisplays()

    @staticmethod
    def get_name(display_index):
        raw = _SDL_GetDisplayName(display_index)
        return raw if raw else ""

    @staticmethod
    def get_current_mode(display_index):
        """Returns dict with keys: format, w, h, refresh_rate, driverdata."""
        buf = ffi.malloc(_SZ_DISPLAYMODE)
        try:
            _check(_SDL_GetCurrentDisplayMode(display_index, buf), "GetCurrentDisplayMode")
            return {
                "format":       _read_u32(buf,  0),
                "w":            _read_i32(buf,  4),
                "h":            _read_i32(buf,  8),
                "refresh_rate": _read_i32(buf, 12),
            }
        finally:
            ffi.free(buf)

    @staticmethod
    def get_desktop_mode(display_index):
        buf = ffi.malloc(_SZ_DISPLAYMODE)
        try:
            _check(_SDL_GetDesktopDisplayMode(display_index, buf), "GetDesktopDisplayMode")
            return {
                "format":       _read_u32(buf,  0),
                "w":            _read_i32(buf,  4),
                "h":            _read_i32(buf,  8),
                "refresh_rate": _read_i32(buf, 12),
            }
        finally:
            ffi.free(buf)

    @staticmethod
    def get_bounds(display_index):
        buf = ffi.malloc(_SZ_RECT)
        try:
            _check(_SDL_GetDisplayBounds(display_index, buf), "GetDisplayBounds")
            return {
                "x": _read_i32(buf,  0),
                "y": _read_i32(buf,  4),
                "w": _read_i32(buf,  8),
                "h": _read_i32(buf, 12),
            }
        finally:
            ffi.free(buf)

    @staticmethod
    def get_dpi(display_index):
        db = ffi.malloc(4); hb = ffi.malloc(4); vb = ffi.malloc(4)
        try:
            _SDL_GetDisplayDPI(display_index, db, hb, vb)
            return {
                "diagonal":   _read_f32(db, 0),
                "horizontal": _read_f32(hb, 0),
                "vertical":   _read_f32(vb, 0),
            }
        finally:
            ffi.free(db); ffi.free(hb); ffi.free(vb)

# ==============================================================================
#  Timer helpers
# ==============================================================================

def get_ticks():
    return _SDL_GetTicks()

def get_ticks64():
    return _SDL_GetTicks64()

def get_performance_counter():
    return _SDL_GetPerformanceCounter()

def get_performance_frequency():
    return _SDL_GetPerformanceFrequency()

def delay(ms):
    _SDL_Delay(ms)

class PerformanceTimer:
    """High-resolution timer using SDL's performance counter."""
    def __init__(self):
        self._freq  = _SDL_GetPerformanceFrequency()
        self._start = _SDL_GetPerformanceCounter()

    def reset(self):
        self._start = _SDL_GetPerformanceCounter()

    def elapsed_ms(self):
        now = _SDL_GetPerformanceCounter()
        return (now - self._start) * 1000 // self._freq

    def elapsed_us(self):
        now = _SDL_GetPerformanceCounter()
        return (now - self._start) * 1000000 // self._freq

    def elapsed_s(self):
        now = _SDL_GetPerformanceCounter()
        return (now - self._start) / self._freq

# ==============================================================================
#  Core functions
# ==============================================================================

def init(flags=SDL_INIT_EVERYTHING):
    """Initialize SDL2. Call once at startup."""
    ret = _SDL_Init(flags)
    if ret < 0:
        raise SDLError(format_str("SDL_Init failed: {get_error()}"))

def init_subsystem(flags):
    _check(_SDL_InitSubSystem(flags), "InitSubSystem")

def quit_subsystem(flags):
    _SDL_QuitSubSystem(flags)

def quit():
    """Shut down SDL2. Call at program exit."""
    _SDL_Quit()

def was_init(flags=SDL_INIT_EVERYTHING):
    return _SDL_WasInit(flags)

def set_hint(name, value):
    _SDL_SetHint(_enc(name), _enc(value))

def get_hint(name):
    raw = _SDL_GetHint(_enc(name))
    return raw if raw else ""

def get_version():
    """Returns (major, minor, patch) tuple."""
    buf = ffi.malloc(4)   # SDL_version: uint8 major, minor, patch + 1 pad
    try:
        _SDL_GetVersion(buf)
        raw = ffi.buffer_to_bytes(buf, 4)
        return (raw[0], raw[1], raw[2])
    finally:
        ffi.free(buf)

def get_revision():
    raw = _SDL_GetRevision()
    return raw if raw else ""

def pump_events():
    _SDL_FlushEvents(0, SDL_LASTEVENT)

def flush_event(event_type):
    _SDL_FlushEvent(event_type)

def register_events(n=1):
    """Reserve n custom event types. Returns the first assigned type id."""
    return _SDL_RegisterEvents(n)

# ==============================================================================
#  LVGL ↔ SDL2 backend
# ==============================================================================

class LVGLSDLBackend:
    """
    Connects LVGL to SDL2.  Creates an SDL2 window+renderer, a streaming
    texture as LVGL's framebuffer, and wires up flush and input callbacks.

    Usage:
        import lvgl
        import sdl as sdl

        sdl.init()
        lvgl.init()

        backend = sdl.LVGLSDLBackend(800, 480, "My App")
        backend.attach()      # registers display + indev with LVGL

        ev = sdl.Event()
        running = True
        while running:
            while ev.poll():
                if ev.type == sdl.SDL_QUIT:
                    running = False
                backend.feed_event(ev)

            lvgl.tick_inc(backend.tick())
            lvgl.timer_handler()

        backend.destroy()
        sdl.quit()
    """
    def __init__(self, width, height, title="LVGL", window_flags=SDL_WINDOW_SHOWN | SDL_WINDOW_RESIZABLE):
        self.w = width
        self.h = height

        self.window   = Window(title, width, height, window_flags)
        self.renderer = Renderer(self.window, flags=0)

        # LVGL renders ARGB8888 by default for 32-bit color depth
        self.texture  = Texture(self.renderer, SDL_PIXELFORMAT_ARGB8888, SDL_TEXTUREACCESS_STREAMING, width, height)

        # Allocate LVGL draw buffer (1/10 screen)
        self._buf_size = width * height // 10 * 4
        self._buf1     = ffi.malloc(self._buf_size)
        self._buf2     = ffi.malloc(self._buf_size)

        self._flush_cb_ref   = None
        self._input_cb_ref   = None
        self._lv_display     = None
        self._lv_indev       = None
        self._last_ticks     = get_ticks()

        # Mouse state for LVGL indev
        self._mouse_x     = 0
        self._mouse_y     = 0
        self._mouse_down  = False

    def attach(self):
        """Register this backend with LVGL. Call after lvgl.init()."""
        import lvgl

        # --- Display ---
        self._lv_display = lvgl.Display(self.w, self.h)
        self._lv_display.set_buffers(self._buf1, self._buf2, self._buf_size, lvgl.LV_DISPLAY_RENDER_MODE_PARTIAL)

        def _flush(disp_ptr, area_ptr, px_map):
            # Read area (lv_area_t: int32 x1, y1, x2, y2)
            x1 = ffi.read_memory_with_offset(area_ptr,  0, ffi.c_int32)
            y1 = ffi.read_memory_with_offset(area_ptr,  4, ffi.c_int32)
            x2 = ffi.read_memory_with_offset(area_ptr,  8, ffi.c_int32)
            y2 = ffi.read_memory_with_offset(area_ptr, 12, ffi.c_int32)
            w  = x2 - x1 + 1
            h  = y2 - y1 + 1

            self.texture.update((x1, y1, w, h), px_map, w * 4)
            self.renderer.set_draw_color(0, 0, 0, 255)
            self.renderer.clear()
            self.renderer.copy(self.texture)
            self.renderer.present()

            lvgl.Display.flush_ready_ptr(disp_ptr)

        self._flush_cb_ref = ffi.callback(_flush, None, [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p])
        self._lv_display.set_flush_cb(_flush)

        # --- Input device ---
        self._lv_indev = lvgl.InputDevice(lvgl.LV_INDEV_TYPE_POINTER)
        self._lv_indev.set_display(self._lv_display)

        def _read_input(indev_ptr, data_ptr):
            lvgl.InputDevice.write_pointer(data_ptr, self._mouse_x, self._mouse_y, lvgl.LV_INDEV_STATE_PRESSED if self._mouse_down else lvgl.LV_INDEV_STATE_RELEASED)

        self._input_cb_ref = ffi.callback(_read_input, None, [ffi.c_void_p, ffi.c_void_p])
        self._lv_indev.set_read_cb(_read_input)

    def feed_event(self, event):
        """
        Feed an SDL event into the LVGL input state.
        Call from your event loop for every event before lvgl.timer_handler().
        """
        t = event.type
        if t == SDL_MOUSEMOTION:
            self._mouse_x = event.x
            self._mouse_y = event.y
        elif t == SDL_MOUSEBUTTONDOWN:
            if event.button == SDL_BUTTON_LEFT:
                self._mouse_down = True
                self._mouse_x    = event.x
                self._mouse_y    = event.y
        elif t == SDL_MOUSEBUTTONUP:
            if event.button == SDL_BUTTON_LEFT:
                self._mouse_down = False
        elif t == SDL_FINGERDOWN or t == SDL_FINGERMOTION:
            # Scale normalized [0..1] touch coords to display pixels
            self._mouse_x    = int(event.finger_x * self.w)
            self._mouse_y    = int(event.finger_y * self.h)
            self._mouse_down = True
        elif t == SDL_FINGERUP:
            self._mouse_down = False

    def tick(self):
        """
        Returns ms elapsed since last call.
        Pass to lvgl.tick_inc() in your main loop.
        """
        now    = get_ticks()
        delta  = now - self._last_ticks
        self._last_ticks = now
        return delta

    def get_mouse(self):
        """Return (x, y, is_pressed) tuple."""
        return (self._mouse_x, self._mouse_y, self._mouse_down)

    def destroy(self):
        if self._lv_indev:
            self._lv_indev.delete()
            self._lv_indev = None
        if self._lv_display:
            self._lv_display.delete()
            self._lv_display = None
        if self._buf1:
            ffi.free(self._buf1)
            self._buf1 = None
        if self._buf2:
            ffi.free(self._buf2)
            self._buf2 = None
        self.texture.destroy()
        self.renderer.destroy()
        self.window.destroy()

# ==============================================================================
#  Nuklear ↔ SDL2 input bridge
# ==============================================================================

class NuklearSDLBridge:
    """
    Feeds SDL2 events into a Nuklear context.
    The Nuklear wrapper (nk) must be imported separately.
    This bridge only handles the SDL → nk_input translation.

    Usage:
        bridge = NuklearSDLBridge(nk_ctx_ptr)
        ev = sdl.Event()
        while True:
            bridge.begin()
            while ev.poll():
                if ev.type == sdl.SDL_QUIT: break
                bridge.feed(ev)
            bridge.end()
            # ... nk layout + render ...
    """
    def __init__(self, nk_ctx_ptr, nk_module=None):
        self._ctx = nk_ctx_ptr
        self._nk  = nk_module  # the nk wrapper — set later via set_nk()
        self._scroll_x = 0.0
        self._scroll_y = 0.0

    def set_nk(self, nk_module):
        self._nk = nk_module

    def begin(self):
        if self._nk:
            self._nk.input_begin(self._ctx)
        self._scroll_x = 0.0
        self._scroll_y = 0.0

    def end(self):
        if self._nk:
            self._nk.input_scroll(self._ctx, self._scroll_x, self._scroll_y)
            self._nk.input_end(self._ctx)

    def feed(self, event):
        """
        Translate one SDL event into nk_input_* calls.
        If nk module is not yet set, the event state is still tracked
        so it can be replayed once set_nk() is called.
        """
        nk = self._nk
        if nk is None:
            return None

        t = event.type

        if t == SDL_KEYDOWN or t == SDL_KEYUP:
            down  = (t == SDL_KEYDOWN)
            mod   = event.mod
            sc    = event.scancode
            sym   = event.sym

            ctrl  = (mod & KMOD_CTRL)  != 0
            shift = (mod & KMOD_SHIFT) != 0

            if sc == SDL_SCANCODE_LSHIFT or sc == SDL_SCANCODE_RSHIFT:
                nk.input_key(self._ctx, nk.NK_KEY_SHIFT, down)
            elif sc == SDL_SCANCODE_DELETE:
                nk.input_key(self._ctx, nk.NK_KEY_DEL, down)
            elif sc == SDL_SCANCODE_RETURN or sc == SDL_SCANCODE_KP_ENTER:
                nk.input_key(self._ctx, nk.NK_KEY_ENTER, down)
            elif sc == SDL_SCANCODE_TAB:
                nk.input_key(self._ctx, nk.NK_KEY_TAB, down)
            elif sc == SDL_SCANCODE_BACKSPACE:
                nk.input_key(self._ctx, nk.NK_KEY_BACKSPACE, down)
            elif sc == SDL_SCANCODE_HOME:
                nk.input_key(self._ctx, nk.NK_KEY_TEXT_START, down)
                nk.input_key(self._ctx, nk.NK_KEY_SCROLL_START, down)
            elif sc == SDL_SCANCODE_END:
                nk.input_key(self._ctx, nk.NK_KEY_TEXT_END, down)
                nk.input_key(self._ctx, nk.NK_KEY_SCROLL_END, down)
            elif sc == SDL_SCANCODE_PAGEDOWN:
                nk.input_key(self._ctx, nk.NK_KEY_SCROLL_DOWN, down)
            elif sc == SDL_SCANCODE_PAGEUP:
                nk.input_key(self._ctx, nk.NK_KEY_SCROLL_UP, down)
            elif sc == SDL_SCANCODE_Z and ctrl:
                nk.input_key(self._ctx, nk.NK_KEY_TEXT_UNDO, down)
            elif sc == SDL_SCANCODE_Y and ctrl:
                nk.input_key(self._ctx, nk.NK_KEY_TEXT_REDO, down)
            elif sc == SDL_SCANCODE_C and ctrl:
                nk.input_key(self._ctx, nk.NK_KEY_COPY, down)
            elif sc == SDL_SCANCODE_V and ctrl:
                nk.input_key(self._ctx, nk.NK_KEY_PASTE, down)
            elif sc == SDL_SCANCODE_X and ctrl:
                nk.input_key(self._ctx, nk.NK_KEY_CUT, down)
            elif sc == SDL_SCANCODE_A and ctrl:
                nk.input_key(self._ctx, nk.NK_KEY_TEXT_SELECT_ALL, down)
            elif sc == SDL_SCANCODE_LEFT:
                if ctrl:
                    nk.input_key(self._ctx, nk.NK_KEY_TEXT_WORD_LEFT, down)
                else:
                    nk.input_key(self._ctx, nk.NK_KEY_LEFT, down)
            elif sc == SDL_SCANCODE_RIGHT:
                if ctrl:
                    nk.input_key(self._ctx, nk.NK_KEY_TEXT_WORD_RIGHT, down)
                else:
                    nk.input_key(self._ctx, nk.NK_KEY_RIGHT, down)
            elif sc == SDL_SCANCODE_UP:
                nk.input_key(self._ctx, nk.NK_KEY_UP, down)
            elif sc == SDL_SCANCODE_DOWN:
                nk.input_key(self._ctx, nk.NK_KEY_DOWN, down)

        elif t == SDL_MOUSEBUTTONDOWN or t == SDL_MOUSEBUTTONUP:
            down = (t == SDL_MOUSEBUTTONDOWN)
            b    = event.button
            x    = event.x
            y    = event.y
            if b == SDL_BUTTON_LEFT:
                if event.clicks == 2:
                    nk.input_button(self._ctx, nk.NK_BUTTON_DOUBLE, x, y, down)
                nk.input_button(self._ctx, nk.NK_BUTTON_LEFT, x, y, down)
            elif b == SDL_BUTTON_MIDDLE:
                nk.input_button(self._ctx, nk.NK_BUTTON_MIDDLE, x, y, down)
            elif b == SDL_BUTTON_RIGHT:
                nk.input_button(self._ctx, nk.NK_BUTTON_RIGHT, x, y, down)

        elif t == SDL_MOUSEMOTION:
            nk.input_motion(self._ctx, event.x, event.y)

        elif t == SDL_MOUSEWHEEL:
            self._scroll_x = self._scroll_x + event.wheel_x
            self._scroll_y = self._scroll_y + event.wheel_y

        elif t == SDL_TEXTINPUT:
            text = event.text
            for ch in text:
                nk.input_char(self._ctx, ch)

# ==============================================================================
#  Convenience: create a standard window + renderer pair
# ==============================================================================

def create_window_renderer(title, w, h, win_flags=SDL_WINDOW_SHOWN | SDL_WINDOW_RESIZABLE, ren_flags=0):
    """
    One-call setup for the most common case.
    Returns (Window, Renderer).
    """
    win = Window(title, w, h, win_flags)
    ren = Renderer(win, flags=ren_flags)
    return (win, ren)

def create_gl_window(title, w, h, major=3, minor=3, profile=SDL_GL_CONTEXT_PROFILE_CORE, depth_bits=24):
    """
    One-call setup for an OpenGL window.
    Returns (Window, GLContext).
    """
    GL.set_attribute(SDL_GL_CONTEXT_MAJOR_VERSION, major)
    GL.set_attribute(SDL_GL_CONTEXT_MINOR_VERSION, minor)
    GL.set_attribute(SDL_GL_CONTEXT_PROFILE_MASK,  profile)
    GL.set_attribute(SDL_GL_DEPTH_SIZE,             depth_bits)
    GL.set_attribute(SDL_GL_DOUBLEBUFFER,           1)

    win = Window(title, w, h, SDL_WINDOW_OPENGL | SDL_WINDOW_SHOWN)
    ctx = GLContext(win)
    ctx.make_current(win)
    GL.set_swap_interval(1)   # vsync
    return (win, ctx)