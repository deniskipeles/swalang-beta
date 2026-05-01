import mongoose

# We use a global flag to know when the async request is done
request_finished = False

def handle_response(conn, response):
    global request_finished
    print(format_str("✅ Server Responded! \n{response.body}"))
    request_finished = True

mgr = mongoose.Manager()

try:
    print("🚀 Connecting to http://localhost:8000/api/test...")
    mgr.http_connect("http://localhost:8000/api/test", handle_response)
    
    # Event loop
    while not request_finished:
        mgr.poll(100)
        
except Exception as e:
    print(e)
finally:
    mgr.free()