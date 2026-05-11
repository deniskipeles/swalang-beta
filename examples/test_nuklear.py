import sdl2
import nuklear

print("🧪 Initializing SDL2 & Nuklear...")
# Initialize video and events
sdl2.init(sdl2.SDL_INIT_VIDEO | sdl2.SDL_INIT_EVENTS)

win = sdl2.Window("Nuklear UI via Swalang", 800, 600)
ren = sdl2.Renderer(win)
ev = sdl2.Event()

# Initialize Nuklear Context and Input helper
ctx = nuklear.Context()
ctx.init() # Must call init to set up internal memory
nk_input = nuklear.NuklearSDLInput(ctx)

print("🚀 UI Loop started. Close the window to exit.")

clicks = 0
running = True

while running:
    # --- Input Gathering ---
    nk_input.begin()
    while ev.poll():
        if ev.type == sdl2.SDL_QUIT:
            running = False
        # Feed the SDL event to the Nuklear input handler
        nk_input.feed(ev)
    nk_input.end()
    
    # --- UI Building ---
    # Use NK_ prefixed constants as defined in the library
    window_flags = nuklear.NK_WINDOW_BORDER | nuklear.NK_WINDOW_MOVABLE | nuklear.NK_WINDOW_TITLE
    
    if ctx.begin("Demo Window", 50, 50, 250, 200, window_flags):
        
        ctx.layout_row_dynamic(30, 1)
        ctx.label(format_str("Clicks: {clicks}"), nuklear.NK_TEXT_CENTERED)
        
        ctx.layout_row_dynamic(30, 1)
        if ctx.button_label("Click Me!"):
            clicks = clicks + 1
            print(format_str("👉 Button clicked! Total: {clicks}"))
            
    ctx.end()
    
    # --- Rendering ---
    ren.set_draw_color(30, 30, 30, 255)
    ren.clear()
    
    # Note: True rendering requires a backend renderer. 
    # For now, we clear the context to prepare for the next frame.
    ctx.clear()
    
    ren.present()
    sdl2.delay(16)

print("Sweep 🧹 Cleaning up...")
ctx.free()
ev.free()
ren.destroy()
win.destroy()
sdl2.quit()
print("✅ SDL2 exited gracefully.")