# routes/__page.py (Pylearn)

def get(request):
    # request object will have .url, .method, .headers (a Dict), .body (String/Bytes/None)
    print("Home page logic (__page.py) called for URL:", request.url)
    
    user_agent_header = request.headers.get(String("user-agent")) # Pylearn Dict.get
    user_agent = "Unknown"
    if user_agent_header and isinstance(user_agent_header, String):
        user_agent = user_agent_header.value

    return {
        "title": "Pylearn Framework Home",
        "message": "Welcome to your Pylearn-powered website!",
        "user_agent": user_agent,
        "current_path": request.url
    }