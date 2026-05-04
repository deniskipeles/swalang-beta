import ffi
import sys

def _load_library():
    platform = sys.platform
    candidates = []
    if platform == 'linux':
        candidates = ["bin/x86_64-linux/sdl2/libSDL2-2.0.so.0", "bin/x86_64-linux/sdl2/libSDL2.so", "libSDL2.so"]
    elif platform == 'windows':
        candidates = ["bin/x86_64-windows-gnu/sdl2/SDL2.dll", "SDL2.dll"]
    elif platform == 'darwin':
        candidates = ["libSDL2.dylib"]
    
    for name in candidates:
        try:
            return ffi.CDLL(name)
        except ffi.FFIError:
            pass
    raise ffi.FFIError("Could not load SDL2 shared library. Ensure it is built and in the bin/ directory.")

_lib = _load_library()

# =============================================================================
# Constants
# =============================================================================

# Init Flags
INIT_TIMER          = 0x00000001
INIT_AUDIO          = 0x00000010
INIT_VIDEO          = 0x00000020
INIT_JOYSTICK       = 0x00000200
INIT_HAPTIC         = 0x00001000
INIT_GAMECONTROLLER = 0x00002000
INIT_EVENTS         = 0x00004000
INIT_EVERYTHING     = 0x0000F231

# Window Position
WINDOWPOS_UNDEFINED = 0x1FFF0000
WINDOWPOS_CENTERED  = 0x2FFF0000

# Window Flags
WINDOW_FULLSCREEN   = 0x00000001
WINDOW_OPENGL       = 0x00000002
WINDOW_SHOWN        = 0x00000004
WINDOW_HIDDEN       = 0x00000008
WINDOW_BORDERLESS   = 0x00000010
WINDOW_RESIZABLE    = 0x00000020
WINDOW_MINIMIZED    = 0x00000040
WINDOW_MAXIMIZED    = 0x00000080
WINDOW_INPUT_FOCUS  = 0x00000200
WINDOW_MOUSE_FOCUS  = 0x00000400
WINDOW_HIGHPIXEL    = 0x00002000

# Renderer Flags
RENDERER_SOFTWARE      = 0x00000001
RENDERER_ACCELERATED   = 0x00000002
RENDERER_PRESENTVSYNC  = 0x00000004
RENDERER_TARGETTEXTURE = 0x00000008

# Texture Formats and Access
PIXELFORMAT_ARGB8888   = 0x16362004
TEXTUREACCESS_STREAMING = 1

# Event Types
QUIT                 = 0x100
WINDOWEVENT          = 0x200
SYSWMEVENT           = 0x201
KEYDOWN              = 0x300
KEYUP                = 0x301
TEXTEDITING          = 0x302
TEXTINPUT            = 0x303
MOUSEMOTION          = 0x400
MOUSEBUTTONDOWN      = 0x401
MOUSEBUTTONUP        = 0x402
MOUSEWHEEL           = 0x403

# =============================================================================
# Structs
# =============================================================================

@ffi.Struct
class Rect:
    _fields_ = [
        ('x', ffi.c_int32),
        ('y', ffi.c_int32),
        ('w', ffi.c_int32),
        ('h', ffi.c_int32)
    ]

@ffi.Struct
class Point:
    _fields_ = [
        ('x', ffi.c_int32),
        ('y', ffi.c_int32)
    ]

@ffi.Struct
class Color:
    _fields_ = [
        ('r', ffi.c_uint8),
        ('g', ffi.c_uint8),
        ('b', ffi.c_uint8),
        ('a', ffi.c_uint8)
    ]

# =============================================================================
# C Signatures
# =============================================================================

_SDL_Init = _lib.SDL_Init([ffi.c_uint32], ffi.c_int32)
_SDL_Quit = _lib.SDL_Quit([], None)
_SDL_GetError = _lib.SDL_GetError([], ffi.c_char_p)
_SDL_GetTicks = _lib.SDL_GetTicks([], ffi.c_uint32)
_SDL_Delay = _lib.SDL_Delay([ffi.c_uint32], None)

_SDL_CreateWindow = _lib.SDL_CreateWindow([ffi.c_char_p, ffi.c_int32, ffi.c_int32, ffi.c_int32, ffi.c_int32, ffi.c_uint32], ffi.c_void_p)
_SDL_DestroyWindow = _lib.SDL_DestroyWindow([ffi.c_void_p], None)
_SDL_GetWindowSize = _lib.SDL_GetWindowSize([ffi.c_void_p, ffi.POINTER(ffi.c_int32), ffi.POINTER(ffi.c_int32)], None)
_SDL_GetWindowFlags = _lib.SDL_GetWindowFlags([ffi.c_void_p], ffi.c_uint32)

_SDL_CreateRenderer = _lib.SDL_CreateRenderer([ffi.c_void_p, ffi.c_int32, ffi.c_uint32], ffi.c_void_p)
_SDL_DestroyRenderer = _lib.SDL_DestroyRenderer([ffi.c_void_p], None)
_SDL_RenderClear = _lib.SDL_RenderClear([ffi.c_void_p], ffi.c_int32)
_SDL_RenderPresent = _lib.SDL_RenderPresent([ffi.c_void_p], None)
_SDL_SetRenderDrawColor = _lib.SDL_SetRenderDrawColor([ffi.c_void_p, ffi.c_uint8, ffi.c_uint8, ffi.c_uint8, ffi.c_uint8], ffi.c_int32)
_SDL_RenderFillRect = _lib.SDL_RenderFillRect([ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_SDL_RenderCopy = _lib.SDL_RenderCopy([ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], ffi.c_int32)

_SDL_CreateTexture = _lib.SDL_CreateTexture([ffi.c_void_p, ffi.c_uint32, ffi.c_int32, ffi.c_int32, ffi.c_int32], ffi.c_void_p)
_SDL_UpdateTexture = _lib.SDL_UpdateTexture([ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_int32], ffi.c_int32)
_SDL_DestroyTexture = _lib.SDL_DestroyTexture([ffi.c_void_p], None)

_SDL_PollEvent = _lib.SDL_PollEvent([ffi.c_void_p], ffi.c_int32)
_SDL_GetMouseState = _lib.SDL_GetMouseState([ffi.POINTER(ffi.c_int32), ffi.POINTER(ffi.c_int32)], ffi.c_uint32)
_SDL_GetKeyboardState = _lib.SDL_GetKeyboardState([ffi.POINTER(ffi.c_int32)], ffi.c_void_p)

# =============================================================================
# High-Level API
# =============================================================================

class SDLError(Exception):
    pass

class SDL:
    @staticmethod
    def init(flags=INIT_VIDEO | INIT_EVENTS):
        if _SDL_Init(flags) < 0:
            raise SDLError(format_str("SDL_Init Error: {SDL.get_error()}"))

    @staticmethod
    def quit():
        _SDL_Quit()

    @staticmethod
    def get_error():
        err_ptr = _SDL_GetError()
        if not err_ptr or err_ptr.Address == 0:
            return ""
        return ffi.string_at(err_ptr)

    @staticmethod
    def delay(ms):
        _SDL_Delay(ms)

    @staticmethod
    def get_ticks():
        return _SDL_GetTicks()

    @staticmethod
    def get_mouse_state():
        x_ptr = ffi.malloc(4)
        y_ptr = ffi.malloc(4)
        try:
            mask = _SDL_GetMouseState(x_ptr, y_ptr)
            x = ffi.read_memory(x_ptr, ffi.c_int32)
            y = ffi.read_memory(y_ptr, ffi.c_int32)
            return x, y, mask
        finally:
            ffi.free(x_ptr)
            ffi.free(y_ptr)


class Window:
    def __init__(self, title, w, h, flags=WINDOW_SHOWN):
        self.ptr = _SDL_CreateWindow(
            title.encode('utf-8'), 
            WINDOWPOS_CENTERED, WINDOWPOS_CENTERED, 
            w, h, flags
        )
        if not self.ptr or self.ptr.Address == 0:
            raise SDLError(format_str("SDL_CreateWindow Error: {SDL.get_error()}"))

    @property
    def size(self):
        w_ptr = ffi.malloc(4)
        h_ptr = ffi.malloc(4)
        try:
            _SDL_GetWindowSize(self.ptr, w_ptr, h_ptr)
            w = ffi.read_memory(w_ptr, ffi.c_int32)
            h = ffi.read_memory(h_ptr, ffi.c_int32)
            return w, h
        finally:
            ffi.free(w_ptr)
            ffi.free(h_ptr)

    @property
    def flags(self):
        return _SDL_GetWindowFlags(self.ptr)

    def destroy(self):
        if self.ptr:
            _SDL_DestroyWindow(self.ptr)
            self.ptr = None


class Texture:
    def __init__(self, renderer, format, access, w, h):
        self.ptr = _SDL_CreateTexture(renderer.ptr, format, access, w, h)
        if not self.ptr or self.ptr.Address == 0:
            raise SDLError(format_str("SDL_CreateTexture Error: {SDL.get_error()}"))

    def update(self, rect_ptr, pixels_ptr, pitch):
        _SDL_UpdateTexture(self.ptr, rect_ptr, pixels_ptr, pitch)

    def destroy(self):
        if self.ptr:
            _SDL_DestroyTexture(self.ptr)
            self.ptr = None


class Renderer:
    def __init__(self, window, index=-1, flags=RENDERER_ACCELERATED | RENDERER_PRESENTVSYNC):
        self.ptr = _SDL_CreateRenderer(window.ptr, index, flags)
        if not self.ptr or self.ptr.Address == 0:
            raise SDLError(format_str("SDL_CreateRenderer Error: {SDL.get_error()}"))

    def set_draw_color(self, r, g, b, a=255):
        _SDL_SetRenderDrawColor(self.ptr, r, g, b, a)

    def clear(self):
        _SDL_RenderClear(self.ptr)

    def present(self):
        _SDL_RenderPresent(self.ptr)

    def fill_rect(self, x, y, w, h):
        rect_ptr = ffi.malloc(16)
        try:
            ffi.write_memory_with_offset(rect_ptr, 0, ffi.c_int32, x)
            ffi.write_memory_with_offset(rect_ptr, 4, ffi.c_int32, y)
            ffi.write_memory_with_offset(rect_ptr, 8, ffi.c_int32, w)
            ffi.write_memory_with_offset(rect_ptr, 12, ffi.c_int32, h)
            _SDL_RenderFillRect(self.ptr, rect_ptr)
        finally:
            ffi.free(rect_ptr)

    def copy(self, texture, srcrect_ptr=None, dstrect_ptr=None):
        _SDL_RenderCopy(self.ptr, texture.ptr, srcrect_ptr, dstrect_ptr)

    def destroy(self):
        if self.ptr:
            _SDL_DestroyRenderer(self.ptr)
            self.ptr = None


class Event:
    def __init__(self):
        self._buf = ffi.malloc(56)

    def poll(self):
        return _SDL_PollEvent(self._buf) > 0

    @property
    def type(self):
        return ffi.read_memory(self._buf, ffi.c_uint32)

    # --- Mouse Motion ---
    @property
    def motion_x(self): return ffi.read_memory_with_offset(self._buf, 20, ffi.c_int32)
    @property
    def motion_y(self): return ffi.read_memory_with_offset(self._buf, 24, ffi.c_int32)
    @property
    def motion_xrel(self): return ffi.read_memory_with_offset(self._buf, 28, ffi.c_int32)
    @property
    def motion_yrel(self): return ffi.read_memory_with_offset(self._buf, 32, ffi.c_int32)

    # --- Mouse Button ---
    @property
    def button(self): return ffi.read_memory_with_offset(self._buf, 16, ffi.c_uint8)
    @property
    def button_state(self): return ffi.read_memory_with_offset(self._buf, 17, ffi.c_uint8)
    @property
    def button_x(self): return ffi.read_memory_with_offset(self._buf, 20, ffi.c_int32)
    @property
    def button_y(self): return ffi.read_memory_with_offset(self._buf, 24, ffi.c_int32)

    # --- Keyboard ---
    @property
    def key_scancode(self): return ffi.read_memory_with_offset(self._buf, 16, ffi.c_int32)
    @property
    def key_sym(self): return ffi.read_memory_with_offset(self._buf, 20, ffi.c_int32)
    @property
    def key_mod(self): return ffi.read_memory_with_offset(self._buf, 24, ffi.c_uint16)

    # --- Window ---
    @property
    def window_event(self): return ffi.read_memory_with_offset(self._buf, 12, ffi.c_uint8)
    @property
    def window_data1(self): return ffi.read_memory_with_offset(self._buf, 16, ffi.c_int32)
    @property
    def window_data2(self): return ffi.read_memory_with_offset(self._buf, 20, ffi.c_int32)

    def free(self):
        if self._buf:
            ffi.free(self._buf)
            self._buf = None