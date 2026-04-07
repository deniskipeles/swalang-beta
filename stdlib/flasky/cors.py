# pylearn/stdlib/flasky/cors.py

"""
A basic implementation of Cross-Origin Resource Sharing (CORS) for the
Pylearn Flasky web framework.
"""

import httpserver as _httpserver

class CORS:
    # ... (docstring and __init__ method are the same as above) ...
    def __init__(self, app, origins="*", methods=None, headers=None, allow_credentials=False, max_age=None):
        if not hasattr(app, '__call__'):
             raise TypeError("The 'app' argument must be a callable Pylearn object.")
        
        if isinstance(origins, str):
            origins = [o.strip() for o in origins.split(',')]
        elif origins is None:
            origins = ["*"]
        
        if methods is None:
            methods = ["GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD", "PATCH"]
        elif isinstance(methods, str):
            methods = [methods]
        
        if headers is None:
            headers = ["Accept", "Accept-Language", "Content-Language", "Content-Type", "Authorization"]
        elif isinstance(headers, str):
            headers = [headers]
        
        options = {
            "origins": origins,
            "methods": methods,
            "headers": headers,
            "allow_credentials": allow_credentials,
            "max_age": max_age,
        }
        
        self._native_wrapper = _httpserver._new_cors_wrapper(app, options)

    def __getattr__(self, name):
        """
        This allows the Go backend to access the internal _native_wrapper.
        """
        if name == "_native_wrapper":
            return self._native_wrapper
        raise AttributeError(format_str("'CORS' object has no attribute '{name}'"))