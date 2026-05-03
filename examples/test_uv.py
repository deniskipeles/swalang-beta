import uv
import time

print("===========================================")
print("⏱️ Testing libuv (Asyncio Engine) Wrapper")
print("===========================================")

# 1. Create the Event Loop
loop = uv.Loop()

counter = 0

def on_interval():
    global counter
    counter = counter + 1
    print(format_str("[{time.time()}] 🔄 Interval ticked! ({counter})"))

def on_timeout():
    print(format_str("\n[{time.time()}] 🛑 Timeout reached! Stopping interval timer..."))
    interval_timer.stop()

# 2. Create Timers attached to the loop
interval_timer = uv.Timer(loop)
timeout_timer = uv.Timer(loop)

# 3. Start timers (Non-blocking!)
print("Starting a 500ms repeating interval...")
interval_timer.start(on_interval, 500, 500)

print("Starting a 2.5 second timeout...")
timeout_timer.start(on_timeout, 2500, 0)

print("\n🚀 Entering libuv event loop. Pylearn is now async!")
# 4. Run the loop. It blocks here until all active handles (timers) are stopped.
loop.run()

print("✅ Event loop exited gracefully.")

# 5. Clean up C memory
interval_timer.close()
timeout_timer.close()
loop.close()
