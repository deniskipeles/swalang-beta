# examples/streaming_server.py
from httpserver import serve
import time

def upload_handler(request):
    """
    Handles large file uploads by streaming the request body.
    It calculates the size of the upload without loading it all into memory.
    """
    print("Receiving a large upload...")
    total_size = 0
    # The request body is now an iterator of bytes chunks.
    for chunk in request.body:
        chunk_size = len(chunk)
        print(format_str("  ... received chunk of size: {chunk_size}"))
        total_size = total_size + chunk_size
        
    print(format_str("Upload complete. Total size: {total_size} bytes."))
    return format_str("Received {total_size} bytes.")

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
        j = i+1
        print(j)
        chunk_data = format_str("Chunk {j} of 10. The time is now {time.time()}.\n")
        yield chunk_data.encode("utf-8") # Yield bytes
        
    yield b"----- END STREAM -----\n"
    print("Download stream finished.")


# Our routes dictionary
routes = {
    "/upload": upload_handler,
    "/download": download_generator,
}

async def main_program():
    print("Starting streaming server at http://127.0.0.1:8080")
    print("  POST to /upload with a large file to test request streaming.")
    print("    e.g., curl -X POST --data-binary @/path/to/large/file http://127.0.0.1:8080/upload")
    print("  GET /download to test response streaming.")
    print("    e.g., curl http://127.0.0.1:8080/download")
    
    serve("127.0.0.1:8080", routes)
    
    # Keep the server running
    while True:
        time.sleep(3600)