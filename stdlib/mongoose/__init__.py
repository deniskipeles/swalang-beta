"""
mongoose.py — Production-ready Mongoose networking bindings for Swalang.

Covers:
  - HTTP/HTTPS server and client
  - WebSocket server and client
  - Static file serving  (mg_http_serve_dir)
  - URI glob matching    (mg_http_match_uri)
  - URL decoding         (pure-Swalang fallback)
  - TLS / HTTPS          (mg_tls_init, built-in Mongoose TLS)

Memory layout notes (Mongoose 7.x / 8.x, 64-bit little-endian):
  mg_str  = { char *ptr [8], size_t len [8] }  → 16 bytes
  mg_http_message headers start at offset 112, each header = 32 bytes (2× mg_str)
"""

import ffi
import sys

# ==============================================================================
#  Library Loading
# ==============================================================================

def _load_library():
    platform = sys.platform
    candidates = []
    if platform == 'linux':
        candidates = [
            "bin/x86_64-linux/mongoose/libmongoose.so",
            "libmongoose.so",
        ]
    elif platform == 'windows':
        candidates = [
            "bin/x86_64-windows-gnu/mongoose/mongoose.dll",
            "mongoose.dll",
        ]
    elif platform == 'darwin':
        candidates = ["libmongoose.dylib"]

    for name in candidates:
        try:
            return ffi.CDLL(name)
        except ffi.FFIError:
            pass
    raise ffi.FFIError("Could not load Mongoose shared library")

_lib = _load_library()

MG_EV_ERROR           = 0
MG_EV_OPEN            = 1
MG_EV_POLL            = 2
MG_EV_RESOLVE         = 3
MG_EV_CONNECT         = 4
MG_EV_ACCEPT          = 5
MG_EV_TLS_HANDSHAKING = 6
MG_EV_READ            = 7
MG_EV_WRITE           = 8
MG_EV_CLOSE           = 9
MG_EV_HTTP_MSG        = 10
MG_EV_HTTP_CHUNK      = 11
MG_EV_WS_OPEN         = 12
MG_EV_WS_MSG          = 13
MG_EV_WS_CTL          = 14

# WebSocket opcodes
WEBSOCKET_OP_CONTINUE = 0
WEBSOCKET_OP_TEXT     = 1
WEBSOCKET_OP_BINARY   = 2
WEBSOCKET_OP_CLOSE    = 8
WEBSOCKET_OP_PING     = 9
WEBSOCKET_OP_PONG     = 10

# ==============================================================================
#  C Function Signatures
# ==============================================================================

# Manager lifecycle
_mg_mgr_init = _lib.mg_mgr_init([ffi.c_void_p], None)
_mg_mgr_free = _lib.mg_mgr_free([ffi.c_void_p], None)
_mg_mgr_poll = _lib.mg_mgr_poll([ffi.c_void_p, ffi.c_int32], None)

# HTTP server / client
_mg_http_listen  = _lib.mg_http_listen( [ffi.c_void_p, ffi.c_char_p, ffi.c_void_p, ffi.c_void_p], ffi.c_void_p)
_mg_http_connect = _lib.mg_http_connect([ffi.c_void_p, ffi.c_char_p, ffi.c_void_p, ffi.c_void_p], ffi.c_void_p)

# mg_http_reply is variadic: (conn, status, headers_fmt, body_fmt, body_arg)
_mg_http_reply = _lib.mg_http_reply([ffi.c_void_p, ffi.c_int32, ffi.c_char_p, ffi.c_char_p], None, is_variadic=True)

# Raw send
_mg_send = _lib.mg_send([ffi.c_void_p, ffi.c_void_p, ffi.c_int32], None)

# WebSocket
_mg_ws_connect = _lib.mg_ws_connect([ffi.c_void_p, ffi.c_char_p, ffi.c_void_p, ffi.c_void_p, ffi.c_char_p], ffi.c_void_p, is_variadic=True)
_mg_ws_send    = _lib.mg_ws_send([ffi.c_void_p, ffi.c_void_p, ffi.c_uint64, ffi.c_int32], None)
_mg_ws_upgrade = _lib.mg_ws_upgrade([ffi.c_void_p, ffi.c_void_p, ffi.c_char_p], None, is_variadic=True)

# Static file serving
# void mg_http_serve_dir (mg_connection*, mg_http_message*, const mg_http_serve_opts*)
# void mg_http_serve_file(mg_connection*, mg_http_message*, const char*, const mg_http_serve_opts*)
_mg_http_serve_dir  = _lib.mg_http_serve_dir( [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], None)
_mg_http_serve_file = _lib.mg_http_serve_file([ffi.c_void_p, ffi.c_void_p, ffi.c_char_p, ffi.c_void_p], None)

# TLS — void mg_tls_init(mg_connection*, const mg_tls_opts*)
# Optional: only present when Mongoose is compiled with MG_ENABLE_MBEDTLS / MG_ENABLE_OPENSSL
try:
    _mg_tls_init  = _lib.mg_tls_init([ffi.c_void_p, ffi.c_void_p], None)
    TLS_AVAILABLE = True
except ffi.FFIError:
    _mg_tls_init  = None
    TLS_AVAILABLE = False

# ==============================================================================
#  mg_str / memory helpers
# ==============================================================================

def _is_valid_ptr(ptr):
    return ptr and getattr(ptr, "Address", None)

def _read_mg_str(base_ptr, offset):
    """Read a mg_str at base_ptr+offset and return as Python str."""
    str_ptr = ffi.read_memory_with_offset(base_ptr, offset,     ffi.c_void_p)
    str_len = ffi.read_memory_with_offset(base_ptr, offset + 8, ffi.c_uint64)
    if not _is_valid_ptr(str_ptr) or str_len == 0:
        return ""
    return ffi.string_at(str_ptr, str_len)

def _read_mg_bytes(base_ptr, offset):
    """Read a mg_str at base_ptr+offset and return as Python bytes."""
    str_ptr = ffi.read_memory_with_offset(base_ptr, offset,     ffi.c_void_p)
    str_len = ffi.read_memory_with_offset(base_ptr, offset + 8, ffi.c_uint64)
    if not _is_valid_ptr(str_ptr) or str_len == 0:
        return b""
    return ffi.buffer_to_bytes(str_ptr, str_len)

# ==============================================================================
#  Struct Builders
# ==============================================================================

# ---- mg_http_serve_opts -------------------------------------------------------
# struct mg_http_serve_opts {
#   const char *root_dir;       // offset  0
#   const char *ssi_pattern;    // offset  8  (NULL = disabled)
#   const char *extra_headers;  // offset 16
#   const char *mime_types;     // offset 24  (NULL = built-ins only)
#   const char *page404;        // offset 32  (NULL = default)
#   struct mg_fs *fs;           // offset 40  (NULL = posix fs)
# };
_SERVE_OPTS_SIZE = 64   # padded generously

def _make_serve_opts(root_dir, extra_headers="", mime_types="", page404=""):
    """
    Allocate and populate a mg_http_serve_opts on the heap.
    Returns (opts_ptr, list_of_bufs_to_free).
    Caller is responsible for calling ffi.free on all returned buffers.
    """
    opts = ffi.malloc(_SERVE_OPTS_SIZE)
    bufs = []

    def _put_str(offset, s):
        if s:
            b = s.encode('utf-8') + b'\x00'
            buf = ffi.malloc(len(b))
            ffi.memcpy(buf, ffi.addressof(b), len(b))
            ffi.write_memory_with_offset(opts, offset, ffi.c_void_p, buf)
            bufs.append(buf)
        else:
            # Write null pointer
            ffi.write_memory_with_offset(opts, offset, ffi.c_uint64, 0)

    _put_str(0,  root_dir)
    _put_str(8,  "")            # ssi_pattern — SSI disabled
    _put_str(16, extra_headers)
    _put_str(24, mime_types)
    _put_str(32, page404)
    ffi.write_memory_with_offset(opts, 40, ffi.c_uint64, 0)  # fs = NULL (POSIX)

    return (opts, bufs)


def _free_serve_opts(opts_ptr, bufs):
    for b in bufs:
        ffi.free(b)
    ffi.free(opts_ptr)


# ---- mg_tls_opts -------------------------------------------------------------
# struct mg_tls_opts {
#   const char *ca;     // offset  0  — CA cert PEM or file path
#   const char *cert;   // offset  8  — server/client cert PEM or file path
#   const char *key;    // offset 16  — private key PEM or file path
#   const char *name;   // offset 24  — SNI hostname (client) or NULL (server)
# };
_TLS_OPTS_SIZE = 64   # padded generously

def _make_tls_opts(cert=None, key=None, ca=None, name=None):
    """
    Allocate and populate a mg_tls_opts on the heap.
    Returns (opts_ptr, list_of_bufs_to_free).
    """
    opts = ffi.malloc(_TLS_OPTS_SIZE)
    bufs = []

    def _put_str(offset, s):
        if s:
            b = s.encode('utf-8') + b'\x00'
            buf = ffi.malloc(len(b))
            ffi.memcpy(buf, ffi.addressof(b), len(b))
            ffi.write_memory_with_offset(opts, offset, ffi.c_void_p, buf)
            bufs.append(buf)
        else:
            ffi.write_memory_with_offset(opts, offset, ffi.c_uint64, 0)

    _put_str(0,  ca)
    _put_str(8,  cert)
    _put_str(16, key)
    _put_str(24, name)

    return (opts, bufs)


def _free_tls_opts(opts_ptr, bufs):
    for b in bufs:
        ffi.free(b)
    ffi.free(opts_ptr)

# ==============================================================================
#  URL Decode (pure-Swalang, no extra C binding needed)
# ==============================================================================

def url_decode(s, is_form=True):
    """
    URL-decode a string.
    is_form=True  →  '+' is decoded as space (application/x-www-form-urlencoded).
    is_form=False →  '+' is left as-is.
    """
    if isinstance(s, bytes):
        src = s
    else:
        src = s.encode('utf-8')

    result_chars = []
    i = 0
    while i < len(src):
        c = src[i]
        if is_form and c == 43:       # ord('+') == 43
            result_chars.append(32)   # space
            i = i + 1
        elif c == 37 and i + 2 < len(src):   # ord('%') == 37
            high = src[i + 1]
            low  = src[i + 2]
            h = _hex_nibble(high)
            l = _hex_nibble(low)
            if h >= 0 and l >= 0:
                result_chars.append(h * 16 + l)
                i = i + 3
            else:
                result_chars.append(c)
                i = i + 1
        else:
            result_chars.append(c)
            i = i + 1

    return bytes(result_chars).decode('utf-8', errors='replace')


def _hex_nibble(b):
    """Convert a hex ASCII byte to its integer value, or -1 on failure."""
    if b >= 48 and b <= 57:    # '0'-'9'
        return b - 48
    if b >= 65 and b <= 70:    # 'A'-'F'
        return b - 55
    if b >= 97 and b <= 102:   # 'a'-'f'
        return b - 87
    return -1

# ==============================================================================
#  HttpMessage
# ==============================================================================

class HttpMessage:
    """
    Wraps a Mongoose mg_http_message* pointer.

    mg_http_message layout (Mongoose 7.15+, MG_MAX_HTTP_HEADERS=30):
      offset   0 : method
      offset  16 : uri
      offset  32 : query
      offset  48 : proto
      offset  64 : headers[0..29], each header = name(16) + value(16) = 32 bytes
      offset 1024: body
      offset 1040: head     (request + headers)
      offset 1056: chunk
      offset 1072: message  (request + headers + body)
    """
    def __init__(self, ptr):
        self.ptr = ptr

    @property
    def method(self):    return _read_mg_str(self.ptr, 0)
    @property
    def uri(self):       return _read_mg_str(self.ptr, 16)
    @property
    def query(self):     return _read_mg_str(self.ptr, 32)
    @property
    def proto(self):     return _read_mg_str(self.ptr, 48)

    def get_headers(self):
        """Return all HTTP headers as a lowercase-keyed dict."""
        headers = {}
        off = 64
        for _ in range(30):
            name = _read_mg_str(self.ptr, off)
            if not name:
                break
            value = _read_mg_str(self.ptr, off + 16)
            headers[name.lower()] = value
            off = off + 32
        return headers

    def get_header(self, name):
        """Get a single header value by name (case-insensitive). Returns '' if absent."""
        target = name.lower()
        off = 64
        for _ in range(30):
            h = _read_mg_str(self.ptr, off)
            if not h:
                break
            if h.lower() == target:
                return _read_mg_str(self.ptr, off + 16)
            off = off + 32
        return ""

    @property
    def body(self):      return _read_mg_str(self.ptr, 1024)
    @property
    def body_bytes(self):return _read_mg_bytes(self.ptr, 1024)
    @property
    def head(self):      return _read_mg_str(self.ptr, 1040)
    @property
    def chunk(self):     return _read_mg_str(self.ptr, 1056)
    @property
    def message(self):   return _read_mg_str(self.ptr, 1072)

    def match_uri(self, glob):
        """
        Return True if the request URI matches `glob`.
        Mongoose glob syntax: '*' matches anything, '?' matches one character.
        Example: msg.match_uri('/api/*')
        """
        return _mg_http_match_uri(self.ptr, glob.encode('utf-8'))

# ==============================================================================
#  WsMessage
# ==============================================================================

class WsMessage:
    """
    Wraps a Mongoose mg_ws_message* pointer.

    Layout:
      offset  0 : data  (mg_str)
      offset 16 : flags (uint8) — low nibble = opcode
    """
    def __init__(self, ptr):
        self.ptr = ptr

    @property
    def data(self):       return _read_mg_str(self.ptr, 0)
    @property
    def data_bytes(self): return _read_mg_bytes(self.ptr, 0)
    @property
    def flags(self):      return ffi.read_memory_with_offset(self.ptr, 16, ffi.c_uint8)
    @property
    def op(self):         return self.flags & 0x0F
    @property
    def is_text(self):    return self.op == WEBSOCKET_OP_TEXT
    @property
    def is_binary(self):  return self.op == WEBSOCKET_OP_BINARY
    @property
    def is_ping(self):    return self.op == WEBSOCKET_OP_PING
    @property
    def is_close(self):   return self.op == WEBSOCKET_OP_CLOSE

# ==============================================================================
#  Manager
# ==============================================================================

class Manager:
    """
    Wraps a Mongoose mg_mgr event manager.

    A Manager drives one or more connections (server listeners, client
    connections, WebSocket channels) via a single poll() call.

    All allocated callbacks and TLS buffers are tracked here and freed
    when free() is called.
    """

    def __init__(self):
        self.mgr_ptr   = ffi.malloc(1024)
        _mg_mgr_init(self.mgr_ptr)
        self.callbacks  = []   # keep C callbacks alive
        self._tls_store = []   # keep TLS struct buffers alive

    # ------------------------------------------------------------------
    # Event loop
    # ------------------------------------------------------------------

    def poll(self, ms):
        """Drive the event loop for up to `ms` milliseconds."""
        _mg_mgr_poll(self.mgr_ptr, ms)

    # ------------------------------------------------------------------
    # HTTP listener
    # ------------------------------------------------------------------

    def http_listen(self, url, handler_func, tls_cert=None, tls_key=None, tls_ca=None):
        """
        Start an HTTP (plain) or HTTPS (TLS) listener.

        handler_func(conn_ptr, event_str, data) where event_str is one of:
          "HTTP_REQUEST"  — data is HttpMessage
          "WS_OPEN"       — data is None  (after ws_upgrade)
          "WS_MSG"        — data is WsMessage
          "CLOSE"         — data is None
          "ERROR"         — data is str error message

        For HTTPS pass tls_cert (PEM string or file path) and tls_key.
        tls_ca is optional (CA bundle for mutual TLS).

        Raises if the port cannot be bound.
        """
        tls_opts_ptr = None

        if tls_cert and tls_key:
            if not TLS_AVAILABLE:
                raise Exception("TLS requested but Mongoose was not compiled with TLS support")
            opts_ptr, bufs = _make_tls_opts(cert=tls_cert, key=tls_key, ca=tls_ca)
            tls_opts_ptr = opts_ptr
            self._tls_store.append(opts_ptr)
            for b in bufs:
                self._tls_store.append(b)

        def c_handler(c_ptr, ev, ev_data, fn_data):
            try:
                safe = _is_valid_ptr(ev_data)

                # TLS handshake on every newly accepted connection
                if ev == MG_EV_ACCEPT and tls_opts_ptr:
                    _mg_tls_init(c_ptr, tls_opts_ptr)
                    return None

                if ev == MG_EV_HTTP_MSG and safe:
                    handler_func(c_ptr, "HTTP_REQUEST", HttpMessage(ev_data))
                elif ev == MG_EV_WS_OPEN:
                    handler_func(c_ptr, "WS_OPEN", None)
                elif ev == MG_EV_WS_MSG and safe:
                    handler_func(c_ptr, "WS_MSG", WsMessage(ev_data))
                elif ev == MG_EV_CLOSE:
                    handler_func(c_ptr, "CLOSE", None)
                elif ev == MG_EV_ERROR and safe:
                    handler_func(c_ptr, "ERROR", ffi.string_at(ev_data))
            except Exception as e:
                print(format_str("🔥 [mongoose] Listen handler error: {e}"))

        cb = ffi.callback(c_handler, None, [ffi.c_void_p, ffi.c_int32, ffi.c_void_p, ffi.c_void_p])
        self.callbacks.append(cb)

        conn = _mg_http_listen(self.mgr_ptr, url.encode('utf-8'), cb, None)
        if not _is_valid_ptr(conn):
            raise Exception(format_str("mg_http_listen failed for '{url}' — port in use or bad address?"))

        print(format_str("🚀 Mongoose listening on {url}"))
        return conn

    # ------------------------------------------------------------------
    # HTTP client
    # ------------------------------------------------------------------

    def http_connect(self, url, handler_func, tls_ca=None):
        """
        Open an HTTP (or HTTPS) client connection.

        handler_func(conn_ptr, event_str, data) where event_str is one of:
          "CONNECT"   — data is None  (TCP connected; send your request here)
          "RESPONSE"  — data is HttpMessage
          "ERROR"     — data is str
          "CLOSE"     — data is None

        For HTTPS the TLS handshake is initiated automatically on CONNECT.
        tls_ca is an optional CA certificate PEM/path for peer verification.
        If tls_ca is None the connection uses the Mongoose default (no verification).
        """
        tls_opts_ptr = None

        if url.startswith("https://"):
            if not TLS_AVAILABLE:
                raise Exception("HTTPS requested but Mongoose was not compiled with TLS support")
            hostname = _extract_hostname(url)
            opts_ptr, bufs = _make_tls_opts(ca=tls_ca, name=hostname)
            tls_opts_ptr = opts_ptr
            self._tls_store.append(opts_ptr)
            for b in bufs:
                self._tls_store.append(b)

        def c_handler(c_ptr, ev, ev_data, fn_data):
            try:
                safe = _is_valid_ptr(ev_data)

                if ev == MG_EV_CONNECT:
                    if tls_opts_ptr:
                        _mg_tls_init(c_ptr, tls_opts_ptr)
                    handler_func(c_ptr, "CONNECT", None)
                elif ev == MG_EV_HTTP_MSG and safe:
                    handler_func(c_ptr, "RESPONSE", HttpMessage(ev_data))
                elif ev == MG_EV_ERROR:
                    msg = ffi.string_at(ev_data) if safe else "unknown error"
                    handler_func(c_ptr, "ERROR", msg)
                elif ev == MG_EV_CLOSE:
                    handler_func(c_ptr, "CLOSE", None)
            except Exception as e:
                print(format_str("🔥 [mongoose] Connect handler error: {e}"))

        cb = ffi.callback(c_handler, None, [ffi.c_void_p, ffi.c_int32, ffi.c_void_p, ffi.c_void_p])
        self.callbacks.append(cb)

        conn = _mg_http_connect(self.mgr_ptr, url.encode('utf-8'), cb, None)
        if not _is_valid_ptr(conn):
            raise Exception(format_str("mg_http_connect failed for '{url}'"))
        return conn

    # ------------------------------------------------------------------
    # WebSocket client
    # ------------------------------------------------------------------

    def ws_connect(self, url, handler_func, tls_ca=None):
        """
        Open a WebSocket client connection (ws:// or wss://).

        handler_func(conn_ptr, event_str, data) where event_str is one of:
          "OPEN"    — data is None  (TCP connected)
          "WS_OPEN" — data is None  (WS handshake complete)
          "WS_MSG"  — data is WsMessage
          "ERROR"   — data is str
          "CLOSE"   — data is None
        """
        tls_opts_ptr = None

        if url.startswith("wss://"):
            if not TLS_AVAILABLE:
                raise Exception("WSS requested but Mongoose was not compiled with TLS support")
            hostname = _extract_hostname(url)
            opts_ptr, bufs = _make_tls_opts(ca=tls_ca, name=hostname)
            tls_opts_ptr = opts_ptr
            self._tls_store.append(opts_ptr)
            for b in bufs:
                self._tls_store.append(b)

        def c_handler(c_ptr, ev, ev_data, fn_data):
            try:
                safe = _is_valid_ptr(ev_data)

                if ev == MG_EV_CONNECT and tls_opts_ptr:
                    _mg_tls_init(c_ptr, tls_opts_ptr)
                elif ev == MG_EV_OPEN:
                    handler_func(c_ptr, "OPEN", None)
                elif ev == MG_EV_WS_OPEN:
                    handler_func(c_ptr, "WS_OPEN", None)
                elif ev == MG_EV_WS_MSG and safe:
                    handler_func(c_ptr, "WS_MSG", WsMessage(ev_data))
                elif ev == MG_EV_ERROR:
                    msg = ffi.string_at(ev_data) if safe else "unknown error"
                    handler_func(c_ptr, "ERROR", msg)
                elif ev == MG_EV_CLOSE:
                    handler_func(c_ptr, "CLOSE", None)
            except Exception as e:
                print(format_str("🔥 [mongoose] WS handler error: {e}"))

        cb = ffi.callback(c_handler, None, [ffi.c_void_p, ffi.c_int32, ffi.c_void_p, ffi.c_void_p])
        self.callbacks.append(cb)

        conn = _mg_ws_connect(self.mgr_ptr, url.encode('utf-8'), cb, None, b"")
        if not _is_valid_ptr(conn):
            raise Exception(format_str("mg_ws_connect failed for '{url}'"))
        return conn

    # ------------------------------------------------------------------
    # Cleanup
    # ------------------------------------------------------------------

    def free(self):
        """Shut down the manager and release all C resources."""
        if self.mgr_ptr:
            _mg_mgr_free(self.mgr_ptr)
            ffi.free(self.mgr_ptr)
            self.mgr_ptr = None
        for cb in self.callbacks:
            ffi.free_callback(cb)
        self.callbacks = []
        for buf in self._tls_store:
            ffi.free(buf)
        self._tls_store = []

# ==============================================================================
#  Public Helpers
# ==============================================================================

def http_reply(conn_ptr, status_code, headers, body):
    """
    Send a complete HTTP response.

    headers — extra headers string, e.g. "Content-Type: text/html\r\n"
              (trailing \\r\\n is added automatically if missing)
    body    — str or bytes
    """
    if headers:
        if not headers.endswith("\r\n"):
            headers = headers + "\r\n"
        hdr_bytes = headers.encode('utf-8')
    else:
        hdr_bytes = b""

    if not isinstance(body, (str, bytes)):
        body = str(body)
    body_bytes = body.encode('utf-8') if isinstance(body, str) else body
    _mg_http_reply(conn_ptr, status_code, hdr_bytes, b"%s", body_bytes)


def send(conn_ptr, data):
    """Send raw bytes on a connection (bypasses HTTP framing)."""
    b = data.encode('utf-8') if isinstance(data, str) else data
    _mg_send(conn_ptr, b, len(b))


def ws_send(conn_ptr, data, op=WEBSOCKET_OP_TEXT):
    """Send a WebSocket frame. op defaults to text (1)."""
    b = data.encode('utf-8') if isinstance(data, str) else data
    _mg_ws_send(conn_ptr, b, len(b), op)


def ws_upgrade(conn_ptr, http_msg, extra_headers=""):
    """Upgrade an HTTP connection to WebSocket."""
    _mg_ws_upgrade(conn_ptr, http_msg.ptr, extra_headers.encode('utf-8'))


def http_serve_dir(conn_ptr, http_msg, root_dir, extra_headers="", mime_types="", page404=""):
    """
    Serve a file from `root_dir` matching the request URI via Mongoose's
    built-in static file server (handles Range, ETag, Gzip, MIME types).

    extra_headers — raw headers appended to every response, e.g.
                    "Cache-Control: max-age=3600\r\n"
    mime_types    — extra MIME overrides, e.g. "wasm=application/wasm"
    page404       — path inside root_dir to serve for 404s, e.g. "404.html"
    """
    opts_ptr, bufs = _make_serve_opts(root_dir, extra_headers, mime_types, page404)
    try:
        _mg_http_serve_dir(conn_ptr, http_msg.ptr, opts_ptr)
    except Exception:
        pass
    finally:
        _free_serve_opts(opts_ptr, bufs)


def http_serve_file(conn_ptr, http_msg, file_path, extra_headers="", mime_types=""):
    """Serve a single specific file."""
    opts_ptr, bufs = _make_serve_opts(".", extra_headers, mime_types)
    try:
        _mg_http_serve_file(conn_ptr, http_msg.ptr, file_path.encode('utf-8'), opts_ptr)
    except Exception:
        pass
    finally:
        _free_serve_opts(opts_ptr, bufs)

# ==============================================================================
#  Internal Utilities
# ==============================================================================

def _extract_hostname(url):
    """Extract bare hostname from a URL string (no port, no path)."""
    if "://" in url:
        after = url.split("://")[1]
        host_port = after.split("/")[0]
        return host_port.split(":")[0]
    return ""