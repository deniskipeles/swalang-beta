# pylearn/examples/cors_api_server.py

from flasky import App
from flasky.cors import CORS
from httpserver import serve, json_response
import asyncio

# 1. Create the Flasky application instance
app = App()

# 2. Define some API routes
@app.route("/api/data")
def get_data(request):
    """A simple API endpoint."""
    user_data = {
        "id": 123,
        "name": "Pylearn User",
        "roles": ["admin", "developer"]
    }
    return user_data

@app.route("/api/stream", methods=["GET"])
def download_generator(request):
    """
    A generator function that streams a large response to the client.
    Each `yield` sends a chunk of data immediately.
    """
    print("Starting a large download stream...")
    
    yield b"----- BEGIN STREAM -----\n"
    
    for i in range(10):
        # In a real app, this could be reading from a large file,
        # querying a database, or calling another service.
        time.sleep(0.5) # Simulate work being done to generate the chunk.
        
        chunk_data = f"Chunk {i+1} of 10. The time is now {time.time()}.\n"
        yield chunk_data.encode("utf-8") # Yield bytes
        
    yield b"----- END STREAM -----\n"
    print("Download stream finished.")

@app.route("/api/status", methods=["POST"])
def post_status(request):
    """An endpoint that accepts POST requests."""
    return json_response({"status": "ok", "message": "Data received!"})

# 3. Wrap the application with the CORS middleware
# This allows all origins, and specific methods and headers for preflight.
cors_app = CORS(
    app,
    origins="*", # or ["http://localhost:3000", "https://my-frontend.com"]
    methods=["GET", "POST", "OPTIONS"],
    headers=["Content-Type", "Authorization"],
    allow_credentials=True
)

async def main_program():
    """
    The main entry point. We make it async to keep the event loop
    running, which allows our background web server to continue serving requests.
    """
    print("Starting API server with CORS enabled at http://127.0.0.1:5173")
    print("Try fetching from a different origin (e.g., a simple HTML file opened from your desktop).")
    
    # The `serve` function starts the HTTP server in a non-blocking Go routine.
    serve("127.0.0.1:5173", cors_app)
    
    # Now, we need to keep the main Pylearn program alive.
    # An infinite sleep is a common pattern for this in async applications.
    # while True:
    #     await asyncio.sleep(3600) # Sleep for a long time (1 hour)
    try:
        while True:
            await asyncio.sleep(3600)
    except KeyboardInterrupt:
        print("\nServer shutting down.")