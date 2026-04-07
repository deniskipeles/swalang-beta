# pylearn/stdlib/pycurl.py
"""
A Pylearn wrapper for the libcurl library, built using the Pylearn FFI.

This module provides a Pythonic, object-oriented interface for HTTP operations,
with platform-aware library loading.
"""

# --- CHANGED: Import the high-level ffi module ---
import ffi2 as ffi
import sys
import json as _json # Assuming a Pylearn-compatible json module exists

# --- CHANGED: Use ffi's built-in platform-aware loader ---
# The helper function is now part of the main ffi module
# def _load_library_with_fallbacks(base_name): ... (removed from here)

# --- Constants ---
# These remain the same, defining the numerical codes for libcurl options and info.
# (Keeping the class structure for organization)
class CurlCode:
    """CURL result codes"""
    OK = 0
    # ... (other codes) ...
    SSL_SHUTDOWN_FAILED = 80
    AGAIN = 81
    SSL_CRL_BADFILE = 82
    SSL_ISSUER_ERROR = 83
    FTP_PRET_FAILED = 84
    RTSP_CSEQ_ERROR = 85
    RTSP_SESSION_ERROR = 86
    FTP_BAD_FILE_LIST = 87
    CHUNK_FAILED = 88
    NO_CONNECTION_AVAILABLE = 89
    SSL_PINNEDPUBKEYNOTMATCH = 90
    SSL_INVALIDCERTSTATUS = 91
    HTTP2_STREAM = 92
    RECURSIVE_API_CALL = 93
    AUTH_ERROR = 94
    HTTP3 = 95
    QUIC_CONNECT_ERROR = 96

class CurlOpt:
    """CURL options"""
    URL = 10002
    WRITEDATA = 10001
    WRITEFUNCTION = 20011
    READDATA = 10009
    READFUNCTION = 20012
    USERAGENT = 10018
    HTTPHEADER = 10023
    POSTFIELDS = 10015
    POSTFIELDSIZE = 60
    HTTPGET = 80
    POST = 47
    CUSTOMREQUEST = 10036
    FOLLOWLOCATION = 52
    # ... (other options) ...
    MAXREDIRS = 68
    TIMEOUT = 78
    CONNECTTIMEOUT = 12
    SSLVERIFY = 64
    SSLCERT = 10025
    SSLKEY = 10087
    CAINFO = 10065
    VERBOSE = 41
    NOPROGRESS = 43
    PROGRESSFUNCTION = 20056
    PROGRESSDATA = 10057
    HEADERFUNCTION = 20079
    HEADERDATA = 10029
    COOKIEFILE = 10031
    COOKIEJAR = 10082
    HTTPAUTH = 107
    USERPWD = 10005
    PROXY = 10004
    PROXYPORT = 59
    PROXYAUTH = 111
    PROXYUSERPWD = 10006

class CurlInfo:
    """CURL info options"""
    RESPONSE_CODE = 0x200000 + 2
    TOTAL_TIME = 0x300000 + 3
    NAMELOOKUP_TIME = 0x300000 + 4
    CONNECT_TIME = 0x300000 + 5
    PRETRANSFER_TIME = 0x300000 + 6
    STARTTRANSFER_TIME = 0x300000 + 8
    REDIRECT_TIME = 0x300000 + 9
    REDIRECT_COUNT = 0x200000 + 20
    SIZE_UPLOAD = 0x300000 + 7
    SIZE_DOWNLOAD = 0x300000 + 8
    SPEED_DOWNLOAD = 0x300000 + 9
    SPEED_UPLOAD = 0x300000 + 10
    CONTENT_LENGTH_DOWNLOAD = 0x300000 + 15
    CONTENT_LENGTH_UPLOAD = 0x300000 + 16
    CONTENT_TYPE = 0x100000 + 18
    EFFECTIVE_URL = 0x100000 + 1

class HttpMethod:
    """HTTP methods"""
    GET = "GET"
    POST = "POST"
    PUT = "PUT"
    DELETE = "DELETE"
    HEAD = "HEAD"
    OPTIONS = "OPTIONS"
    PATCH = "PATCH"

# --- Exception Classes ---
class CurlError(Exception):
    """Base exception for curl errors"""
    def __init__(self, code, message):
        self.code = code
        self.message = message
        super().__init__(format_str("Curl error {code}: {message}"))

# --- Response Class ---
class Response:
    """Represents an HTTP response"""
    def __init__(self):
        self.status_code = 0
        self.headers = {}
        self.body = ""
        self.url = ""
        self.total_time = 0.0
        self.connect_time = 0.0
        self.download_size = 0
        self.upload_size = 0
        self.redirect_count = 0

    def json(self):
        """Parse response body as JSON"""
        # This will now correctly raise a JSONDecodeError if content is not valid JSON
        return _json.loads(self.body) # Assuming self.body is a string

    def text(self):
        """Return response body as text"""
        return self.body

    def ok(self):
        """Check if response was successful (2xx status code)"""
        return 200 <= self.status_code < 300

# --- Curl Handle Class ---
class Curl:
    """Wrapper for a CURL handle"""
    def __init__(self):
        if not CURL_AVAILABLE:
            raise CurlError(0, "libcurl not available")
        self._handle = _curl_easy_init()
        if not self._handle:
            raise CurlError(CurlCode.FAILED_INIT, "Failed to initialize CURL handle")
        self._response_buffer = ""
        self._header_buffer = ""
        self._slist = None # This will hold a curl_slist*
        # Set up write callback
        self._setup_callbacks()

    def _setup_callbacks(self):
        """Set up the write and header callbacks"""
        # Define the write callback function
        def write_callback(data_ptr, size, nmemb, user_data):
            total_size = size * nmemb
            if total_size > 0:
                # Use ffi.buffer_to_bytes to get the data from the C pointer
                chunk = ffi.buffer_to_bytes(data_ptr, total_size)
                self._response_buffer = self._response_buffer + chunk.decode('utf-8') # Assuming UTF-8 response
            return total_size # Important: return number of bytes processed

        # Define the header callback function
        def header_callback(data_ptr, size, nmemb, user_data):
            total_size = size * nmemb
            if total_size > 0:
                header_chunk = ffi.buffer_to_bytes(data_ptr, total_size)
                self._header_buffer = self._header_buffer + header_chunk.decode('utf-8') # Assuming UTF-8 headers
            return total_size

        # Create FFI callbacks
        # C signature for write callback: size_t (*)(char *ptr, size_t size, size_t nmemb, void *userdata)
        # We pass None as userdata for simplicity here.
        self._write_callback_func = ffi.callback(write_callback, ffi.c_int64, [ffi.c_char_p, ffi.c_int64, ffi.c_int64, ffi.c_void_p])
        # C signature for header callback: size_t (*)(char *ptr, size_t size, size_t nmemb, void *userdata)
        self._header_callback_func = ffi.callback(header_callback, ffi.c_int64, [ffi.c_char_p, ffi.c_int64, ffi.c_int64, ffi.c_void_p])

        # Set the callbacks using the specific setopt functions
        # CURLOPT_WRITEFUNCTION
        res_w = _setopt_funcptr(self._handle, CurlOpt.WRITEFUNCTION, self._write_callback_func)
        # CURLOPT_WRITEDATA (often CURLOPT_FILE) - pass NULL/userdata if needed, here we rely on closure
        # res_wd = _setopt_funcptr(self._handle, CurlOpt.WRITEDATA, ffi.NULL) # Often not needed if callback handles state

        # CURLOPT_HEADERFUNCTION
        res_h = _setopt_funcptr(self._handle, CurlOpt.HEADERFUNCTION, self._header_callback_func)
        # CURLOPT_HEADERDATA (often CURLOPT_WRITEHEADER)
        # res_hd = _setopt_funcptr(self._handle, CurlOpt.HEADERDATA, ffi.NULL)

        # Check for errors in setting callbacks (optional but good practice)
        if res_w != CurlCode.OK: # or res_h != CurlCode.OK:
             print(format_str("Warning: Failed to set write/header callback: {res_w}")) # or res_h


    # --- CHANGED: Use specific setopt functions for type safety ---
    def setopt_string(self, option, value):
        """Set a string option"""
        # Ensure value is a string
        if not isinstance(value, str):
             raise TypeError("Value for setopt_string must be a string")
        res = _setopt_string(self._handle, option, value)
        if res != CurlCode.OK:
            raise CurlError(res, format_str("Failed to set string option {option}"))

    def setopt_long(self, option, value):
        """Set a long integer option"""
        # Ensure value is an integer
        if not isinstance(value, int):
             raise TypeError("Value for setopt_long must be an integer")
        res = _setopt_long(self._handle, option, value)
        if res != CurlCode.OK:
            raise CurlError(res, format_str("Failed to set long option {option}"))

    def setopt_funcptr(self, option, value):
        """Set a function pointer option (callback)"""
        # The FFI callback object acts like a void* here
        # Ensure value is a callback object? (ffi checks this)
        res = _setopt_funcptr(self._handle, option, value) # value is the ffi callback object
        if res != CurlCode.OK:
            raise CurlError(res, format_str("Failed to set funcptr option {option}"))

    # --- CHANGED: Simplified generic setopt using type checking ---
    # This is optional if you prefer the explicit *_string, *_long methods
    # def setopt(self, option, value):
    #     """Set a CURL option (generic)"""
    #     if isinstance(value, str):
    #         self.setopt_string(option, value)
    #     elif isinstance(value, int):
    #         self.setopt_long(option, value)
    #     elif hasattr(value, '_is_ffi_callback'): # Hypothetical check for ffi callback
    #          self.setopt_funcptr(option, value)
    #     elif value is None: # For options like CURLOPT_WRITEDATA NULL
    #          res = _curl_easy_setopt(self._handle, option, ffi.NULL)
    #          if res != CurlCode.OK:
    #              raise CurlError(res, format_str("Failed to set option {option} to NULL")
    #     else:
    #         # Default to void* (might be a pointer from malloc or another lib call)
    #         # This is risky, better to be explicit
    #         # res = _curl_easy_setopt(self._handle, option, ffi.cast(ffi.c_void_p, id(value))) # Incorrect
    #         raise TypeError(format_str("Unsupported type for setopt: {type(value)}")


    def perform(self):
        """Perform the HTTP request"""
        self._response_buffer = ""
        self._header_buffer = ""
        result = _curl_easy_perform(self._handle)
        # print(result) # Debug print, can be removed

        if result != CurlCode.OK:
            raise CurlError(result, format_str("Request failed with code {result}"))

        # Build response object
        response = Response()
        response.body = self._response_buffer
        # Get response info
        response.status_code = self._get_info_long(CurlInfo.RESPONSE_CODE)
        response.total_time = self._get_info_double(CurlInfo.TOTAL_TIME)
        response.connect_time = self._get_info_double(CurlInfo.CONNECT_TIME)
        response.download_size = self._get_info_double(CurlInfo.SIZE_DOWNLOAD)
        response.upload_size = self._get_info_double(CurlInfo.SIZE_UPLOAD)
        response.redirect_count = self._get_info_long(CurlInfo.REDIRECT_COUNT)
        response.url = self._get_info_string(CurlInfo.EFFECTIVE_URL)
        # Parse headers
        response.headers = self._parse_headers(self._header_buffer)
        return response

    # --- CHANGED: Use ffi.allocate and ffi.dereference for info retrieval ---
    def _get_info_long(self, info):
        """Get long info from CURL handle"""
        # Allocate memory for a c_int64 (long)
        # value_ptr = ffi.malloc(ffi.c_int64) # Alternative if allocate doesn't exist
        # Use ffi.allocate if available (conceptual, might need ffi.malloc)
        # Let's assume we pass the type and get a pointer back conceptually
        # Or pass a pointer to the function. The C signature is usually getinfo(handle, info, void* ptr_to_value)
        # So we need to pass a pointer to where the long value should be stored.
        # Let's allocate memory for an int64 using the new FFI primitives if possible,
        # or use a low-level approach. Assuming ffi supports getting a pointer to a value location.
        # Simplified: Allocate 8 bytes and cast pointer.
        value_ptr = ffi.malloc(8) # Allocate 8 bytes for int64
        if not value_ptr:
             raise CurlError(CurlCode.OUT_OF_MEMORY, "Failed to allocate memory for _get_info_long")
        try:
            result = _curl_easy_getinfo(self._handle, info, value_ptr)
            if result == CurlCode.OK:
                # Read the int64 value from the allocated memory
                # This requires reading memory as a specific type.
                # Assuming ffi.read_memory or similar exists, or casting the pointer.
                # Let's assume a way to read an int64 from a void* pointer.
                # This is a placeholder for the actual FFI read mechanism.
                # If ffi supports reading typed values directly:
                # value_obj = ffi.read_memory(value_ptr, ffi.c_int64)
                # value = value_obj.Value if value_obj is Integer object
                # Or if we need to cast the pointer and dereference in Go:
                # This part depends on how your FFI exposes memory reading.
                # Placeholder logic - needs correct FFI call:
                # --- CORRECTED APPROACH using likely FFI methods ---
                # 1. Cast the void* pointer to the correct type pointer (conceptually)
                # 2. Dereference it. This is usually done inside the Go FFI layer.
                # Let's assume the Go side has a function to read an int64 from a void*.
                # We call that via the FFI.
                # For now, simulate getting the value (this part needs real FFI integration)
                # Simulate reading the value (replace with actual FFI call)
                # This is the tricky part - how does the Pylearn FFI read a value from a pointer?
                # Option 1: If the Go FFI has a helper like pyReadMemory(ptr, type)
                # and returns a Pylearn object. This would require exposing pyReadMemory.
                # Option 2: Modify _curl_easy_getinfo wrapper or add a specific reader.
                # Let's assume a generic memory reader exists in the updated ffi module.
                # This is pseudo-code for the FFI call needed:
                # pylearn_int_obj = ffi._read_primitive_from_ptr(value_ptr, ffi.c_int64) # Hypothetical
                # if pylearn_int_obj and hasattr(pylearn_int_obj, 'Value'):
                #     return pylearn_int_obj.Value
                # else:
                #     return 0 # Or raise error
                # --- REALISTIC APPROACH BASED ON EXISTING PATTERN ---
                # The existing ffi.read_memory takes a Pointer and type.
                # We need to wrap our void* in a Pointer object.
                # But read_memory reads *from* a pointer, we need to interpret the data *at* the pointer.
                # The Go FFI likely needs a specific function or the Go code that calls
                # _curl_easy_getinfo needs to handle the dereferencing.
                # Let's assume the Go backend for _curl_easy_getinfo handles storing the
                # retrieved value into a Pylearn Integer object passed by reference.
                # This is complex. Let's simplify by assuming the Go side returns the value directly
                # if we pass the right arguments, or we need a helper.
                # --- WORKAROUND/SOLUTION ---
                # Pass the pointer and let the Go side handle reading/dereferencing.
                # This requires the Go _curl_easy_getinfo binding to understand it needs
                # to write the result to the memory pointed to by the third arg.
                # The Pylearn wrapper for _curl_easy_getinfo should then return the value read.
                # This implies the Go function needs to be more complex or we need a specific
                # reader function. Let's assume the Go _curl_easy_getinfo binding is correctly
                # implemented to write the long value to the memory location pointed to by value_ptr.
                # Therefore, after the call, the memory at value_ptr contains the long value.
                # We now need to read that value using FFI.
                # How? If ffi.read_memory reads *from* a pointer location into a Pylearn object...
                # We need to tell it to read an int64. But read_memory likely expects a Pointer object.
                # Let's create a temporary Pointer object pointing to our allocated memory.
                temp_pointer_obj = type('TempPointer', (), {'Address': value_ptr.Address if hasattr(value_ptr, 'Address') else value_ptr})()
                # This is getting convoluted. Let's assume ffi has a direct way or the Go binding returns it.
                # --- FINAL ASSUMPTION for this example ---
                # The call to _curl_easy_getinfo modifies the memory pointed to by value_ptr.
                # We need a way to read that memory as an int64.
                # Assuming the FFI has a way (e.g., extending read_memory or adding read_int64_at_ptr)
                # For now, placeholder return. ***THIS NEEDS REAL FFI INTEGRATION***
                # A robust solution would involve modifying the Go FFI bindings or the Pylearn ffi module
                # to correctly handle "output parameters" like this.
                # Let's pretend we have a way to read it:
                # retrieved_value = ffi._internal_read_int64(value_ptr) # Placeholder
                # return retrieved_value
                # --- LETS USE A MORE PLAUSIBLE APPROACH ---
                # Modify the _curl_easy_getinfo binding in Go to return the value directly
                # if the third argument points to an int64 location.
                # This is how ctypes often works.
                # The Pylearn Go wrapper for _curl_easy_getinfo would then need to:
                # 1. Accept the third arg as a special "output pointer" indicator.
                # 2. Call the real curl_easy_getinfo.
                # 3. Read the value from the memory location pointed to by the third arg.
                # 4. Return that value as a Pylearn Integer.
                # This is complex binding logic.
                # For this Pylearn wrapper code, let's assume the Go binding handles it and
                # _curl_easy_getinfo returns the value (or an error).
                # But the C signature is getinfo(handle, info, void*).
                # We need to signal to the Go binding that this void* is an output param for an int64.
                # This usually means the Pylearn call needs to be different, or the Go binding
                # is smart enough (unlikely).
                # --- SIMPLEST WORKING APPROACH ---
                # Assume the Go side has a specific function like _curl_easy_getinfo_long
                # that takes (handle, info) and returns the long value.
                # We won't use the generic _curl_easy_getinfo for long/double/string then.
                # Let's redefine _curl_easy_getinfo_long etc. in the Go init() and expose them.
                # This is the cleanest separation.
                # Placeholder for now, indicating where the FFI integration is needed.
                print("WARNING: _get_info_long not fully implemented due to FFI output param handling.")
                return 0 # Placeholder
            return 0
        finally:
             if value_ptr:
                  ffi.free(value_ptr) # Free the allocated memory

    def _get_info_double(self, info):
        """Get double info from CURL handle"""
        # Similar issues as _get_info_long. Needs FFI integration for reading doubles from ptr.
        # Placeholder
        print("WARNING: _get_info_double not fully implemented due to FFI output param handling.")
        value_ptr = ffi.malloc(8) # Allocate 8 bytes for double
        # ... (similar logic to _get_info_long for calling, reading, freeing)
        # Placeholder return
        if value_ptr:
             ffi.free(value_ptr)
        return 0.0

    def _get_info_string(self, info):
        """Get string info from CURL handle"""
        # C signature: getinfo(handle, info, char**)
        # We need to pass a pointer to a char* (void**).
        # Allocate memory for a pointer (size of void*)
        ptr_to_char_ptr = ffi.malloc(ffi.c_void_p.Size()) # Size of a pointer
        if not ptr_to_char_ptr:
             raise CurlError(CurlCode.OUT_OF_MEMORY, "Failed to allocate memory for _get_info_string")
        try:
            result = _curl_easy_getinfo(self._handle, info, ptr_to_char_ptr)
            if result == CurlCode.OK:
                # Now, ptr_to_char_ptr points to memory that contains a char* (the actual string pointer)
                # We need to read that char* value from the memory location ptr_to_char_ptr.
                # This is the "value" returned by getinfo for string types.
                # Again, this requires FFI support for reading pointer values from memory.
                # Placeholder logic:
                # retrieved_char_ptr_value = ffi._internal_read_pointer(ptr_to_char_ptr) # Hypothetical
                # if retrieved_char_ptr_value and retrieved_char_ptr_value != ffi.NULL:
                #     # Now convert the C char* to a Pylearn string
                #     return ffi.string_from_ptr(retrieved_char_ptr_value) # Hypothetical
                # --- SIMPLER ASSUMPTION ---
                # Assume the Go binding for _curl_easy_getinfo handles the char** case specially
                # and returns the Pylearn string directly.
                # This would mean the Pylearn call site doesn't pass the output buffer pointer,
                # but the Go binding creates it, calls curl, reads the result, frees the C buffer,
                # and returns the Pylearn string. This is cleaner for the Pylearn user.
                # Let's assume this is how it's implemented.
                # So, the Go _curl_easy_getinfo binding needs logic:
                # if info_type_is_string:
                #   call real_curl_easy_getinfo with a local char* var
                #   if success:
                #     convert char* to Pylearn string
                #     return Pylearn string
                #   else:
                #     return FFIError
                # This requires modifying the Go binding generation or the specific Go function.
                # Placeholder for now.
                 print("WARNING: _get_info_string not fully implemented due to FFI output param handling.")
                 return ""
            return ""
        finally:
             if ptr_to_char_ptr:
                  ffi.free(ptr_to_char_ptr) # Free the buffer holding the pointer address

    def _parse_headers(self, header_text):
        """Parse HTTP headers from header text"""
        headers = {}
        # Ensure header_text is a string
        if not isinstance(header_text, str):
             header_text = str(header_text) # Fallback attempt
        lines = header_text.split('\n') # Use \n for line splitting
        for line in lines:
            line = line.strip()
            if line and ':' in line:
                parts = line.split(':', 1) # Split only on the first ':'
                if len(parts) == 2:
                     key = parts[0].strip()
                     value = parts[1].strip()
                     headers[key] = value
        return headers

    def set_headers(self, headers):
        """Set HTTP headers"""
        # Clean up any existing slist
        if self._slist:
            _curl_slist_free_all(self._slist)
            self._slist = None

        # Create new slist
        current_list = ffi.NULL # Start with a NULL list
        for key, value in headers.items():
            header_line = format_str("{key}: {value}")
            # Append to the list. This returns the new head of the list.
            # The returned pointer must be kept track of.
            current_list = _curl_slist_append(current_list, header_line)
            # Check for allocation failure (current_list would be NULL)
            if current_list is ffi.NULL or (hasattr(current_list, 'Address') and current_list.Address == 0):
                 # Clean up any partially built list
                 if current_list is not ffi.NULL:
                      _curl_slist_free_all(current_list)
                 raise CurlError(CurlCode.OUT_OF_MEMORY, "Failed to allocate curl_slist")

        self._slist = current_list # Store the final head of the list

        # Set the HTTPHEADER option to use this list
        if self._slist:
            # Pass the pointer to the head of the list
            # The pointer object itself acts like the void* here.
            self.setopt_funcptr(CurlOpt.HTTPHEADER, self._slist) # Use funcptr as it's a pointer option


    def cleanup(self):
        """Clean up CURL handle and resources"""
        if self._slist:
            _curl_slist_free_all(self._slist)
            self._slist = None
        if self._handle:
            _curl_easy_cleanup(self._handle)
            self._handle = ffi.NULL # Or None, depending on how NULL is represented

# --- CHANGED: Use ffi.CDLL for loading ---
# Load libcurl
_curl = None
CURL_AVAILABLE = False
try:
    # Use the platform-aware loader from the ffi module
    _curl = ffi._load_library_with_fallbacks("curl") # Or "libcurl" depending on search logic
    CURL_AVAILABLE = True
except ffi.FFIError as e:
    print(format_str("Warning: Failed to load libcurl: {e}"))
    print("HTTP functionality will not be available.")

# --- Global curl Functions ---
# Define function pointers for libcurl functions
_curl_global_init = None
_curl_global_cleanup = None
_curl_easy_init = None
_curl_easy_cleanup = None
# --- CHANGED: Specific setopt functions for type safety ---
_setopt_long = None
_setopt_string = None
_setopt_funcptr = None # For callbacks and pointer options like HTTPHEADER
# Generic setopt (less type-safe, relies on FFI conversion)
# _curl_easy_setopt = None
_curl_easy_perform = None
# --- CHANGED: Specific getinfo functions for type safety and output param handling ---
# _curl_easy_getinfo = None # Generic one is tricky with output params
_curl_easy_getinfo_long = None # Hypothetical specific one
_curl_easy_getinfo_double = None # Hypothetical specific one
_curl_easy_getinfo_string = None # Hypothetical specific one

_curl_slist_append = None
_curl_slist_free_all = None

if CURL_AVAILABLE:
    # --- CHANGED: Use ffi.CDLL function access and configuration ---
    # CURLcode curl_global_init(long flags);
    _curl_global_init = _curl.curl_global_init([ffi.c_int64], ffi.c_int32)
    # void curl_global_cleanup(void);
    _curl_global_cleanup = _curl.curl_global_cleanup([], None)
    # CURL *curl_easy_init(void);
    # Returns a pointer, likely opaque. Represent as c_void_p.
    _curl_easy_init = _curl.curl_easy_init([], ffi.c_void_p)
    # void curl_easy_cleanup(CURL *curl);
    _curl_easy_cleanup = _curl.curl_easy_cleanup([ffi.c_void_p], None)

    # --- CHANGED: Define separate setopt functions for each C type ---
    # For integer options: CURLcode curl_easy_setopt(CURL *handle, CURLoption option, long value);
    _setopt_long = _curl.curl_easy_setopt([ffi.c_void_p, ffi.c_int32, ffi.c_int64], ffi.c_int32)
    # For string options: CURLcode curl_easy_setopt(CURL *handle, CURLoption option, const char *value);
    _setopt_string = _curl.curl_easy_setopt([ffi.c_void_p, ffi.c_int32, ffi.c_char_p], ffi.c_int32)
    # For pointer options (like callbacks, HTTPHEADER list): CURLcode curl_easy_setopt(CURL *handle, CURLoption option, void *value);
    # Also used for CURLOPT_WRITEDATA etc. if passing a pointer.
    _setopt_funcptr = _curl.curl_easy_setopt([ffi.c_void_p, ffi.c_int32, ffi.c_void_p], ffi.c_int32)
    # Generic setopt (if needed, less safe) - commented out as specific ones are preferred
    # _curl_easy_setopt = _curl.curl_easy_setopt([ffi.c_void_p, ffi.c_int32, ffi.c_void_p], ffi.c_int32)

    # --- CHANGED: Use ffi.CDLL for curl_easy_perform ---
    # CURLcode curl_easy_perform(CURL *curl);
    _curl_easy_perform = _curl.curl_easy_perform([ffi.c_void_p], ffi.c_int32)

    # --- CHANGED: Specific getinfo functions (conceptual, need Go FFI support) ---
    # Getting info is tricky because C uses output parameters (pointers to values).
    # Conceptually, we'd want functions like:
    # CURLcode curl_easy_getinfo_long(CURL *curl, CURLINFO info) -> returns long value or error
    # CURLcode curl_easy_getinfo_double(CURL *curl, CURLINFO info) -> returns double value or error
    # CURLcode curl_easy_getinfo_string(CURL *curl, CURLINFO info) -> returns string value or error
    # These would require special Go FFI binding logic to handle the output parameters correctly.
    # For now, we'll try to use the generic one and note the limitations.
    # CURLcode curl_easy_getinfo(CURL *curl, CURLINFO info, ...);
    # The ... makes it variadic. We need to tell FFI the specific type expected.
    # This is where variadic FFI support (if added) would be useful, or specific wrappers.
    # Let's define it to take a void* for the output parameter.
    # _curl_easy_getinfo = _curl.curl_easy_getinfo([ffi.c_void_p, ffi.c_int32, ffi.c_void_p], ffi.c_int32)
    # Placeholder definitions for specific ones - they need Go backend changes
    # def _dummy_getinfo(ctx, args): return ffi.FFIError(ffi.ErrBadSignature, "Not implemented")
    # _curl_easy_getinfo_long = type('DummyFunc', (), {'__call__': lambda self, h, i: _dummy_getinfo(None, [])})()
    # _curl_easy_getinfo_double = _curl_easy_getinfo_long
    # _curl_easy_getinfo_string = _curl_easy_getinfo_long
    # For this update, we'll leave the generic one and mark the _get_info_* methods as needing work.

    # struct curl_slist *curl_slist_append(struct curl_slist *list, const char *data);
    # Returns a pointer to the (new) head of the list.
    _curl_slist_append = _curl.curl_slist_append([ffi.c_void_p, ffi.c_char_p], ffi.c_void_p)
    # void curl_slist_free_all(struct curl_slist *list);
    _curl_slist_free_all = _curl.curl_slist_free_all([ffi.c_void_p], None)

# --- High-level HTTP Functions ---
def get(url, headers=None, timeout=30):
    """Perform an HTTP GET request"""
    curl = Curl()
    try:
        # --- CHANGED: Use the new typed methods ---
        curl.setopt_string(CurlOpt.URL, url)
        curl.setopt_long(CurlOpt.HTTPGET, 1)
        curl.setopt_long(CurlOpt.TIMEOUT, timeout)
        curl.setopt_long(CurlOpt.FOLLOWLOCATION, 1)
        # --- END FIX ---
        if headers:
            curl.set_headers(headers)
        return curl.perform()
    finally:
        curl.cleanup()

def post(url, data="", headers=None, timeout=30):
    """Perform an HTTP POST request"""
    curl = Curl()
    try:
        curl.setopt_string(CurlOpt.URL, url)
        curl.setopt_long(CurlOpt.POST, 1)
        curl.setopt_long(CurlOpt.TIMEOUT, timeout)
        curl.setopt_long(CurlOpt.FOLLOWLOCATION, 1)
        if data:
            if isinstance(data, str):
                curl.setopt_string(CurlOpt.POSTFIELDS, data)
                curl.setopt_long(CurlOpt.POSTFIELDSIZE, len(data))
        if headers:
            curl.set_headers(headers)
        return curl.perform()
    finally:
        curl.cleanup()

def put(url, data="", headers=None, timeout=30):
    """Perform an HTTP PUT request"""
    curl = Curl()
    try:
        curl.setopt_string(CurlOpt.URL, url)
        curl.setopt_string(CurlOpt.CUSTOMREQUEST, "PUT") # Use CUSTOMREQUEST for PUT
        curl.setopt_long(CurlOpt.TIMEOUT, timeout)
        curl.setopt_long(CurlOpt.FOLLOWLOCATION, 1)
        if data:
            if isinstance(data, str):
                curl.setopt_string(CurlOpt.POSTFIELDS, data) # PUT can also use POSTFIELDS
                curl.setopt_long(CurlOpt.POSTFIELDSIZE, len(data))
        if headers:
            curl.set_headers(headers)
        return curl.perform()
    finally:
        curl.cleanup()

def delete(url, headers=None, timeout=30):
    """Perform an HTTP DELETE request"""
    curl = Curl()
    try:
        curl.setopt_string(CurlOpt.URL, url)
        curl.setopt_string(CurlOpt.CUSTOMREQUEST, "DELETE")
        curl.setopt_long(CurlOpt.TIMEOUT, timeout)
        curl.setopt_long(CurlOpt.FOLLOWLOCATION, 1)
        if headers:
            curl.set_headers(headers)
        return curl.perform()
    finally:
        curl.cleanup()

def head(url, headers=None, timeout=30):
    """Perform an HTTP HEAD request"""
    curl = Curl()
    try:
        curl.setopt_string(CurlOpt.URL, url)
        curl.setopt_string(CurlOpt.CUSTOMREQUEST, "HEAD")
        curl.setopt_long(CurlOpt.TIMEOUT, timeout)
        curl.setopt_long(CurlOpt.FOLLOWLOCATION, 1)
        if headers:
            curl.set_headers(headers)
        return curl.perform()
    finally:
        curl.cleanup()

# --- Session Class ---
class Session:
    """HTTP session for reusing connections and maintaining state (simplified)"""
    def __init__(self):
        self.curl = Curl()
        self.default_headers = {}

    def get(self, url, headers=None, timeout=30):
        """Perform GET request in this session"""
        return self._request(HttpMethod.GET, url, headers=headers, timeout=timeout)

    def post(self, url, data="", headers=None, timeout=30):
        """Perform POST request in this session"""
        return self._request(HttpMethod.POST, url, data=data, headers=headers, timeout=timeout)

    def put(self, url, data="", headers=None, timeout=30):
        """Perform PUT request in this session"""
        return self._request(HttpMethod.PUT, url, data=data, headers=headers, timeout=timeout)

    def delete(self, url, headers=None, timeout=30):
        """Perform DELETE request in this session"""
        return self._request(HttpMethod.DELETE, url, headers=headers, timeout=timeout)

    def head(self, url, headers=None, timeout=30):
        """Perform HEAD request in this session"""
        return self._request(HttpMethod.HEAD, url, headers=headers, timeout=timeout)

    def _request(self, method, url, data=None, headers=None, timeout=30):
        """Internal method to perform requests"""
        # Merge default headers with request headers
        all_headers = {}
        all_headers.update(self.default_headers)
        if headers:
            all_headers.update(headers)

        self.curl.setopt_string(CurlOpt.URL, url)
        self.curl.setopt_long(CurlOpt.TIMEOUT, timeout)
        self.curl.setopt_long(CurlOpt.FOLLOWLOCATION, 1)

        if method == HttpMethod.GET:
            self.curl.setopt_long(CurlOpt.HTTPGET, 1)
        elif method == HttpMethod.POST:
            self.curl.setopt_long(CurlOpt.POST, 1)
        else:
            self.curl.setopt_string(CurlOpt.CUSTOMREQUEST, method)

        if data and isinstance(data, str):
            # PUT/POST data
            self.curl.setopt_string(CurlOpt.POSTFIELDS, data)
            self.curl.setopt_long(CurlOpt.POSTFIELDSIZE, len(data))

        if all_headers:
            self.curl.set_headers(all_headers)

        return self.curl.perform()

    def close(self):
        """Close the session and clean up resources"""
        if self.curl:
            self.curl.cleanup()
            self.curl = None

# --- Initialization Functions ---
def init():
    """Initialize the curl library globally"""
    if CURL_AVAILABLE:
        # CURL_GLOBAL_ALL is usually 3
        result = _curl_global_init(3) # CURL_GLOBAL_ALL
        if result != CurlCode.OK:
            raise CurlError(result, "Failed to initialize libcurl")
    else:
        print("libcurl not available, init() skipped.")

def cleanup():
    """Clean up curl library resources"""
    if CURL_AVAILABLE:
        _curl_global_cleanup()
    else:
        print("libcurl not available, cleanup() skipped.")

# --- Utility Functions ---
def version():
    """Get curl version information (placeholder)"""
    if CURL_AVAILABLE:
        # This would normally call curl_version()
        return "libcurl (version info not implemented in this wrapper)"
    else:
        return "libcurl not available"

def escape(url):
    """URL escape a string (placeholder)"""
    # Simplified - real implementation would use curl_easy_escape
    # Requires handling of output string memory allocation/freeing via FFI
    import urllib.parse # Assuming Pylearn has this or similar
    return urllib.parse.quote(url) # Placeholder using standard lib if available

# --- Example Usage (if run directly) ---
if __name__ == "__main__":
    init()
    try:
        response = get("http://example.com")
        print(format_str("Status: {response.status_code}"))
        print(format_str("Body: {response.text()}"))
        # print(format_str("JSON: {response.json()}") # Only if body is JSON
    except CurlError as e:
        print(format_str("Error: {e}"))
    finally:
        cleanup()
