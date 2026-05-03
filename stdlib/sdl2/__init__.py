# stdlib/sdl2/__init__.py

import ffi
import sys

def _load_library():
    platform = sys.platform
    candidates = []
    if platform == 'linux':
        candidates = ["bin/x86_64-linux/sdl2/libSDL2-2.0.so", "libSDL2-2.0.so.0", "libSDL2.so"]
    elif platform == 'windows':
        candidates = ["bin/x86_64-windows-gnu/sdl2/SDL2.dll", "SDL2.dll"]

    for name in candidates:
        try:
            return ffi.CDLL(name)
        except ffi.FFIError:
            pass
    raise ffi.FFIError("Could not load SDL2 shared library")

_lib = _load_library()

# --- Constants ---
SDL_INIT_TIMER          = 0x00000001
SDL_INIT_AUDIO          = 0x00000010
SDL_INIT_VIDEO          = 0x00000020
SDL_INIT_JOYSTICK       = 0x00000200
SDL_INIT_HAPTIC         = 0x00001000
SDL_INIT_GAMECONTROLLER = 0x00002000
SDL_INIT_EVENTS         = 0x00004000
SDL_INIT_SENSOR         = 0x00008000
SDL_INIT_EVERYTHING     = 0x0000FFFF

SDL_WINDOW_FULLSCREEN = 0x00000001
SDL_WINDOW_OPENGL = 0x00000002
SDL_WINDOW_SHOWN = 0x00000004
SDL_WINDOW_HIDDEN = 0x00000008
SDL_WINDOW_BORDERLESS = 0x00000010
SDL_WINDOW_RESIZABLE = 0x00000020
SDL_WINDOW_MINIMIZED = 0x00000040
SDL_WINDOW_MAXIMIZED = 0x00000080
SDL_WINDOW_INPUT_GRABBED = 0x00000100
SDL_WINDOW_INPUT_FOCUS = 0x00000200
SDL_WINDOW_MOUSE_FOCUS = 0x00000400
SDL_WINDOW_FULLSCREEN_DESKTOP = 0x00001001
SDL_WINDOW_ALLOW_HIGHDPI = 0x00002000

SDL_WINDOWPOS_UNDEFINED = 0x1FFF0000
SDL_WINDOWPOS_CENTERED = 0x2FFF0000

SDL_RENDERER_SOFTWARE = 0x00000001
SDL_RENDERER_ACCELERATED = 0x00000002
SDL_RENDERER_PRESENTVSYNC = 0x00000004
SDL_RENDERER_TARGETTEXTURE = 0x00000008

# Event Types
SDL_QUIT = 0x100
SDL_WINDOWEVENT = 0x200
SDL_KEYDOWN = 0x300
SDL_KEYUP = 0x301
SDL_MOUSEMOTION = 0x400
SDL_MOUSEBUTTONDOWN = 0x401
SDL_MOUSEBUTTONUP = 0x402
SDL_MOUSEWHEEL = 0x403

# Window Event IDs
SDL_WINDOWEVENT_NONE = 0
SDL_WINDOWEVENT_SHOWN = 1
SDL_WINDOWEVENT_HIDDEN = 2
SDL_WINDOWEVENT_EXPOSED = 3
SDL_WINDOWEVENT_MOVED = 4
SDL_WINDOWEVENT_RESIZED = 5
SDL_WINDOWEVENT_SIZE_CHANGED = 6
SDL_WINDOWEVENT_MINIMIZED = 7
SDL_WINDOWEVENT_MAXIMIZED = 8
SDL_WINDOWEVENT_RESTORED = 9
SDL_WINDOWEVENT_ENTER = 10
SDL_WINDOWEVENT_LEAVE = 11
SDL_WINDOWEVENT_FOCUS_GAINED = 12
SDL_WINDOWEVENT_FOCUS_LOST = 13
SDL_WINDOWEVENT_CLOSE = 14

# Button states
SDL_PRESSED = 1
SDL_RELEASED = 0

# --- FFI Function Definitions ---
_SDL_Init = _lib.SDL_Init([ffi.c_uint32], ffi.c_int32)
_SDL_Quit = _lib.SDL_Quit([], None)
_SDL_GetError = _lib.SDL_GetError([], ffi.c_char_p)

_SDL_CreateWindow = _lib.SDL_CreateWindow([ffi.c_char_p, ffi.c_int32, ffi.c_int32, ffi.c_int32, ffi.c_int32, ffi.c_uint32], ffi.c_void_p)
_SDL_DestroyWindow = _lib.SDL_DestroyWindow([ffi.c_void_p], None)
_SDL_SetWindowTitle = _lib.SDL_SetWindowTitle([ffi.c_void_p, ffi.c_char_p], None)

_SDL_CreateRenderer = _lib.SDL_CreateRenderer([ffi.c_void_p, ffi.c_int32, ffi.c_uint32], ffi.c_void_p)
_SDL_DestroyRenderer = _lib.SDL_DestroyRenderer([ffi.c_void_p], None)
_SDL_RenderClear = _lib.SDL_RenderClear([ffi.c_void_p], ffi.c_int32)
_SDL_RenderPresent = _lib.SDL_RenderPresent([ffi.c_void_p], None)
_SDL_SetRenderDrawColor = _lib.SDL_SetRenderDrawColor([ffi.c_void_p, ffi.c_uint8, ffi.c_uint8, ffi.c_uint8, ffi.c_uint8], ffi.c_int32)
_SDL_RenderDrawRect = _lib.SDL_RenderDrawRect([ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_SDL_RenderFillRect = _lib.SDL_RenderFillRect([ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_SDL_RenderDrawLine = _lib.SDL_RenderDrawLine([ffi.c_void_p, ffi.c_int32, ffi.c_int32, ffi.c_int32, ffi.c_int32], ffi.c_int32)
_SDL_RenderDrawPoint = _lib.SDL_RenderDrawPoint([ffi.c_void_p, ffi.c_int32, ffi.c_int32], ffi.c_int32)
_SDL_RenderCopy = _lib.SDL_RenderCopy([ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], ffi.c_int32)

_SDL_CreateTexture = _lib.SDL_CreateTexture([ffi.c_void_p, ffi.c_uint32, ffi.c_int32, ffi.c_int32, ffi.c_int32], ffi.c_void_p)
_SDL_DestroyTexture = _lib.SDL_DestroyTexture([ffi.c_void_p], None)
_SDL_UpdateTexture = _lib.SDL_UpdateTexture([ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_int32], ffi.c_int32)

_SDL_PollEvent = _lib.SDL_PollEvent([ffi.c_void_p], ffi.c_int32)
_SDL_WaitEvent = _lib.SDL_WaitEvent([ffi.c_void_p], ffi.c_int32)

_SDL_Delay = _lib.SDL_Delay([ffi.c_uint32], None)
_SDL_GetTicks = _lib.SDL_GetTicks([], ffi.c_uint32)

# --- Classes ---
class Error(Exception):
    pass

def init(flags=SDL_INIT_EVERYTHING):
    if _SDL_Init(flags) != 0:
        raise Error(_SDL_GetError())

def quit():
    _SDL_Quit()

def get_error():
    return _SDL_GetError()

class Rect:
    def __init__(self, x, y, w, h):
        self.ptr = ffi.malloc(16)
        ffi.write_memory_with_offset(self.ptr, 0, ffi.c_int32, x)
        ffi.write_memory_with_offset(self.ptr, 4, ffi.c_int32, y)
        ffi.write_memory_with_offset(self.ptr, 8, ffi.c_int32, w)
        ffi.write_memory_with_offset(self.ptr, 12, ffi.c_int32, h)

    @property
    def x(self): return ffi.read_memory_with_offset(self.ptr, 0, ffi.c_int32)
    @x.setter
    def x(self, val): ffi.write_memory_with_offset(self.ptr, 0, ffi.c_int32, val)

    @property
    def y(self): return ffi.read_memory_with_offset(self.ptr, 4, ffi.c_int32)
    @y.setter
    def y(self, val): ffi.write_memory_with_offset(self.ptr, 4, ffi.c_int32, val)

    @property
    def w(self): return ffi.read_memory_with_offset(self.ptr, 8, ffi.c_int32)
    @w.setter
    def w(self, val): ffi.write_memory_with_offset(self.ptr, 8, ffi.c_int32, val)

    @property
    def h(self): return ffi.read_memory_with_offset(self.ptr, 12, ffi.c_int32)
    @h.setter
    def h(self, val): ffi.write_memory_with_offset(self.ptr, 12, ffi.c_int32, val)

class Window:
    def __init__(self, title, x=SDL_WINDOWPOS_CENTERED, y=SDL_WINDOWPOS_CENTERED, w=800, h=600, flags=SDL_WINDOW_SHOWN):
        self.ptr = _SDL_CreateWindow(title, x, y, w, h, flags)
        if not self.ptr:
            raise Error(_SDL_GetError())

    def set_title(self, title):
        _SDL_SetWindowTitle(self.ptr, title)

    def destroy(self):
        if hasattr(self, 'ptr') and self.ptr:
            _SDL_DestroyWindow(self.ptr)
            self.ptr = None

class Renderer:
    def __init__(self, window, index=-1, flags=SDL_RENDERER_ACCELERATED):
        self.ptr = _SDL_CreateRenderer(window.ptr, index, flags)
        if not self.ptr:
            raise Error(_SDL_GetError())

    def destroy(self):
        if hasattr(self, 'ptr') and self.ptr:
            _SDL_DestroyRenderer(self.ptr)
            self.ptr = None

    def clear(self):
        _SDL_RenderClear(self.ptr)

    def present(self):
        _SDL_RenderPresent(self.ptr)

    def set_draw_color(self, r, g, b, a=255):
        _SDL_SetRenderDrawColor(self.ptr, r, g, b, a)

    def draw_rect(self, rect):
        _SDL_RenderDrawRect(self.ptr, rect.ptr)

    def fill_rect(self, rect):
        _SDL_RenderFillRect(self.ptr, rect.ptr)

    def draw_line(self, x1, y1, x2, y2):
        _SDL_RenderDrawLine(self.ptr, x1, y1, x2, y2)

    def draw_point(self, x, y):
        _SDL_RenderDrawPoint(self.ptr, x, y)

    def copy(self, texture, srcrect=None, dstrect=None):
        src_ptr = srcrect.ptr if srcrect else None
        dst_ptr = dstrect.ptr if dstrect else None
        return _SDL_RenderCopy(self.ptr, texture.ptr, src_ptr, dst_ptr)

class Texture:
    def __init__(self, renderer, format, access, w, h):
        self.ptr = _SDL_CreateTexture(renderer.ptr, format, access, w, h)
        if not self.ptr:
            raise Error(_SDL_GetError())

    def destroy(self):
        if hasattr(self, 'ptr') and self.ptr:
            _SDL_DestroyTexture(self.ptr)
            self.ptr = None

    def update(self, rect, pixels, pitch):
        r_ptr = rect.ptr if rect else None
        return _SDL_UpdateTexture(self.ptr, r_ptr, pixels, pitch)

class Event:
    def __init__(self):
        self.ptr = ffi.malloc(64)

    def poll(self):
        return _SDL_PollEvent(self.ptr) != 0

    def wait(self):
        return _SDL_WaitEvent(self.ptr) != 0

    @property
    def type(self):
        return ffi.read_memory_with_offset(self.ptr, 0, ffi.c_uint32)

    @property
    def window(self):
        return {
            'type': ffi.read_memory_with_offset(self.ptr, 0, ffi.c_uint32),
            'timestamp': ffi.read_memory_with_offset(self.ptr, 4, ffi.c_uint32),
            'windowID': ffi.read_memory_with_offset(self.ptr, 8, ffi.c_uint32),
            'event': ffi.read_memory_with_offset(self.ptr, 12, ffi.c_uint8),
            'data1': ffi.read_memory_with_offset(self.ptr, 16, ffi.c_int32),
            'data2': ffi.read_memory_with_offset(self.ptr, 20, ffi.c_int32),
        }

    @property
    def key(self):
        return {
            'type': ffi.read_memory_with_offset(self.ptr, 0, ffi.c_uint32),
            'timestamp': ffi.read_memory_with_offset(self.ptr, 4, ffi.c_uint32),
            'windowID': ffi.read_memory_with_offset(self.ptr, 8, ffi.c_uint32),
            'state': ffi.read_memory_with_offset(self.ptr, 12, ffi.c_uint8),
            'repeat': ffi.read_memory_with_offset(self.ptr, 13, ffi.c_uint8),
            'scancode': ffi.read_memory_with_offset(self.ptr, 16, ffi.c_int32),
            'sym': ffi.read_memory_with_offset(self.ptr, 20, ffi.c_int32),
            'mod': ffi.read_memory_with_offset(self.ptr, 24, ffi.c_uint16),
        }

    @property
    def motion(self):
        return {
            'type': ffi.read_memory_with_offset(self.ptr, 0, ffi.c_uint32),
            'timestamp': ffi.read_memory_with_offset(self.ptr, 4, ffi.c_uint32),
            'windowID': ffi.read_memory_with_offset(self.ptr, 8, ffi.c_uint32),
            'which': ffi.read_memory_with_offset(self.ptr, 12, ffi.c_uint32),
            'state': ffi.read_memory_with_offset(self.ptr, 16, ffi.c_uint32),
            'x': ffi.read_memory_with_offset(self.ptr, 20, ffi.c_int32),
            'y': ffi.read_memory_with_offset(self.ptr, 24, ffi.c_int32),
            'xrel': ffi.read_memory_with_offset(self.ptr, 28, ffi.c_int32),
            'yrel': ffi.read_memory_with_offset(self.ptr, 32, ffi.c_int32),
        }

    @property
    def button(self):
        return {
            'type': ffi.read_memory_with_offset(self.ptr, 0, ffi.c_uint32),
            'timestamp': ffi.read_memory_with_offset(self.ptr, 4, ffi.c_uint32),
            'windowID': ffi.read_memory_with_offset(self.ptr, 8, ffi.c_uint32),
            'which': ffi.read_memory_with_offset(self.ptr, 12, ffi.c_uint32),
            'button': ffi.read_memory_with_offset(self.ptr, 16, ffi.c_uint8),
            'state': ffi.read_memory_with_offset(self.ptr, 17, ffi.c_uint8),
            'clicks': ffi.read_memory_with_offset(self.ptr, 18, ffi.c_uint8),
            'x': ffi.read_memory_with_offset(self.ptr, 20, ffi.c_int32),
            'y': ffi.read_memory_with_offset(self.ptr, 24, ffi.c_int32),
        }

def get_ticks():
    return _SDL_GetTicks()

def delay(ms):
    _SDL_Delay(ms)
