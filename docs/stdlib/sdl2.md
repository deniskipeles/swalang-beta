# `sdl2` Module Reference

The `sdl2` module provides high-level bindings to the Simple DirectMedia Layer (SDL2) library.

## Constants

### Initialization Flags
- `INIT_TIMER`
- `INIT_AUDIO`
- `INIT_VIDEO`
- `INIT_JOYSTICK`
- `INIT_HAPTIC`
- `INIT_GAMECONTROLLER`
- `INIT_EVENTS`
- `INIT_EVERYTHING`

### Window Position
- `WINDOWPOS_UNDEFINED`
- `WINDOWPOS_CENTERED`

### Window Flags
- `WINDOW_FULLSCREEN`
- `WINDOW_OPENGL`
- `WINDOW_SHOWN`
- `WINDOW_HIDDEN`
- `WINDOW_BORDERLESS`
- `WINDOW_RESIZABLE`
- `WINDOW_MINIMIZED`
- `WINDOW_MAXIMIZED`
- `WINDOW_INPUT_FOCUS`
- `WINDOW_MOUSE_FOCUS`
- `WINDOW_HIGHPIXEL`

### Renderer Flags
- `RENDERER_SOFTWARE`
- `RENDERER_ACCELERATED`
- `RENDERER_PRESENTVSYNC`
- `RENDERER_TARGETTEXTURE`

### Event Types
- `QUIT`
- `WINDOWEVENT`
- `SYSWMEVENT`
- `KEYDOWN`
- `KEYUP`
- `TEXTEDITING`
- `TEXTINPUT`
- `MOUSEMOTION`
- `MOUSEBUTTONDOWN`
- `MOUSEBUTTONUP`
- `MOUSEWHEEL`

## Classes

### `SDL` (Static Methods)
- `init(flags=INIT_VIDEO | INIT_EVENTS)`: Initializes SDL.
- `quit()`: Cleans up SDL.
- `get_error()`: Returns the last SDL error message.
- `delay(ms)`: Pauses execution for `ms` milliseconds.
- `get_ticks()`: Returns the number of milliseconds since SDL initialization.
- `get_mouse_state()`: Returns `(x, y, button_mask)`.

### `Window(title, w, h, flags=WINDOW_SHOWN)`
Creates a new window.
- `ptr`: The underlying C pointer to the window.
- `size`: (Property) Returns `(w, h)` of the window.
- `flags`: (Property) Returns the window flags.
- `destroy()`: Destroys the window.

### `Renderer(window, index=-1, flags=RENDERER_ACCELERATED | RENDERER_PRESENTVSYNC)`
Creates a 2D rendering context for a window.
- `ptr`: The underlying C pointer to the renderer.
- `set_draw_color(r, g, b, a=255)`: Sets the drawing color.
- `clear()`: Clears the current rendering target with the drawing color.
- `present()`: Updates the screen with any rendering performed since the previous call.
- `fill_rect(x, y, w, h)`: Fills a rectangle on the current rendering target.
- `copy(texture, srcrect_ptr=None, dstrect_ptr=None)`: Copies a portion of the texture to the current rendering target.
- `destroy()`: Destroys the renderer.

### `Texture(renderer, format, access, w, h)`
Creates a texture for a rendering context.
- `ptr`: The underlying C pointer to the texture.
- `update(rect_ptr, pixels_ptr, pitch)`: Updates the texture with new pixel data.
- `destroy()`: Destroys the texture.

### `Event()`
An object used to poll and hold data about SDL events.
- `poll()`: Returns `True` if there is a pending event and populates the object.
- `type`: (Property) The type of the event.
- **Mouse Motion:** `motion_x`, `motion_y`, `motion_xrel`, `motion_yrel`.
- **Mouse Button:** `button`, `button_state`, `button_x`, `button_y`.
- **Keyboard:** `key_scancode`, `key_sym`, `key_mod`.
- **Window:** `window_event`, `window_data1`, `window_data2`.
- `free()`: Frees internal memory used by the event object.

## Structs

- `Rect`: Fields `x`, `y`, `w`, `h` (all `c_int32`).
- `Point`: Fields `x`, `y` (both `c_int32`).
- `Color`: Fields `r`, `g`, `b`, `a` (all `c_uint8`).
