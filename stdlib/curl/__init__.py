# pylearn/stdlib/curl/__init__.py
"""
A Pylearn wrapper for the libcurl library, built using the Pylearn FFI.
This module provides a Pythonic, object-oriented interface for HTTP operations.
"""

import ffi
import sys
import json as _json
import time

# ==============================================================================
#  Platform-Aware Library Loading
# ==============================================================================

def _load_library_with_fallbacks(base_name):
    """Tries to load a shared library using common platform-specific names."""
    platform = sys.platform
    
    try:
        if platform == 'windows':
            return ffi.CDLL(format_str("lib{base_name}.dll"))
        return ffi.CDLL(base_name)
    except ffi.FFIError as e:
        print(format_str("Warning: Failed to load libcurl with standard name: {e}"))
        raise e

_curl = None
CURL_AVAILABLE = False
try:
    _curl = _load_library_with_fallbacks("curl")
    CURL_AVAILABLE = True
except ffi.FFIError as e:
    print(format_str("Warning: Failed to load libcurl: {e}"))
    print("HTTP functionality in the 'curl' module will not be available.")


# ==============================================================================
#  Define C Function Signatures
# ==============================================================================

if CURL_AVAILABLE:
    # --- Global Functions ---
    _curl_global_init = _curl.curl_global_init([ffi.c_long], ffi.c_int32)
    _curl_global_cleanup = _curl.curl_global_cleanup([], None)

    # --- Easy Handle Functions ---
    _curl_easy_init = _curl.curl_easy_init([], ffi.c_void_p)
    _curl_easy_cleanup = _curl.curl_easy_cleanup([ffi.c_void_p], None)
    _curl_easy_perform = _curl.curl_easy_perform([ffi.c_void_p], ffi.c_int32)
    _curl_easy_strerror = _curl.curl_easy_strerror([ffi.c_int32], ffi.c_char_p)

    # --- Setopt Variants ---
    _setopt_ptr = _curl.curl_easy_setopt([ffi.c_void_p, ffi.c_int32, ffi.c_void_p], ffi.c_int32)
    _setopt_string = _curl.curl_easy_setopt([ffi.c_void_p, ffi.c_int32, ffi.c_char_p], ffi.c_int32)
    _setopt_long = _curl.curl_easy_setopt([ffi.c_void_p, ffi.c_int32, ffi.c_long], ffi.c_int32)

    # --- Getinfo Variants ---
    _getinfo_ptr = _curl.curl_easy_getinfo([ffi.c_void_p, ffi.c_int32, ffi.POINTER(ffi.c_void_p)], ffi.c_int32)
    _getinfo_string = _curl.curl_easy_getinfo([ffi.c_void_p, ffi.c_int32, ffi.POINTER(ffi.c_char_p)], ffi.c_int32)
    _getinfo_long = _curl.curl_easy_getinfo([ffi.c_void_p, ffi.c_int32, ffi.POINTER(ffi.c_long)], ffi.c_int32)
    _getinfo_double = _curl.curl_easy_getinfo([ffi.c_void_p, ffi.c_int32, ffi.POINTER(ffi.c_double)], ffi.c_int32)

    # --- Slist (for headers) ---
    _curl_slist_append = _curl.curl_slist_append([ffi.c_void_p, ffi.c_char_p], ffi.c_void_p)
    _curl_slist_free_all = _curl.curl_slist_free_all([ffi.c_void_p], None)

    # --- Multi Handle Functions ---
    _curl_multi_init = _curl.curl_multi_init([], ffi.c_void_p)
    _curl_multi_cleanup = _curl.curl_multi_cleanup([ffi.c_void_p], ffi.c_int32)
    _curl_multi_add_handle = _curl.curl_multi_add_handle([ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
    _curl_multi_remove_handle = _curl.curl_multi_remove_handle([ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
    _curl_multi_perform = _curl.curl_multi_perform([ffi.c_void_p, ffi.POINTER(ffi.c_int32)], ffi.c_int32)
    _curl_multi_info_read = _curl.curl_multi_info_read([ffi.c_void_p, ffi.POINTER(ffi.c_int32)], ffi.c_void_p)
else:
    def _dummy_func(*args, **kwargs):
        raise CurlError(-1, "libcurl is not available on this system.")
    _curl_easy_init = _curl_easy_cleanup = _setopt_ptr = _setopt_string = _setopt_long = _dummy_func


# ==============================================================================
#  Constants and Exception Classes
# ==============================================================================

class CurlCode:
    OK = 0
    AGAIN = 81 # Not an error, used by multi_perform

class CurlOpt:
    URL = 10002
    WRITEDATA = 10001
    WRITEFUNCTION = 20011
    HEADERFUNCTION = 20079
    HEADERDATA = 10029
    CUSTOMREQUEST = 10036
    HTTPHEADER = 10023
    POSTFIELDS = 10015
    POSTFIELDSIZE = 60
    FOLLOWLOCATION = 52
    TIMEOUT = 13

class CurlInfo:
    RESPONSE_CODE = 0x200000 + 2
    EFFECTIVE_URL = 0x100000 + 1
    TOTAL_TIME = 0x300000 + 5

class CurlError(Exception):
    def __init__(self, code, message=""):
        self.code = code
        if not message and CURL_AVAILABLE:
            err_str_ptr = _curl_easy_strerror(code)
            self.message = ffi.string_at(err_str_ptr)
        else:
            self.message = message
        super().__init__(format_str("[Code {self.code}] {self.message}"))


# ==============================================================================
#  Response and Internal Data Buffer
# ==============================================================================

class Response:
    def __init__(self):
        self.status_code = 0
        self.headers = {}
        self._content_buffer = None
        self.url = ""
        self.reason = ""
        self.elapsed = 0.0
        self._easy_handle = None

    @property
    def text(self):
        return self.content.decode('utf-8')

    @property
    def content(self):
        if self._content_buffer is None:
            if self._easy_handle is not None:
                self._content_buffer = b"".join(list(self.iter_content()))
            else:
                self._content_buffer = b''
        return self._content_buffer

    def json(self):
        return _json.loads(self.text)

    def ok(self):
        return self.status_code < 400

    def __repr__(self):
        return format_str("<Response [{self.status_code}]>")

    def iter_content(self, chunk_size=8192):
        if self._easy_handle is None:
            if self._content_buffer is not None:
                yield self._content_buffer
            return None

        for chunk in self._easy_handle.iter_content(chunk_size):
            yield chunk

    def close(self):
        if self._easy_handle is not None:
            for chunk in self.iter_content():
                pass
            self._easy_handle.cleanup()
            self._easy_handle = None

class _DataBuffer:
    def __init__(self):
        self.chunks = []
        self.size = 0

    def write(self, chunk):
        self.chunks.append(chunk)
        self.size = self.size + len(chunk)

    def get_bytes(self):
        return b"".join(self.chunks)

# ==============================================================================
#  Core cURL Wrapper Classes
# ==============================================================================

class _CurlEasy:
    def __init__(self):
        self.handle = _curl_easy_init()
        if not self.handle:
            raise CurlError(-1, "curl_easy_init() failed.")
        
        self.response = Response()
        self.body_buffer = _DataBuffer()
        self.header_buffer = _DataBuffer()
        
        self._write_cb = None
        self._header_cb = None
        
        self._setup_callbacks()

    def _setup_callbacks(self):
        def write_callback(ptr, size, nmemb, userdata):
            chunk_size = size * nmemb
            chunk = ffi.buffer_to_bytes(ptr, chunk_size)
            self.body_buffer.write(chunk)
            return chunk_size

        def header_callback(ptr, size, nmemb, userdata):
            chunk_size = size * nmemb
            chunk = ffi.buffer_to_bytes(ptr, chunk_size)
            self.header_buffer.write(chunk)
            return chunk_size

        self._write_cb = ffi.callback(write_callback, ffi.c_int64, [ffi.c_void_p, ffi.c_int64, ffi.c_int64, ffi.c_void_p])
        self._header_cb = ffi.callback(header_callback, ffi.c_int64, [ffi.c_void_p, ffi.c_int64, ffi.c_int64, ffi.c_void_p])

        self._setopt(CurlOpt.WRITEFUNCTION, self._write_cb)
        self._setopt(CurlOpt.HEADERFUNCTION, self._header_cb)

    def _setopt(self, option, value):
        res = -1
        if hasattr(value, 'is_callback'):
             res = _setopt_ptr(self.handle, option, value)
        elif isinstance(value, str):
            res = _setopt_string(self.handle, option, value)
        elif isinstance(value, bytes):
            res = _setopt_string(self.handle, option, value)
        elif isinstance(value, int):
            res = _setopt_long(self.handle, option, value)
        elif isinstance(value, ffi.Pointer):
             res = _setopt_ptr(self.handle, option, value)
        else:
            raise TypeError(format_str("Unsupported type for setopt: {type(value)}"))
        
        if res != CurlCode.OK:
            raise CurlError(res)

    def _getinfo(self, info_code, typ):
        ptr_to_result = ffi.malloc(typ.Size())
        try:
            res = -1
            if typ == ffi.c_long:
                res = _getinfo_long(self.handle, info_code, ptr_to_result)
            elif typ == ffi.c_double:
                res = _getinfo_double(self.handle, info_code, ptr_to_result)
            elif typ == ffi.c_char_p:
                res = _getinfo_string(self.handle, info_code, ptr_to_result)
            else:
                raise TypeError(format_str("Unsupported type for getinfo: {typ.Inspect()}"))

            if res != CurlCode.OK:
                return None
            
            if typ == ffi.c_char_p:
                string_ptr = ffi.read_memory(ptr_to_result, ffi.c_void_p)
                if string_ptr is None or string_ptr.Address == 0:
                    return ""
                return ffi.string_at(string_ptr)
            else:
                pylearn_obj = ffi.read_memory(ptr_to_result, typ)
                if pylearn_obj is None:
                    return None
                # str(pylearn_obj) calls its Inspect(), returning "1".
                # int("1") returns the number 1.
                if typ in (ffi.c_long, ffi.c_int32):
                    return int(str(pylearn_obj))
                if typ == ffi.c_double:
                    return float(str(pylearn_obj))
                if hasattr(pylearn_obj, 'Value'):
                    print("DEBUGLINE: ",pylearn_obj)
                    return pylearn_obj.Value
                return pylearn_obj
        finally:
            ffi.free(ptr_to_result)
            
    def perform(self):
        res = _curl_easy_perform(self.handle)
        if res != CurlCode.OK:
            raise CurlError(res)

        self.response.status_code = self._getinfo(CurlInfo.RESPONSE_CODE, ffi.c_long)
        self.response.url = self._getinfo(CurlInfo.EFFECTIVE_URL, ffi.c_char_p)
        self.response.elapsed = self._getinfo(CurlInfo.TOTAL_TIME, ffi.c_double)
        self.response._content_buffer = self.body_buffer.get_bytes()
        self._parse_headers()
        return self.response

    def _parse_headers(self):
        header_text = self.header_buffer.get_bytes().decode('utf-8')
        lines = header_text.split('\r\n')
        for line in lines:
            if ':' in line:
                key_value = line.split(':', 1)
                key = key_value[0].strip()
                value = key_value[1].strip()
                self.response.headers[key.strip()] = value.strip()
            elif line.startswith('HTTP/'):
                parts = line.split(' ', 2)
                if len(parts) > 2:
                    self.response.reason = parts[2]

    def cleanup(self):
        if self.handle:
            _curl_easy_cleanup(self.handle)
            ffi.free_callback(self._write_cb)
            ffi.free_callback(self._header_cb)
            self.handle = None

class _CurlStreamer(_CurlEasy):
    def __init__(self):
        self.handle = _curl_easy_init()
        if not self.handle:
            raise CurlError(-1, "curl_easy_init() failed.")
        
        self.response = Response()
        self.body_buffer = _DataBuffer()
        self.header_buffer = _DataBuffer()
        
        self.multi_handle = _curl_multi_init()
        _curl_multi_add_handle(self.multi_handle, self.handle)
        
        self._headers_ready = False
        self._is_finished = False
        
        self._setup_streaming_callbacks()

    def _setup_streaming_callbacks(self):
        def write_callback(ptr, size, nmemb, userdata):
            chunk_size = size * nmemb
            chunk = ffi.buffer_to_bytes(ptr, chunk_size)
            self.body_buffer.write(chunk)
            return chunk_size

        def header_callback(ptr, size, nmemb, userdata):
            chunk_size = size * nmemb
            chunk = ffi.buffer_to_bytes(ptr, chunk_size)
            self.header_buffer.write(chunk)
            if chunk == b'\r\n':
                self._headers_ready = True
            return chunk_size

        self._write_cb = ffi.callback(write_callback, ffi.c_int64, [ffi.c_void_p, ffi.c_int64, ffi.c_int64, ffi.c_void_p])
        self._header_cb = ffi.callback(header_callback, ffi.c_int64, [ffi.c_void_p, ffi.c_int64, ffi.c_int64, ffi.c_void_p])

        self._setopt(CurlOpt.WRITEFUNCTION, self._write_cb)
        self._setopt(CurlOpt.HEADERFUNCTION, self._header_cb)

    def _wait_for_activity(self):
        time.sleep(0.001)

    def _receive_headers_and_populate_response(self):
        still_running_ptr = ffi.malloc(ffi.c_int32.Size())
        try:
            while not self._headers_ready:
                res = _curl_multi_perform(self.multi_handle, still_running_ptr)
                if res != CurlCode.OK and res != CurlCode.AGAIN:
                    raise CurlError(res)
                
                still_running_obj = ffi.read_memory(still_running_ptr, ffi.c_int32)
                # print("STILL RUNNING DEBUGLINE: ",still_running_obj)
                if still_running_obj == 0: # <<< FIX
                    self._is_finished = True
                    break
                self._wait_for_activity()
            
            self.response.status_code = self._getinfo(CurlInfo.RESPONSE_CODE, ffi.c_long)
            self.response.url = self._getinfo(CurlInfo.EFFECTIVE_URL, ffi.c_char_p)
            self.response.elapsed = self._getinfo(CurlInfo.TOTAL_TIME, ffi.c_double)
            self._parse_headers()
        finally:
            ffi.free(still_running_ptr)

    def iter_content(self, chunk_size=8192):
        try:
            if self.body_buffer.size > 0:
                yield self.body_buffer.get_bytes()
                self.body_buffer = _DataBuffer()

            still_running_ptr = ffi.malloc(ffi.c_int32.Size())
            try:
                while not self._is_finished:
                    res = _curl_multi_perform(self.multi_handle, still_running_ptr)
                    if res != CurlCode.OK and res != CurlCode.AGAIN:
                        raise CurlError(res)

                    if self.body_buffer.size > 0:
                        yield self.body_buffer.get_bytes()
                        self.body_buffer = _DataBuffer()
                    
                    still_running_obj = ffi.read_memory(still_running_ptr, ffi.c_int32)
                    if still_running_obj == 0: # <<< FIX
                        self._is_finished = True
                        msgs_in_queue_ptr = ffi.malloc(ffi.c_int32.Size())
                        try:
                            while True:
                                msg_ptr = _curl_multi_info_read(self.multi_handle, msgs_in_queue_ptr)
                                if not msg_ptr.Address:
                                    break
                        finally:
                            ffi.free(msgs_in_queue_ptr)
                    
                    if not self._is_finished:
                        self._wait_for_activity()
            finally:
                ffi.free(still_running_ptr)
        finally:
            self.cleanup()

    def cleanup(self):
        if self.multi_handle:
            if self.handle:
                _curl_multi_remove_handle(self.multi_handle, self.handle)
            _curl_multi_cleanup(self.multi_handle)
            self.multi_handle = None
        super().cleanup()

# ==============================================================================
#  High-Level API Functions
# ==============================================================================
def request(method, url, headers=None, data=None, json=None, timeout=30, stream=False):
    if not CURL_AVAILABLE:
        raise CurlError(-1, "libcurl is not available.")
    
    easy = _CurlStreamer() if stream else _CurlEasy()
    slist = None

    try:
        easy._setopt(CurlOpt.URL, url)
        easy._setopt(CurlOpt.TIMEOUT, int(timeout)) # Timeout must be an integer
        easy._setopt(CurlOpt.FOLLOWLOCATION, 1)
        easy._setopt(CurlOpt.CUSTOMREQUEST, method.upper())

        has_content_type = False
        if headers:
            for key, value in headers.items():
                if key.lower() == 'content-type':
                    has_content_type = True
                slist = _curl_slist_append(slist, format_str("{key}: {value}"))
            easy._setopt(CurlOpt.HTTPHEADER, slist)

        if data is not None:
            body = data if isinstance(data, bytes) else data.encode('utf-8')
            easy._setopt(CurlOpt.POSTFIELDS, body)
            easy._setopt(CurlOpt.POSTFIELDSIZE, len(body))
        elif json is not None:
            if not has_content_type:
                slist = _curl_slist_append(slist, "Content-Type: application/json")
                easy._setopt(CurlOpt.HTTPHEADER, slist)
            
            body = _json.dumps(json).encode('utf-8')
            easy._setopt(CurlOpt.POSTFIELDS, body)
            easy._setopt(CurlOpt.POSTFIELDSIZE, len(body))
        
        if stream:
            easy._receive_headers_and_populate_response()
            easy.response._easy_handle = easy
            return easy.response
        else:
            response = easy.perform()
            easy.cleanup()
            return response

    except Exception as e:
        easy.cleanup()
        raise e
    finally:
        if slist:
            _curl_slist_free_all(slist)

def get(url, stream=False, **kwargs):
    return request('GET', url, stream=stream, **kwargs)

def post(url, data=None, json=None, stream=False, **kwargs):
    return request('POST', url, data=data, json=json, stream=stream, **kwargs)

def put(url, data=None, stream=False, **kwargs):
    return request('PUT', url, data=data, stream=stream, **kwargs)

def delete(url, **kwargs):
    return request('DELETE', url, **kwargs)

# ==============================================================================
#  Global Setup and Cleanup
# ==============================================================================
if CURL_AVAILABLE:
    CURL_GLOBAL_DEFAULT = 3
    res = _curl_global_init(CURL_GLOBAL_DEFAULT)
    if res != CurlCode.OK:
        print(format_str("Warning: curl_global_init() failed with code {res}"))