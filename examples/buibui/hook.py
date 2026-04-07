# hook.py (Pylearn)

# This is a very simplified middleware concept for Pylearn.
# A real middleware system would involve classes and a more formal stack.

# The middleware factory takes the 'next_app' (our original handler)
# and returns a new handler function that wraps it.
def logging_middleware_factory(next_app_handler):
    print("Logging middleware factory called. Wrapping handler:", next_app_handler)

    def middleware_handler(request):
        print("Middleware: Request received for URL:", request.url, "Method:", request.method)
        
        # Example: Modify request (if Pylearn Request objects are mutable or can be replaced)
        # request.headers.SetObjectItem(String("X-Middleware-Added"), String("true"))

        # Call the next handler in the chain (or the main app handler)
        response = next_app_handler(request) # Pylearn function call

        # Example: Modify response (if Pylearn Response objects are mutable or can be replaced)
        response_status_code = "N/A"
        if response.StatusCode:
            response_status_code = response.StatusCode.value
        print("Middleware: Response status for URL:", request.url, "is", response_status_code)
        # if isinstance(response, object.ServerResponse) and response.Headers:
        #    response.Headers.SetObjectItem(String("X-Middleware-Processed"), String("true"))
        
        return response
    
    return middleware_handler

# The 'hook' dictionary is what app.py will look for.
hook = {
    "middleware": logging_middleware_factory
}

print("hook.py executed and 'hook' dictionary defined.")