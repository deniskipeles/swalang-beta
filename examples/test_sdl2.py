import sdl2
import time

print("🧪 Initializing SDL2...")
sdl2.SDL.init()

win = sdl2.Window("Swalang SDL2 Engine", 800, 600)
ren = sdl2.Renderer(win)
ev = sdl2.Event()

print("🚀 Loop started. Close the window to exit.")

running = True
while running:
    # 1. Handle Events
    while ev.poll():
        if ev.type == sdl2.QUIT:
            running = False
            break

    # 2. Rendering
    # Set background to a dark "Swalang" purple
    ren.set_draw_color(45, 20, 85)
    ren.clear()
    
    # Present the frame
    ren.present()
    
    # Small sleep to prevent 100% CPU usage for this test
    # (In a real app, use SDL_Delay or vsync)
    # time.sleep(0.01)

print("🧹 Cleaning up...")
ev.free()
ren.destroy()
win.destroy()
sdl2.SDL.quit()
print("✅ SDL2 exited gracefully.")