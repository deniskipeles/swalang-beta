# pylearn/stdlib/pycurl.py
"""
A simple, high-level wrapper around libcurl for making HTTP requests.
"""
import ffi
import json as _json

# --- Load libcurl and define constants ---

try:
    _lib = ffi.CDLL("libcurl.so.4")
except ffi.FFIError:
    try:
        _lib = ffi.CDLL("libcurl.dylib")
    except ffi.FFIError:
        raise ImportError("Could not find libcurl.so.4 or libcurl.dylib. Please ensure libcurl is installed.")

# Constants from curl.h for general options
CURL_GLOBAL_DEFAULT = 3
CURLOPT_URL = 10002
CURLOPT_WRITEFUNCTION = 20011
CURLOPT_FOLLOWLOCATION = 52
CURLOPT_HTTPHEADER = 10023 

# Constants for request methods and data
CURLOPT_CUSTOMREQUEST = 10036
CURLOPT_POSTFIELDS = 10015
CURLOPT_POSTFIELDSIZE = 60
CURLOPT_POST = 47
CURLOPT_HTTPGET = 80 # <<< ADDED: Explicitly set GET method

# Constants for getting info after a request
CURLINFO_RESPONSE_CODE = 2097154
CURLINFO_EFFECTIVE_URL = 1048578


# --- Public API ---

class Response:
    """A container for the response from an HTTP request."""
    def __init__(self):
        self.content = b''
        self.status_code = 0
        self.url = ""

    def text(self, encoding='utf-8'):
        """The response body as a string, decoded with the given encoding."""
        return str(self.content.decode(encoding)) # This relies on the default str() for bytes

    def json(self):
        """Parses the response content as JSON and returns a Pylearn object."""
        # This will now correctly raise a JSONDecodeError if content is not valid JSON
        return _json.loads(self.content)


def request(method, url, data=None, headers=None, follow_redirects=True):
    """
    Performs a generic HTTP request and returns a Response object.
    """
    _lib.curl_global_init([ffi.c_int64], ffi.c_int32)(CURL_GLOBAL_DEFAULT)
    data_chunks = []

    def write_callback(ptr, size, nmemb, userdata):
        total_size = size * nmemb
        if total_size > 0:
            chunk = ffi.buffer_to_bytes(ptr, total_size)
            data_chunks.append(chunk)
        return total_size

    write_cb_ptr = ffi.callback(write_callback, ffi.c_int64, [ffi.c_char_p, ffi.c_int64, ffi.c_int64, ffi.c_void_p])
    handle = _lib.curl_easy_init([], ffi.c_void_p)()
    if handle is None:
        raise ffi.FFIError("curl_easy_init() failed")

    setopt_string = _lib.curl_easy_setopt([ffi.c_void_p, ffi.c_int32, ffi.c_char_p], ffi.c_int32)
    setopt_long = _lib.curl_easy_setopt([ffi.c_void_p, ffi.c_int32, ffi.c_int64], ffi.c_int32)
    setopt_funcptr = _lib.curl_easy_setopt([ffi.c_void_p, ffi.c_int32, ffi.c_void_p], ffi.c_int32)

    try:
        # --- Configure the handle ---
        setopt_string(handle, CURLOPT_URL, url)
        setopt_funcptr(handle, CURLOPT_WRITEFUNCTION, write_cb_ptr)
        if follow_redirects:
            setopt_long(handle, CURLOPT_FOLLOWLOCATION, 1)

        # --- THIS IS THE FIX FOR EXPLICITLY SETTING THE HTTP METHOD ---
        method_upper = method.upper()
        if method_upper == 'GET':
            # Explicitly tell curl to perform a GET. This resets other methods.
            setopt_long(handle, CURLOPT_HTTPGET, 1)
        elif method_upper == 'POST':
            # Use CURLOPT_POST for POST requests.
            setopt_long(handle, CURLOPT_POST, 1)
        else:
            # For all other methods (PUT, DELETE, etc.), use CUSTOMREQUEST.
            setopt_string(handle, CURLOPT_CUSTOMREQUEST, method_upper)
        # --- END OF METHOD FIX ---

        request_data_bytes = None
        final_headers = {}
        if headers is not None:
             for key in headers:
                 final_headers[key] = headers[key]

        if data is not None:
            if isinstance(data, dict):
                request_data_bytes = _json.dumps(data).encode('utf-8')
                final_headers['Content-Type'] = 'application/json'
            elif isinstance(data, bytes):
                request_data_bytes = data
            else:
                raise TypeError("request `data` must be a dict or bytes object")

        if request_data_bytes is not None:
            data_ptr = ffi.addressof(request_data_bytes)
            setopt_funcptr(handle, CURLOPT_POSTFIELDS, data_ptr)
            setopt_long(handle, CURLOPT_POSTFIELDSIZE, len(request_data_bytes))

        # Header handling would go here...

        # --- Perform and Get Info ---
        perform = _lib.curl_easy_perform([ffi.c_void_p], ffi.c_int32)
        res = perform(handle)
        if res != 0:
            raise ffi.FFIError(format_str("curl_easy_perform() failed: code {res}"))

        response = Response()
        status_code_ptr = ffi.malloc(ffi.c_int64.Size())
        getinfo_long = _lib.curl_easy_getinfo([ffi.c_void_p, ffi.c_int32, ffi.c_void_p], ffi.c_int32)
        getinfo_long(handle, CURLINFO_RESPONSE_CODE, status_code_ptr)
        response.status_code = ffi.read_memory(status_code_ptr, ffi.c_int64)
        ffi.free(status_code_ptr)

        url_ptr_ptr = ffi.malloc(ffi.c_void_p.Size())
        getinfo_string = _lib.curl_easy_getinfo([ffi.c_void_p, ffi.c_int32, ffi.c_void_p], ffi.c_int32)
        getinfo_string(handle, CURLINFO_EFFECTIVE_URL, url_ptr_ptr)
        final_url_ptr = ffi.read_memory(url_ptr_ptr, ffi.c_void_p)
        if final_url_ptr is not None:
            temp_bytes = ffi.buffer_to_bytes(final_url_ptr, 4096)
            null_pos = -1
            for i in range(len(temp_bytes)):
                if temp_bytes[i] == 0:
                    null_pos = i
                    break
            if null_pos != -1:
                response.url = str(temp_bytes[:null_pos])
            else:
                 response.url = str(temp_bytes)
        ffi.free(url_ptr_ptr)

    finally:
        cleanup = _lib.curl_easy_cleanup([ffi.c_void_p], None)
        cleanup(handle)

    final_content = b''
    for chunk in data_chunks:
        final_content = final_content + chunk
    response.content = final_content
    
    return response

# --- High-Level Wrapper Functions ---

def get(url, headers=None, follow_redirects=True):
    return request('GET', url, headers=headers, follow_redirects=follow_redirects)

def post(url, data=None, headers=None, follow_redirects=True):
    return request('POST', url, data=data, headers=headers, follow_redirects=follow_redirects)

def put(url, data=None, headers=None, follow_redirects=True):
    return request('PUT', url, data=data, headers=headers, follow_redirects=follow_redirects)

def patch(url, data=None, headers=None, follow_redirects=True):
    return request('PATCH', url, data=data, headers=headers, follow_redirects=follow_redirects)

def delete(url, headers=None, follow_redirects=True):
    return request('DELETE', url, headers=headers, follow_redirects=follow_redirects)



# Clean up libcurl globally when the module is (conceptually) unloaded.
# Note: In Pylearn, there's no module unload mechanism, so this is for correctness.
# _lib.curl_global_cleanup([], None)()
# Example usage function for testing
def test():
    print("--- Testing pycurl wrapper ---")
    try:
        resp = get("https://httpbin.org/get")
        print("Successfully fetched response from httpbin.org/get.")
        print("Response Content (first 100 bytes):")
        print(resp.content)
    except Exception as e:
        print("An error occurred:", e)