from httpserver import serve
import asyncio

async def echo_handler(request, websocket):
    """
    This async function will be called for each new WebSocket connection.
    """
    print(format_str("New WebSocket connection established: {websocket}"))
    try:
        # Loop forever, echoing messages back to the client.
        while True:
            # await waits for a message to arrive from the client.
            message = await websocket.receive()
            print(format_str("Received message: {message.inspect()}"))
            
            # Send the message back.
            response = format_str("Echo: {message.strip()}")
            await websocket.send(response)
            
    except WebSocketClosedError as e:
        print(format_str("Connection closed: {e}"))
    except Exception as e:
        print(format_str("An unexpected error occurred in handler: {e}"))
    # finally:
    #     print("WebSocket handler finished.")

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

async def main_program():
    routes = {
        "/": handle_http_root,
        "/echo_ws": echo_handler,
    }
    
    print("Starting WebSocket echo server at ws://127.0.0.1:5173/ws")
    serve("127.0.0.1:5173", routes)
    
    while True:
        await asyncio.sleep(3600)