"""
Production-ready HTTP + WebSocket server framework for Swalang.

Features:
  - Parameterised routing      /users/:id  /posts/:id/comments/:cid
  - Wildcard routes            /static/*
  - HTTP method filtering      GET, POST, PUT, PATCH, DELETE, ...
  - Before/after hooks         (middleware chain)
  - Custom error handlers      @app.error_handler(404)
  - Static file directories    app.static('/assets', './public')
  - WebSocket support          @app.ws_route('/ws')
  - WebSocket broadcast        app.broadcast('/ws', 'hello everyone')
  - HTTPS/WSS                  app.run(tls_cert=..., tls_key=...)
  - Route groups               app.group('/api/v1')
  - Cookie helpers             Request.cookies, Response.set_cookie()
  - Query string parsing       Request.args  (dict)
  - JSON body parsing          Request.json()
  - Structured Response        Response.json() / .html() / .redirect() / ...
  - HttpError                  raise HttpError(403, 'Forbidden')

Example
-------
    app = Server(port=8080)

    @app.route('/')
    def index(req):
        return Response.html('<h1>Hello, world!</h1>')

    @app.get('/users/:id')
    def get_user(req):
        uid = req.params['id']
        return Response.json({'id': uid, 'name': 'Alice'})

    @app.post('/users')
    def create_user(req):
        body = req.json()
        # ... persist ...
        return Response.json({'created': True}, status=201)

    @app.error_handler(404)
    def not_found(req):
        return Response.html('<h1>404 — Not Found</h1>', status=404)

    @app.ws_route('/chat')
    def chat_ws(conn, event, data):
        if event == 'MSG':
            app.broadcast('/chat', data.data)   # echo to everyone

    app.static('/static', './public')
    app.run()                                    # HTTP on port 8080
    # app.run(tls_cert='server.crt', tls_key='server.key')  # HTTPS
"""

import mongoose
import cjson

# ==============================================================================
#  URL / Query String / Cookie Utilities
# ==============================================================================

def parse_qs(query_str):
    """
    Parse a URL query string into a dict.
    Repeated keys become lists.

    '?name=alice&age=30&tag=x&tag=y'
        → {'name': 'alice', 'age': '30', 'tag': ['x', 'y']}
    """
    result = {}
    if not query_str:
        return result

    src = query_str
    if src.startswith("?"):
        src = src[1:]

    for pair in src.split("&"):
        if not pair:
            continue
        if "=" in pair:
            idx = pair.find("=")
            key = mongoose.url_decode(pair[:idx])
            val = mongoose.url_decode(pair[idx + 1:])
        else:
            key = mongoose.url_decode(pair)
            val = ""

        if key in result:
            existing = result[key]
            if isinstance(existing, list):
                existing.append(val)
            else:
                result[key] = [existing, val]
        else:
            result[key] = val

    return result


def parse_cookies(cookie_header):
    """
    Parse the value of a Cookie header into a dict.
    'session=abc; user=alice' → {'session': 'abc', 'user': 'alice'}
    """
    cookies = {}
    if not cookie_header:
        return cookies
    for part in cookie_header.split(";"):
        part = part.strip()
        if "=" in part:
            idx   = part.find("=")
            name  = part[:idx].strip()
            value = part[idx + 1:].strip()
            if name:
                cookies[name] = value
    return cookies


def _build_set_cookie(name, value, max_age=None, path="/", domain=None, http_only=True, secure=False, same_site="Lax"):
    """Build a Set-Cookie header value string."""
    parts = [format_str("{name}={value}")]
    if path:
        parts.append(format_str("Path={path}"))
    if domain:
        parts.append(format_str("Domain={domain}"))
    if max_age is not None:
        parts.append(format_str("Max-Age={max_age}"))
    if secure:
        parts.append("Secure")
    if http_only:
        parts.append("HttpOnly")
    if same_site:
        parts.append(format_str("SameSite={same_site}"))
    return "; ".join(parts)

# ==============================================================================
#  Response
# ==============================================================================

class Response:
    """
    A complete HTTP response.

    Create via factory methods or directly:
        Response("Hello", 200, {"X-Custom": "val"})
        Response.json({"ok": True})
        Response.html("<h1>Hi</h1>")
        Response.redirect("/new-path")
        Response.error("Forbidden", 403)
    """

    def __init__(self, body="", status=200, headers=None):
        self.body    = body
        self.status  = status
        self.headers = headers if headers is not None else {}

    # ---- factory methods ------------------------------------------------

    @staticmethod
    def json(data, status=200, headers=None):
        h = {"Content-Type": "application/json; charset=utf-8"}
        if headers:
            for k in headers:
                h[k] = headers[k]
        return Response(cjson.dumps(data), status, h)

    @staticmethod
    def html(content, status=200, headers=None):
        h = {"Content-Type": "text/html; charset=utf-8"}
        if headers:
            for k in headers:
                h[k] = headers[k]
        return Response(content, status, h)

    @staticmethod
    def text(content, status=200, headers=None):
        h = {"Content-Type": "text/plain; charset=utf-8"}
        if headers:
            for k in headers:
                h[k] = headers[k]
        return Response(content, status, h)

    @staticmethod
    def redirect(location, status=302):
        """HTTP redirect (301 permanent, 302 temporary, etc.)."""
        return Response("", status, {"Location": location})

    @staticmethod
    def empty(status=204):
        """No-Content or similar empty response."""
        return Response("", status, {})

    @staticmethod
    def error(message="Internal Server Error", status=500):
        return Response(message, status, {"Content-Type": "text/plain; charset=utf-8"})

    @staticmethod
    def bytes(data, content_type="application/octet-stream", status=200):
        return Response(data, status, {"Content-Type": content_type})

    # ---- cookie helper --------------------------------------------------

    def set_cookie(self, name, value, max_age=None, path="/", domain=None, http_only=True, secure=False, same_site="Lax"):
        """Attach a Set-Cookie header. Returns self for chaining."""
        self.headers["Set-Cookie"] = _build_set_cookie(name, value, max_age=max_age, path=path, domain=domain, http_only=http_only, secure=secure, same_site=same_site)
        return self

    def delete_cookie(self, name, path="/"):
        """Expire a cookie immediately."""
        self.headers["Set-Cookie"] = _build_set_cookie(name, "", max_age=0, path=path, http_only=True)
        return self

    # ---- internal -------------------------------------------------------

    def _headers_str(self):
        parts = []
        for k in self.headers:
            parts.append(format_str("{k}: {self.headers[k]}\r\n"))
        return "".join(parts)

# ==============================================================================
#  URL Route Pattern Matching
# ==============================================================================

class _RoutePattern:
    """
    Match URL paths and extract named parameters.

    Syntax:
      /users/:id                → params['id']
      /posts/:pid/comments/:cid → params['pid'], params['cid']
      /files/*                  → params['*'] = remaining path
      /exact/path               → exact match only

    Method filtering: pass methods=['GET', 'POST'] to restrict.
    An empty methods list matches any method.
    """

    def __init__(self, pattern, methods=None):
        self.raw_pattern = pattern
        self.methods     = []
        if methods:
            for m in methods:
                self.methods.append(m.upper())
        # Normalise: strip trailing slash except for root
        p = pattern
        if p != "/" and p.endswith("/"):
            p = p[:-1]
        self._parts       = p.split("/")
        self._has_wildcard = "*" in self._parts

    def match(self, url, method="GET"):
        """
        Returns (True, params_dict) on match, (False, {}) otherwise.
        url may include a query string — it is stripped before matching.
        """
        if self.methods and method.upper() not in self.methods:
            return (False, {})

        # Strip query string
        path = url.split("?")[0]
        # Normalise trailing slash
        if path != "/" and path.endswith("/"):
            path = path[:-1]

        url_parts = path.split("/")
        pat_parts = self._parts
        params    = {}

        if self._has_wildcard:
            wc_idx = pat_parts.index("*")
            if len(url_parts) < wc_idx:
                return (False, {})
            for i in range(wc_idx):
                pp = pat_parts[i]
                up = url_parts[i] if i < len(url_parts) else ""
                if pp.startswith(":"):
                    params[pp[1:]] = up
                elif pp != up:
                    return (False, {})
            # Collect wildcard tail
            params["*"] = "/".join(url_parts[wc_idx:])
            return (True, params)

        if len(pat_parts) != len(url_parts):
            return (False, {})

        for i in range(len(pat_parts)):
            pp = pat_parts[i]
            up = url_parts[i]
            if pp.startswith(":"):
                params[pp[1:]] = up
            elif pp != up:
                return (False, {})

        return (True, params)

# ==============================================================================
#  Request
# ==============================================================================

class Request:
    """
    A parsed HTTP request.

    Attributes:
      method   — 'GET', 'POST', etc.
      path     — URI without query string, e.g. '/users/42'
      url      — full URI including query, e.g. '/users/42?verbose=1'
      query    — raw query string, e.g. 'verbose=1'
      args     — parsed query dict (lazy), e.g. {'verbose': '1'}
      headers  — lowercase-keyed dict of HTTP headers
      cookies  — parsed Cookie header (lazy dict)
      params   — URL parameter dict populated by the router, e.g. {'id': '42'}
      text     — request body as str
      content  — request body as bytes
    """

    def __init__(self, msg, params=None):
        self.method  = msg.method
        self.path    = msg.uri.split("?")[0]
        self.url     = msg.uri
        self.query   = msg.query
        self.headers = msg.get_headers()
        self.text    = msg.body
        self.content = msg.body_bytes
        self.params  = params if params is not None else {}
        self._msg    = msg
        self._args   = None
        self._cookies = None

    # ---- body helpers ---------------------------------------------------

    def json(self):
        """Deserialise the request body as JSON. Returns None if empty."""
        if not self.text:
            return None
        return cjson.loads(self.text)

    def form(self):
        """
        Parse application/x-www-form-urlencoded body into a dict.
        Returns {} if the body is empty or wrong content-type.
        """
        if not self.text:
            return {}
        return parse_qs(self.text)

    def get_var(self, name):
        """
        Shorthand: look up `name` in args first, then form body.
        Returns '' if not found.
        """
        val = self.args.get(name, "")
        if not val:
            val = self.form().get(name, "")
        return val

    # ---- lazy properties ------------------------------------------------

    @property
    def args(self):
        """Parsed query string dict (computed once per request)."""
        if self._args is None:
            self._args = parse_qs(self.query)
        return self._args

    @property
    def cookies(self):
        """Parsed Cookie header dict (computed once per request)."""
        if self._cookies is None:
            self._cookies = parse_cookies(self.headers.get("cookie", ""))
        return self._cookies

    # ---- content-type helpers -------------------------------------------

    def is_json(self):
        ct = self.headers.get("content-type", "")
        return "application/json" in ct

    def is_form(self):
        ct = self.headers.get("content-type", "")
        return "application/x-www-form-urlencoded" in ct

    def is_multipart(self):
        ct = self.headers.get("content-type", "")
        return "multipart/form-data" in ct

    def is_websocket_upgrade(self):
        return self.headers.get("upgrade", "").lower() == "websocket"

    def header(self, name, default=""):
        """Case-insensitive header lookup."""
        return self.headers.get(name.lower(), default)

# ==============================================================================
#  HttpError — raise from a route handler to send a structured error
# ==============================================================================

class HttpError(Exception):
    """
    Raise from a route handler to respond with a specific HTTP error code.

    Usage:
        raise HttpError(403, 'Forbidden')
        raise HttpError(404)
    """
    def __init__(self, status_code, message=""):
        self.status_code = status_code
        self.message = message if message else str(status_code)
        super().__init__(self.message)

# ==============================================================================
#  Internal: Route Entry and Static Mount
# ==============================================================================

class _Route:
    def __init__(self, pattern_str, handler, methods=None):
        self.pattern = _RoutePattern(pattern_str, methods)
        self.handler = handler

class _StaticMount:
    def __init__(self, url_prefix, dir_path, extra_headers="", mime_types="", page404=""):
        self.url_prefix    = url_prefix
        self.dir_path      = dir_path
        self.extra_headers = extra_headers
        self.mime_types    = mime_types
        self.page404       = page404

# ==============================================================================
#  HTTP Client
# ==============================================================================

class ClientResponse:
    def __init__(self, status_code, headers, body):
        self.status_code = status_code
        self.headers = headers
        self.body = body
        self.content = body.encode('utf-8') if isinstance(body, str) else body

    def json(self):
        if not self.body:
            return None
        return cjson.loads(self.body)

def request(method, url, headers=None, data=None, json=None, timeout=10.0):
    method = method.upper()
    mgr = mongoose.Manager()
    
    resp_data = {"status": 0, "headers": {}, "body": "", "done": False, "error": None}
    
    # Extract path and host
    host = ""
    path = "/"
    if "://" in url:
        parts = url.split("://")
        scheme = parts[0]
        rest = parts[1]
        if "/" in rest:
            slash_idx = rest.find("/")
            host = rest[:slash_idx]
            path = rest[slash_idx:]
        else:
            host = rest
    else:
        host = url

    def handler(conn, ev, ev_data):
        if ev == "CONNECT":
            req_lines = [format_str("{method} {path} HTTP/1.0"), format_str("Host: {host}")]
            body_str = ""
            if json is not None:
                body_str = cjson.dumps(json)
                req_lines.append("Content-Type: application/json")
                req_lines.append(format_str("Content-Length: {len(body_str)}"))
            elif data is not None:
                body_str = data if isinstance(data, str) else str(data)
                req_lines.append(format_str("Content-Length: {len(body_str)}"))
                
            if headers:
                for k in headers:
                    req_lines.append(format_str("{k}: {headers[k]}"))
                    
            req_lines.append("")
            req_lines.append(body_str)
            
            full_req = "\r\n".join(req_lines)
            mongoose.send(conn, full_req)
            
        elif ev == "RESPONSE":
            status_code = 200
            if ev_data.uri:
                try:
                    status_code = int(ev_data.uri)
                except Exception:
                    pass
            
            resp_data["status"] = status_code
            resp_data["headers"] = ev_data.get_headers()
            resp_data["body"] = ev_data.body
            resp_data["done"] = True
            
        elif ev == "ERROR":
            resp_data["error"] = ev_data
            resp_data["done"] = True
            
        elif ev == "CLOSE":
            resp_data["done"] = True

    try:
        mgr.http_connect(url, handler)
        import time
        start_time = time.time()
        while not resp_data["done"]:
            mgr.poll(50)
            if time.time() - start_time > timeout:
                resp_data["error"] = "Timeout"
                break
    finally:
        mgr.free()

    if resp_data["error"]:
        raise HttpError(500, format_str("Request failed: {resp_data['error']}"))
        
    return ClientResponse(resp_data["status"], resp_data["headers"], resp_data["body"])

def get(url, headers=None, timeout=10.0):
    return request("GET", url, headers=headers, timeout=timeout)

def post(url, headers=None, data=None, json=None, timeout=10.0):
    return request("POST", url, headers=headers, data=data, json=json, timeout=timeout)

def put(url, headers=None, data=None, json=None, timeout=10.0):
    return request("PUT", url, headers=headers, data=data, json=json, timeout=timeout)

def patch(url, headers=None, data=None, json=None, timeout=10.0):
    return request("PATCH", url, headers=headers, data=data, json=json, timeout=timeout)

def delete(url, headers=None, timeout=10.0):
    return request("DELETE", url, headers=headers, timeout=timeout)

# ==============================================================================
#  WebSocket Connection Registry
# ==============================================================================

class _WsRegistry:
    """
    Tracks live WebSocket connections grouped by route path.
    Used for broadcasting.
    """

    def __init__(self):
        self._by_path = {}     # path → { conn_key → conn_ptr }
        self._to_path = {}     # conn_key → path

    def add(self, path, conn_key, conn_ptr):
        if path not in self._by_path:
            self._by_path[path] = {}
        self._by_path[path][conn_key] = conn_ptr
        self._to_path[conn_key] = path

    def remove(self, conn_key):
        path = self._to_path.get(conn_key)
        if path:
            if path in self._by_path and conn_key in self._by_path[path]:
                del self._by_path[path][conn_key]
            del self._to_path[conn_key]

    def path_of(self, conn_key):
        return self._to_path.get(conn_key)

    def conns_on(self, path):
        """Return dict of conn_key → conn_ptr for a given WS path."""
        return self._by_path.get(path, {})

    def count(self, path):
        return len(self._by_path.get(path, {}))

# ==============================================================================
#  Route Group (prefix helper)
# ==============================================================================

class _RouteGroup:
    """
    Temporarily sets a URL prefix on the server so decorators use it.

    Usage (if Swalang supports 'with'):
        with app.group('/api/v1') as api:
            @api.get('/users')
            def users(req): ...

    Usage (without 'with'):
        api = app.group('/api/v1')
        api.__enter__()
        @app.get('/users')
        def users(req): ...
        api.__exit__(None, None, None)
    """

    def __init__(self, server, prefix):
        self._server     = server
        self._prefix     = prefix
        self._saved      = ""

    def __enter__(self):
        self._saved = self._server._prefix
        self._server._prefix = self._saved + self._prefix
        return self._server

    def __exit__(self, exc_type, exc_val, exc_tb):
        self._server._prefix = self._saved
        return False

# ==============================================================================
#  Response Coercion
# ==============================================================================

def _coerce(result):
    """
    Convert any value returned by a route handler into a Response.

    Supported return values:
      str                      → 200 text/html
      dict                     → 200 application/json
      Response                 → used as-is
      (body, int)              → body with given status code
      (body, dict)             → body with additional headers dict
      (body, int, dict)        → all three
      bytes                    → 200 application/octet-stream
      anything else            → str(result) as text/plain
    """
    if isinstance(result, Response):
        return result

    if isinstance(result, str):
        return Response.html(result)

    if isinstance(result, dict):
        return Response.json(result)

    if isinstance(result, bytes):
        return Response.bytes(result)

    if isinstance(result, tuple):
        n = len(result)
        if n == 2:
            body, second = result[0], result[1]
            if isinstance(second, int):
                if isinstance(body, dict):
                    return Response.json(body, status=second)
                return Response.html(str(body), status=second)
            if isinstance(second, dict):
                if isinstance(body, dict):
                    r = Response.json(body)
                else:
                    r = Response.html(str(body))
                for k in second:
                    r.headers[k] = second[k]
                return r
        if n == 3:
            body, status, extra_headers = result[0], result[1], result[2]
            if isinstance(body, dict):
                r = Response.json(body, status=status)
            else:
                r = Response.html(str(body), status=status)
            if isinstance(extra_headers, dict):
                for k in extra_headers:
                    r.headers[k] = extra_headers[k]
            return r

    return Response.text(str(result))

# ==============================================================================
#  Server
# ==============================================================================

class Server:
    """
    Production-ready HTTP + WebSocket server.
    See module docstring for usage examples.
    """

    def __init__(self, host="0.0.0.0", port=8000):
        self.host    = host
        self.port    = port
        self._mgr    = None
        self._prefix = ""          # current route prefix (for group())

        self._routes          = []  # list of _Route
        self._ws_routes       = {}  # path → handler
        self._static_mounts   = []  # list of _StaticMount
        self._before_hooks    = []  # list of callables(req) → Response or None
        self._after_hooks     = []  # list of callables(req, resp) → Response
        self._error_handlers  = {}  # int status → callable(req) → Response
        self._ws_registry     = _WsRegistry()

    # ------------------------------------------------------------------
    # Route decorators
    # ------------------------------------------------------------------

    def route(self, path, methods=None):
        """Register a route handler for one or more HTTP methods."""
        full = self._prefix + path
        def decorator(handler):
            self._routes.append(_Route(full, handler, methods))
            return handler
        return decorator

    def get(self, path):
        """Register a GET-only route."""
        return self.route(path, methods=["GET"])

    def post(self, path):
        """Register a POST-only route."""
        return self.route(path, methods=["POST"])

    def put(self, path):
        """Register a PUT-only route."""
        return self.route(path, methods=["PUT"])

    def patch(self, path):
        """Register a PATCH-only route."""
        return self.route(path, methods=["PATCH"])

    def delete(self, path):
        """Register a DELETE-only route."""
        return self.route(path, methods=["DELETE"])

    def ws_route(self, path):
        """
        Register a WebSocket route handler.

        handler(conn_ptr, event, data) where event is one of:
          'OPEN'  — data is the initial Request (for headers/cookies)
          'MSG'   — data is mongoose.WsMessage
          'CLOSE' — data is None
        """
        full = self._prefix + path
        def decorator(handler):
            self._ws_routes[full] = handler
            return handler
        return decorator

    def error_handler(self, status_code):
        """Register a custom handler for a given HTTP error status code."""
        def decorator(handler):
            self._error_handlers[status_code] = handler
            return handler
        return decorator

    # ------------------------------------------------------------------
    # Middleware / hooks
    # ------------------------------------------------------------------

    def before_request(self, func):
        """
        Register a function to run before every HTTP request.
        Return a Response to short-circuit (e.g. for auth checks).
        Return None to pass through to the route handler.

        Example:
            @app.before_request
            def require_auth(req):
                if not req.header('Authorization'):
                    raise HttpError(401, 'Unauthorized')
        """
        self._before_hooks.append(func)
        return func

    def after_request(self, func):
        """
        Register a function to run after every HTTP response is built,
        before it is sent.  Must return the (possibly modified) Response.

        Example:
            @app.after_request
            def add_cors(req, resp):
                resp.headers['Access-Control-Allow-Origin'] = '*'
                return resp
        """
        self._after_hooks.append(func)
        return func

    # ------------------------------------------------------------------
    # Static file serving
    # ------------------------------------------------------------------

    def static(self, url_prefix, dir_path, extra_headers="", mime_types="", page404=""):
        """
        Serve static files from `dir_path` for any request whose path
        starts with `url_prefix`.

        extra_headers — appended to every file response
        mime_types    — extra MIME overrides: "ext=type,ext2=type2"
        page404       — relative path inside dir_path for 404 responses
        """
        self._static_mounts.append(_StaticMount(url_prefix, dir_path, extra_headers=extra_headers, mime_types=mime_types, page404=page404))

    # ------------------------------------------------------------------
    # Route groups
    # ------------------------------------------------------------------

    def group(self, prefix):
        """
        Return a context manager that prefixes all routes registered
        inside its body.

        Usage (with 'with'):
            with app.group('/api/v1') as api:
                @api.get('/users')
                def users(req): ...

        Usage (without 'with'):
            g = app.group('/api/v1')
            g.__enter__()
            @app.get('/users')
            def users(req): ...
            g.__exit__(None, None, None)
        """
        return _RouteGroup(self, prefix)

    # ------------------------------------------------------------------
    # WebSocket utilities
    # ------------------------------------------------------------------

    def broadcast(self, path, data, op=mongoose.WEBSOCKET_OP_TEXT):
        """
        Send `data` to every active WebSocket client on `path`.
        Dead connections are silently skipped.
        """
        conns = self._ws_registry.conns_on(path)
        for conn_key in conns:
            conn_ptr = conns[conn_key]
            try:
                mongoose.ws_send(conn_ptr, data, op)
            except Exception as e:
                print(format_str("⚠️  [server] broadcast error on {conn_key}: {e}"))

    def ws_count(self, path):
        """Return the number of connected WebSocket clients on `path`."""
        return self._ws_registry.count(path)

    def ws_send(self, conn_ptr, data, op=mongoose.WEBSOCKET_OP_TEXT):
        """Send to a single WebSocket connection."""
        mongoose.ws_send(conn_ptr, data, op)

    # ------------------------------------------------------------------
    # Internal dispatch
    # ------------------------------------------------------------------

    def _send_response(self, conn_ptr, req, resp):
        """Apply after-hooks and transmit the response."""
        for hook in self._after_hooks:
            try:
                resp = hook(req, resp)
            except Exception as e:
                print(format_str("⚠️  [server] after_request hook error: {e}"))

        mongoose.http_reply(conn_ptr, resp.status, resp._headers_str(), resp.body)

    def _send_error(self, conn_ptr, req, status_code, message=""):
        """Send an error response, using a custom handler if registered."""
        if status_code in self._error_handlers:
            try:
                resp = self._error_handlers[status_code](req)
                self._send_response(conn_ptr, req, resp)
                return None
            except Exception as e:
                print(format_str("⚠️  [server] error handler ({status_code}) raised: {e}"))

        default_msg = message if message else str(status_code)
        mongoose.http_reply(conn_ptr, status_code, "Content-Type: text/plain; charset=utf-8\r\n", default_msg)

    def _dispatch_http(self, conn_ptr, raw_msg):
        req = Request(raw_msg)

        # before_request middleware
        for hook in self._before_hooks:
            try:
                early = hook(req)
                if early is not None:
                    resp = _coerce(early)
                    self._send_response(conn_ptr, req, resp)
                    return None
            except HttpError as e:
                self._send_error(conn_ptr, req, e.status_code, e.message)
                return None
            except Exception as e:
                print(format_str("⚠️  [server] before_request hook error: {e}"))
                self._send_error(conn_ptr, req, 500, format_str("Middleware error: {e}"))
                return None

        # Static file mounts (checked before dynamic routes)
        for mount in self._static_mounts:
            if req.path.startswith(mount.url_prefix):
                mongoose.http_serve_dir(conn_ptr, raw_msg, mount.dir_path, extra_headers=mount.extra_headers, mime_types=mount.mime_types, page404=mount.page404)
                return None

        # Dynamic routes (first match wins)
        for route in self._routes:
            matched, params = route.pattern.match(req.url, req.method)
            if matched:
                req.params = params
                try:
                    result = route.handler(req)
                    resp   = _coerce(result)
                    self._send_response(conn_ptr, req, resp)
                except HttpError as e:
                    self._send_error(conn_ptr, req, e.status_code, e.message)
                except Exception as e:
                    print(format_str("🔥 [server] route handler error: {e}"))
                    self._send_error(conn_ptr, req, 500, format_str("Internal Server Error: {e}"))
                return None

        # 404
        self._send_error(conn_ptr, req, 404, format_str("Not Found: {req.path}"))

    def _dispatch_ws_upgrade(self, conn_ptr, raw_msg):
        req = Request(raw_msg)
        for ws_path in self._ws_routes:
            matched, params = _RoutePattern(ws_path).match(req.url)
            if matched:
                req.params = params
                handler   = self._ws_routes[ws_path]
                conn_key  = str(conn_ptr)
                mongoose.ws_upgrade(conn_ptr, raw_msg, "")
                self._ws_registry.add(ws_path, conn_key, conn_ptr)
                try:
                    handler(conn_ptr, "OPEN", req)
                except Exception as e:
                    print(format_str("🔥 [server] WS OPEN handler error: {e}"))
                return None
        # No WS route matched
        mongoose.http_reply(conn_ptr, 404, "", "WebSocket route not found")

    def _internal_handler(self, conn_ptr, ev_type, msg):
        conn_key = str(conn_ptr)

        if ev_type == "HTTP_REQUEST":
            if Request(msg).is_websocket_upgrade():
                self._dispatch_ws_upgrade(conn_ptr, msg)
            else:
                self._dispatch_http(conn_ptr, msg)

        elif ev_type == "WS_MSG":
            path = self._ws_registry.path_of(conn_key)
            if path and path in self._ws_routes:
                try:
                    self._ws_routes[path](conn_ptr, "MSG", msg)
                except Exception as e:
                    print(format_str("🔥 [server] WS MSG handler error: {e}"))

        elif ev_type == "CLOSE":
            path = self._ws_registry.path_of(conn_key)
            if path and path in self._ws_routes:
                try:
                    self._ws_routes[path](conn_ptr, "CLOSE", None)
                except Exception as e:
                    print(format_str("🔥 [server] WS CLOSE handler error: {e}"))
                self._ws_registry.remove(conn_key)

        elif ev_type == "ERROR":
            print(format_str("⚠️  [server] connection error: {msg}"))

    # ------------------------------------------------------------------
    # Run
    # ------------------------------------------------------------------

    def run(self, tls_cert=None, tls_key=None, tls_ca=None, poll_ms=50):
        """
        Start the server event loop (blocking until Ctrl-C or exception).

        Plain HTTP:
            app.run()

        HTTPS (requires Mongoose compiled with TLS):
            app.run(tls_cert='server.crt', tls_key='server.key')
            app.run(tls_cert='path/or/pem', tls_key='...', tls_ca='ca.pem')

        poll_ms — how long each poll() call waits.  50 ms is a good default
                  (balances CPU usage vs. response latency).
        """
        self._mgr = mongoose.Manager()
        scheme    = "https" if (tls_cert and tls_key) else "http"
        url       = format_str("{scheme}://{self.host}:{self.port}")

        self._mgr.http_listen(url, self._internal_handler, tls_cert=tls_cert, tls_key=tls_key, tls_ca=tls_ca)

        print(format_str("🌐 Server running on {url}  (Ctrl-C to stop)"))
        try:
            while True:
                self._mgr.poll(poll_ms)
        except Exception as e:
            print(format_str("Server stopped: {e}"))
        finally:
            self._mgr.free()
            self._mgr = None