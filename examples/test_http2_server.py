# file: test_http2_server.py

import httpserver
from net import http2,BaseRequestHandler

# Define a simple request handler
# BaseHTTPRequestHandler = httpserver.BaseHTTPRequestHandler
class MyHandler(BaseRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header('Content-type', 'text/html')
        self.end_headers()
        # The protocol version will be available on the request object
        protocol = self.request_version
        body = f"<h1>Hello from Pylearn!</h1><p>Protocol: {protocol}</p>"
        self.wfile.write(body.encode('utf-8'))

# --- Main Server Setup ---
async def main_program():
    host = 'localhost'
    port = 8443  # HTTP/2 requires TLS

    # 1. Create a standard HTTP server instance
    server = httpserver.HTTPServer((host, port), MyHandler)
    
    # 2. **Enable HTTP/2 on the server instance**
    try:
        net.http2.configure_server(server)
        print("HTTP/2 configured successfully on the server.")
    except net.http2.HTTP2Error as e:
        print(f"Failed to configure HTTP/2: {e}")
        return None

    # 3. Set up TLS (SSL) context, as browsers require it for HTTP/2
    # This assumes you have `server.crt` and `server.key` files.
    # You can generate them with:
    # openssl req -new -x509 -days 365 -nodes -out server.crt -keyout server.key
    try:
        server.socket = httpserver.ssl_wrap_socket(
            server.socket,
            keyfile="server.key",
            certfile="server.crt"
        )
        print("Server socket wrapped with TLS.")
    except Exception as e:
        print(f"Error setting up TLS: {e}")
        print("Please generate server.key and server.crt files.")
        return None

    print(f"Starting server on https://{host}:{port}")
    print("Use a modern browser or curl with --http2 to test.")
    # Example curl command: curl -k --http2 https://localhost:8443
    
    # 4. Run the server forever
    server.serve_forever()