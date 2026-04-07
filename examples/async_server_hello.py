# examples/async_server_hello.py

import httpserver
import aio # Assuming your async utilities are here (for aio.sleep if used)
# import time # Pylearn's standard time for comparison or if aio.sleep not used

# This is the global environment for the script.
# The httpserver will use a context derived from this for handlers.
MY_GLOBAL_VAR = "This is global"

async def handle_hello_async(request):
    print(f("Async Handler: Hello endpoint called. URL: {request.url}"))
    print(f("Async Handler: Accessing global: {MY_GLOBAL_VAR}"))

    # Simulate some non-blocking async work
    if hasattr(aio, "sleep"): # Check if our cooperative sleep exists
        print("Async Handler: Sleeping for 0.5 seconds cooperatively...")
        await aio.sleep(0.5)
        print("Async Handler: Resumed after sleep.")
    else:
        print("Async Handler: aio.sleep not available, proceeding synchronously for this part.")
        # time.sleep(0.5) # This would be a blocking sleep

    return "Hello, async Pylearn world!" # Pylearn String

async def handle_root_async(request):
    print(f("Async Handler: Root endpoint called. URL: {request.url}"))
    # You could do other async operations here
    # external_data_promise = http.get("some_api_url") # If http.get is async
    # external_data = await external_data_promise
    return "Welcome to the Async Pylearn Root!"

def handle_sync_ping(request):
    print("Sync Handler: Ping called.")
    return "pong_sync"

routes = {
    "/hello": handle_hello_async, # Registering an async handler
    "/": handle_root_async,       # Another async handler
    "/ping": handle_sync_ping,    # A synchronous handler
}

def main():
    address = "127.0.0.1:8080" # Use a different port if 5173 is busy
    print(f("Pylearn: Attempting to start ASYNC-AWARE server on {address}"))

    # The `httpserver.serve` function's Go implementation now needs
    # to correctly handle both async and sync Pylearn handlers.
    httpserver.serve(address, routes)
    
    print(f("Pylearn: Server setup initiated. Check browser/curl at http://{address}"))
    print("Pylearn: Main script will wait here. Server runs in background (Go goroutine).")
    # The Go server runs indefinitely until the Go program is terminated (e.g., Ctrl+C).
    # This input() is just for the Pylearn script itself to pause.
    try:
        input("Pylearn: Press Enter in this Pylearn script to 'finish' (Go server continues)...")
    except Exception as e: # Catch EOF if running non-interactively
        print("Pylearn: Script continuing non-interactively (Go server continues).")
        # In a real non-interactive script, you might have an infinite loop or another way to keep main alive
        # while the Go server runs, or just let main exit if server is truly detached.
        # For now, let server run until Ctrl+C.
        import time
        while True:
            time.sleep(3600) # Keep Pylearn main alive for a long time


# if __name__ == "__main__": # Pylearn needs this support
main()