import ffi
import sys

def _load_library():
    platform = sys.platform
    candidates =[]
    if platform == 'linux':
        candidates =["bin/x86_64-linux/sdl2/libSDL2-2.0.so.0", "bin/x86_64-linux/sdl2/libSDL2.so", "libSDL2.so"]
    elif platform == 'windows':
        candidates =["bin/x86_64-windows-gnu/sdl2/SDL2.dll", "SDL2.dll"]
    elif platform == 'darwin':
        candidates = ["libSDL2.dylib"]
    
    for name in candidates:
        try:
            return ffi.CDLL(name)
        except ffi.FFIError:
            pass
    raise ffi.FFIError("Could not load SDL2 shared library. Ensure it is built and in the bin/ directory.")

_lib = _load_library()

# --- Constants ---
INIT_TIMER          = 0x00000001
INIT_AUDIO          = 0x00000010
INIT_VIDEO          = 0x00000020
INIT_JOYSTICK       = 0x00000200
INIT_HAPTIC         = 0x00001000
INIT_GAMECONTROLLER = 0x00002000
INIT_EVENTS         = 0x00004000
INIT_EVERYTHING     = 0x0000F231

WINDOWPOS_CENTERED  = 0x2FFF0000
WINDOW_SHOWN        = 0x00000004
WINDOW_OPENGL       = 0x00000002
WINDOW_RESIZABLE    = 0x00000020

RENDERER_SOFTWARE      = 0x00000001
RENDERER_ACCELERATED   = 0x00000002
RENDERER_PRESENTVSYNC  = 0x00000004
PIXELFORMAT_ARGB8888   = 0x16362004
TEXTUREACCESS_STREAMING = 1

QUIT                 = 0x100
MOUSEMOTION          = 0x400
MOUSEBUTTONDOWN      = 0x401
MOUSEBUTTONUP        = 0x402
KEYDOWN              = 0x300

# --- C Signatures ---
_SDL_Init = _lib.SDL_Init([ffi.c_uint32], ffi.c_int32)
_SDL_Quit = _lib.SDL_Quit([], None)
_SDL_GetError = _lib.SDL_GetError([], ffi.c_void_p)
_SDL_Delay = _lib.SDL_Delay([ffi.c_uint32], None)
_SDL_GetTicks = _lib.SDL_GetTicks([], ffi.c_uint32)

_SDL_CreateWindow = _lib.SDL_CreateWindow([ffi.c_char_p, ffi.c_int32, ffi.c_int32, ffi.c_int32, ffi.c_int32, ffi.c_uint32], ffi.c_void_p)
_SDL_DestroyWindow = _lib.SDL_DestroyWindow([ffi.c_void_p], None)

_SDL_CreateRenderer = _lib.SDL_CreateRenderer([ffi.c_void_p, ffi.c_int32, ffi.c_uint32], ffi.c_void_p)
_SDL_DestroyRenderer = _lib.SDL_DestroyRenderer([ffi.c_void_p], None)
_SDL_RenderClear = _lib.SDL_RenderClear([ffi.c_void_p], ffi.c_int32)
_SDL_RenderPresent = _lib.SDL_RenderPresent([ffi.c_void_p], None)
_SDL_SetRenderDrawColor = _lib.SDL_SetRenderDrawColor([ffi.c_void_p, ffi.c_uint8, ffi.c_uint8, ffi.c_uint8, ffi.c_uint8], ffi.c_int32)
_SDL_RenderCopy = _lib.SDL_RenderCopy([ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], ffi.c_int32)

_SDL_CreateTexture = _lib.SDL_CreateTexture([ffi.c_void_p, ffi.c_uint32, ffi.c_int32, ffi.c_int32, ffi.c_int32], ffi.c_void_p)
_SDL_UpdateTexture = _lib.SDL_UpdateTexture([ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_int32], ffi.c_int32)
_SDL_DestroyTexture = _lib.SDL_DestroyTexture([ffi.c_void_p], None)

_SDL_PollEvent = _lib.SDL_PollEvent([ffi.c_void_p], ffi.c_int32)

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

class Window:
    def __init__(self, title, w, h, flags=WINDOW_SHOWN):
        self.ptr = _SDL_CreateWindow(title.encode('utf-8'), WINDOWPOS_CENTERED, WINDOWPOS_CENTERED, w, h, flags)
        if not self.ptr or self.ptr.Address == 0:
            raise SDLError(format_str("SDL_CreateWindow Error: {SDL.get_error()}"))

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
            # --- FIX: Fallback to software renderer for cloud/headless environments ---
            self.ptr = _SDL_CreateRenderer(window.ptr, index, RENDERER_SOFTWARE)
            if not self.ptr or self.ptr.Address == 0:
                raise SDLError(format_str("SDL_CreateRenderer Error: {SDL.get_error()}"))

    def set_draw_color(self, r, g, b, a=255):
        _SDL_SetRenderDrawColor(self.ptr, r, g, b, a)

    def clear(self):
        _SDL_RenderClear(self.ptr)

    def present(self):
        _SDL_RenderPresent(self.ptr)

    def copy(self, texture, srcrect=None, dstrect=None):
        _SDL_RenderCopy(self.ptr, texture.ptr, srcrect, dstrect)

    def destroy(self):
        if self.ptr:
            _SDL_DestroyRenderer(self.ptr)
            self.ptr = None

class Event:
    def __init__(self):
        self._buf = ffi.malloc(64) # Increased for safety across platforms
    def poll(self):
        return _SDL_PollEvent(self._buf) > 0
    @property
    def type(self):
        return ffi.read_memory(self._buf, ffi.c_uint32)
    @property
    def motion_x(self): return ffi.read_memory_with_offset(self._buf, 20, ffi.c_int32)
    @property
    def motion_y(self): return ffi.read_memory_with_offset(self._buf, 24, ffi.c_int32)
    @property
    def button(self): return ffi.read_memory_with_offset(self._buf, 16, ffi.c_uint8)
    def free(self):
        if self._buf:
            ffi.free(self._buf)
            self._buf = None