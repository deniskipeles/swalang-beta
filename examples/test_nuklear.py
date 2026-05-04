import sdl2
import nuklear
import time

print("🧪 Initializing SDL2 & Nuklear...")
sdl2.SDL.init()
win = sdl2.Window("Nuklear UI via Swalang", 800, 600)
ren = sdl2.Renderer(win)
ev = sdl2.Event()

# Initialize Nuklear Context
ctx = nuklear.Context()

print("🚀 UI Loop started. Interact with the invisible button around (50, 100). Close the window to exit.")

clicks = 0
running = True

while running:
    # --- Input Gathering ---
    ctx.input_begin()
    while ev.poll():
        if ev.type == sdl2.QUIT:
            running = False
        # Feed the raw SDL event to Nuklear!
        ctx.handle_event(ev)
    ctx.input_end()
    
    # --- UI Building ---
    # Draw a window. Pass standard Python primitive types!
    if ctx.begin("Demo Window", 50, 50, 250, 200, nuklear.WINDOW_BORDER | nuklear.WINDOW_MOVABLE | nuklear.WINDOW_TITLE):
        
        ctx.layout_row_dynamic(30, 1)
        ctx.label(format_str("Clicks: {clicks}"), nuklear.TEXT_CENTERED)
        
        ctx.layout_row_dynamic(30, 1)
        if ctx.button_label("Click Me!"):
            clicks = clicks + 1
            print(format_str("👉 Button clicked! Total: {clicks}"))
            
    ctx.end()
    
    # --- Rendering ---
    ren.set_draw_color(30, 30, 30)
    ren.clear()
    
    # Note: True rendering requires iterating the nk_command buffer or using a C backend.
    # The commands are safely generated in memory but not drawn.
    # This script proves the C-Struct passing and Callback integration works!

    ren.present()
    
    # Clean up Nuklear command queues for the next frame
    ctx.clear()
    
    sdl2.SDL.delay(16) # ~60 FPS

print("🧹 Cleaning up...")
ctx.free()
ev.free()
ren.destroy()
win.destroy()
sdl2.SDL.quit()
print("✅ Nuklear exited gracefully.")