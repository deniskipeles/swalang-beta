import mongoose

def handle_request(conn, request):
    print(format_str("\n📥 Received {request.method} request to {request.uri}"))
    
    if request.method == "POST":
        print(format_str("📦 Body Data: {request.body}"))
        response_body = format_str("Success! You posted: {request.body}")
    else:
        response_body = format_str("<h1>Hello!</h1><p>You visited: {request.uri}</p>")
    
    mongoose.http_reply(conn, 200, "Content-Type: text/html\r\n", response_body)

mgr = mongoose.Manager()

try:
    mgr.http_listen("http://0.0.0.0:5000", handle_request)
    print("Press Ctrl+C to stop.")
    
    while True:
        mgr.poll(1000)
except Exception as e:
    print(e)
finally:
    mgr.free()