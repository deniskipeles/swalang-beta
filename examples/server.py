# examples/server.py
# A simple web server example using the built-in httpserver module.

import httpserver
import sys # To print arguments (optional)
import os # To get env vars (optional)

print("--- PyLearn Simple Web Server ---")

# Define the function that will handle incoming requests
def request_handler(req):
    print(">> Handling Request:")
    print("   Method:", req.method)
    print("   Path:", req.path)
    # Print headers (if available and Dict iteration works)
    # print("   Headers:", req.headers)
    # Example: Accessing a specific header (case-insensitive key assumed due to Go implementation)
    user_agent = req.headers["user-agent"] # Access header (might KeyError if missing)
    print("   User-Agent approx:", user_agent)
    print("   Body:", req.body)

    # Basic Routing based on path
    if req.path == "/":
        # Simple HTML response
        print('entered /')
        body = """
        <html>
          <head><title>PyLearn Server</title></head>
          <body>
            <h1>Welcome to the PyLearn Server!</h1>
            <p>You requested the root path (/).</p>
            <p><a href="/hello">Say Hello</a></p>
            <p><a href="/info">Show Info</a></p>
            <p><a href="/other">Other Path</a></p>
            <p>Try sending a POST request to /submit</p>
          </body>
        </html>
        """
        return body # Return HTML string directly (implies 200 OK)

    elif req.path == "/hello":
        name = os.getenv("USERNAME", "Guest") # Example using 'os' module
        return "<h1>Hello, " + name + "!</h1><p>From PyLearn.</p>"

    elif req.path == "/info":
        # Display some info from the request and sys module
        info_body = "<h2>Request Info</h2>"
        info_body = info_body + "<p>Method: " + req.method + "</p>"
        info_body = info_body + "<p>Path: " + req.path + "</p>"
        info_body = info_body + "<p>Body Length: " + str(len(req.body)) + "</p>" # Example using len()
        info_body = info_body + "<h2>System Info</h2>"
        info_body = info_body + "<p>Platform: " + sys.platform + "</p>"
        info_body = info_body + "<p>Argv Count: " + str(len(sys.argv)) + "</p>"
        # Note: Accessing sys.argv elements would require list indexing support
        return info_body

    elif req.path == "/submit" and req.method == "POST":
        # Handle POST data
        return "Received your POST data: " + req.body

    else:
        # Simple "Not Found" for other paths
        # TODO: Implement returning 404 status code when handler can return more than just body
        return "<h1>404 - Not Found</h1><p>Path '" + req.path + "' not handled.</p>"

# --- Server Configuration ---
# Get host/port from environment or use defaults
host = os.getenv("PYLEARN_HOST", "127.0.0.1") # Default to localhost
port_str = os.getenv("PYLEARN_PORT", "8080") # Default to 8080
address = host + ":" + port_str
# TODO: Add error handling if port_str is not a valid number

# --- Start the Server ---
try:
    print("Starting server on address:", address)
    # The 'serve' function starts the server in the background
    # and this script continues (and likely finishes)
    httpserver.serve(address, request_handler)

    # Keep the main script alive (optional, Go routine keeps server running)
    # Without this, the script might exit but the Go server goroutine might continue.
    # Depending on your interpreter's design, this might be needed or not.
    # A simple way to wait indefinitely (requires input support):
    print("Server started. Press Ctrl+C to stop.")
    input("Waiting...") # This would block here

    # Or just let the script finish, the Go process stays alive due to the server goroutine
    print("Server start command issued. Main script finishing.")
    # The Go process will only exit if the server goroutine errors fatally
    # or if you explicitly call sys.exit() or send a signal (Ctrl+C).

except Exception as e:
    # Catch potential errors during serve() setup if any were possible
    # (e.g., invalid address format before Go server starts - unlikely with current code)
    print("Error starting server. Check address format and permissions.",e)
    # import sys # Import here if needed
    # sys.exit(1)