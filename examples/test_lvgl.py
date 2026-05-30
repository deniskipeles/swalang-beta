import sdl2
import lvgl
import ffi

mouse  = {"x": 0, "y": 0, "pressed": False}
clicks = 0

print("Initializing SDL2 & LVGL...")
sdl2.init(sdl2.SDL_INIT_VIDEO | sdl2.SDL_INIT_EVENTS)
lvgl.init()

W, H = 800, 600
win     = sdl2.Window("LVGL UI via Swalang", W, H)
ren     = sdl2.Renderer(win)
texture = sdl2.Texture(ren, sdl2.SDL_PIXELFORMAT_RGB565, sdl2.SDL_TEXTUREACCESS_STREAMING, W, H)

disp     = lvgl.Display(W, H)
buf_size = W * H * 2
draw_buf = ffi.malloc(buf_size)
disp.set_buffers(draw_buf, None, buf_size, lvgl.LV_DISPLAY_RENDER_MODE_PARTIAL)

# ── Flush: copy pixels into our own buffer, THEN call flush_ready ─────────────
pixel_buf     = ffi.malloc(buf_size)
pending_flush = {"active": False, "x1": 0, "y1": 0, "w": 0, "h": 0}

def _flush_fn(disp_ptr, area_ptr, px_map_ptr):
    x1 = ffi.read_memory_with_offset(area_ptr,  0, ffi.c_int32)
    y1 = ffi.read_memory_with_offset(area_ptr,  4, ffi.c_int32)
    x2 = ffi.read_memory_with_offset(area_ptr,  8, ffi.c_int32)
    y2 = ffi.read_memory_with_offset(area_ptr, 12, ffi.c_int32)
    w  = (x2 - x1) + 1
    h  = (y2 - y1) + 1
    
    ffi.memcpy(pixel_buf, px_map_ptr, w * h * 2)
    pending_flush["active"] = True
    pending_flush["x1"]     = x1
    pending_flush["y1"]     = y1
    pending_flush["w"]      = w
    pending_flush["h"]      = h
    disp.flush_ready()

disp.set_flush_cb(_flush_fn)

# ── Input device ──────────────────────────────────────────────────────────────
indev = lvgl.InputDevice(lvgl.LV_INDEV_TYPE_POINTER)
indev.set_display(disp)

def _read_fn(indev_ptr_arg, data_ptr):
    state = lvgl.LV_INDEV_STATE_PRESSED if mouse["pressed"] else lvgl.LV_INDEV_STATE_RELEASED
    lvgl.InputDevice.write_pointer(data_ptr, mouse["x"], mouse["y"], state)

indev.set_read_cb(_read_fn)

# ── UI ────────────────────────────────────────────────────────────────────────
scr = lvgl.screen_active()
scr.set_size(W, H)

btn = lvgl.Button(scr.ptr)
btn.set_size(200, 60)
btn.align(lvgl.LV_ALIGN_CENTER, 0, 0)

lbl = lvgl.Label(btn.ptr)
lbl.set_text("Click Me! (Swalang)")
lbl.align(lvgl.LV_ALIGN_CENTER, 0, 0)

# Force initial layout calculations
lvgl.tick_inc(100)
lvgl.timer_handler()

print(f("btn hit area: x={btn.get_x()} to {btn.get_x()+btn.get_width()}  y={btn.get_y()} to {btn.get_y()+btn.get_height()}"))

# ── Event callbacks ───────────────────────────────────────────────────────────
def btn_clicked(event_ptr):
    global clicks
    clicks = clicks + 1
    print(f("*** BUTTON CLICKED! Total: {clicks} ***"))
    lbl.set_text(f("Clicked {clicks} times"))

def btn_pressed(event_ptr):
    print("  [evt] PRESSED")

def btn_released(event_ptr):
    print("  [evt] RELEASED")

btn.on(lvgl.LV_EVENT_CLICKED, btn_clicked)
btn.on(lvgl.LV_EVENT_PRESSED, btn_pressed)
btn.on(lvgl.LV_EVENT_RELEASED, btn_released)

# ── Main loop ─────────────────────────────────────────────────────────────────
print("UI Loop started. Close the window to exit.")

ev      = sdl2.Event()
running = True
last_tick = sdl2.get_ticks()

while running:
    while ev.poll():
        t = ev.type

        if t == sdl2.SDL_QUIT:
            running = False

        elif t == sdl2.SDL_MOUSEMOTION:
            mouse["x"] = ev.x
            mouse["y"] = ev.y

        elif t == sdl2.SDL_MOUSEBUTTONDOWN:
            mouse["x"]       = ev.x
            mouse["y"]       = ev.y
            mouse["pressed"] = True
            print(f("  [SDL] DOWN x={ev.x} y={ev.y}"))
            
            # FIX: Force LVGL to poll the input immediately by jumping 33ms into the future!
            lvgl.tick_inc(33)
            lvgl.timer_handler()

        elif t == sdl2.SDL_MOUSEBUTTONUP:
            mouse["x"]       = ev.x
            mouse["y"]       = ev.y
            mouse["pressed"] = False
            print(f("  [SDL] UP   x={ev.x} y={ev.y}"))
            
            # FIX: Force LVGL to poll the input immediately by jumping 33ms into the future!
            lvgl.tick_inc(33)
            lvgl.timer_handler()

    now       = sdl2.get_ticks()
    elapsed   = now - last_tick
    last_tick = now
    
    if elapsed > 0:
        lvgl.tick_inc(elapsed)
    lvgl.timer_handler()

    # Upload any pending flush to SDL texture
    if pending_flush["active"]:
        texture.update(
            (pending_flush["x1"], pending_flush["y1"], pending_flush["w"], pending_flush["h"]),
            pixel_buf,
            pending_flush["w"] * 2
        )
        pending_flush["active"] = False
        
        ren.clear()
        ren.copy(texture)
        ren.present()

    sdl2.delay(5)

# ── Cleanup ───────────────────────────────────────────────────────────────────
print("Cleaning up...")
ev.free()
ffi.free(draw_buf)
ffi.free(pixel_buf)
texture.destroy()
ren.destroy()
win.destroy()
sdl2.quit()
print("Done.")