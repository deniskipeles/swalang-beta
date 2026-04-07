# examples/net_stdlib_test.py

"""
High-level, framework-like networking interface for Pylearn.

This module provides a powerful, Pythonic, object-oriented API for building
network clients and servers. It wraps the native Go-based `_net` implementation,
offering everything from low-level TCP sockets to a high-level, decorator-based
web server inspired by frameworks like Flask.

Core Features:
- A simple, elegant web server with decorator-based routing (`@app.route`).
- Support for middleware to process requests.
- Out-of-the-box support for HTTP/2 and TLS.
- High-performance, memory-efficient streaming responses.
- Low-level TCP client and server primitives (`dial`, `listen`).
- WebSocket client functionality.
- SOCKS5 proxy support for client connections.
- Utility submodules for ICMP (ping), IDNA, Public Suffix lookups, and more.
"""

import _net as _native_net
import os  # Used for the streaming example

# ==============================================================================
#  Re-export Core Classes and Functions from the Native Backend
# ==============================================================================

# --- High-Level Server Framework ---
App = _native_net.App
Request = _native_net.Request
Response = _native_net.Response

# --- Low-Level Networking Primitives ---
Connection = _native_net.Connection
Listener = _native_net.Listener
Addr = _native_net.Addr
dial = _native_net.dial
listen = _native_net.listen
StreamingBody = _native_net.StreamingBody

# --- Error Classes (re-exported from the top-level native module) ---
# <<< START OF FIX >>>
# The native error classes are registered at the top level of the _net module,
# not inside the submodules. This corrects the AttributeError.
ConnectionError = _native_net.ConnectionError
TimeoutError = _native_net.TimeoutError
ProxyError = _native_net.ProxyError
ICMPError = _native_net.ICMPError
HTTP2Error = _native_net.HTTP2Error
# The WebSocketError is part of the WebSocket submodule's environment.
WebSocketError = _native_net.websocket.WebSocketError
# <<< END OF FIX >>>

# --- Submodules ---
# Access advanced functionality via these namespaces.
websocket = _native_net.websocket
http2 = _native_net.http2
publicsuffix = _native_net.publicsuffix
idna = _native_net.idna
icmp = _native_net.icmp
proxy = _native_net.proxy

# ==============================================================================
#  High-Level Helper Classes and Functions
# ==============================================================================

# class StreamingBody:
#     """
#     A wrapper for file-like objects to enable streaming responses in the web framework.
#     When a handler returns an instance of this class, the server will
#     stream the content of the file-like object instead of buffering it, which is
#     highly memory-efficient for large files.

#     Example:
#         @app.route("/download")
#         def download_file(req, res):
#             file_handle = open("large_video.mp4", "rb")
#             res.header("Content-Type", "video/mp4")
#             return StreamingBody(file_handle)
#     """
#     def __init__(self, file_like_object):
#         if not hasattr(file_like_object, 'read'):
#             raise TypeError("Object passed to StreamingBody must have a 'read' method.")
#         self.source = file_like_object

class TCPServer:
    """
    A high-level TCP server that uses a handler class to process requests.
    This is similar to Python's standard `socketserver.TCPServer`.
    """
    def __init__(self, server_address, RequestHandlerClass):
        if not isinstance(server_address, tuple):
            raise TypeError("server_address must be a (host, port) tuple")
        self.server_address = server_address
        self.RequestHandlerClass = RequestHandlerClass
        self.socket = listen("tcp", f"{server_address[0]}:{server_address[1]}")

    def serve_forever(self):
        """Handle one request at a time until shutdown."""
        print(f"TCP server listening on {self.server_address[0]}:{self.server_address[1]}...")
        while True:
            try:
                conn, addr = self.socket.accept()
                handler = self.RequestHandlerClass(conn, addr, self)
            except Exception as e:
                print(f"Error in server loop: {e}")

class BaseRequestHandler:
    """
    Base class for TCP request handler classes.
    Users should subclass this and override the handle() method.
    """
    def __init__(self, request, client_address, server):
        self.request = request
        self.client_address = client_address
        self.server = server
        self.setup()
        try:
            self.handle()
        finally:
            self.finish()

    def setup(self):
        pass

    def handle(self):
        pass

    def finish(self):
        self.request.close()


# # ==============================================================================
# #  Runnable Demonstration
# # ==============================================================================

# async def main_program1():
#     """
#     This function is executed when this file is run directly.
#     It demonstrates the key features of the `net` library.
#     """
#     print("--- Running `net` library demonstrations ---")

#     # --- 1. High-Level Web Server (Flask/Django style) ---
#     print("\n[1] Starting high-level web server demonstration...")
#     app = App()

#     @app.middleware
#     def logging_middleware(next_handler):
#         def wrapper(req, res):
#             print(f"Request received: {req.method} {req.path}")
#             next_handler(req, res)
#             print(f"Response sent for: {req.path}")
#         return wrapper

#     @app.route("/", "GET")
#     def index(req, res):
#         res.header("Content-Type", "text/html")
#         res.body(b"<h1>Hello from the Pylearn `net.App`!</h1><p>Try visiting /json or /stream</p>")

#     @app.route("/json", "GET")
#     def json_route(req, res):
#         res.json({
#             "message": "This is a JSON response!",
#             "status": "ok",
#             "items": [1, "two", 3.0, True]
#         })

#     # --- START SERVER (for testing with curl) ---
#     # Since main_program is async and doesn't block, we can start the server.
#     # The interpreter will keep running because of the other demos.
#     # In a real app, app.run() would be the last line of a synchronous script.
#     print("Starting server in background for curl tests...")
#     # NOTE: This is a simplified way to run in the background. A real async server
#     # would integrate with the main event loop.
#     # For now, we will just start it and let the script continue.
#     # A dedicated `app.start()` and `app.stop()` would be better for async apps.
#     # Let's run the server in a non-blocking way for the demo.
#     # The current `app.run` is blocking, so we cannot call it directly here
#     # without stopping the rest of the script.
#     print("Web server app configured. Run this script without other demos to test the blocking server.")
#     print("Example: Comment out demos 2, 3, 4 and call app.run('localhost:8080')")


#     # --- 2. WebSocket Client ---
#     print("\n[2] WebSocket client demonstration...")
#     try:
#         ws_url = "wss://echo.websocket.events"
#         print(f"Connecting to WebSocket: {ws_url}")
#         headers = {"User-Agent": "Pylearn-Net-Client/1.0"}
#         conn = websocket.connect(ws_url, headers)
#         print("WebSocket connection successful.")
#         message_to_send = "Hello from a more robust Pylearn client!"
#         conn.send(message_to_send)
#         print(f"Sent: '{message_to_send}'")
#         response = conn.recv()
#         print(f"Received: '{response}'")
#         conn.close()
#         print("WebSocket connection closed.")
#     except WebSocketError as e:
#         print(f"WebSocket Error: {e}")
#     except ConnectionError as e:
#         print(f"Could not connect to WebSocket server: {e}")

#     # --- 3. ICMP Ping ---
#     print("\n[3] ICMP Ping demonstration...")
#     try:
#         target = "8.8.8.8"
#         print(f"Pinging {target}...")
#         rtt = icmp.ping(target, timeout=1.0)
#         if rtt is None:
#             print("Ping timed out.")
#         else:
#             print(f"Success! Reply from {target} in {rtt:.2f}ms")
#     except ICMPError as e:
#         print(f"ICMP Error: {e}")
#         print("    (Note: Pinging often requires administrator/root privileges to create raw sockets.)")
#         print("    (This is an expected error if not running with sudo.)")

#     # --- 4. Proxied TCP Connection ---
#     print("\n[4] SOCKS5 Proxy demonstration...")
#     try:
#         print("Creating a SOCKS5 dialer for localhost:9050...")
#         socks_dialer = proxy.socks5("tcp", "localhost:9050")
#         target_host = "www.google.com:80"
#         print(f"Dialing {target_host} through the proxy...")
#         conn = dial("tcp", target_host, dialer=socks_dialer)
#         peer = conn.getpeername()
#         print(f"Successfully connected to {peer.string()} via proxy!")
#         conn.close()
#     except ProxyError as e:
#         print(f"Proxy Error: {e}. Is a SOCKS5 proxy running on localhost:9050?")
#     except ConnectionError as e:
#          print(f"Proxy Connection Error: {e}")

#     print("\n--- All demonstrations complete. ---")

# async def main_program():
#     """
#     This is an example of a complete web server using the new streaming feature.
#     To test it, uncomment the `app.run` call at the end and run this file.
#     Then, in another terminal, run:
    
#     curl http://localhost:8080/stream -o /dev/null
    
#     You will see the server printing progress without using a lot of memory.
#     """
#     print("--- Pynet Streaming Server Demonstration ---")

#     app = App()

#     # Create a large dummy file for the streaming demonstration
#     DUMMY_FILE_NAME = "large_dummy_file.dat"
#     print(f"Creating a large dummy file: {DUMMY_FILE_NAME}")
#     with open(DUMMY_FILE_NAME, "wb") as f:
#         # Create a 100MB file
#         chunk = b'\x00' * 1024 * 1024 # 1MB chunk
#         for i in range(100):
#             f.write(chunk)
#     print("Dummy file created.")


#     @app.route("/", "GET")
#     def index(req, res):
#         res.header("Content-Type", "text/html")
#         res.body(f"""
#             <h1>Pylearn Streaming Server</h1>
#             <p>This server demonstrates memory-efficient file streaming.</p>
#             <p>Try downloading the file: <a href="/stream">/stream</a></p>
#             <p>In your terminal, you can run:</p>
#             <code>curl http://localhost:8080/stream -o downloaded_file.dat</code>
#         """.encode('utf-8'))

#     @app.route("/stream", "GET")
#     def stream_file(req, res):
#         print(f"Request received for /stream. Preparing to stream {DUMMY_FILE_NAME}...")
        
#         try:
#             file_handle = open(DUMMY_FILE_NAME, "rb")
            
#             # Set headers before returning the streaming body
#             res.header("Content-Type", "application/octet-stream")
#             res.header("Content-Disposition", f'attachment; filename="{DUMMY_FILE_NAME}"')
            
#             # Returning this object triggers the native streaming logic
#             return StreamingBody(file_handle)
            
#         except OSError as e:
#             res.status(500)
#             res.body(f"Error opening file: {e}".encode('utf-8'))


#     # To run this example, you would make this the only active part of the script
#     # and call it from a synchronous context.
#     print("\nTo run the streaming server, comment out other demos in your test file")
#     print("and add this line to the end of main_program:")
#     print("app.run('localhost:8080')")
    
#     # Example of how to run it (comment out other demos to use this):
#     app.run('localhost:8080')