import httpserver
import json
import aio
import template
import time

# --- In-Memory "Database" ---
# A dictionary to hold our chat rooms.
# Each room contains a list of messages and a list of active WebSocket connections.
chat_rooms = {
    "general": {
        "messages": [
            {"user": "Admin", "text": "Welcome to the #general channel!"}
        ],
        "connections": [] # List of active ws objects
    },
    "random": {
        "messages": [
            {"user": "Admin", "text": "This is the #random channel. Anything goes!"}
        ],
        "connections": []
    }
}

# --- Helper Functions ---

# Broadcasts a message to all connected clients in a specific room.
async def broadcast_message(room_id, message_data):
    # We need to make sure the room exists.
    if not chat_rooms.__contains__(room_id):
        return None

    # Create the HTML snippet that will be sent to all clients.
    # HTMX will use the hx-swap-oob attribute to append this to the #messages div.
    html_to_send = format_str("""
    <div id="messages" hx-swap-oob="beforeend">
        <p><strong>{message_data['user']}:</strong> {message_data['text']}</p>
    </div>
    """)

    # Create a list of awaitable tasks (one for each connection).
    send_tasks = []
    for ws_conn in chat_rooms[room_id]["connections"]:
        send_tasks.append(ws_conn.send(html_to_send))
    
    # Run all send tasks concurrently.
    if len(send_tasks) > 0:
        await aio.gather(*send_tasks)

# --- Route Handlers ---

# The main WebSocket handler. One instance of this function runs per connected client.
async def websocket_chat_handler(request, ws):
    print(request.url)
    path_parts = request.url.split("/")
    room_id = path_parts[-1] if len(path_parts) > 2 else "general"

    # Ensure the room exists in our DB, create it if not.
    if not chat_rooms.__contains__(room_id):
        chat_rooms[room_id] = {"messages": [], "connections": []}
        print(format_str("Created new chat room: {room_id}"))

    # Add the new connection to the room's list.
    chat_rooms[room_id]["connections"].append(ws)
    print(format_str("New connection to room '{room_id}'. Total connections: {len(chat_rooms[room_id]['connections'])}"))

    try:
        # Send chat history to the newly connected client.
        history_html = ""
        for msg in chat_rooms[room_id]["messages"]:
            p = "<p><strong>" + msg['user'] + ":</strong> " + msg['text'] + "</p>"
            history_html = history_html + p
        
        # Wrap the history in a div that HTMX can swap.
        initial_payload = format_str("""
        <div id="messages" hx-swap-oob="innerHTML">
            {history_html}
        </div>
        """)
        await ws.send(initial_payload)

        # Main loop to listen for messages from this client.
        while True:
            json_string = await ws.receive()
            print(json_string)
            # The 'await ws.receive()' will raise a WebSocketClosedError when the client
            # disconnects, which will be caught by the 'except' block below.

            data = json.loads(json_string)
            message_text = data.get("message", "")

            if message_text != "":
                new_message = {"user": "User", "text": message_text} # In a real app, user would come from a session
                
                # Add the new message to the room's history.
                chat_rooms[room_id]["messages"].append(new_message)
                
                # Broadcast the new message to everyone in the room.
                await broadcast_message(room_id, new_message)

    except httpserver.WebSocketClosedError as e:
        print(format_str("Connection closed for room '{room_id}': {e}"))
    
    # finally:
    # This block is crucial! It ensures the connection is removed from our list
    # when the client disconnects, preventing memory leaks and errors.
    if ws in chat_rooms[room_id]["connections"]:
        chat_rooms[room_id]["connections"].remove(ws)
    print(format_str("Cleaned up connection for room '{room_id}'. Total connections: {len(chat_rooms[room_id]['connections'])}"))


# The HTTP handler to serve the main chat page.
def chat_page_handler(request):
    # Extract room ID from a URL like /chat/general
    path_parts = request.url.split("/")
    room_id = path_parts[-1] if len(path_parts) > 1 and path_parts[-1] != "" else "general"

    # In a real app, you would check if the room is valid.
    # Here, we just pass the name to the template.

    template_context = {
        "room_id": room_id,
        "available_rooms": chat_rooms.keys()
    }
    # template.render_file returns a Pylearn String, which the server can send.
    # return template.render_file("chat_page.html", template_context)
    return template.render_file("chat_page.html", template_context, "pylearn_custom", "./examples/templates")

rooms = {"available_rooms": chat_rooms.keys()}
def root_handler(request):
    return template.render_file("home.html", rooms, "pylearn_custom", "./examples/templates")

# --- Main Server Setup ---
def main():
    # Import the template module which is now needed by our handler
    
    routes = {
        "/": root_handler,
        "/chat/*": chat_page_handler,      # Serves the HTML page for any room
        "/chat_ws/*": websocket_chat_handler # The WebSocket endpoint
    }
    
    address = "127.0.0.1:5173"
    print(format_str("Starting Pylearn Chat Server on http://{address}/chat/general"))
    
    httpserver.serve(address, routes)
    
    # Keep the script running so the server stays alive.
    print("Server is running. Press Ctrl+C to stop.")
    while True:
        time.sleep(3600)

# Run the server
main()