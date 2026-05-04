import sdl2
import lvgl
import ffi

# -------------------------------------------------------------------------
# Step 1: Initialize SDL2 & LVGL
# -------------------------------------------------------------------------
print("🧪 Initializing SDL2 & LVGL...")
sdl2.SDL.init()
lvgl.init()

W, H = 800, 600
win = sdl2.Window("LVGL UI via Swalang", W, H)
ren = sdl2.Renderer(win)

# Create an SDL Texture that we will update with LVGL's pixel data
texture = sdl2.Texture(ren, sdl2.PIXELFORMAT_ARGB8888, sdl2.TEXTUREACCESS_STREAMING, W, H)

# -------------------------------------------------------------------------
# Step 2: Setup LVGL Display & Flush Callback
# -------------------------------------------------------------------------
disp = lvgl.Display(W, H)

# Allocate a draw buffer for LVGL. (W * H * 4 bytes per pixel)
buf_size = W * H * 4
draw_buf = ffi.malloc(buf_size)

# Tell LVGL to use this buffer
disp.set_buffers(draw_buf, None, buf_size, lvgl.LV_DISPLAY_RENDER_MODE_PARTIAL)

def flush_callback(disp_ptr, area_ptr, px_map_ptr):
    """Called by LVGL when it wants to push rendered pixels to the screen."""
    # Read the updated area
    x1 = ffi.read_memory_with_offset(area_ptr, 0, ffi.c_int32)
    y1 = ffi.read_memory_with_offset(area_ptr, 4, ffi.c_int32)
    x2 = ffi.read_memory_with_offset(area_ptr, 8, ffi.c_int32)
    y2 = ffi.read_memory_with_offset(area_ptr, 12, ffi.c_int32)
    
    w = (x2 - x1) + 1
    h = (y2 - y1) + 1

    # Create an SDL_Rect for the update area
    rect_ptr = ffi.malloc(16)
    ffi.write_memory_with_offset(rect_ptr, 0, ffi.c_int32, x1)
    ffi.write_memory_with_offset(rect_ptr, 4, ffi.c_int32, y1)
    ffi.write_memory_with_offset(rect_ptr, 8, ffi.c_int32, w)
    ffi.write_memory_with_offset(rect_ptr, 12, ffi.c_int32, h)

    # Update the SDL Texture with the raw pixels from LVGL
    # Pitch is width * 4 bytes
    texture.update(rect_ptr, px_map_ptr, w * 4)
    
    ffi.free(rect_ptr)

    # Tell LVGL we are done flushing
    disp.flush_ready()

disp.set_flush_cb(flush_callback)

# -------------------------------------------------------------------------
# Step 3: Setup LVGL Input Device (Mouse)
# -------------------------------------------------------------------------
mouse_x = 0
mouse_y = 0
mouse_pressed = False

indev = lvgl.InputDevice()

def read_callback(indev_ptr, data_ptr):
    """Called by LVGL to get the current mouse state."""
    state = lvgl.LV_INDEV_STATE_PRESSED if mouse_pressed else lvgl.LV_INDEV_STATE_RELEASED
    lvgl.InputDevice.write_pointer_data(data_ptr, mouse_x, mouse_y, state)

indev.set_read_cb(read_callback)

# -------------------------------------------------------------------------
# Step 4: Build the LVGL UI
# -------------------------------------------------------------------------
scr = lvgl.Screen.active()

btn = lvgl.Button(scr)
btn.set_size(200, 60)
btn.align(lvgl.LV_ALIGN_CENTER, 0, 0)

lbl = lvgl.Label(btn.ptr)
lbl.set_text("Click Me! (Swalang)")
lbl.align(lvgl.LV_ALIGN_CENTER, 0, 0)

clicks = 0
def btn_clicked(event_ptr):
    global clicks
    clicks = clicks + 1
    print(format_str("👉 Button clicked! Total: {clicks}"))
    lbl.set_text(format_str("Clicked {clicks} times"))

btn.add_event_cb(btn_clicked, lvgl.LV_EVENT_CLICKED)

# -------------------------------------------------------------------------
# Step 5: Main Loop
# -------------------------------------------------------------------------
print("🚀 UI Loop started. Close the window to exit.")

ev = sdl2.Event()
running = True

while running:
    # 1. Gather SDL Events
    while ev.poll():
        if ev.type == sdl2.QUIT:
            running = False
        elif ev.type == sdl2.MOUSEMOTION:
            mouse_x = ev.motion_x
            mouse_y = ev.motion_y
        elif ev.type == sdl2.MOUSEBUTTONDOWN:
            mouse_pressed = True
        elif ev.type == sdl2.MOUSEBUTTONUP:
            mouse_pressed = False

    # 2. Tell LVGL time has passed (e.g., 10ms)
    lvgl.tick_inc(10)

    # 3. Run LVGL tasks (this triggers flush_cb if drawing is needed)
    lvgl.timer_handler()

    # 4. Render the updated Texture to the screen
    ren.clear()
    ren.copy(texture)
    ren.present()

    sdl2.SDL.delay(10)

print("🧹 Cleaning up...")
ev.free()
ffi.free(draw_buf)
texture.destroy()
ren.destroy()
win.destroy()
sdl2.SDL.quit()
print("✅ LVGL exited gracefully.")