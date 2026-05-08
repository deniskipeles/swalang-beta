import sdl2

print("🧪 Initializing SDL2...")
# Initialize ONLY video and events to prevent crashes on systems without audio devices
sdl2.init(sdl2.SDL_INIT_VIDEO | sdl2.SDL_INIT_EVENTS)

win = sdl2.Window("Swalang SDL2 Animation", 800, 600)
ren = sdl2.Renderer(win)
ev = sdl2.Event()

print("🚀 Animation loop started. Press ESC or close window to exit.")

running = True
box_x = 100
box_y = 100
box_dx = 5
box_dy = 4

while running:
    # 1. Handle Events
    while ev.poll():
        if ev.type == sdl2.SDL_QUIT:
            running = False
        elif ev.type == sdl2.SDL_KEYDOWN:
            if ev.scancode == sdl2.SDL_SCANCODE_ESCAPE:
                running = False

    # 2. Update Physics
    box_x = box_x + box_dx
    box_y = box_y + box_dy

    # Bounce off the walls (window is 800x600, box is 50x50)
    if box_x <= 0 or box_x >= (800 - 50):
        box_dx = -box_dx
    if box_y <= 0 or box_y >= (600 - 50):
        box_dy = -box_dy

    # 3. Rendering
    # Background (Dark Purple)
    ren.set_draw_color(45, 20, 85, 255)
    ren.clear()
    
    # Bouncing Box (Cyan)
    ren.set_draw_color(0, 255, 255, 255)
    ren.fill_rect(box_x, box_y, 50, 50)
    
    # Present the frame
    ren.present()
    
    # Cap at ~60 FPS
    sdl2.delay(16)

print("🧹 Cleaning up...")
ev.free()
ren.destroy()
win.destroy()
sdl2.quit()
print("✅ SDL2 exited gracefully.")