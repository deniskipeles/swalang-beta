"""
libuv.py — Production-ready libuv binding for Swalang.

The libuv event loop powers Node.js, uvloop, Julia's networking stack, and more.
This wrapper exposes the full networking and async I/O surface:

  Loop          — the event loop (drive with loop.run())
  Timer         — one-shot or repeating async timer
  TCP           — TCP server (listen/accept) and client (connect)
  UDP           — UDP send/recv (unicast, broadcast, multicast)
  Pipe          — named pipe / UNIX domain socket
  Signal        — OS signal handler (SIGINT, SIGTERM, …)
  Idle          — callback fired every iteration when the loop has nothing to do
  Check         — callback fired once per loop iteration (after I/O)
  Prepare       — callback fired once per loop iteration (before I/O)
  dns_getaddrinfo  — async DNS resolution
  getaddrinfo   — synchronous DNS resolution (blocking helper)

Shutdown pattern:
  Always call handle.close() then loop.run(UV_RUN_NOWAIT) before loop.close().
  The helper shutdown_loop(loop) does this correctly.

Buffer ownership:
  All buffers passed to write/send are copied into libuv-owned memory
  by the uv_write / uv_udp_send calls.  You may free your buffer as soon
  as the call returns.

Error handling:
  libuv returns negative integers for errors.  This wrapper translates them
  to UvError exceptions.  uv_strerror() is used to obtain human-readable text.
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
            "bin/x86_64-linux/libuv/libuv.so",
            "libuv.so.1",
            "libuv.so",
        ]
    elif platform == 'windows':
        candidates = [
            "bin/x86_64-windows-gnu/libuv/uv.dll",
            "libuv.dll",
            "uv.dll",
        ]
    elif platform == 'darwin':
        candidates = ["libuv.dylib", "libuv.1.dylib"]

    for name in candidates:
        try:
            return ffi.CDLL(name)
        except ffi.FFIError:
            pass
    raise ffi.FFIError("Could not load libuv shared library")

_lib = _load_library()

# ==============================================================================
#  Constants
# ==============================================================================

# uv_run_mode
UV_RUN_DEFAULT = 0
UV_RUN_ONCE    = 1
UV_RUN_NOWAIT  = 2

# uv_handle_type  (from uv.h enum)
UV_UNKNOWN_HANDLE = 0
UV_ASYNC          = 1
UV_CHECK          = 2
UV_FS_EVENT       = 3
UV_FS_POLL        = 4
UV_HANDLE         = 5
UV_IDLE           = 6
UV_NAMED_PIPE     = 7
UV_POLL           = 8
UV_PREPARE        = 9
UV_PROCESS        = 10
UV_STREAM         = 11
UV_TCP            = 12
UV_TIMER          = 14
UV_TTY            = 15
UV_UDP            = 16
UV_SIGNAL         = 17
UV_FILE           = 18

# uv_req_type
UV_UNKNOWN_REQ    = 0
UV_REQ            = 1
UV_CONNECT        = 2
UV_WRITE          = 3
UV_SHUTDOWN       = 4
UV_UDP_SEND       = 5
UV_FS             = 6
UV_WORK           = 7
UV_GETADDRINFO    = 8
UV_GETNAMEINFO    = 9

# UDP flags
UV_UDP_IPV6ONLY    = 1
UV_UDP_PARTIAL     = 2
UV_UDP_REUSEADDR   = 4
UV_UDP_MMSG_CHUNK  = 8
UV_UDP_MMSG_FREE   = 16
UV_UDP_LINUX_RECVERR = 64
UV_UDP_RECVMMSG    = 256

# Address family constants (match AF_INET / AF_INET6)
AF_UNSPEC = 0
AF_INET   = 2
AF_INET6  = 10   # Linux; 30 on macOS — but uv helpers accept strings

# ==============================================================================
#  Error Class
# ==============================================================================

class UvError(Exception):
    def __init__(self, code, context=""):
        self.code    = code
        self.context = context
        msg = _uv_strerror(code) if code != 0 else "ok"
        if context:
            super().__init__(format_str("[{context}] libuv error {code}: {msg}"))
        else:
            super().__init__(format_str("libuv error {code}: {msg}"))


def _check(ret, context=""):
    """Raise UvError if ret < 0."""
    if isinstance(ret, int) and ret < 0:
        raise UvError(ret, context)
    return ret

# ==============================================================================
#  C Function Signatures
# ==============================================================================

# ---- Error / version --------------------------------------------------------
_uv_strerror   = _lib.uv_strerror(  [ffi.c_int32], ffi.c_char_p)
_uv_err_name   = _lib.uv_err_name(  [ffi.c_int32], ffi.c_char_p)
_uv_version    = _lib.uv_version(   [], ffi.c_uint32)

# ---- Size queries -----------------------------------------------------------
_uv_loop_size   = _lib.uv_loop_size(  [], ffi.c_uint64)
_uv_handle_size = _lib.uv_handle_size([ffi.c_int32], ffi.c_uint64)
_uv_req_size    = _lib.uv_req_size(   [ffi.c_int32], ffi.c_uint64)

# ---- Loop -------------------------------------------------------------------
_uv_loop_init      = _lib.uv_loop_init(     [ffi.c_void_p], ffi.c_int32)
_uv_loop_close     = _lib.uv_loop_close(    [ffi.c_void_p], ffi.c_int32)
_uv_run            = _lib.uv_run(           [ffi.c_void_p, ffi.c_int32], ffi.c_int32)
_uv_stop           = _lib.uv_stop(          [ffi.c_void_p], None)
_uv_loop_alive     = _lib.uv_loop_alive(    [ffi.c_void_p], ffi.c_int32)
_uv_loop_configure = _lib.uv_loop_configure([ffi.c_void_p, ffi.c_int32, ffi.c_int32], ffi.c_int32)
_uv_backend_fd     = _lib.uv_backend_fd(    [ffi.c_void_p], ffi.c_int32)
_uv_backend_timeout = _lib.uv_backend_timeout([ffi.c_void_p], ffi.c_int32)
_uv_now            = _lib.uv_now(           [ffi.c_void_p], ffi.c_uint64)

# ---- Handles (generic) ------------------------------------------------------
_uv_close              = _lib.uv_close(          [ffi.c_void_p, ffi.c_void_p], None)
_uv_is_closing         = _lib.uv_is_closing(     [ffi.c_void_p], ffi.c_int32)
_uv_is_active          = _lib.uv_is_active(      [ffi.c_void_p], ffi.c_int32)
_uv_handle_get_loop    = _lib.uv_handle_get_loop([ffi.c_void_p], ffi.c_void_p)
_uv_ref                = _lib.uv_ref(            [ffi.c_void_p], None)
_uv_unref              = _lib.uv_unref(          [ffi.c_void_p], None)
_uv_has_ref            = _lib.uv_has_ref(        [ffi.c_void_p], ffi.c_int32)
_uv_recv_buffer_size   = _lib.uv_recv_buffer_size([ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_uv_send_buffer_size   = _lib.uv_send_buffer_size([ffi.c_void_p, ffi.c_void_p], ffi.c_int32)

# ---- Timer ------------------------------------------------------------------
_uv_timer_init  = _lib.uv_timer_init( [ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_uv_timer_start = _lib.uv_timer_start([ffi.c_void_p, ffi.c_void_p, ffi.c_uint64, ffi.c_uint64], ffi.c_int32)
_uv_timer_stop  = _lib.uv_timer_stop( [ffi.c_void_p], ffi.c_int32)
_uv_timer_again = _lib.uv_timer_again([ffi.c_void_p], ffi.c_int32)
_uv_timer_set_repeat = _lib.uv_timer_set_repeat([ffi.c_void_p, ffi.c_uint64], None)
_uv_timer_get_repeat = _lib.uv_timer_get_repeat([ffi.c_void_p], ffi.c_uint64)

# ---- Idle / Prepare / Check -------------------------------------------------
_uv_idle_init    = _lib.uv_idle_init(   [ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_uv_idle_start   = _lib.uv_idle_start(  [ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_uv_idle_stop    = _lib.uv_idle_stop(   [ffi.c_void_p], ffi.c_int32)

_uv_prepare_init  = _lib.uv_prepare_init( [ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_uv_prepare_start = _lib.uv_prepare_start([ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_uv_prepare_stop  = _lib.uv_prepare_stop( [ffi.c_void_p], ffi.c_int32)

_uv_check_init   = _lib.uv_check_init(  [ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_uv_check_start  = _lib.uv_check_start( [ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_uv_check_stop   = _lib.uv_check_stop(  [ffi.c_void_p], ffi.c_int32)

# ---- Signal -----------------------------------------------------------------
_uv_signal_init  = _lib.uv_signal_init( [ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_uv_signal_start = _lib.uv_signal_start([ffi.c_void_p, ffi.c_void_p, ffi.c_int32], ffi.c_int32)
_uv_signal_stop  = _lib.uv_signal_stop( [ffi.c_void_p], ffi.c_int32)

# ---- Stream (base for TCP and Pipe) -----------------------------------------
# uv_read_start(stream, alloc_cb, read_cb)
_uv_read_start   = _lib.uv_read_start(  [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_uv_read_stop    = _lib.uv_read_stop(   [ffi.c_void_p], ffi.c_int32)
# uv_write(req, stream, bufs, nbufs, write_cb)
_uv_write        = _lib.uv_write(       [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_uint32, ffi.c_void_p], ffi.c_int32)
# uv_shutdown(req, stream, cb)
_uv_shutdown     = _lib.uv_shutdown(    [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_uv_listen       = _lib.uv_listen(      [ffi.c_void_p, ffi.c_int32, ffi.c_void_p], ffi.c_int32)
_uv_accept       = _lib.uv_accept(      [ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_uv_is_readable  = _lib.uv_is_readable( [ffi.c_void_p], ffi.c_int32)
_uv_is_writable  = _lib.uv_is_writable( [ffi.c_void_p], ffi.c_int32)
_uv_stream_set_blocking = _lib.uv_stream_set_blocking([ffi.c_void_p, ffi.c_int32], ffi.c_int32)

# ---- TCP --------------------------------------------------------------------
_uv_tcp_init         = _lib.uv_tcp_init(        [ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_uv_tcp_open         = _lib.uv_tcp_open(        [ffi.c_void_p, ffi.c_int32], ffi.c_int32)
_uv_tcp_nodelay      = _lib.uv_tcp_nodelay(     [ffi.c_void_p, ffi.c_int32], ffi.c_int32)
_uv_tcp_keepalive    = _lib.uv_tcp_keepalive(   [ffi.c_void_p, ffi.c_int32, ffi.c_uint32], ffi.c_int32)
_uv_tcp_simultaneous_accepts = _lib.uv_tcp_simultaneous_accepts([ffi.c_void_p, ffi.c_int32], ffi.c_int32)
_uv_tcp_bind         = _lib.uv_tcp_bind(        [ffi.c_void_p, ffi.c_void_p, ffi.c_uint32], ffi.c_int32)
# uv_tcp_connect(req, handle, addr, connect_cb)
_uv_tcp_connect      = _lib.uv_tcp_connect(     [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_uv_tcp_getpeername  = _lib.uv_tcp_getpeername( [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_uv_tcp_getsockname  = _lib.uv_tcp_getsockname( [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], ffi.c_int32)

# ---- UDP --------------------------------------------------------------------
_uv_udp_init          = _lib.uv_udp_init(         [ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_uv_udp_open          = _lib.uv_udp_open(         [ffi.c_void_p, ffi.c_int32], ffi.c_int32)
_uv_udp_bind          = _lib.uv_udp_bind(         [ffi.c_void_p, ffi.c_void_p, ffi.c_uint32], ffi.c_int32)
_uv_udp_connect       = _lib.uv_udp_connect(      [ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_uv_udp_getpeername   = _lib.uv_udp_getpeername(  [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_uv_udp_getsockname   = _lib.uv_udp_getsockname(  [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_uv_udp_set_membership      = _lib.uv_udp_set_membership(     [ffi.c_void_p, ffi.c_char_p, ffi.c_char_p, ffi.c_int32], ffi.c_int32)
_uv_udp_set_multicast_loop  = _lib.uv_udp_set_multicast_loop( [ffi.c_void_p, ffi.c_int32], ffi.c_int32)
_uv_udp_set_multicast_ttl   = _lib.uv_udp_set_multicast_ttl(  [ffi.c_void_p, ffi.c_int32], ffi.c_int32)
_uv_udp_set_broadcast       = _lib.uv_udp_set_broadcast(      [ffi.c_void_p, ffi.c_int32], ffi.c_int32)
_uv_udp_set_ttl             = _lib.uv_udp_set_ttl(            [ffi.c_void_p, ffi.c_int32], ffi.c_int32)
# uv_udp_send(req, handle, bufs, nbufs, addr, send_cb)
_uv_udp_send          = _lib.uv_udp_send(         [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_uint32, ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
# uv_udp_recv_start(handle, alloc_cb, recv_cb)
_uv_udp_recv_start    = _lib.uv_udp_recv_start(   [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p], ffi.c_int32)
_uv_udp_recv_stop     = _lib.uv_udp_recv_stop(    [ffi.c_void_p], ffi.c_int32)

# ---- Pipe -------------------------------------------------------------------
_uv_pipe_init      = _lib.uv_pipe_init(     [ffi.c_void_p, ffi.c_void_p, ffi.c_int32], ffi.c_int32)
_uv_pipe_open      = _lib.uv_pipe_open(     [ffi.c_void_p, ffi.c_int32], ffi.c_int32)
_uv_pipe_bind      = _lib.uv_pipe_bind(     [ffi.c_void_p, ffi.c_char_p], ffi.c_int32)
# uv_pipe_connect(req, handle, name, connect_cb)
_uv_pipe_connect   = _lib.uv_pipe_connect(  [ffi.c_void_p, ffi.c_void_p, ffi.c_char_p, ffi.c_void_p], None)
_uv_pipe_getpeername = _lib.uv_pipe_getpeername([ffi.c_void_p, ffi.c_char_p, ffi.c_void_p], ffi.c_int32)
_uv_pipe_getsockname = _lib.uv_pipe_getsockname([ffi.c_void_p, ffi.c_char_p, ffi.c_void_p], ffi.c_int32)

# ---- Network address helpers ------------------------------------------------
# uv_ip4_addr(ip_str, port, out_sockaddr_in*)
_uv_ip4_addr    = _lib.uv_ip4_addr(   [ffi.c_char_p, ffi.c_int32, ffi.c_void_p], ffi.c_int32)
_uv_ip6_addr    = _lib.uv_ip6_addr(   [ffi.c_char_p, ffi.c_int32, ffi.c_void_p], ffi.c_int32)
_uv_ip4_name    = _lib.uv_ip4_name(   [ffi.c_void_p, ffi.c_char_p, ffi.c_uint64], ffi.c_int32)
_uv_ip6_name    = _lib.uv_ip6_name(   [ffi.c_void_p, ffi.c_char_p, ffi.c_uint64], ffi.c_int32)
_uv_inet_ntop   = _lib.uv_inet_ntop(  [ffi.c_int32, ffi.c_void_p, ffi.c_char_p, ffi.c_uint64], ffi.c_int32)
_uv_inet_pton   = _lib.uv_inet_pton(  [ffi.c_int32, ffi.c_char_p, ffi.c_void_p], ffi.c_int32)

# ---- DNS (async) ------------------------------------------------------------
# uv_getaddrinfo(loop, req, getaddrinfo_cb, node, service, hints)
_uv_getaddrinfo  = _lib.uv_getaddrinfo( [ffi.c_void_p, ffi.c_void_p, ffi.c_void_p, ffi.c_char_p, ffi.c_char_p, ffi.c_void_p], ffi.c_int32)
_uv_freeaddrinfo = _lib.uv_freeaddrinfo([ffi.c_void_p], None)

# ==============================================================================
#  Size Constants (resolved at runtime so the wrapper works with any libuv version)
# ==============================================================================

_LOOP_SIZE    = _uv_loop_size()
_TCP_SIZE     = _uv_handle_size(UV_TCP)
_UDP_SIZE     = _uv_handle_size(UV_UDP)
_TIMER_SIZE   = _uv_handle_size(UV_TIMER)
_IDLE_SIZE    = _uv_handle_size(UV_IDLE)
_PREPARE_SIZE = _uv_handle_size(UV_PREPARE)
_CHECK_SIZE   = _uv_handle_size(UV_CHECK)
_SIGNAL_SIZE  = _uv_handle_size(UV_SIGNAL)
_PIPE_SIZE    = _uv_handle_size(UV_NAMED_PIPE)

_CONNECT_REQ_SIZE = _uv_req_size(UV_CONNECT)
_WRITE_REQ_SIZE   = _uv_req_size(UV_WRITE)
_SHUTDOWN_REQ_SIZE= _uv_req_size(UV_SHUTDOWN)
_UDP_SEND_REQ_SIZE= _uv_req_size(UV_UDP_SEND)
_GETADDR_REQ_SIZE = _uv_req_size(UV_GETADDRINFO)

# sockaddr_in (IPv4): 16 bytes; sockaddr_in6 (IPv6): 28 bytes
_SOCKADDR_STORAGE = 128   # generous; fits both

# uv_buf_t = { base: char*, len: size_t } = 16 bytes on 64-bit
_UVBUF_SIZE = 16

# ==============================================================================
#  Internal helpers
# ==============================================================================

def _make_sockaddr(host, port):
    """
    Build a sockaddr_in or sockaddr_in6 on the heap and return (ptr, is_ipv6).
    Caller must ffi.free(ptr).
    """
    sa = ffi.malloc(_SOCKADDR_STORAGE)
    if ":" in host:
        ret = _uv_ip6_addr(host.encode('utf-8'), port, sa)
        _check(ret, format_str("uv_ip6_addr({host}:{port})"))
        return (sa, True)
    else:
        ret = _uv_ip4_addr(host.encode('utf-8'), port, sa)
        _check(ret, format_str("uv_ip4_addr({host}:{port})"))
        return (sa, False)


def _make_uvbuf(data_bytes):
    """
    Allocate a uv_buf_t on the heap pointing at a heap copy of data_bytes.
    Returns (uvbuf_ptr, data_ptr) — caller must free both.
    """
    data_ptr = ffi.malloc(len(data_bytes))
    ffi.memcpy(data_ptr, ffi.addressof(data_bytes), len(data_bytes))

    uvbuf = ffi.malloc(_UVBUF_SIZE)
    ffi.write_memory_with_offset(uvbuf, 0, data_ptr,         ffi.c_void_p)
    ffi.write_memory_with_offset(uvbuf, 8, len(data_bytes),  ffi.c_uint64)
    return (uvbuf, data_ptr)


def _alloc_cb_for(buf_size=65536):
    """
    Build a standard uv_alloc_cb that hands out a freshly malloc'd buffer.
    Returns the ffi.callback object (keep it alive in self.callbacks).

    Signature: void alloc_cb(uv_handle_t*, size_t, uv_buf_t*)
    """
    def alloc_cb(handle_ptr, suggested_size, buf_ptr):
        data = ffi.malloc(buf_size)
        ffi.write_memory_with_offset(buf_ptr, 0, data,     ffi.c_void_p)
        ffi.write_memory_with_offset(buf_ptr, 8, buf_size, ffi.c_uint64)

    return ffi.callback(alloc_cb, None, [ffi.c_void_p, ffi.c_uint64, ffi.c_void_p])

# ==============================================================================
#  Loop
# ==============================================================================

class Loop:
    """
    The libuv event loop.

    Typical usage:
        loop = Loop()
        # ... attach handles ...
        loop.run()          # blocks until all handles are closed / loop.stop()
        loop.close()

    For incremental driving (e.g., inside a larger poll loop):
        loop.run(UV_RUN_NOWAIT)   # process ready events without blocking
        loop.run(UV_RUN_ONCE)     # block until at least one event, then return
    """

    def __init__(self):
        self.ptr       = ffi.malloc(_LOOP_SIZE)
        self.callbacks = []       # keep C callbacks alive
        self._reqs     = []       # keep request buffers alive until callbacks fire
        _check(_uv_loop_init(self.ptr), "uv_loop_init")

    def run(self, mode=UV_RUN_DEFAULT):
        """Drive the event loop.  Returns non-zero if there are pending handles."""
        return _uv_run(self.ptr, mode)

    def stop(self):
        """Signal the loop to stop after the current iteration."""
        _uv_stop(self.ptr)

    def alive(self):
        """True if the loop has active handles or requests."""
        return _uv_loop_alive(self.ptr) != 0

    def now(self):
        """Return the loop's cached timestamp in milliseconds (uint64)."""
        return _uv_now(self.ptr)

    def backend_fd(self):
        """Return the backend file descriptor (epoll fd on Linux)."""
        return _uv_backend_fd(self.ptr)

    def backend_timeout(self):
        """Return the poll timeout the backend will use on the next tick (ms)."""
        return _uv_backend_timeout(self.ptr)

    def close(self):
        """Close the loop.  All handles must be closed first."""
        if self.ptr:
            _uv_loop_close(self.ptr)
            ffi.free(self.ptr)
            self.ptr = None
        for cb in self.callbacks:
            ffi.free_callback(cb)
        self.callbacks = []

    def _keep(self, cb):
        """Register a C callback to keep alive for the loop's lifetime."""
        self.callbacks.append(cb)
        return cb

    def _keep_req(self, req_ptr):
        """Keep a request buffer alive until the callback fires."""
        self._reqs.append(req_ptr)
        return req_ptr

    def _release_req(self, req_ptr):
        """Release a request buffer after its callback has fired."""
        new_reqs = []
        for r in self._reqs:
            if r != req_ptr:
                new_reqs.append(r)
        self._reqs = new_reqs
        ffi.free(req_ptr)

# ==============================================================================
#  Shutdown helper
# ==============================================================================

def shutdown_loop(loop):
    """
    Properly drain and close a loop.
    Call this instead of loop.close() directly when handles may still be open.
    """
    loop.stop()
    loop.run(UV_RUN_NOWAIT)
    loop.close()

# ==============================================================================
#  Timer
# ==============================================================================

class Timer:
    """
    Async timer handle.

    Usage:
        t = Timer(loop)
        t.start(callback, timeout_ms=1000, repeat_ms=500)   # fires every 500ms after 1s
        t.stop()
        t.close()
    """

    def __init__(self, loop):
        self.loop = loop
        self.ptr  = ffi.malloc(_TIMER_SIZE)
        _check(_uv_timer_init(loop.ptr, self.ptr), "uv_timer_init")

    def start(self, callback, timeout_ms=0, repeat_ms=0):
        """
        Start the timer.
        timeout_ms — initial delay in milliseconds (0 = fire as soon as possible)
        repeat_ms  — interval for repeating (0 = one-shot)
        callback() — called with no arguments each time the timer fires
        """
        def _c_cb(handle_ptr):
            try:
                callback()
            except Exception as e:
                print(format_str("⚠️  [Timer] callback error: {e}"))

        cb = ffi.callback(_c_cb, None, [ffi.c_void_p])
        self.loop._keep(cb)
        _check(_uv_timer_start(self.ptr, cb, timeout_ms, repeat_ms), "uv_timer_start")

    def stop(self):
        _uv_timer_stop(self.ptr)

    def again(self):
        """Restart the timer using the same repeat interval."""
        _check(_uv_timer_again(self.ptr), "uv_timer_again")

    def set_repeat(self, repeat_ms):
        _uv_timer_set_repeat(self.ptr, repeat_ms)

    def get_repeat(self):
        return _uv_timer_get_repeat(self.ptr)

    def close(self, callback=None):
        if self.ptr and not _uv_is_closing(self.ptr):
            if callback:
                def _c_close(h):
                    try:
                        callback()
                    except Exception:
                        pass
                cb = ffi.callback(_c_close, None, [ffi.c_void_p])
                self.loop._keep(cb)
                _uv_close(self.ptr, cb)
            else:
                _uv_close(self.ptr, None)
            ffi.free(self.ptr)
            self.ptr = None

# ==============================================================================
#  Idle / Prepare / Check
# ==============================================================================

class Idle:
    """
    Called every loop iteration when there are no other I/O events pending.
    Useful for background tasks that should not block I/O.
    """

    def __init__(self, loop):
        self.loop = loop
        self.ptr  = ffi.malloc(_IDLE_SIZE)
        _check(_uv_idle_init(loop.ptr, self.ptr), "uv_idle_init")

    def start(self, callback):
        def _c_cb(h):
            try:
                callback()
            except Exception as e:
                print(format_str("⚠️  [Idle] {e}"))
        cb = ffi.callback(_c_cb, None, [ffi.c_void_p])
        self.loop._keep(cb)
        _check(_uv_idle_start(self.ptr, cb), "uv_idle_start")

    def stop(self):  _uv_idle_stop(self.ptr)

    def close(self):
        if self.ptr and not _uv_is_closing(self.ptr):
            _uv_close(self.ptr, None)
            ffi.free(self.ptr)
            self.ptr = None


class Prepare:
    """Callback fired once per loop iteration BEFORE I/O polling."""

    def __init__(self, loop):
        self.loop = loop
        self.ptr  = ffi.malloc(_PREPARE_SIZE)
        _check(_uv_prepare_init(loop.ptr, self.ptr), "uv_prepare_init")

    def start(self, callback):
        def _c_cb(h):
            try:
                callback()
            except Exception as e:
                print(format_str("⚠️  [Prepare] {e}"))
        cb = ffi.callback(_c_cb, None, [ffi.c_void_p])
        self.loop._keep(cb)
        _check(_uv_prepare_start(self.ptr, cb), "uv_prepare_start")

    def stop(self):  _uv_prepare_stop(self.ptr)

    def close(self):
        if self.ptr and not _uv_is_closing(self.ptr):
            _uv_close(self.ptr, None)
            ffi.free(self.ptr)
            self.ptr = None


class Check:
    """Callback fired once per loop iteration AFTER I/O polling."""

    def __init__(self, loop):
        self.loop = loop
        self.ptr  = ffi.malloc(_CHECK_SIZE)
        _check(_uv_check_init(loop.ptr, self.ptr), "uv_check_init")

    def start(self, callback):
        def _c_cb(h):
            try:
                callback()
            except Exception as e:
                print(format_str("⚠️  [Check] {e}"))
        cb = ffi.callback(_c_cb, None, [ffi.c_void_p])
        self.loop._keep(cb)
        _check(_uv_check_start(self.ptr, cb), "uv_check_start")

    def stop(self):  _uv_check_stop(self.ptr)

    def close(self):
        if self.ptr and not _uv_is_closing(self.ptr):
            _uv_close(self.ptr, None)
            ffi.free(self.ptr)
            self.ptr = None

# ==============================================================================
#  Signal
# ==============================================================================

class Signal:
    """
    OS signal handler.

    Usage:
        sig = Signal(loop)
        sig.start(handler, signal_number)   # e.g. signal_number=2 for SIGINT
        sig.stop()
        sig.close()

    Common signal numbers (POSIX):
        SIGINT  = 2
        SIGTERM = 15
        SIGHUP  = 1
        SIGUSR1 = 10
        SIGUSR2 = 12
    """

    SIGINT  = 2
    SIGTERM = 15
    SIGHUP  = 1
    SIGUSR1 = 10
    SIGUSR2 = 12

    def __init__(self, loop):
        self.loop = loop
        self.ptr  = ffi.malloc(_SIGNAL_SIZE)
        _check(_uv_signal_init(loop.ptr, self.ptr), "uv_signal_init")

    def start(self, callback, signum):
        """
        callback(signum) — called when the signal is received.
        """
        def _c_cb(handle_ptr, sig):
            try:
                callback(sig)
            except Exception as e:
                print(format_str("⚠️  [Signal] {e}"))
        cb = ffi.callback(_c_cb, None, [ffi.c_void_p, ffi.c_int32])
        self.loop._keep(cb)
        _check(_uv_signal_start(self.ptr, cb, signum), "uv_signal_start")

    def stop(self):  _uv_signal_stop(self.ptr)

    def close(self):
        if self.ptr and not _uv_is_closing(self.ptr):
            _uv_close(self.ptr, None)
            ffi.free(self.ptr)
            self.ptr = None

# ==============================================================================
#  TCP
# ==============================================================================

class TCP:
    """
    TCP handle.  Serves as both a server listener and a client connection.

    --- Server pattern ---
        server = TCP(loop)
        server.bind('0.0.0.0', 8080)
        server.listen(on_connect)

        def on_connect(client_tcp):
            client_tcp.read_start(on_data)

        def on_data(data, error):
            if error:
                client_tcp.close()
                return
            client_tcp.write(data)   # echo

    --- Client pattern ---
        client = TCP(loop)
        client.connect('93.184.216.34', 80, on_connected)

        def on_connected(error):
            if error:
                print('connection failed:', error)
                return
            client.read_start(on_data)
            client.write(b'GET / HTTP/1.0\\r\\n\\r\\n')
    """

    def __init__(self, loop):
        self.loop = loop
        self.ptr  = ffi.malloc(_TCP_SIZE)
        _check(_uv_tcp_init(loop.ptr, self.ptr), "uv_tcp_init")
        self._alloc_cb = None

    # ---- socket options -----------------------------------------------------

    def nodelay(self, enable=True):
        """Disable Nagle's algorithm."""
        _check(_uv_tcp_nodelay(self.ptr, 1 if enable else 0), "uv_tcp_nodelay")
        return self

    def keepalive(self, enable=True, delay_seconds=60):
        """Enable TCP keep-alive probes."""
        _check(_uv_tcp_keepalive(self.ptr, 1 if enable else 0, delay_seconds), "uv_tcp_keepalive")
        return self

    def simultaneous_accepts(self, enable=True):
        """Enable simultaneous async accept requests (Windows only)."""
        _uv_tcp_simultaneous_accepts(self.ptr, 1 if enable else 0)
        return self

    # ---- server side --------------------------------------------------------

    def bind(self, host, port, ipv6only=False):
        """Bind to a local address.  Must be called before listen()."""
        sa, is_ipv6 = _make_sockaddr(host, port)
        flags = UV_UDP_IPV6ONLY if (is_ipv6 and ipv6only) else 0
        try:
            _check(_uv_tcp_bind(self.ptr, sa, flags), format_str("uv_tcp_bind({host}:{port})"))
        except Exception:
            pass
        finally:
            ffi.free(sa)
        return self

    def listen(self, on_connect, backlog=128):
        """
        Start accepting connections.
        on_connect(client_tcp, error) — called for each incoming connection.
        If error is not None, it is a UvError instance.
        """
        def _c_cb(server_ptr, status):
            try:
                if status < 0:
                    on_connect(None, UvError(status, "listen"))
                    return None
                client = TCP(self.loop)
                ret = _uv_accept(self.ptr, client.ptr)
                if ret < 0:
                    client.close()
                    on_connect(None, UvError(ret, "uv_accept"))
                    return None
                on_connect(client, None)
            except Exception as e:
                print(format_str("🔥 [TCP.listen] on_connect error: {e}"))

        cb = ffi.callback(_c_cb, None, [ffi.c_void_p, ffi.c_int32])
        self.loop._keep(cb)
        _check(_uv_listen(self.ptr, backlog, cb), "uv_listen")
        return self

    # ---- client side --------------------------------------------------------

    def connect(self, host, port, on_connect):
        """
        Initiate a TCP connection.
        on_connect(error) — called when connected (error=None) or failed.
        """
        sa, _ = _make_sockaddr(host, port)
        req   = ffi.malloc(_CONNECT_REQ_SIZE)
        self.loop._keep_req(req)

        def _c_cb(req_ptr, status):
            self.loop._release_req(req_ptr)
            ffi.free(sa)
            try:
                if status < 0:
                    on_connect(UvError(status, format_str("connect({host}:{port})")))
                else:
                    on_connect(None)
            except Exception as e:
                print(format_str("🔥 [TCP.connect] callback error: {e}"))

        cb = ffi.callback(_c_cb, None, [ffi.c_void_p, ffi.c_int32])
        self.loop._keep(cb)
        _check(_uv_tcp_connect(req, self.ptr, sa, cb), format_str("uv_tcp_connect({host}:{port})"))
        return self

    # ---- stream I/O ---------------------------------------------------------

    def read_start(self, on_data, buf_size=65536):
        """
        Begin reading from this stream.
        on_data(data_bytes, error) — data_bytes is bytes or None on EOF/error.
        error is a UvError or None.
        """
        alloc_cb = _alloc_cb_for(buf_size)
        self._alloc_cb = alloc_cb
        self.loop._keep(alloc_cb)

        def _read_cb(stream_ptr, nread, buf_ptr):
            base = ffi.read_memory_with_offset(buf_ptr, 0, ffi.c_void_p)
            try:
                if nread < 0:
                    if base:
                        ffi.free(base)
                    # EOF = UV_EOF = -4095; treat as clean close
                    on_data(None, None if nread == -4095 else UvError(nread, "read"))
                    return None
                if nread == 0:
                    if base:
                        ffi.free(base)
                    return None
                data = ffi.buffer_to_bytes(base, nread)
                on_data(data, None)
            except Exception as e:
                print(format_str("🔥 [TCP.read] callback error: {e}"))
            finally:
                if base:
                    try:
                        ffi.free(base)
                    except Exception:
                        pass

        cb = ffi.callback(_read_cb, None, [ffi.c_void_p, ffi.c_int64, ffi.c_void_p])
        self.loop._keep(cb)
        _check(_uv_read_start(self.ptr, alloc_cb, cb), "uv_read_start")
        return self

    def read_stop(self):
        _uv_read_stop(self.ptr)

    def write(self, data, on_write=None):
        """
        Queue data for writing.
        on_write(error) — optional completion callback.
        """
        b = data.encode('utf-8') if isinstance(data, str) else data
        uvbuf, data_ptr = _make_uvbuf(b)
        req = ffi.malloc(_WRITE_REQ_SIZE)
        self.loop._keep_req(req)

        def _write_cb(req_ptr, status):
            ffi.free(uvbuf)
            ffi.free(data_ptr)
            self.loop._release_req(req_ptr)
            if on_write:
                try:
                    on_write(UvError(status, "write") if status < 0 else None)
                except Exception as e:
                    print(format_str("⚠️  [TCP.write] callback error: {e}"))

        cb = ffi.callback(_write_cb, None, [ffi.c_void_p, ffi.c_int32])
        self.loop._keep(cb)
        _check(_uv_write(req, self.ptr, uvbuf, 1, cb), "uv_write")
        return self

    def shutdown(self, on_shutdown=None):
        """
        Gracefully shut down the write side of the connection.
        on_shutdown(error) — called when all queued writes have completed.
        """
        req = ffi.malloc(_SHUTDOWN_REQ_SIZE)
        self.loop._keep_req(req)

        def _cb(req_ptr, status):
            self.loop._release_req(req_ptr)
            if on_shutdown:
                try:
                    on_shutdown(UvError(status, "shutdown") if status < 0 else None)
                except Exception as e:
                    print(format_str("⚠️  [TCP.shutdown] callback error: {e}"))

        cb = ffi.callback(_cb, None, [ffi.c_void_p, ffi.c_int32])
        self.loop._keep(cb)
        _check(_uv_shutdown(req, self.ptr, cb), "uv_shutdown")

    # ---- addressing ---------------------------------------------------------

    def getsockname(self):
        """Return (host, port) of the local bound address."""
        return _sockname(_uv_tcp_getsockname, self.ptr)

    def getpeername(self):
        """Return (host, port) of the remote peer address."""
        return _sockname(_uv_tcp_getpeername, self.ptr)

    # ---- socket buffer sizes ------------------------------------------------

    def set_recv_buffer(self, size):
        buf = ffi.malloc(4)
        ffi.write_memory(buf, ffi.c_int32, size)
        _uv_recv_buffer_size(self.ptr, buf)
        ffi.free(buf)

    def set_send_buffer(self, size):
        buf = ffi.malloc(4)
        ffi.write_memory(buf, ffi.c_int32, size)
        _uv_send_buffer_size(self.ptr, buf)
        ffi.free(buf)

    # ---- lifecycle ----------------------------------------------------------

    def close(self, callback=None):
        if self.ptr and not _uv_is_closing(self.ptr):
            if callback:
                def _c_close(h):
                    try:
                        callback()
                    except Exception:
                        pass
                cb = ffi.callback(_c_close, None, [ffi.c_void_p])
                self.loop._keep(cb)
                _uv_close(self.ptr, cb)
            else:
                _uv_close(self.ptr, None)
            ffi.free(self.ptr)
            self.ptr = None

# ==============================================================================
#  UDP
# ==============================================================================

class UDP:
    """
    UDP datagram socket.

    --- Server (recv) ---
        sock = UDP(loop)
        sock.bind('0.0.0.0', 9000)
        sock.recv_start(on_recv)

        def on_recv(data, addr, error):
            if error or data is None: return
            sock.send(data, addr[0], addr[1])   # echo

    --- Client (send) ---
        sock = UDP(loop)
        sock.send(b'hello', '127.0.0.1', 9000)

    --- Broadcast ---
        sock = UDP(loop)
        sock.set_broadcast(True)
        sock.send(b'ping', '255.255.255.255', 9000)

    --- Multicast ---
        sock = UDP(loop)
        sock.bind('0.0.0.0', 9000)
        sock.set_membership('239.0.0.1', UV_JOIN_GROUP)
        sock.recv_start(on_recv)
    """

    UV_JOIN_GROUP  = 1
    UV_LEAVE_GROUP = 2

    def __init__(self, loop):
        self.loop = loop
        self.ptr  = ffi.malloc(_UDP_SIZE)
        _check(_uv_udp_init(loop.ptr, self.ptr), "uv_udp_init")

    def bind(self, host, port, flags=0):
        sa, _ = _make_sockaddr(host, port)
        try:
            _check(_uv_udp_bind(self.ptr, sa, flags), format_str("uv_udp_bind({host}:{port})"))
        finally:
            ffi.free(sa)
        return self

    def set_broadcast(self, enable=True):
        _check(_uv_udp_set_broadcast(self.ptr, 1 if enable else 0), "uv_udp_set_broadcast")
        return self

    def set_ttl(self, ttl):
        _check(_uv_udp_set_ttl(self.ptr, ttl), "uv_udp_set_ttl")
        return self

    def set_multicast_loop(self, enable=True):
        _check(_uv_udp_set_multicast_loop(self.ptr, 1 if enable else 0), "uv_udp_set_multicast_loop")
        return self

    def set_multicast_ttl(self, ttl):
        _check(_uv_udp_set_multicast_ttl(self.ptr, ttl), "uv_udp_set_multicast_ttl")
        return self

    def set_membership(self, multicast_addr, membership, interface_addr=""):
        _check(_uv_udp_set_membership(self.ptr, multicast_addr.encode('utf-8'), interface_addr.encode('utf-8') if interface_addr else None, membership), "uv_udp_set_membership")
        return self

    def send(self, data, host, port, on_send=None):
        """
        Send a datagram to (host, port).
        on_send(error) — optional completion callback.
        """
        b = data.encode('utf-8') if isinstance(data, str) else data
        sa, _ = _make_sockaddr(host, port)
        uvbuf, data_ptr = _make_uvbuf(b)
        req = ffi.malloc(_UDP_SEND_REQ_SIZE)
        self.loop._keep_req(req)

        def _cb(req_ptr, status):
            ffi.free(uvbuf)
            ffi.free(data_ptr)
            ffi.free(sa)
            self.loop._release_req(req_ptr)
            if on_send:
                try:
                    on_send(UvError(status, "udp_send") if status < 0 else None)
                except Exception as e:
                    print(format_str("⚠️  [UDP.send] callback error: {e}"))

        cb = ffi.callback(_cb, None, [ffi.c_void_p, ffi.c_int32])
        self.loop._keep(cb)
        _check(_uv_udp_send(req, self.ptr, uvbuf, 1, sa, cb), "uv_udp_send")
        return self

    def recv_start(self, on_recv, buf_size=65536):
        """
        Begin receiving datagrams.
        on_recv(data_bytes, (host, port), error) — called for each datagram.
        data_bytes may be None on error.
        """
        alloc_cb = _alloc_cb_for(buf_size)
        self.loop._keep(alloc_cb)

        def _recv_cb(handle_ptr, nread, buf_ptr, sa_ptr, flags):
            base = ffi.read_memory_with_offset(buf_ptr, 0, ffi.c_void_p)
            try:
                if nread < 0:
                    if base: ffi.free(base)
                    on_recv(None, None, UvError(nread, "udp_recv"))
                    return None
                if nread == 0:
                    if base: ffi.free(base)
                    return None
                data = ffi.buffer_to_bytes(base, nread)
                addr = _parse_sockaddr(sa_ptr) if sa_ptr else ("", 0)
                on_recv(data, addr, None)
            except Exception as e:
                print(format_str("🔥 [UDP.recv] callback error: {e}"))
            finally:
                if base:
                    try:
                        ffi.free(base)
                    except Exception:
                        pass

        cb = ffi.callback(_recv_cb, None, [ffi.c_void_p, ffi.c_int64, ffi.c_void_p, ffi.c_void_p, ffi.c_uint32])
        self.loop._keep(cb)
        _check(_uv_udp_recv_start(self.ptr, alloc_cb, cb), "uv_udp_recv_start")
        return self

    def recv_stop(self):
        _uv_udp_recv_stop(self.ptr)

    def getsockname(self):
        return _sockname(_uv_udp_getsockname, self.ptr)

    def close(self, callback=None):
        if self.ptr and not _uv_is_closing(self.ptr):
            if callback:
                def _c_close(h):
                    try:
                        callback()
                    except Exception:
                        pass
                cb = ffi.callback(_c_close, None, [ffi.c_void_p])
                self.loop._keep(cb)
                _uv_close(self.ptr, cb)
            else:
                _uv_close(self.ptr, None)
            ffi.free(self.ptr)
            self.ptr = None

# ==============================================================================
#  Pipe  (UNIX domain socket / Windows named pipe)
# ==============================================================================

class Pipe:
    """
    Named-pipe / UNIX domain socket handle.

    Server:
        p = Pipe(loop)
        p.bind('/tmp/my.sock')
        p.listen(on_connect)

    Client:
        p = Pipe(loop)
        p.connect('/tmp/my.sock', on_connected)
    """

    def __init__(self, loop, ipc=False):
        self.loop = loop
        self.ptr  = ffi.malloc(_PIPE_SIZE)
        _check(_uv_pipe_init(loop.ptr, self.ptr, 1 if ipc else 0), "uv_pipe_init")

    def bind(self, name):
        _check(_uv_pipe_bind(self.ptr, name.encode('utf-8')), format_str("uv_pipe_bind({name})"))
        return self

    def connect(self, name, on_connect):
        """on_connect(error) — error is None on success."""
        req = ffi.malloc(_CONNECT_REQ_SIZE)
        self.loop._keep_req(req)

        def _cb(req_ptr, status):
            self.loop._release_req(req_ptr)
            try:
                on_connect(UvError(status, format_str("pipe_connect({name})")) if status < 0 else None)
            except Exception as e:
                print(format_str("🔥 [Pipe.connect] {e}"))

        cb = ffi.callback(_cb, None, [ffi.c_void_p, ffi.c_int32])
        self.loop._keep(cb)
        _uv_pipe_connect(req, self.ptr, name.encode('utf-8'), cb)
        return self

    def listen(self, on_connect, backlog=128):
        """on_connect(client_pipe, error)"""
        def _c_cb(server_ptr, status):
            try:
                if status < 0:
                    on_connect(None, UvError(status, "pipe_listen"))
                    return None
                client = Pipe(self.loop)
                ret = _uv_accept(self.ptr, client.ptr)
                if ret < 0:
                    client.close()
                    on_connect(None, UvError(ret, "pipe_accept"))
                    return None
                on_connect(client, None)
            except Exception as e:
                print(format_str("🔥 [Pipe.listen] {e}"))

        cb = ffi.callback(_c_cb, None, [ffi.c_void_p, ffi.c_int32])
        self.loop._keep(cb)
        _check(_uv_listen(self.ptr, backlog, cb), "uv_pipe_listen")
        return self

    def read_start(self, on_data, buf_size=65536):
        alloc_cb = _alloc_cb_for(buf_size)
        self.loop._keep(alloc_cb)

        def _read_cb(stream_ptr, nread, buf_ptr):
            base = ffi.read_memory_with_offset(buf_ptr, 0, ffi.c_void_p)
            try:
                if nread < 0:
                    if base: ffi.free(base)
                    on_data(None, None if nread == -4095 else UvError(nread, "pipe_read"))
                    return None
                if nread == 0:
                    if base: ffi.free(base)
                    return None
                data = ffi.buffer_to_bytes(base, nread)
                on_data(data, None)
            except Exception as e:
                print(format_str("🔥 [Pipe.read] {e}"))
            finally:
                if base:
                    try:
                        ffi.free(base)
                    except Exception:
                        pass

        cb = ffi.callback(_read_cb, None, [ffi.c_void_p, ffi.c_int64, ffi.c_void_p])
        self.loop._keep(cb)
        _check(_uv_read_start(self.ptr, alloc_cb, cb), "uv_read_start(pipe)")
        return self

    def read_stop(self): _uv_read_stop(self.ptr)

    def write(self, data, on_write=None):
        b = data.encode('utf-8') if isinstance(data, str) else data
        uvbuf, data_ptr = _make_uvbuf(b)
        req = ffi.malloc(_WRITE_REQ_SIZE)
        self.loop._keep_req(req)

        def _write_cb(req_ptr, status):
            ffi.free(uvbuf)
            ffi.free(data_ptr)
            self.loop._release_req(req_ptr)
            if on_write:
                try:
                    on_write(UvError(status, "pipe_write") if status < 0 else None)
                except Exception as e:
                    print(format_str("⚠️  [Pipe.write] {e}"))

        cb = ffi.callback(_write_cb, None, [ffi.c_void_p, ffi.c_int32])
        self.loop._keep(cb)
        _check(_uv_write(req, self.ptr, uvbuf, 1, cb), "uv_write(pipe)")
        return self

    def close(self, callback=None):
        if self.ptr and not _uv_is_closing(self.ptr):
            if callback:
                def _c_close(h):
                    try:
                        callback()
                    except Exception:
                        pass
                cb = ffi.callback(_c_close, None, [ffi.c_void_p])
                self.loop._keep(cb)
                _uv_close(self.ptr, cb)
            else:
                _uv_close(self.ptr, None)
            ffi.free(self.ptr)
            self.ptr = None

# ==============================================================================
#  DNS — Async getaddrinfo
# ==============================================================================

class _AddrInfo:
    """Result of a successful DNS lookup."""
    def __init__(self, family, addr, port):
        self.family = family   # AF_INET or AF_INET6
        self.addr   = addr     # str IP address
        self.port   = port     # int

    def __repr__(self):
        return format_str("AddrInfo(family={self.family}, addr={self.addr}, port={self.port})")


def dns_getaddrinfo(loop, hostname, service_or_port, on_resolved):
    """
    Asynchronously resolve `hostname`.

    on_resolved(results, error)
      results — list of _AddrInfo, or None on error
      error   — UvError or None

    service_or_port may be a string ('http', 'https') or an integer port number.
    """
    req = ffi.malloc(_GETADDR_REQ_SIZE)
    loop._keep_req(req)

    if isinstance(service_or_port, int):
        svc_str = str(service_or_port).encode('utf-8')
    else:
        svc_str = service_or_port.encode('utf-8')

    def _cb(req_ptr, status, res_ptr):
        loop._release_req(req_ptr)
        try:
            if status < 0:
                on_resolved(None, UvError(status, format_str("getaddrinfo({hostname})")))
                return None

            results = _parse_addrinfo_list(res_ptr)
            _uv_freeaddrinfo(res_ptr)
            on_resolved(results, None)
        except Exception as e:
            print(format_str("🔥 [dns_getaddrinfo] callback error: {e}"))

    cb = ffi.callback(_cb, None, [ffi.c_void_p, ffi.c_int32, ffi.c_void_p])
    loop._keep(cb)

    _check(_uv_getaddrinfo(loop.ptr, req, cb, hostname.encode('utf-8'), svc_str, None), format_str("uv_getaddrinfo({hostname})"))


def getaddrinfo(loop, hostname, service_or_port, timeout_ms=5000):
    """
    Synchronous (blocking) DNS resolution using the async API + manual poll.
    Returns a list of _AddrInfo.  Raises UvError on failure or timeout.
    """
    results = [None]
    errors  = [None]
    done    = [False]

    def on_resolved(res, err):
        results[0] = res
        errors[0]  = err
        done[0]    = True

    dns_getaddrinfo(loop, hostname, service_or_port, on_resolved)

    elapsed = 0
    while not done[0] and elapsed < timeout_ms:
        loop.run(UV_RUN_ONCE)
        elapsed = elapsed + 1

    if not done[0]:
        raise UvError(-12345, format_str("getaddrinfo({hostname}) timed out"))
    if errors[0]:
        raise errors[0]
    return results[0]

# ==============================================================================
#  Address helpers
# ==============================================================================

def _parse_sockaddr(sa_ptr):
    """
    Parse a struct sockaddr* (IPv4 or IPv6) and return (host_str, port).
    sockaddr_in:
      uint16_t sin_family [0]
      uint16_t sin_port   [2]  (network byte order)
      uint8_t  sin_addr   [4]  (4 bytes)
    sockaddr_in6:
      uint16_t sin6_family   [0]
      uint16_t sin6_port     [2]
      uint32_t sin6_flowinfo [4]
      uint8_t  sin6_addr     [8]  (16 bytes)
    """
    family = ffi.read_memory_with_offset(sa_ptr, 0, ffi.c_uint16)
    port_n = ffi.read_memory_with_offset(sa_ptr, 2, ffi.c_uint16)
    # Convert big-endian network port to host order
    port = ((port_n & 0xFF) << 8) | ((port_n >> 8) & 0xFF)

    buf = ffi.malloc(64)
    try:
        if family == 10:   # AF_INET6
            _uv_ip6_name(sa_ptr, buf, 64)
        else:              # AF_INET (2)
            _uv_ip4_name(sa_ptr, buf, 64)
        host = ffi.string_at(buf)
    except Exception:
        pass
    finally:
        ffi.free(buf)

    return (host, port)


def _sockname(fn, handle_ptr):
    """Call fn(handle_ptr, sa_buf, &len) and parse the result."""
    sa_buf  = ffi.malloc(_SOCKADDR_STORAGE)
    len_buf = ffi.malloc(4)
    ffi.write_memory(len_buf, ffi.c_int32, _SOCKADDR_STORAGE)
    try:
        ret = fn(handle_ptr, sa_buf, len_buf)
        if ret < 0:
            return ("", 0)
        return _parse_sockaddr(sa_buf)
    finally:
        ffi.free(sa_buf)
        ffi.free(len_buf)


# addrinfo struct offsets (Linux / macOS 64-bit):
#   int      ai_flags     [0]
#   int      ai_family    [4]
#   int      ai_socktype  [8]
#   int      ai_protocol  [12]
#   socklen_t ai_addrlen  [16]  (4 bytes on Linux, 4 on macOS)
#   struct sockaddr* ai_addr [24]  (pointer, 8 bytes)
#   char*    ai_canonname  [32]
#   addrinfo* ai_next      [40]

def _parse_addrinfo_list(head_ptr):
    """Walk a linked list of struct addrinfo and return list of _AddrInfo."""
    results = []
    cur = head_ptr
    while cur:
        family  = ffi.read_memory_with_offset(cur, 4, ffi.c_int32)
        addr_ptr= ffi.read_memory_with_offset(cur, 24, ffi.c_void_p)
        if addr_ptr:
            try:
                host, port = _parse_sockaddr(addr_ptr)
                results.append(_AddrInfo(family, host, port))
            except Exception:
                pass
        cur = ffi.read_memory_with_offset(cur, 40, ffi.c_void_p)
    return results

# ==============================================================================
#  Convenience: version string
# ==============================================================================

def version():
    """Return the libuv version as a string, e.g. '1.44.2'."""
    v = _uv_version()
    major = (v >> 16) & 0xFF
    minor = (v >>  8) & 0xFF
    patch =  v        & 0xFF
    return format_str("{major}.{minor}.{patch}")