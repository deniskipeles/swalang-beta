import mongoose

print("===========================================")
print("🌐 Testing Mongoose FFI Wrapper")
print("===========================================")

# Mongoose Event ID for HTTP Messages (Usually 5 or 6 depending on the version)
# We will print the event ID in the console so you can see the connection lifecycle!
MG_EV_HTTP_MSG = 5 

def my_http_handler(conn_ptr, ev, ev_data_ptr):
    # Print the event number to see Mongoose working (0 = POLL, 1 = ACCEPT, 2 = READ, etc.)
    if ev != 0: # Ignore the constant POLL events for clean logging
        print(format_str("🔔 Mongoose Event Received: {ev}"))
    
    # Event 5 in Mongoose v7.x is MG_EV_HTTP_MSG
    if ev == 5:
        print("✅ Received HTTP Request! Sending response...")
        
        body = "<h1>Hello from Mongoose via Swalang FFI!</h1><p>The wrapper works perfectly.</p>"
        headers = "Content-Type: text/html\r\n"
        
        # Send the response back to the browser
        mongoose.http_reply(conn_ptr, 200, headers, body)

# Initialize the Mongoose Manager
mgr = mongoose.Manager()

try:
    # Start the server
    mgr.http_listen("http://0.0.0.0:5000", my_http_handler)
    print("👉 Open your browser and go to http://localhost:5000")
    print("Press Ctrl+C to stop.")
    
    # Run the event loop forever
    while True:
        mgr.poll(1000) # Block for up to 1000ms waiting for network traffic
        
except Exception as e:
    print(format_str("Server stopped: {e}"))
finally:
    mgr.free()
    print("✅ Mongoose resources freed.")