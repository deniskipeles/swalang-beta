# examples/websocket_server_test.py
import httpserver
import aio # For aio.sleep if needed by handler
import time # For logging

MY_GLOBAL_WS_VAR = "WebSocket Global"

async def handle_echo_websocket(request, ws):
    print(format_str("WS Handler: Connection established from {request.headers.get('Remote-Addr', 'unknown')}"))
    print(format_str("WS Handler: Accessing global: {MY_GLOBAL_WS_VAR}"))
    
    try:
        await ws.send("Welcome to the Pylearn Echo WebSocket!")
        
        while True:
            print("WS Handler: Waiting for message...")
            try:
                message = await ws.receive()
            except Exception as receive_error:
                # Handle WebSocket closure or other receive errors
                print(format_str("WS Handler: Connection closed or error during receive: {receive_error}"))
                break
            
            if isinstance(message, str):
                print(format_str("WS Handler: Received TEXT: '{message}'"))
                if message == "error_me":
                    raise ValueError("Test error in handler")
                if message == "close_me":
                    print("WS Handler: Client asked to close. Closing with normal status.")
                    await ws.close(1000, "Client requested close")
                    break # Explicitly break after closing
                await ws.send(format_str("Echo (str): {message}"))
            elif isinstance(message, bytes):
                print(format_str("WS Handler: Received BINARY data, length: {len(message)}"))
                # Try to echo back the exact binary data received
                try:
                    await ws.send(message)  # Echo back the same binary data
                except Exception as send_error:
                    print(format_str("WS Handler: Error sending binary data: {send_error}"))
                    # Fallback to text response
                    response_text = format_str("Echo (bytes): Received {len(message)} bytes but couldn't echo binary")
                    await ws.send(response_text)
            else:
                print(format_str("WS Handler: Received unexpected message type: {type(message)}"))
                break # Break on unexpected type
    
    except ValueError as e:
        # Handle the specific "error_me" test case
        print(format_str("WS Handler: Test error occurred: {e}"))
        if hasattr(ws, "close"):
            await ws.close(1011, "Internal Server Error")
    
    except Exception as e:
        # This catches other unexpected errors
        print(format_str("WS Handler: An unexpected error occurred: {e} (type: {type(e)})"))
        if hasattr(ws, "close"):
            await ws.close(1011, "Internal Server Error")

    print(format_str("WS Handler: Loop finished for {request.headers.get('Remote-Addr', 'unknown')}"))


def handle_http_root(request):
    return """
    <h1>Pylearn WebSocket Test Server</h1>
    <p>Connect to /echo_ws with a WebSocket client.</p>
    <script>
        let ws = new WebSocket("wss://" + window.location.host + "/echo_ws");
        ws.onopen = () => { 
            console.log("WebSocket opened!"); 
            ws.send("Hello from browser!");
            ws.send(new Uint8Array([1,2,3,4,5])); // Send binary
        };
        ws.onmessage = (event) => {
            if (event.data instanceof Blob) {
                event.data.arrayBuffer().then(buffer => {
                    const view = new Uint8Array(buffer);
                    console.log("Received BINARY:", view);
                });
            } else {
                console.log("Received TEXT:", event.data);
            }
        };
        ws.onclose = (event) => { console.log("WebSocket closed:", event.code, event.reason); };
        ws.onerror = (event) => { console.error("WebSocket error:", event); };
        window.ws = ws; // For console interaction: ws.send("message"), ws.close()
    </script>
    """

routes = {
    "/echo_ws": handle_echo_websocket, # This handler MUST be async
    "/": handle_http_root,
}

def main():
    address = "127.0.0.1:5173" # Use a different port
    print(format_str("Pylearn: Attempting to start WebSocket-aware server on {address}"))
    httpserver.serve(address, routes)
    print(format_str("Pylearn: Server setup initiated. Check browser/curl at http://{address}"))
    
    try:
        input("Pylearn: Server running. Press Enter in this Pylearn script to 'finish'...")
    except EOFError as e:
        print("Pylearn: Script continuing non-interactively (Go server continues).")
        print(e)
        while True: 
            time.sleep(3600)

if __name__ == "__main__":
    main()












# # examples/websocket_server_test.py
# import httpserver
# import aio # For aio.sleep if needed by handler
# import time # For logging

# MY_GLOBAL_WS_VAR = "WebSocket Global"

# async def handle_echo_websocket(request, ws):
#     print(format_str("WS Handler: Connection established from {request.headers.get('Remote-Addr', 'unknown')}"))
#     print(format_str("WS Handler: Accessing global: {MY_GLOBAL_WS_VAR}"))
    
#     try:
#         await ws.send("Welcome to the Pylearn Echo WebSocket!")
        
#         while True:
#             print("WS Handler: Waiting for message...")
#             try:
#                 message = await ws.receive()
#             except Exception as receive_error:
#                 # Handle WebSocket closure or other receive errors
#                 print(format_str("WS Handler: Connection closed or error during receive: {receive_error}"))
#                 break
            
#             if isinstance(message, str):
#                 print(format_str("WS Handler: Received TEXT: '{message}'"))
#                 if message == "error_me":
#                     raise ValueError("Test error in handler")
#                 if message == "close_me":
#                     print("WS Handler: Client asked to close. Closing with normal status.")
#                     await ws.close(1000, "Client requested close")
#                     break # Explicitly break after closing
#                 await ws.send(format_str("Echo (str): {message}"))
#             elif isinstance(message, bytes):
#                 print(format_str("WS Handler: Received BINARY data, length: {len(message)}"))
#                 await ws.send(b"Echo (bytes): " + message)
#             else:
#                 print(format_str("WS Handler: Received unexpected message type: {type(message)}"))
#                 break # Break on unexpected type
    
#     except ValueError as e:
#         # Handle the specific "error_me" test case
#         print(format_str("WS Handler: Test error occurred: {e}"))
#         if hasattr(ws, "close"):
#             await ws.close(1011, "Internal Server Error")
    
#     except Exception as e:
#         # This catches other unexpected errors
#         print(format_str("WS Handler: An unexpected error occurred: {e} (type: {type(e)})"))
#         if hasattr(ws, "close"):
#             await ws.close(1011, "Internal Server Error")

#     print(format_str("WS Handler: Loop finished for {request.headers.get('Remote-Addr', 'unknown')}"))


# def handle_http_root(request):
#     return """
#     <h1>Pylearn WebSocket Test Server</h1>
#     <p>Connect to /echo_ws with a WebSocket client.</p>
#     <script>
#         let ws = new WebSocket("wss://" + window.location.host + "/echo_ws");
#         ws.onopen = () => { 
#             console.log("WebSocket opened!"); 
#             ws.send("Hello from browser!");
#             ws.send(new Uint8Array([1,2,3,4,5])); // Send binary
#         };
#         ws.onmessage = (event) => {
#             if (event.data instanceof Blob) {
#                 event.data.arrayBuffer().then(buffer => {
#                     const view = new Uint8Array(buffer);
#                     console.log("Received BINARY:", view);
#                 });
#             } else {
#                 console.log("Received TEXT:", event.data);
#             }
#         };
#         ws.onclose = (event) => { console.log("WebSocket closed:", event.code, event.reason); };
#         ws.onerror = (event) => { console.error("WebSocket error:", event); };
#         window.ws = ws; // For console interaction: ws.send("message"), ws.close()
#     </script>
#     """

# routes = {
#     "/echo_ws": handle_echo_websocket, # This handler MUST be async
#     "/": handle_http_root,
# }

# def main():
#     address = "127.0.0.1:5173" # Use a different port
#     print(format_str("Pylearn: Attempting to start WebSocket-aware server on {address}"))
#     httpserver.serve(address, routes)
#     print(format_str("Pylearn: Server setup initiated. Check browser/curl at http://{address}"))
    
#     try:
#         input("Pylearn: Server running. Press Enter in this Pylearn script to 'finish'...")
#     except EOFError as e:
#         print("Pylearn: Script continuing non-interactively (Go server continues).")
#         print(e)
#         while True: 
#             time.sleep(3600)

# if __name__ == "__main__":
#     main()