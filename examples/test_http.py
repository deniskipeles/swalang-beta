import http
import mongoose

print("===========================================")
print("🌐 Testing Swalang Complete HTTP Library")
print("===========================================")

print("\n1. Testing HTTP Client (GET)...")
try:
    resp = http.get("https://jsonplaceholder.typicode.com/todos/1")
    print(format_str("👉 Status: {resp.status_code}"))
    print(format_str("👉 JSON Response: {resp.json()}"))
except Exception as e:
    print(format_str("HTTP Client Error: {e}"))

print("\n2. Testing HTTP Server & WebSockets...")
print("Starting server on port 5000...")

app = http.Server(port=5000)

@app.route("/")
def index(req):
    return "<h1>Welcome to Swalang HTTP Server!</h1><p>Try <a href='/api'>/api</a> or connect to /ws via WebSocket client.</p>"

@app.route("/api")
def api(req):
    return {"message": "Hello API", "method": req.method, "query": req.query}

@app.ws_route("/ws")
def ws_handler(conn, event, msg):
    if event == "OPEN":
        print("🟢 WebSocket Client Connected!")
        mongoose.ws_send(conn, "Welcome to Swalang WebSocket Server!")
    elif event == "MSG":
        print(format_str("📩 WS Received: {msg.data}"))
        # Echo back
        mongoose.ws_send(conn, format_str("Echo: {msg.data}"))
    elif event == "CLOSE":
        print("🔴 WebSocket Client Disconnected")

# This will block the thread forever, answering requests
app.run()