# app.py (Pylearn)

import httpserver
import pylearn_importlib
import os
path = os.path
# import object # Access the submodule
import template
import json # For default error responses

# --- Configuration ---
ROUTES_DIR = "./routes" # Relative to where app.py is run from
TEMPLATES_COMMON_DIR = "./templates" # For error pages etc.
DEFAULT_SYNTAX = "pylearn_custom" # or "go"

# --- Global Store for Discovered Routes ---
# Structure: { "/path": {"py": "abs/path/to/__page.py", "html": "abs/path/to/__page.html", "route_dir": "abs/path/to/route_dir"} }
discovered_routes = dict()

# --- Helper Functions ---

"""
Recursively walks the routes directory to find __page.py and __page.html files.
current_dir_abs: Absolute path to the directory currently being scanned.
current_path_segment: The URL path segment corresponding to current_dir_abs (e.g., "/", "/services/").
base_routes_dir_abs: Absolute path to the root 'routes' directory.
"""
def find_routes_recursive(current_dir_abs, current_path_segment, base_routes_dir_abs):
    discovered_routes = dict()
    # print("Scanning directory:", current_dir_abs, "for path segment:", current_path_segment)

    try:
        entries = os.listdir(current_dir_abs)
    # except Exception as e:
    except: # Pylearn might use a generic Exception
        print("Error listing directory", current_dir_abs, ":", "ERROR<<<<<<")
        return None

    py_file = os.path.join(current_dir_abs, "__page.py")
    html_file = os.path.join(current_dir_abs, "__page.html")

    if os.path.exists(py_file) and os.path.exists(html_file):
        # Ensure current_path_segment ends with a slash if it's not just "/"
        # And normalize: "/" for root, "/path/" for others
        url_path = current_path_segment
        if not url_path.endswith("/"):
            url_path = url_path + "/"
        if url_path == "//": # From joining "/" and "/"
             url_path = "/"
        
        # For root, path is just "/"
        if current_dir_abs == base_routes_dir_abs:
            url_path = "/"
        
        # Normalize double slashes from join, except for root
        temp_url_path = url_path.replace("//", "/")
        if temp_url_path != "/" and not temp_url_path.endswith("/"):
            temp_url_path = temp_url_path + "/"
        if temp_url_path == "//": 
            temp_url_path = "/"

        # A simpler way for URL path generation based on relative structure
        rel_dir_path = os.path.relpath(current_dir_abs, base_routes_dir_abs)
        if rel_dir_path == ".": # Root of routes
            url_path = "/"
        else:
            url_path = "/" + rel_dir_path.replace("\\", "/") + "/" # Ensure slashes and trailing
            if url_path == "//": 
                url_path = "/" # Normalize for root case if relpath was "."

        print("Found route:", url_path, "-> py:", py_file, "html:", html_file)
        discovered_routes[url_path] = {
            "py_path": os.path.abspath(py_file),
            "html_path": os.path.abspath(html_file),
            "route_dir": current_dir_abs # Absolute path to the route's own directory
        }

    for entry_name in entries:
        entry_path_abs = os.path.join(current_dir_abs, entry_name)
        if os.path.isdir(entry_path_abs):
            # Construct next path segment
            next_path_segment = current_path_segment
            if not next_path_segment.endswith("/"):
                next_path_segment = next_path_segment + "/"
            next_path_segment = next_path_segment + entry_name
            find_routes_recursive(entry_path_abs, next_path_segment, base_routes_dir_abs)

def find_nearest_layout(route_dir_abs, base_routes_dir_abs):
    """
    Looks for __layout.html starting from route_dir_abs and going up to base_routes_dir_abs.
    Returns the absolute path to the layout, or None.
    """
    current_check_dir = route_dir_abs
    # print("Finding layout for route_dir:", route_dir_abs, "base_routes:", base_routes_dir_abs)
    while True:
        layout_path = os.path.join(current_check_dir, "__layout.html")
        # print("Checking for layout at:", layout_path)
        if os.path.exists(layout_path):
            return os.path.abspath(layout_path)
        
        # Stop if we've reached the parent of base_routes_dir_abs or an invalid path
        if current_check_dir == base_routes_dir_abs or os.path.dirname(current_check_dir) == current_check_dir:
            break
        current_check_dir = os.path.dirname(current_check_dir)
    return None

def render_error_page(status_code, error_message, base_template_dir):
    """Renders a generic error page or a specific one if found."""
    print("Rendering error page:", status_code, error_message)
    title = "Error"
    message = error_message
    if status_code == 404:
        title = "Page Not Found"
        message = "The page you requested could not be found."
    elif status_code == 500:
        title = "Server Error"
        # In production, you might not want to show the detailed error_message
        # message = "An internal server error occurred."

    error_page_path = os.path.join(base_template_dir, "errors", str(status_code) + ".html")
    context = {"title": title, "error_code": status_code, "message": message}
    html_content = ""
    
    try:
        if os.path.exists(error_page_path):
            # Use default 'go' syntax for error pages for simplicity, or pass DEFAULT_SYNTAX
            # Loader needs to know where common templates are.
            # Create a temporary loader for common templates.
            tmpl = template.new("error_page_render", DEFAULT_SYNTAX, base_template_dir)
            tmpl.load(os.path.join("errors", str(status_code) + ".html")) # Relative to TEMPLATES_COMMON_DIR/errors
            html_content = tmpl.execute(context)
        else:
            # Basic fallback HTML
            html_content = "<h1>{}</h1><p>{}</p><p><i>Detailed error: {}</i></p>".format(title, message, error_message)
    # except Exception as e:
    except:
        print("Critical error rendering error page:", "ERROR<<<<<<")
        html_content = "<h1>Critical Server Error</h1><p>Failed to render error page.</p>"
        status_code = 500 # Ensure status is 500 if error page itself fails

    # return object.ServerResponse(
    #     body=html_content,
    #     status_code=status_code,
    #     content_type="text/html; charset=utf-8"
    # )
    # return httpserver.text_response(body=html_content, status_code=status_code, content_type="text/html; charset=utf-8")
    return httpserver.text_response(html_content, "text/html; charset=utf-8")

def handle_request_internal(request):
    """Internal request handler, before middleware."""
    print("Request received for URL:", request.url, "Method:", request.method)
    
    # Normalize request path for matching (ensure trailing slash, unless root)
    req_path = request.url
    if not req_path.endswith("/") and req_path != "/":
        req_path = req_path + "/"
    if req_path == "//": 
        req_path = "/" # Handle potential double slash if URL was "/"

    route_info = discovered_routes.get(req_path)

    if not route_info:
        return render_error_page(404, "Route {} not found.".format(request.url), TEMPLATES_COMMON_DIR)

    py_path = route_info.get("py_path")
    html_path_short = os.path.basename(route_info.get("html_path")) # e.g., "__page.html"
    route_dir = route_info.get("route_dir")

    try:
        # 1. Load and execute __page.py
        page_module = pylearn_importlib.load_module_from_path(py_path)
        if not page_module or isinstance(page_module, object.Error): 
            # Check for Pylearn error object
            pm = None
            if page_module:
                pm = page_module.inspect()
            print("Failed to load page module:", py_path, pm )
            return render_error_page(500, "Error loading page logic.", TEMPLATES_COMMON_DIR)

        page_get_func = page_module.Env.get("get") # Assumes module.Env access, or use getattr
        if not page_get_func: # or not is_callable(page_get_func)
            print("No 'get' function in page module:", py_path)
            return render_error_page(500, "'get' function not found in page logic.", TEMPLATES_COMMON_DIR)

        # Execute the get function
        # Context for executing 'get' is its own module's env, but builtins should be available.
        # This depends on how ExecutionContext and Eval handle builtins for dynamic modules.
        # For now, assume 'request' object is passed correctly.
        page_context = page_get_func(request) # Pylearn call
        if isinstance(page_context, object.Error): # Check if 'get' returned an error
             print("Error from page_module.get():", page_context.inspect())
             return render_error_page(500, "Error in page logic execution: " + page_context.message, TEMPLATES_COMMON_DIR)
        if not isinstance(page_context, object.Dict):
             print("page_module.get() did not return a Dict, got:", page_context.type())
             return render_error_page(500, "Page logic must return a dictionary context.", TEMPLATES_COMMON_DIR)

        # 2. Find layout
        base_routes_dir = os.path.abspath(ROUTES_DIR)
        layout_abs_path = find_nearest_layout(route_dir, base_routes_dir)

        # 3. Render __page.html
        # The base_path for the template loader should be the route's own directory
        # to allow `{% include "reusableComponent.html" %}` within that route.
        page_tmpl = template.new("page_render", DEFAULT_SYNTAX, route_dir)
        page_tmpl.load(html_path_short) # Load relative to route_dir (implicitly via loader)
        
        page_html_content = page_tmpl.execute(page_context)
        if isinstance(page_html_content, object.Error):
            print("Error rendering page template:", html_path_short, page_html_content.inspect())
            return render_error_page(500, "Error rendering page template: " + page_html_content.message, TEMPLATES_COMMON_DIR)

        # 4. Render layout if found
        if layout_abs_path:
            layout_dir = os.path.dirname(layout_abs_path)
            layout_filename_short = os.path.basename(layout_abs_path)
            
            layout_tmpl = template.new("layout_render", DEFAULT_SYNTAX, layout_dir)
            layout_tmpl.load(layout_filename_short)

            # Add page content to context for layout
            # Pylearn dict update:
            # layout_context = page_context.copy() # if copy method exists
            layout_context_pairs = dict()
            pc = page_context.Pairs.keys()
            for k_hash in pc: # Assuming Dict.Pairs is accessible
                pair = page_context.Pairs.get(k_hash)
                layout_context_pairs[pair.Key] = pair.Value
            layout_context_pairs[object.String(Value="content")] = page_html_content # Add content

            # Create new dict for layout (safer than modifying page_context)
            final_layout_context_dict = object.Dict(Pairs=dict())
            lcp = layout_context_pairs.keys()
            for k_obj in lcp:
                final_layout_context_dict.SetObjectItem(k_obj, layout_context_pairs.get(k_obj))

            final_html = layout_tmpl.execute(final_layout_context_dict)
            if isinstance(final_html, object.Error):
                print("Error rendering layout template:", layout_abs_path, final_html.inspect())
                return render_error_page(500, "Error rendering layout: " + final_html.message, TEMPLATES_COMMON_DIR)
            
            return object.ServerResponse(body=final_html, status_code=200)
        else:
            return object.ServerResponse(body=page_html_content, status_code=200)

    # except Exception as e:
    except: # Pylearn generic exception
        print("Unhandled error in handle_request_internal:", "ERROR<<<<<<")
        # Potentially try to get more info from 'e' if Pylearn errors have attributes
        return render_error_page(500, "An unexpected server error occurred: " + str("e"), TEMPLATES_COMMON_DIR)


# --- Middleware Handling (Simplified) ---
# Pylearn doesn't have Starlette's BaseHTTPMiddleware.
# We'll make a simple wrapper chain.
def apply_middleware(handler_func, middleware_func_or_list):
    if middleware_func_or_list == None:
        return handler_func

    # For simplicity, assume middleware_func_or_list is a single function for now.
    # A list would require chaining them.
    if isinstance(middleware_func_or_list, object.Function) or isinstance(middleware_func_or_list, object.Builtin): # check if callable
        
        def wrapped_handler(request):
            # The middleware function is expected to take (app_callable) and return (new_app_callable)
            # where app_callable takes (request) and returns response.
            # Starlette middleware: middleware_instance(app) -> new_app
            # new_app(scope, receive, send)
            # Simplified: Our middleware takes (request, call_next)
            
            # This needs a Pylearn callable that middleware can call.
            # Let's assume the hook returns a function that takes (request, next_handler)
            # and itself returns a response.
            
            # Simpler: hook returns a function that WRAPS the handler.
            # hook = {"middleware": lambda original_handler: new_wrapped_handler}
            
            # Let's assume hook.middleware is the actual new handler if present
            # The hook itself needs to be structured to take the 'handler_func' (our 'app')
            # and return the callable that takes 'request'.
            
            # If the hook.py defines a middleware class like Starlette:
            #   middleware_class = hook_module.middleware_stack[0] # if it's a list
            #   app_with_middleware = middleware_class(handler_func) # instantiates with our app
            #   return app_with_middleware.dispatch(request) # assumes a dispatch method
            # This is too complex without Pylearn classes fully matching Python.

            # Simplest possible middleware hook:
            # hook.py returns a function that IS the middleware.
            # This middleware must call the original handler itself.
            # e.g., hook_middleware_func(request, original_handler_func)
            
            # Let's assume middleware_func_or_list *is* the callable that takes (request, call_next)
            # This implies the hook.py must be written very carefully.
            # A more common pattern: middleware_func(app) returns a new app.
            # So, middleware_func_or_list IS the new app/handler.
            return middleware_func_or_list(request) # This means middleware is the outermost handler
            
        # This logic is tricky without a proper middleware stack / class system.
        # The hook.py example implies a BaseHTTPMiddleware structure.
        # For Pylearn, a simpler function chain might be:
        # app = handler
        # for mw_factory in reversed(middleware_list_from_hook):
        #     app = mw_factory(app) # Each mw_factory returns a new handler that calls the previous app
        # return app(request)
        
        # For now, if hook.middleware is a single function, it becomes the main handler.
        # It's the hook's job to call the original `handler_func` if it wants to proceed.
        # This is not standard middleware stack. Let's keep it simple:
        # The hook.py provides THE handler that the server will call.
        # The hook can then call our handle_request_internal.
        
        # Correct approach: The 'hook' function from hook.py should be a factory.
        # middleware_factory = middleware_func_or_list (from hook["middleware"])
        # actual_handler_to_run = middleware_factory(handler_func) # handler_func is handle_request_internal
        # return actual_handler_to_run(request)
        print("Middleware found, attempting to apply.")
        actual_handler_to_run = middleware_func_or_list(handler_func) # Apply the factory
        return actual_handler_to_run(request)


    print("Warning: Middleware from hook.py is not a recognized callable or list.")
    return handler_func(request) # Fallback to original handler

# --- Main Application Logic ---
def main_app_handler(request):
    # This function will be wrapped by middleware if hook.py exists.
    print("Main app handler called.")
    print("Request received for URL:", "Method:", request.method)
    return handle_request_internal(request)

final_handler_to_serve = main_app_handler

hook_path = "./hook.py" #"/teamspace/studios/this_studio/go-programs/practice/pylearn/examples/buibui/hook.py"
# print(type(hook_path))
try:
    import hook
    hook_module = hook
    # hook_module = pylearn_importlib.load_module_from_path(hook_path)
    # hook_module = importlib.load_module_from_path("./examples/buibui/hook.py")
    print("Looking for hook.py...")
except:
    hook_module = hook

try:
    print(type(hook_module))
    print("Looking for hook.py...")
    if hook_module and not isinstance(hook_module, object.Error):
        print("hook.py found and loaded.")
        hook_dict = hook_module.Env.get("hook")
        if hook_dict and isinstance(hook_dict, object.Dict):
            middleware_factory = hook_dict.get(object.String(Value="middleware")) # Pylearn dict.get
            if middleware_factory and isPylearnCallable(middleware_factory):
                print("Applying middleware from hook.py...")
                # The middleware_factory should take the next app and return a new app
                final_handler_to_serve = middleware_factory(main_app_handler) 
            else:
                print("No valid 'middleware' function found in hook.py's 'hook' dictionary.")
        else:
            print("No 'hook' dictionary found or it's not a Dict in hook.py.")
    elif isinstance(hook_module, object.Error):
        print("Error loading hook.py:", hook_module.inspect())
    else:
        print("hook.py not found or empty.")
    # if hook_module and not isinstance(hook_module, object.Error):
    #     print("hook.py found and loaded.")
    #     hook_dict = hook_module.Env.get("hook")
    #     if hook_dict and isinstance(hook_dict, object.Dict):
    #         middleware_factory = hook_dict.get(object.String(Value="middleware")) # Pylearn dict.get
    #         if middleware_factory and isPylearnCallable(middleware_factory):
    #             print("Applying middleware from hook.py...")
    #             # The middleware_factory should take the next app and return a new app
    #             final_handler_to_serve = middleware_factory(main_app_handler) 
    #         else:
    #             print("No valid 'middleware' function found in hook.py's 'hook' dictionary.")
    #     else:
    #         print("No 'hook' dictionary found or it's not a Dict in hook.py.")
    # elif isinstance(hook_module, object.Error):
    #     print("Error loading hook.py:", hook_module.inspect())
    # else:
    #     print("hook.py not found or empty.")
# except Exception as e:
except: # Pylearn needs its own import error or file not found error
    print("No hook.py found or error loading it:", "ERROR<<<<<<")


# --- Discover routes once at startup ---
abs_routes_dir = os.path.abspath(ROUTES_DIR)
if not os.path.exists(abs_routes_dir) or not os.path.isdir(abs_routes_dir):
    print("FATAL: Routes directory '{}' not found or not a directory.".format(abs_routes_dir))
    # sys.exit(1) # If Pylearn has sys.exit
else:
    print("Discovering routes in:", abs_routes_dir)
    find_routes_recursive(abs_routes_dir, "/", abs_routes_dir)
    print("Discovered routes:", discovered_routes) # Pylearn dict's Inspect might be verbose

# --- Start Server ---
address = "0.0.0.0:5173" # Make configurable later
print("Starting Pylearn Web Framework on", address)
httpserver.serve(address, final_handler_to_serve) # Pass the (potentially wrapped) handler
input("Press Enter to exit...")

print("Pylearn server setup seems complete. It should be running in the background (Go goroutine).")
print("This Pylearn script will now exit if httpserver.serve is non-blocking.")