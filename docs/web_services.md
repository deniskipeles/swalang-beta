# Web Services and Networking

Swalang makes it easy to interact with the web and build networked applications using its built-in `http` and `mongoose` modules.

## 1. Using the HTTP Client

For consuming external APIs, the `http` module provides a simple, high-level interface similar to the Python `requests` library.

### Making a GET Request

```python
import http

try:
    response = http.get("https://api.github.com/repos/deniskipeles/pylearn")
    print(format_str("Status: {response.status_code}"))

    # Parse JSON automatically
    data = response.json()
    print(format_str("Repo Name: {data['name']}"))
    print(format_str("Stars: {data['stargazers_count']}"))
except Exception as e:
    print(format_str("Request failed: {e}"))
```

### Making a POST Request

```python
import http

payload = {"title": "Hello Swalang", "body": "Building great things."}
response = http.post("https://jsonplaceholder.typicode.com/posts", json=payload)

if response.status_code == 201:
    print("Post created successfully!")
    print(response.json())
```

## 2. Building a Web Server

Swalang includes a flexible web server framework built on the high-performance **Mongoose** library.

### Basic Web Server

The `http.Server` class allows you to define routes using decorators.

```python
import http

app = http.Server(port=8080)

@app.route("/")
def home(request):
    return "<h1>Welcome to Swalang</h1><p>The fastest way to build native apps.</p>"

@app.route("/api/health")
def health_check(request):
    # Returning a dictionary automatically converts it to a JSON response
    return {"status": "ok", "version": "1.0.0"}

print("Server starting on http://localhost:8080")
app.run()
```

### Handling Parameters and Methods

```python
@app.route("/greet")
def greet(request):
    name = request.query.get("name", "Stranger")
    return format_str("Hello, {name}!")

@app.route("/data", methods=["POST"])
def receive_data(request):
    print(format_str("Received body: {request.body}"))
    return {"received": True}
```

## 3. Real-time Communication with WebSockets

Swalang supports WebSockets for building interactive applications like chat systems or live dashboards.

```python
import http
import mongoose

app = http.Server(port=5000)

@app.ws_route("/ws")
def handle_ws(conn, event, msg):
    if event == "OPEN":
        print("Client connected!")
        mongoose.ws_send(conn, "Welcome to the real-time stream.")
    elif event == "MSG":
        print(format_str("Client sent: {msg.data}"))
        # Echo the message back
        mongoose.ws_send(conn, format_str("You said: {msg.data}"))
    elif event == "CLOSE":
        print("Client disconnected")

app.run()
```

## 4. Advanced Networking with Mongoose

For low-level control, you can use the `mongoose` module directly. This is useful for building custom protocols or highly optimized services.

```python
import mongoose

def on_http_event(conn, request):
    mongoose.http_reply(conn, 200, "Content-Type: text/plain\r\n", "Direct Mongoose Response")

mgr = mongoose.Manager()
mgr.http_listen("http://0.0.0.0:3000", on_http_event)

while True:
    mgr.poll(1000) # Poll every 1 second
```

## Best Practices

- **Async Loops**: When building highly concurrent servers, use the `asyncio` loop to handle multiple requests without blocking.
- **Error Handling**: Always wrap network calls in `try...except` blocks to handle timeouts, DNS failures, or server errors gracefully.
- **Security**: When using Mongoose in production, ensure you configure SSL/TLS settings via the `mbedtls` integration.
