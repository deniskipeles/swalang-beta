# examples/http_test_app.py

import httpserver
import http # Pylearn's ASYNC http client module
import json # Pylearn's json module
import aio  # Hypothetical Pylearn module for async utilities like gather
import time # Pylearn's time module (assuming it has a cooperative async sleep if needed)

# --- Global Data (Simple In-Memory Store) ---
users_db = {
    "1": {"id": "1", "name": "Alice", "email": "alice@example.com"},
    "2": {"id": "2", "name": "Bob", "email": "bob@example.com"}
}
next_user_id = 3

# --- Handler Functions ---
import template # Assuming template engine is synchronous

# --- Template Setup (remains synchronous for this example) ---
print("--- Testing Go Syntax (Template Setup) ---")
template_file = './examples/templates/index.html'
# In a real async app, file I/O for templates might also be async,
# but template parsing/rendering is often CPU-bound or done at startup.
try:
    template_data_engine = template.new(template_file, "pylearn_custom")
    html_file = open(template_file, 'r')   # Assuming Pylearn's open supports with
    template_data_engine.parse(html_file.read())
except Exception as e:
    print(format_str("Error loading template: {e}"))
    template_data_engine = None
# --- End Template Setup ---

async def handle_root(request):
    print(format_str("Async Handler: handle_root called for URL: {request.url}"))
    
    # Example: Fetch some external data asynchronously to include in the page
    external_todo = {}
    try:
        # http.get is now async and needs to be awaited
        todo_response = http.get("https://jsonplaceholder.typicode.com/todos/1")
        print(todo_response)
        print(todo_response.status_code)
        if todo_response.status_code == 200:
            external_todo = todo_response.json() # .json() is sync after response is received
        else:
            print(format_str("Failed to fetch external_todo: {todo_response.status_code}"))
    except Exception as e:
        print(format_str("Error fetching external todo:"))

    # Prepare template data
    current_users_list = []
    for user_id_str in users_db.keys():
        current_users_list.append(users_db.get(user_id_str))
    
    etg = "N/A"
    if external_todo:
        etg = external_todo.get("title")
        
    template_context = {
        "users": current_users_list, # Pylearn List of Pylearn Dicts
        "external_todo_title": etg
    }
    
    # Assuming template_data_engine.execute is synchronous
    if template_data_engine:
        html_output = template_data_engine.execute(template_context)
    else:
        html_output = "<h1>Template Error</h1>"

    # Build the rest of the body
    body_parts = [
        html_output, # This is already a Pylearn String
        format_str("""
        <hr>
        <h2>Welcome to the ASYNC Pylearn HTTP Server Test!</h2>
        <p>Request Method: {request.method}</p>
        <p>Request URL: {request.url}</p>
        <p>Request Headers:</p>
        <ul>
        """)
    ]

    if request.headers:
        for key in request.headers.keys():
             body_parts.append(format_str("<li><b>{key}:</b> {request.headers.get(key)}</li>"))
    body_parts.append("</ul>")
    
    body_parts.append("""
    <p>Try these links:</p>
    <ul>
        <li><a href="/hello_async">/hello_async</a></li>
        <li><a href="/data_async">/data_async (returns JSON)</a></li>
        <li><a href="/users_async">/users_async (GET all users)</a></li>
        <li><a href="/users_async/1">/users_async/1 (GET user by ID)</a></li>
        <li><a href="/concurrent_fetch">/concurrent_fetch</a></li>
        <li>POST to /users_async with JSON like {"name": "Charlie", "email": "charlie@example.com"} to add a user.</li>
    </ul>
    """)
    return "".join(body_parts) # Concatenate list of strings


async def handle_hello_async(request):
    print("Async Handler: handle_hello_async called")
    # Simulate some async work, e.g., a non-blocking sleep
    if hasattr(aio, "sleep"): # Check if our hypothetical async sleep exists
        await aio.sleep(0.1) # Cooperative sleep for 100ms
    
    user_agent = "Unknown"
    if request.headers and request.headers.get("user-agent"):
        user_agent = request.headers.get("user-agent")
    return format_str("Hello from ASYNC Pylearn! Your User-Agent is: {user_agent}")

async def fetch_url_task(url):
    print(format_str("Fetching {url}..."))
    response =  http.get(url) # Assuming http.get is async
    if response and response.status_code == 200:
        # print(format_str("Successfully fetched {url}, content length: {len(response.content)}"))
        return response
    else:
        status = "N/A"
        if response: 
            status = response.status_code
        print(format_str("Failed to fetch {url}, status: {status}"))
        return {"url": url, "error": format_str("Failed with status {status}")}

async def handle_data_async(request):
    print("Async Handler: handle_data_async called")
    # Simulate fetching data from two sources concurrently
    url1 = "https://jsonplaceholder.typicode.com/todos/3"
    url2 = "https://jsonplaceholder.typicode.com/todos/4"

    try:
        # These calls to http.get() return "awaitables" (AsyncResultWrappers)
        todo3_promise = fetch_url_task(url1)
        todo4_promise = fetch_url_task(url2)

        # Use aio.gather (or equivalent) to run them concurrently
        # results will be a list: [response_for_todo3, response_for_todo4]
        responses = await aio.gather(todo3_promise, todo4_promise)
        print(responses)
        # return json.dumps(responses)
        todo3_data = None
        todo4_data = None
        http_results = []
        for response in responses:
            if response and response.status_code == 200:
                print(format_str("- {response}")) # These will be Pylearn Dicts
                res = response.json()
                http_results.append(res)
            else:
                pass


        # print(type(responses))
        # print(type(responses[0]))

        # if responses[0] and responses[0].status_code == 200:
        #     todo3_data = responses[0].json()
        # if responses[1] and responses[1].status_code == 200:
        #     todo4_data = responses[1].json()
            
    except Exception as e:
        print(format_str("Error in handle_data_async fetching: {e}"))
        return {"error": str(1111), "message": "Failed to fetch concurrent data"}

    data_dict = {
        "message": "This is ASYNC JSON data from Pylearn!",
        "status": "success",
        "concurrent_items": http_results,
        "items": [1, "two", True, None, {"nested": "data"}]
    }
    return data_dict # Pylearn Dict, server will JSONify

# Synchronous versions (can coexist or be replaced)
def handle_get_users(request): # Renamed from original
    print("Sync Handler: handle_get_users called")
    user_list = list()
    for user_id_str in users_db.keys():
        user_list.append(users_db.get(user_id_str))
    return json.dumps(user_list)

async def handle_get_users_async(request):
    print("Async Handler: handle_get_users_async called")
    # This specific handler doesn't do async I/O, but it can be async.
    # For demonstration, let's add a small cooperative sleep.
    if hasattr(aio, "sleep"): 
        await aio.sleep(0.01)
    
    user_list = []
    for user_id_str in users_db.keys():
        user_list.append(users_db.get(user_id_str))
    # Server will JSON-serialize a Pylearn List of Dicts automatically if returned directly
    return user_list

async def handle_get_user_by_id_async(request):
    print(format_str("Async Handler: handle_get_user_by_id_async called for URL: {request.url}"))
    path_parts = request.url.split("/")
    user_id = None
    if len(path_parts) >= 3:
        user_id = path_parts[2]
    
    # Simulate an async database lookup if this were a real DB
    # For now, db access is sync
    if hasattr(aio, "sleep"): 
        await aio.sleep(0.02) 

    if user_id and users_db.get(user_id):
        return users_db.get(user_id) # Pylearn Dict
    else:
        obj = {"error": "User not found", "user_id": user_id}
        # Using a hypothetical helper for cleaner response
        return httpserver.json_response( obj, status_code=404)

async def handle_create_user_async(request):
    global next_user_id
    print("Async Handler: handle_create_user_async called")
    if request.method != "POST":
        return object.ServerResponse(body={"error": "Method Not Allowed"}, status_code=405, headers={"Allow": "POST"})
    
    # Request body parsing is typically synchronous once the body is received
    try:
        request_data_str = ""
        # In a true async server, request.body might be an awaitable stream.
        # For now, assume Pylearn's http.Request provides body as string/bytes.
        if isinstance(request.body, str):
            request_data_str = request.body
        elif isinstance(request.body, bytes):
            request_data_str = request.body.decode("utf-8")
        elif isinstance(request.body, dict): # If server pre-parses JSON
             data = request.body
        else:
            if not request_data_str: # Should check before isinstance(dict)
                 return object.ServerResponse(body={"error": "Request body is empty or not a string/bytes/dict"}, status_code=400)
            data = json.loads(request_data_str)
        
        if not isinstance(request.body, dict): # If it wasn't already a dict
            data = json.loads(request_data_str)

    except Exception as e: # Pylearn's generic exception
        print(format_str("Error decoding JSON from async request body: {e}"))
        return object.ServerResponse(body={"error": "Invalid JSON data"}, status_code=400)

    name = data.get("name")
    email = data.get("email")

    if not name or not email:
        return object.ServerResponse(body={"error": "Missing name or email"}, status_code=400)

    # Simulate async database write
    if hasattr(aio, "sleep"): 
        await aio.sleep(0.05) 
    
    new_id_str = str(next_user_id)
    users_db[new_id_str] = {"id": new_id_str, "name": name, "email": email}
    next_user_id = next_user_id + 1
    
    return object.ServerResponse(
        body=users_db[new_id_str], 
        status_code=201,
        headers={"Location": format_str("/users_async/{new_id_str}")}
    )

async def handle_users_router_async(req):
    if req.method == "GET": 
        return {'users': await handle_get_users_async(req)} # Await if the sub-handler is async
    elif req.method == "POST":
        return {'user created':await handle_create_user_async(req)} # Await if the sub-handler is async
    else:
        return {"body":"Method Not Allowed", "status_code":405}

async def handle_concurrent_fetch(request):
    print("Async Handler: handle_concurrent_fetch called")
    urls = [
        "https://jsonplaceholder.typicode.com/todos/5",
        "https://jsonplaceholder.typicode.com/todos/6",
        "https://jsonplaceholder.typicode.com/todos/7"
    ]
    
    tasks = []
    for url in urls:
        tasks.append(await http.get(url)) # http.get returns an awaitable/promise

    # await aio.gather to run them concurrently
    responses = await aio.gather(tasks) # * unpacks list into arguments for gather
    

    results = []
    for i, resp in enumerate(responses):
        if resp and resp.status_code == 200:
            r = resp.json()
            results.append(r)
        else:
            ress = "N/A"
            if resp:
                ress = resp.status_code
            results.append({"error": format_str("Failed to fetch {urls[i]}"), "status": ress})
            
    return {"source": "concurrent_fetch", "results": results}

# --- Routes Definition ---
routes = {
    "/": handle_root, # Can be async
    "/data": handle_hello_async,
    "/data_async": handle_data_async,
    "/users_async": handle_users_router_async, # This router is async, calls other async handlers
    "/users_async/1": handle_get_user_by_id_async, # Example specific async route
    "/users_async/2": handle_get_user_by_id_async,
    "/users_async/nonexistent": handle_get_user_by_id_async,
    "/concurrent_fetch": handle_concurrent_fetch,

    # Keep sync routes for comparison or if some parts are not async
    "/users_sync": handle_get_users,
}

# --- Main Server Start (remains synchronous from Pylearn script's perspective) ---
def main():
    address = "127.0.0.1:5173"
    print(format_str("Attempting to start Pylearn server (with async handler support) on {address}"))
    
    httpserver.serve(address, routes) # httpserver.serve's Go impl needs to handle async handlers
    
    print(format_str("Pylearn server setup initiated. Check your browser/curl at http://{address}"))
    input("Server running (potentially with async handlers). Press Enter to stop script (server might continue in background if Go part is detached)...")
    # In a real scenario, the Go server runs until interrupted (Ctrl+C).
    # The input() here is just for this script example to pause.

# Assume `aio.gather` and `aio.sleep` are provided by Pylearn's async stdlib.
# If not, these parts would need to be adapted or would run sequentially despite `async def`.
if __name__ == "__main__":
    main()





# # examples/http_test_app.py

# import httpserver
# import http # Pylearn's http client module
# import json # Pylearn's json module

# # --- Global Data (Simple In-Memory Store) ---
# users_db = {
#     "1": {"id": "1", "name": "Alice", "email": "alice@example.com"},
#     "2": {"id": "2", "name": "Bob", "email": "bob@example.com"}
# }
# next_user_id = 3

# # --- Handler Functions ---
# import template

# # --- Test Go Syntax (default) ---
# print("--- Testing Go Syntax ---")
# tempate_file = './examples/templates/index.html'
# template_data = template.new(tempate_file,"pylearn_custom")
# # template_data = open(tempate_file)
# html_file = open(tempate_file,'r')  
# template_data.parse(html_file.read())
# # For this test, let's directly construct the list of Pylearn Dicts
# user_list = list()
# for user_id_str in users_db.keys(): # users_db is a Python dict here, not Pylearn
#     user_list.append(users_db.get(user_id_str))
# template_data = template_data.execute({"users":users_db})
# def handle_root(request):
#     print("Handler: handle_root called for URL:", request.url)
#     print('request>>>',request)
#     body = """
#     <h1>Welcome to the Pylearn HTTP Server Test!</h1>
#     <p>Request Method: {}</p>
#     <p>Request URL: {}</p>
#     <p>Request Headers:</p>
#     <ul>
#     """.format(request.method, request.url)

#     if request.headers:
#         for key in request.headers.keys(): # Assuming headers is a Dict with keys()
#              body = body + "<li><b>{}:</b> {}</li>".format(key, request.headers.get(key))
#     body = body + "</ul>"
    
#     body = body + """
#     <p>Try these links:</p>
#     <ul>
#         <li><a href="/hello">/hello</a></li>
#         <li><a href="/data">/data (returns JSON)</a></li>
#         <li><a href="/users">/users (GET all users)</a></li>
#         <li><a href="/users/1">/users/1 (GET user by ID)</a></li>
#         <li>POST to /users with JSON like {"name": "Charlie", "email": "charlie@example.com"} to add a user.</li>
#     </ul>
#     """
#     # Return a simple string (HTML)
#     return template_data


# def handle_hello(request):
#     print("Handler: handle_hello called")
#     user_agent = "Unknown"
#     if request.headers and request.headers.get("user-agent"):
#         user_agent = request.headers.get("user-agent")

#     return "Hello from Pylearn! Your User-Agent is: {}".format(user_agent)

# def handle_data(request):
#     print("Handler: handle_data called")
#     # Return a Pylearn dictionary, which the server will try to JSON-serialize
#     data_dict = {
#         "message": "This is JSON data from Pylearn!",
#         "status": "success",
#         "items": [1, "two", True, None, {"nested": "data"}]
#     }
#     return data_dict

# def handle_get_users(request):
#     print("Handler: handle_get_users called")
#     # Convert users_db values to a list for JSON response
#     user_list = list()
#     # Assuming users_db.values() gives a list of the dictionary's values
#     # If not, iterate through keys and get values.
#     # For Pylearn Dict, we might need to iterate keys and build the list.
    
#     # Let's assume users_db's values are directly listable for simplicity for now
#     # A proper Pylearn Dict would need a .values() method returning a list
#     # temp_user_values = []
#     # for k in users_db.keys(): # Assuming users_db is a Pylearn Dict
#     #    temp_user_values.append(users_db.get(k))
#     # return temp_user_values # Return list of Pylearn Dicts

#     # For this test, let's directly construct the list of Pylearn Dicts
#     for user_id_str in users_db.keys(): # users_db is a Python dict here, not Pylearn
#         user_list.append(users_db.get(user_id_str))
#     return json.dumps(user_list) # Server should convert list of dicts to JSON array of objects

# def handle_get_user_by_id(request):
#     print("Handler: handle_get_user_by_id called for URL:", request.url)
#     # Basic path parsing - a real router would handle this better
#     path_parts = request.url.split("/") # Assumes String.split() method exists
#     user_id = None
#     if len(path_parts) >= 3: # e.g. /users/1
#         user_id = path_parts[2]
    
#     if user_id and users_db.get(user_id):
#         return users_db.get(user_id)
#     else:
#         # Use the ServerResponse object to return a 404
#         # return object.ServerResponse(
#         #     body={"error": "User not found", "user_id": user_id},
#         #     status_code=404,
#         #     content_type="application/json" 
#         #     # Headers could be added here too: headers={"X-Custom": "value"}
#         # )
#         return object.ServerResponse(body={"error": "User not found", "user_id": user_id}, status_code=404, content_type="application/json")

# def handle_create_user(request):
#     global next_user_id # To modify the global
#     print("Handler: handle_create_user called")
#     if request.method != "POST":
#         # return object.ServerResponse(
#         #     body={"error": "Method Not Allowed"},
#         #     status_code=405,
#         #     headers={"Allow": "POST"}
#         # )
#         return object.ServerResponse(body={"error": "Method Not Allowed"}, status_code=405, headers={"Allow": "POST"})
    
#     try:
#         # Assume request.body is a Pylearn String if content type was JSON
#         # and the server pre-parses it, or we parse it here.
#         # For now, let's assume request.body might be String or Bytes.
#         request_data_str = ""
#         if isinstance(request.body, str): # Pylearn str
#             request_data_str = request.body
#         elif isinstance(request.body, bytes): # Pylearn bytes
#             request_data_str = request.body.decode("utf-8") # Assuming decode method
#         else:
#             # If it's already a Dict (if server pre-parses JSON for us), great!
#             if isinstance(request.body, dict): # Pylearn dict
#                 data = request.body
#             else: # Try to parse if it's a string.
#                 if not request_data_str:
#                      return object.ServerResponse(body={"error": "Request body is empty or not a string/bytes"}, status_code=400)
#                 data = json.loads(request_data_str) # Use Pylearn's json.loads

#     # except: # Generic catch, Pylearn would need its own error types
#     except: # Generic catch, Pylearn would need its own error types
#         print("Error decoding JSON from request body:", e)
#         return object.ServerResponse(body={"error": "Invalid JSON data"}, status_code=400)

#     name = data.get("name")
#     email = data.get("email")

#     if not name or not email:
#         return object.ServerResponse(body={"error": "Missing name or email"}, status_code=400)

#     new_id_str = str(next_user_id)
#     users_db[new_id_str] = {"id": new_id_str, "name": name, "email": email}
#     next_user_id = next_user_id + 1
    
#     # return object.ServerResponse(
#     #     body=users_db[new_id_str], 
#     #     status_code=201, # Created
#     #     headers={"Location": "/users/" + new_id_str}
#     # )
#     return object.ServerResponse(body=users_db[new_id_str], status_code=201, headers={"Location": "/users/" + new_id_str})

# # --- Routes Definition ---
# # This will be a Pylearn Dictionary
# def handle_request(req):
#     """
#     Handles HTTP requests based on the request method.

#     Args:
#         req (object): The HTTP request object.

#     Returns:
#         object: The server response object.
#     """
#     if req.method == "GET": 
#         return handle_get_users(req) 
#     elif req.method == "POST":
#         return handle_create_user(req)
#     else:
#         return object.ServerResponse(body="Method Not Allowed", status_code=405)
    
# routes = {
#     "/": handle_root,
#     "/hello": handle_hello,
#     "/data": handle_data,
#     "/users": handle_request,
#     # A more advanced router would handle /users/{id}
#     # For now, we'll use a simple prefix match or specific paths
#     "/users/1": handle_get_user_by_id, # Specific path for testing
#     "/users/2": handle_get_user_by_id,
#     "/users/nonexistent": handle_get_user_by_id, # Test 404
# }

# # --- Main Server Start ---
# def main():
#     address = "127.0.0.1:5173"
#     print("Attempting to start Pylearn server on " + address)
    
#     # This is where we would register Pylearn's object.ServerResponse type
#     # if the server needed to know about it explicitly for type checking.
#     # For now, the server's Go code will check the type of the returned object.
    
#     # Start the server (this function in Pylearn will likely be a blocking call
#     # or start a background process/goroutine in Go)
#     # The `serve` function will run indefinitely in Go.
#     # Pylearn's `httpserver.serve` should return None immediately after starting the server.
#     httpserver.serve(address, routes)
    
#     print("Pylearn server setup initiated. Check your browser/curl at http://" + address)
#     print("The Pylearn script will continue if serve() is non-blocking, or hang here if it's blocking.")
#     print("Typically, a web server runs until interrupted (Ctrl+C).")
#     input("Waiting...") # This would block here

#     # --- Test HTTP Client (after server starts, if non-blocking) ---
#     # Note: If serve() is blocking, these client tests won't run until server stops.
#     # For a real test, client calls should be in a separate script or thread.
#     # For this example, we assume serve() might be non-blocking or we'd run client later.
    
#     # Add a small delay if `serve` is non-blocking to give the server time to start
#     # This is a hack for testing; a real app wouldn't do this.
#     # import time # Pylearn would need a time module
#     # time.sleep(1) # Pylearn time.sleep(seconds_float_or_int)

#     print("\n--- Testing HTTP Client ---")
#     base_url = "http://" + address
    
#     # Test GET
#     try:
#         print("Client: GET /hello")
#         response_hello = http.get(base_url + "/hello", headers={"X-Pylearn-Test": "true"})
#         print("Status:", response_hello.status_code)
#         print("Reason:", response_hello.reason)
#         print("Headers (Content-Type):", response_hello.headers.get("content-type"))
#         print("Text:", response_hello.text)
#         print("Content (bytes):", response_hello.content) # Should be bytes
#         response_hello.raise_for_status() # Should not raise error for 2xx
#     # except:
#     except:
#         print("Client Error (GET /hello): error<<<<<<<")

#     # Test GET JSON
#     try:
#         print("\nClient: GET /data")
#         response_data = http.get(base_url + "/data")
#         print("Status:", response_data.status_code)
#         json_data = response_data.json() # Assumes response_data.json() exists
#         print("JSON Data:", json_data)
#         if json_data and json_data.get("message"):
#             print("JSON Message:", json_data.get("message"))
#     # except:
#     except:
#         print("Client Error (GET /data): error<<<<<<<")

#     # Test POST JSON
#     try:
#         print("\nClient: POST /users (create Charlie)")
#         user_payload = {"name": "Charlie", "email": "charlie@example.com"}
#         # http.post expects options dict
#         post_options = {
#             "json": user_payload, # Send Pylearn dict, pyhttp will convert
#             "headers": {"X-Client": "PylearnHTTPTest"}
#         }
#         response_post = http.post(base_url + "/users", post_options)
#         print("Status:", response_post.status_code)
#         print("Location Header:", response_post.headers.get("location"))
#         created_user = response_post.json()
#         print("Created User (JSON):", created_user)
#         if created_user and created_user.get("id"):
#             print("New User ID:", created_user.get("id"))
#     # except:
#     except:
#         print("Client Error (POST /users): ERROR<<<<<<<")

#     # Test GET all users after POST
#     try:
#         print("\nClient: GET /users (after POST)")
#         response_users_after = http.get(base_url + "/users")
#         print("Status:", response_users_after.status_code)
#         all_users = response_users_after.json()
#         print("All Users (JSON):", all_users)
#     # except:
#     except:
#         print("Client Error (GET /users after POST): ERROR<<<<<")

#     # Test 404 with ServerResponse
#     try:
#         print("\nClient: GET /users/nonexistent")
#         response_404 = http.get(base_url + "/users/nonexistent")
#         print("Status for 404:", response_404.status_code)
#         print("404 Text:", response_404.text)
#         # response_404.raise_for_status() # This should raise an error
#     # except: # Catch the expected HTTPError from raise_for_status
#     except: # Catch the expected HTTPError from raise_for_status
#         print("Client Error (GET /users/nonexistent - as expected from raise_for_status): ERROR<<<<<<<")


# # To run this script:
# # 1. Pylearn interpreter must be working.
# # 2. The `httpserver` and `http` Pylearn modules must be implemented and registered.
# # 3. Pylearn's `json` module (loads, dumps) should be available.
# # 4. Pylearn object methods like `String.split()`, `Dict.keys()`, `Dict.get()`, `Bytes.decode()` need to exist.
# # 5. `object.ServerResponse` needs to be a known type by the Pylearn `httpserver` if it's used directly as a return.
# #    (The current Go code for the server checks the type of the return value.)
# # 6. Pylearn needs `isinstance(obj, type)` built-in.
# # 7. Pylearn needs basic `try...except` handling.

# # if __name__ == "__main__": # Pylearn would need __name__ == "__main__" support
# main()