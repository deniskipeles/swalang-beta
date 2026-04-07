# examples/chat_app.py
import httpserver
import template
import time
import json

# ==================== DATABASE (Using dictionaries, lists, sets) ====================

# Chat rooms database
chat_rooms = {
    "general": {
        "id": "general",
        "name": "General Chat",
        "description": "General discussion room",
        "created_at": time.time(),
        "messages": []
    },
    "tech": {
        "id": "tech",
        "name": "Tech Talk",
        "description": "Technology discussions",
        "created_at": time.time(),
        "messages": []
    },
    "random": {
        "id": "random",
        "name": "Random",
        "description": "Random conversations",
        "created_at": time.time(),
        "messages": []
    }
}

# Users database (set for online users, dict for user data)
online_users = set()
user_data = {}

# Message counter for unique IDs
message_counter = {"count": 0}

# ==================== HELPER FUNCTIONS ====================

def get_next_message_id():
    message_counter["count"] = 1
    return message_counter["count"]

def format_timestamp(timestamp):
    """Convert timestamp to readable format"""
    # Simple time formatting (you could enhance this)
    return format_str("{int(timestamp)}")

def add_message(room_id, username, message_text, message_type="text"):
    """Add a message to a chat room"""
    if room_id not in chat_rooms:
        return False
    
    message = {
        "id": get_next_message_id(),
        "username": username,
        "text": message_text,
        "type": message_type,
        "timestamp": time.time(),
        "formatted_time": format_timestamp(time.time())
    }
    
    chat_rooms[room_id]["messages"].append(message)
    
    # Keep only last 100 messages per room
    if len(chat_rooms[room_id]["messages"]) > 100:
        chat_rooms[room_id]["messages"] = chat_rooms[room_id]["messages"][-100:]
    
    return True

def get_room_messages(room_id, limit=50):
    """Get recent messages from a room"""
    if room_id not in chat_rooms:
        return []
    
    messages = chat_rooms[room_id]["messages"]
    return messages[-limit:] if len(messages) > limit else messages

def add_user(username):
    """Add user to online users and user data"""
    online_users.add(username)
    if username not in user_data:
        user_data[username] = {
            "username": username,
            "joined_at": time.time(),
            "last_seen": time.time(),
            "message_count": 0
        }
    user_data[username]["last_seen"] = time.time()

def remove_user(username):
    """Remove user from online users"""
    online_users.discard(username)

def get_online_users_list():
    """Get list of online users"""
    return list(online_users)

# ==================== TEMPLATE DEFINITIONS ====================

# Main chat page template
CHAT_PAGE_TEMPLATE = '''

'''

# ==================== REQUEST HANDLERS ====================

def handle_home(request):
    """Redirect to general chat room"""
    obj = {
        "status_code": 302,
        "headers": {"Location": "/chat/general"},
        "body": ""
    }
    return obj


def handle_chat_room(request):
    """Handle chat room display"""
    # Extract room ID from URL path
    path_parts = request.url.split('/')
    room_id = path_parts[-1] if len(path_parts) > 2 else "general"
    
    # Get username from cookie (simple implementation)
    username = None
    if hasattr(request, 'headers') and request.headers:
        cookie_header = request.headers.get('cookie', '')
        if 'username=' in cookie_header:
            for cookie in cookie_header.split(';'):
                if cookie.strip().startswith('username='):
                    username = cookie.strip().split('=')[1]
                    break
    
    # Ensure room exists
    if room_id not in chat_rooms:
        room_id = "general"
    
    current_room = chat_rooms[room_id]
    
    # Add user to online users if they have a username
    if username:
        add_user(username)
    
    # Get messages for this room
    messages = get_room_messages(room_id)
    
    # Prepare template data
    template_data = {
        "current_room": current_room,
        "rooms": list(chat_rooms.values()),
        "messages": messages,
        "online_users": get_online_users_list(),
        "username": username if username else None,
        "user_data": user_data if username else None
    }
    
    # Render template
    html_content = template.render(CHAT_PAGE_TEMPLATE, template_data, "pylearn_custom")
    html_content = template.render_file("chat.html", template_data, "pylearn_custom", "./examples/templates")
    
    return html_content

def handle_set_username(request):
    """Handle username setting"""
    if request.method == "POST":
        # Parse form data (simple implementation)
        body = request.body if hasattr(request, 'body') and request.body else ""
        username = None
        print(body)
        
        if isinstance(body, str) and 'username' in body:
            for param in body.split('&'):
                if param.startswith('username='):
                    username = param.split('=')[1].replace('%20', ' ')
                    # Simple URL decode for spaces
                    break
        
        if username and len(username.strip()) > 0:
            username = username.strip()[:20]  # Limit length
            
            # Add system message
            add_message("general", "System", format_str("{username} joined the chat!"), "system")
            
            # Set cookie and redirect
            obj = {
                "status_code": 302,
                "headers": {
                    "Location": "/chat/general",
                    "Set-Cookie": format_str("username={username}; Path=/; Max-Age=86400")
                },
                "body": ""
            }
            return obj
    
    # Redirect back if invalid
    obj = {
        "status_code": 302,
        "headers": {"Location": "/chat/general"},
        "body": ""
    }
    return obj


def handle_send_message(request):
    """Handle message sending"""
    if request.method == "POST":
        # Get username from cookie
        username = None
        if hasattr(request, 'headers') and request.headers:
            cookie_header = request.headers.get('cookie', '')
            if 'username=' in cookie_header:
                for cookie in cookie_header.split(';'):
                    if cookie.strip().startswith('username='):
                        username = cookie.strip().split('=')[1]
                        break
        
        if not username:
            obj = {
                "status_code": 302,
                "headers": {"Location": "/chat/general"},
                "body": ""
            }
            return obj
        
        # Parse form data
        body = request.body if hasattr(request, 'body') and request.body else ""
        room_id = "general"
        message_text = ""
        
        if isinstance(body, str):
            for param in body.split('&'):
                if param.startswith('room_id='):
                    room_id = param.split('=')[1]
                elif param.startswith('message='):
                    message_text = param.split('=')[1].replace('%20', ' ').replace('+', ' ')
                    # Simple URL decode
        
        if message_text.strip():
            add_message(room_id, username, message_text.strip())
            
            # Update user message count
            if username in user_data:
                user_data[username]["message_count"] = user_data[username]["message_count"] +1
        
        # Redirect back to room
        obj = {
            "status_code": 302,
            "headers": {"Location": format_str("/chat/{room_id}")},
            "body": ""
        }
        return obj
    
    obj = {
        "status_code": 302,
        "headers": {"Location": "/chat/general"},
        "body": ""
    }
    return obj

def handle_api_messages(request):
    """API endpoint to get messages (for potential AJAX updates)"""
    path_parts = request.url.split('/')
    room_id = path_parts[-1] if len(path_parts) > 3 else "general"
    
    if room_id not in chat_rooms:
        room_id = "general"
    
    messages = get_room_messages(room_id, 20)  # Last 20 messages
    
    # Convert to simple dict for JSON
    messages_data = []
    for msg in messages:
        messages_data.append({
            "id": msg["id"],
            "username": msg["username"],
            "text": msg["text"],
            "type": msg["type"],
            "timestamp": msg["timestamp"]
        })
    
    return {
        "status_code": 200,
        "headers": {"Content-Type": "application/json"},
        "body": messages_data  # This will be JSON serialized
    }

# ==================== ROUTES AND SERVER SETUP ====================

# Add some initial messages
add_message("general", "System", "Welcome to Pylearn Chat! 🚀", "system")
add_message("general", "System", "This chat app is built with Pylearn's template engine and uses dictionaries as database!", "system")
add_message("tech", "System", "Welcome to Tech Talk! Share your programming insights here. 💻", "system")
add_message("random", "System", "Random chat room - talk about anything! 🎲", "system")

# Define routes
routes = {
    "/": handle_home,
    "/chat/general": handle_chat_room,
    "/chat/tech": handle_chat_room,
    "/chat/random": handle_chat_room,
    "/set_username": handle_set_username,
    "/send_message": handle_send_message,
    "/api/messages/general": handle_api_messages,
    "/api/messages/tech": handle_api_messages,
    "/api/messages/random": handle_api_messages,
}

def main():
    address = "127.0.0.1:5173"
    x_val = [len(room['messages']) for room in chat_rooms.values()]
    room_messages = None

    print(format_str("🚀 Starting Pylearn Chat Server on {address}"))
    print(format_str("💬 Visit http://{address} to start chatting!"))
    print("📊 Database Status:")
    print(format_str("   - Chat Rooms: {len(chat_rooms)}"))
    print(format_str("   - Total Messages: {x_val}"))
    print(format_str("   - Online Users: {len(online_users)}"))
    
    httpserver.serve(address, routes)
    
    try:
        input("\n🔥 Chat server is running! Press Enter to stop...\n")
    except EOFError:
        print("🔄 Server running in background...")
        while True:
            time.sleep(3600)

if __name__ == "__main__":
    main()