# pylearn/examples/test_uv.py

import uv
import time

# Counter for our one-shot timer
one_shot_fired = False

def on_repeat_timer():
    """This function is called every second."""
    print(format_str("Repeating timer fired at: {time.time()}"))

def on_one_shot_timer():
    """This function is called once after 3 seconds, then stops the loop."""
    global one_shot_fired
    print("One-shot timer fired! Stopping the event loop.")
    one_shot_fired = True
    # We need a way to stop the loop. A clean way is to close the handles.
    # When all handles are closed, uv_run will exit.
    repeating_timer.close()
    one_shot_timer.close()

# --- Main application logic ---
if not uv.UV_AVAILABLE:
    print("libuv is not available, cannot run the example.")
else:
    print("libuv event loop example started.")
    
    # 1. Create a new event loop
    main_loop = uv.Loop()

    # 2. Create a repeating timer
    repeating_timer = uv.Timer(main_loop)
    # Start it to fire every 1000ms (1 second)
    repeating_timer.start(on_repeat_timer, 1000, 1000)
    print("Started a timer that repeats every second.")

    # 3. Create a one-shot timer to stop the loop after 3 seconds
    one_shot_timer = uv.Timer(main_loop)
    # Start it to fire once after 3000ms
    one_shot_timer.start(on_one_shot_timer, 3000, 0)
    print("Started a one-shot timer that will fire in 3 seconds and stop the loop.")

    # 4. Run the event loop. This will block until all timers are closed.
    main_loop.run()
    
    # 5. Clean up the loop itself
    main_loop.close()

    print("\nEvent loop finished.")

    # Simple assertion to verify the test ran correctly
    assert one_shot_fired, "The one-shot timer did not fire!"
    print("Test successful.")

    