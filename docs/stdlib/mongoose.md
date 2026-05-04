# `mongoose` Module Reference

The `mongoose` module provides high-level bindings to the Mongoose embedded networking library.

## Global Functions

- `http_reply(conn_ptr, status_code, headers, body)`: Sends an HTTP response through the given connection.

## Constants

- `MG_EV_OPEN`, `MG_EV_POLL`, `MG_EV_RESOLVE`, `MG_EV_CONNECT`, `MG_EV_ACCEPT`, `MG_EV_READ`, `MG_EV_WRITE`, `MG_EV_CLOSE`, `MG_EV_ERROR`, `MG_EV_HTTP_MSG`, `MG_EV_HTTP_CHUNK`

## Classes

### `Manager()`
The main event manager.
- `poll(ms)`: Checks for network events for the specified timeout.
- `http_listen(url, handler_func)`: Starts an HTTP server. `handler_func` is called with `(conn_ptr, HttpMessage)`.
- `http_connect(url, handler_func)`: Connects as an HTTP client. `handler_func` is called with `(conn_ptr, HttpMessage)`.
- `free()`: Frees the manager and associated callbacks.

### `HttpMessage(ptr)`
Wraps a C `mg_http_message` struct.
- `method`: (Property) HTTP method (e.g., "GET").
- `uri`: (Property) Target URI.
- `query`: (Property) Query string.
- `body`: (Property) Request/response body.
