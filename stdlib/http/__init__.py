import mongoose
import cjson

class Response:
    def __init__(self):
        self.status_code = None
        self.headers = {}
        self.content = b""
        self.text = ""
        self.url = ""
        self.error = None

    def json(self):
        if not self.text:
            return None
        return cjson.loads(self.text)

class RequestError(Exception):
    pass

def _do_request(method, url, headers=None, data=None, json=None, timeout=10000):
    mgr = mongoose.Manager()
    response = Response()
    response.url = url
    
    # State tracking: [is_done]
    state = [False] 
    
    req_headers = ""
    if headers:
        for pair in headers.items():
            k = pair[0]
            v = pair[1]
            req_headers = req_headers + format_str("{k}: {v}\r\n")
            
    body = b""
    if json is not None:
        body = cjson.dumps(json).encode('utf-8')
        if "Content-Type" not in req_headers and "content-type" not in req_headers:
            req_headers = req_headers + "Content-Type: application/json\r\n"
    elif data is not None:
        if isinstance(data, str):
            body = data.encode('utf-8')
        elif isinstance(data, bytes):
            body = data
        else:
            raise TypeError("data must be string or bytes")

    if body:
        req_headers = req_headers + format_str("Content-Length: {len(body)}\r\n")
        
    parts = url.split("://")
    host = ""
    if len(parts) > 1:
        host_parts = parts[1].split("/")
        host = host_parts[0]
    
    if "Host" not in req_headers and "host" not in req_headers and host:
        req_headers = req_headers + format_str("Host: {host}\r\n")

    req_str = format_str("{method} {url} HTTP/1.1\r\n{req_headers}\r\n").encode('utf-8') + body

    def handler(conn, ev_type, ev_data):
        if ev_type == "CONNECT":
            mongoose.send(conn, req_str)
        elif ev_type == "RESPONSE":
            resp_msg = ev_data
            
            head_str = resp_msg.head
            parts = head_str.split(" ")
            if len(parts) >= 2:
                try:
                    response.status_code = int(parts[1])
                except Exception:
                    pass
            
            response.headers = resp_msg.get_headers()
            response.text = resp_msg.body
            response.content = resp_msg.body_bytes
            
            state[0] = True # Mark done
        elif ev_type == "ERROR":
            response.error = ev_data
            state[0] = True # Mark done
        elif ev_type == "CLOSE":
            # If closed before response/error, it usually means TLS issue or server reject
            if not state[0]:
                response.error = "Connection closed prematurely (If HTTPS, ensure TLS support is enabled in Mongoose)"
                state[0] = True

    mgr.http_connect(url, handler)
    
    elapsed = 0
    while not state[0] and elapsed < timeout:
        mgr.poll(10)
        elapsed = elapsed + 10
        
    mgr.free()
    
    if response.error:
        raise RequestError(response.error)
    if not state[0]:
        raise RequestError("Request timed out")
        
    return response

def get(url, headers=None, timeout=10000):
    return _do_request("GET", url, headers=headers, timeout=timeout)

def post(url, data=None, json=None, headers=None, timeout=10000):
    return _do_request("POST", url, headers=headers, data=data, json=json, timeout=timeout)

def put(url, data=None, json=None, headers=None, timeout=10000):
    return _do_request("PUT", url, headers=headers, data=data, json=json, timeout=timeout)

def delete(url, headers=None, timeout=10000):
    return _do_request("DELETE", url, headers=headers, timeout=timeout)


class Request:
    def __init__(self, msg):
        self.method = msg.method
        self.url = msg.uri
        self.query = msg.query
        self.headers = msg.get_headers()
        self.text = msg.body
        self.content = msg.body_bytes

    def json(self):
        if not self.text:
            return None
        return cjson.loads(self.text)

class Server:
    def __init__(self, host="0.0.0.0", port=8000):
        self.host = host
        self.port = port
        self.mgr = mongoose.Manager()
        self.routes = {}
        self.ws_routes = {}
        self.active_ws = {}

    def route(self, path):
        def decorator(handler):
            self.routes[path] = handler
            return handler
        return decorator

    def ws_route(self, path):
        def decorator(handler):
            self.ws_routes[path] = handler
            return handler
        return decorator

    def _internal_handler(self, conn, ev_type, msg):
        addr_key = str(conn)
        
        if ev_type == "HTTP_REQUEST":
            req = Request(msg)
            
            headers = req.headers
            if "upgrade" in headers and headers["upgrade"].lower() == "websocket":
                handler = self.ws_routes.get(req.url)
                if handler:
                    mongoose.ws_upgrade(conn, msg, "")
                    self.active_ws[addr_key] = handler
                    handler(conn, "OPEN", None)
                    return None
                else:
                    mongoose.http_reply(conn, 404, "", "WS Not Found")
                    return None

            handler = self.routes.get(req.url)
            if handler:
                try:
                    response = handler(req)
                    if isinstance(response, str):
                        mongoose.http_reply(conn, 200, "Content-Type: text/html\r\n", response)
                    elif isinstance(response, dict):
                        body = cjson.dumps(response)
                        mongoose.http_reply(conn, 200, "Content-Type: application/json\r\n", body)
                    elif isinstance(response, tuple) and len(response) == 2:
                        body = response[0]
                        status = response[1]
                        if isinstance(body, dict):
                            body_str = cjson.dumps(body)
                            mongoose.http_reply(conn, status, "Content-Type: application/json\r\n", body_str)
                        else:
                            mongoose.http_reply(conn, status, "Content-Type: text/html\r\n", str(body))
                    else:
                        mongoose.http_reply(conn, 500, "", "Invalid response type")
                except Exception as e:
                    mongoose.http_reply(conn, 500, "", format_str("Internal Server Error: {e}"))
            else:
                mongoose.http_reply(conn, 404, "", "Not Found")
        
        elif ev_type == "WS_MSG":
            handler = self.active_ws.get(addr_key)
            if handler:
                handler(conn, "MSG", msg)
        
        elif ev_type == "CLOSE":
            handler = self.active_ws.get(addr_key)
            if handler:
                handler(conn, "CLOSE", None)

    def run(self):
        url = format_str("http://{self.host}:{self.port}")
        self.mgr.http_listen(url, self._internal_handler)
        
        print(format_str("Server running on {url}... Press Ctrl+C to stop."))
        try:
            while True:
                self.mgr.poll(1000)
        except Exception as e:
            print(format_str("Server stopped: {e}"))
        finally:
            self.mgr.free()