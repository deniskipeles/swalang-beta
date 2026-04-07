# examples/flasky_app.py
# Import our new framework and the http server library
import time
import flasky
import httpserver

# 1. Create the application object
app = flasky.App()

# The desired API in Pylearn
@app.route("/data", methods=["GET", "POST"])
def handle_data(request):
    if request.method == "POST":
        # Handle posted data
        return {"status": "created", "data": request.body}
    else:
        # Handle getting data
        return {"some": "data"}

# 2. Define a route for the root URL "/" using a decorator
@app.route("/",)
def index(request):
    # 3. Handlers receive a request object and return a response
    # We can return a simple string, and the framework will wrap it.
    return "<h1>Hello, World!</h1>"

# Define another route with a dynamic part
@app.route("/hello/<name>")
def hello_name(request,name):
    # The framework will automatically pass the <name> part as an argument
    return f("""
    <h1>Hello: {name}!</h1>
    <h2>Age: {age}</h2>
    """)
# Define another route with a dynamic part
@app.route("/hello/<name>/<age>")
def hello_name_age(request,name, age):
    # The framework will automatically pass the <name> part as an argument
    return f("""
    <h1>Hello: {name}!</h1>
    <h2>Age: {age}</h2>
    """)

@app.route("/json")
def json_value(request):
    return {"foo": "bar"}

# An async route example
@app.route("/async_stuff")
async def async_stuff(request):
    # You can use async features inside your handlers
    return "This was handled asynchronously!"


# Test 1: Simple string response (should work)
@app.route("/test/string")
def test_string(request):
    return "Hello, World!"

# Test 2: Simple dictionary (the failing case)
@app.route("/test/dict")
def test_dict(request):
    return {"foo": "bar"}

# Test 3: More complex dictionary
@app.route("/test/complex")
def test_complex(request):
    return {
        "name": "test",
        "age": 25,
        "active": True,
        "items": ["apple", "banana", "cherry"]
    }

# Test 4: Empty dictionary
@app.route("/test/empty")
def test_empty(request):
    return {}

# Test 5: List response
@app.route("/test/list")
def test_list(request):
    return ["apple", "banana", "cherry"]

# Test 6: Debug what we actually get
@app.route("/test/debug")
def test_debug(request):
    # Let's see what type of object we're creating
    test_dict = {"foo": "bar"}
    print(f("Type of test_dict: {type(test_dict)}"))
    print(f("test_dict: {test_dict}"))
    return test_dict

# 4. Run the app
# The app object itself will be a callable that the http server can use.
while True:
    print("Starting Flasky server at http://127.0.0.1:5173")
    print("Starting debug server on http://127.0.0.1:8080")
    print("Test these URLs:")
    print("  http://127.0.0.1:8080/test/string")
    print("  http://127.0.0.1:8080/test/dict")
    print("  http://127.0.0.1:8080/test/complex")
    print("  http://127.0.0.1:8080/test/empty")
    print("  http://127.0.0.1:8080/test/list")
    print("  http://127.0.0.1:8080/test/debug")
    httpserver.serve("127.0.0.1:5173", app)
    time.sleep(3600)