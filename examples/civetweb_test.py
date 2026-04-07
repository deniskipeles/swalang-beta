# examples/civetweb_test.py
"""
A test and demonstration script for the Pylearn Civetweb FFI wrapper.
... (rest of docstring) ...
"""

import civetweb
import json
import sys # Import sys to use sys.exit

# Check if the Civetweb library was successfully loaded by the FFI.
if not civetweb.CIVETWEB_AVAILABLE:
    print("Error: The Civetweb shared library (libcivetweb.so) could not be loaded.")
    print("Please ensure it has been built correctly using the build script and is in the bin/ directory.")
    sys.exit(1)

# --- Server Configuration ---
PORT = 8088

# <<< START OF FIX >>>
# Civetweb is multi-threaded by default. When a C worker thread calls back
# into Go to run our Pylearn handler, it causes a crash because the thread is
# not managed by the Go runtime.
# Setting 'num_threads' to '1' forces Civetweb to handle all requests on the
# main thread, which is safe for our FFI callbacks.
server_options = {
    "listening_ports": str(PORT),
    "document_root": ".",
    "enable_keep_alive": "yes",
    "num_threads": "1", # CRITICAL: This forces single-threaded mode.
}
# <<< END OF FIX >>>

# --- Server and Route Definitions ---
# (No changes needed for the rest of this file)
# ... (rest of civetweb_test.py remains the same) ...
# Create a server instance with our options.
server = civetweb.Server(options=server_options)

@server.route("/", "GET")
def handle_root(req):
    """Handler for the root URL."""
    print(format_str("[{req.method}] Request for URI: {req.uri}"))
    # Return a simple HTML page as a string.
    return """
    <html>
        <head><title>Pylearn Civetweb Test</title></head>
        <body>
            <h1>Hello from Pylearn + Civetweb!</h1>
            <p>This server is running inside the Pylearn interpreter.</p>
            <ul>
                <li><a href="/api/data">GET JSON Data</a></li>
                <li><a href="/greeting?name=Pylearn">GET with Query String</a></li>
                <li>Try sending a POST request to /submit</li>
            </ul>
        </body>
    </html>
    """

@server.route("/api/data", "GET")
def handle_api(req):
    """Handler that returns a JSON response."""
    print(format_str("[{req.method}] Request for URI: {req.uri}"))
    data = {
        "framework": "Civetweb",
        "language": "Pylearn",
        "status": "ok",
        "items": [1, "two", 3.0, True]
    }
    # Use the standard json library to serialize the dictionary to a string.
    return json.dumps(data)

@server.route("/greeting", "GET")
def handle_greeting(req):
    """Handler that demonstrates parsing a query string."""
    print(format_str("[{req.method}] Request for URI: {req.uri}"))
    
    # Manually parse the query string for a 'name' parameter.
    name = "Guest" # Default value
    if req.query_string:
        parts = req.query_string.split('=')
        if len(parts) == 2 and parts[0] == 'name':
            name = parts[1]

    return format_str("<h1>Hello, {name}!</h1>")

@server.route("/submit", "POST")
def handle_submit(req):
    """Handler for POST requests that echoes the request body."""
    print(format_str("[{req.method}] Request for URI: {req.uri}"))
    
    # Read the body of the POST request.
    body_bytes = req.read()
    
    print(format_str("Received POST data: {body_bytes}"))
    
    # Echo the received data back to the client.
    return format_str("Received the following data:\n\n{body_bytes.decode('utf-8')}")

# --- Main Execution Block ---

def main_program():
    """The main entry point for the script."""
    print("--- Pylearn Civetweb Server Test ---")
    print(format_str("Server configured to run on port {PORT}"))
    print(format_str("Access the server at: http://localhost:{PORT}/"))
    print("Press Ctrl+C to stop the server.")

    # This is a blocking call that starts the server and waits for requests.
    try:
        server.start()
        import time
        while True:
            time.sleep(3600)
    except Exception as e:
        print(format_str("Server stopped due to an error: {e}"))

# The interpreter will execute this function if the script is run directly.
if __name__ == "__main__":
    main_program()