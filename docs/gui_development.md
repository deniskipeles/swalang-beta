# GUI Development with Swalang

Swalang provides powerful capabilities for building graphical user applications by leveraging industry-standard C libraries like **SDL2** (Simple DirectMedia Layer) and **LVGL** (Light and Versatile Graphics Library).

## 1. Low-Level Graphics with SDL2

The `sdl2` module is perfect for games, custom rendering engines, or applications where you need direct control over windows, textures, and input.

### Basic SDL2 Boilerplate

```python
import sdl2
import time

# Initialize SDL2
sdl2.SDL.init()

# Create a window and a renderer
win = sdl2.Window("My Swalang App", 800, 600)
ren = sdl2.Renderer(win)
ev = sdl2.Event()

running = True
while running:
    # 1. Handle Events
    while ev.poll():
        if ev.type == sdl2.QUIT:
            running = False
            break

    # 2. Update Logic
    # (Update your game state here)

    # 3. Rendering
    ren.set_draw_color(45, 20, 85) # Purple background
    ren.clear()

    # Draw a rectangle
    ren.set_draw_color(255, 255, 255) # White
    # Note: Renderer methods like fill_rect take Pointer to SDL_Rect
    # For simple clear/present, no complex pointers are needed.

    ren.present()

    # Control frame rate
    sdl2.SDL.delay(16) # ~60 FPS

# Cleanup
ev.free()
ren.destroy()
win.destroy()
sdl2.SDL.quit()
```

## 2. Advanced UI with LVGL

For complex user interfaces with buttons, sliders, and charts, Swalang includes high-level bindings for **LVGL**. LVGL usually runs on top of a display driver, and in Swalang, we typically use SDL2 as that driver.

### LVGL Concepts

- **Screens**: The root containers for UI elements.
- **Widgets**: Buttons, Labels, Sliders, etc.
- **Events**: Callbacks triggered by user interaction.
- **Timer Handler**: The heart of LVGL that needs to be called regularly.

### LVGL & SDL2 Integration Example

This example demonstrates a complete setup with a button that updates a label.

```python
import sdl2
import lvgl
import ffi

# 1. Initialization
sdl2.SDL.init()
lvgl.init()

W, H = 800, 600
win = sdl2.Window("LVGL + Swalang", W, H)
ren = sdl2.Renderer(win)

# Create a texture for LVGL to draw into
texture = sdl2.Texture(ren, sdl2.PIXELFORMAT_ARGB8888, sdl2.TEXTUREACCESS_STREAMING, W, H)

# 2. LVGL Display Setup
disp = lvgl.Display(W, H)
buf_size = W * H * 4
draw_buf = ffi.malloc(buf_size)
disp.set_buffers(draw_buf, None, buf_size, lvgl.LV_DISPLAY_RENDER_MODE_PARTIAL)

# Flush callback: Transfers LVGL pixels to SDL texture
def flush_callback(disp_ptr, area_ptr, px_map_ptr):
    # Logic to extract area and update texture
    # (See examples/test_lvgl.py for full implementation details)
    disp.flush_ready()

disp.set_flush_cb(flush_callback)

# 3. Create UI Elements
scr = lvgl.Screen.active()

btn = lvgl.Button(scr)
btn.set_size(200, 60)
btn.align(lvgl.LV_ALIGN_CENTER, 0, 0)

lbl = lvgl.Label(btn.ptr)
lbl.set_text("Click Me!")
lbl.align(lvgl.LV_ALIGN_CENTER, 0, 0)

# Event handling
def on_click(event_ptr):
    print("Button was clicked!")
    lbl.set_text("Clicked!")

btn.add_event_cb(on_click, lvgl.LV_EVENT_CLICKED)

# 4. Main Loop
ev = sdl2.Event()
running = True
while running:
    while ev.poll():
        if ev.type == sdl2.QUIT:
            running = False

    lvgl.tick_inc(10)      # Tell LVGL 10ms passed
    lvgl.timer_handler()   # Let LVGL do its work

    ren.clear()
    ren.copy(texture)
    ren.present()
    sdl2.SDL.delay(10)

# Cleanup (Omitting for brevity)
```

## Best Practices

1.  **Resource Management**: Always call `.destroy()` or `.free()` on SDL2 objects and `ffi.malloc` pointers when they are no longer needed to prevent memory leaks.
2.  **Frame Rate**: Use `sdl2.SDL.delay()` or vertical sync to keep CPU usage low.
3.  **Concurrency**: Avoid making GUI calls from background threads. Use a thread-safe message queue or the `asyncio` loop to schedule UI updates on the main thread.
