# examples/run_server.py

import net
import os
import json # Use the standard json library for creating the response

# This code demonstrates a complete web server using the pynet framework.
app = net.App()

# Create a dummy file for the streaming demonstration
DUMMY_FILE_NAME = "large_dummy_file.dat"
FILE_SIZE_MB = 10 # Using a smaller size for quicker startup
if not os.path.exists(DUMMY_FILE_NAME):
    print(f"Creating a {FILE_SIZE_MB}MB dummy file: {DUMMY_FILE_NAME}")
    with open(DUMMY_FILE_NAME, "wb") as f:
        chunk = b'\x00' * 1024 * 1024 # 1MB chunk
        for i in range(FILE_SIZE_MB):
            f.write(chunk)
    print("Dummy file created.")

# --- GET Route (Home Page) ---
@app.route("/", "GET")
def index(req, res):
    res.header("Content-Type", "text/html")
    body = (b"""
        <h1>Pylearn `net.App` Server</h1>
        <p>This server demonstrates several features of the pynet framework.</p>
        <h2>GET Endpoints:</h2>
        <ul>
            <li><a href="/">/</a> - This page</li>
            <li><a href="/stream">/stream</a> - Download a large file via streaming</li>
        </ul>
        <h2>POST Endpoints:</h2>
        <p>Use a tool like `curl` to test these.</p>
        <ul>
            <li><b>/reflect</b> (POST with JSON)
                <pre><code>curl -X POST -H "Content-Type: application/json" -d '{"name": "pylearn", "value": 123}' http://localhost:8080/reflect</code></pre>
            </li>
            <li><b>/upload</b> (POST with binary file)
                <pre><code>curl -X POST --data-binary @large_dummy_file.dat http://localhost:8080/upload</code></pre>
            </li>
        </ul>
    """)
    print(type(body))
    res.body(body)

# --- GET Route (Streaming) ---
@app.route("/stream", "GET")
def stream_file(req, res):
    print(f"Request for /stream received. Preparing to stream...")
    try:
        file_stats = os.stat(DUMMY_FILE_NAME)
        file_size = file_stats['st_size']
        
        file_handle = open(DUMMY_FILE_NAME, "rb")
        
        res.header("Content-Type", "application/octet-stream")
        res.header("Content-Length", str(file_size))
        res.header("Content-Disposition", f'attachment; filename="{DUMMY_FILE_NAME}"')
        
        return net.StreamingBody(file_handle)
    except OSError as e:
        res.status(500)
        res.body(f"Error: {e}".encode('utf-8'))

# --- POST Route (JSON) ---
@app.route("/reflect", "POST")
def reflect_json(req, res):
    print("Request for /reflect received.")
    try:
        # The req.json() method automatically reads the body and parses it
        data = req.json()
        print(f"  - Received JSON data: {data}")

        # Prepare a response
        response_data = {
            "status": "success",
            "received_data": data,
            "message": "Data received and reflected successfully."
        }
        
        # The res.json() method automatically sets the Content-Type
        # and serializes the Pylearn dictionary to a JSON string.
        res.json(response_data)
        print("  - Sent JSON response.")
        
    except json.JSONDecodeError as e:
        res.status(400) # Bad Request
        res.json({"status": "error", "message": f"Invalid JSON: {e}"})
        print(f"  - Error: Invalid JSON received.")
    except Exception as e:
        res.status(500) # Internal Server Error
        res.json({"status": "error", "message": f"An internal error occurred: {e}"})

# --- POST Route (Binary Upload) ---
@app.route("/upload", "POST")
def upload_binary(req, res):
    print("Request for /upload received.")
    try:
        # The req.body() method reads the entire raw request body as bytes.
        data = req.body()
        size_kb = len(data) / 1024
        print(f"  - Received {size_kb} KB of binary data.")

        # Send a confirmation response
        res.json({
            "status": "success",
            "bytes_received": len(data),
            "message": "File uploaded successfully."
        })
        
    except Exception as e:
        res.status(500)
        res.json({"status": "error", "message": f"An internal error occurred during upload: {e}"})

# This is a blocking call that starts the server.
app.run('localhost:5173')