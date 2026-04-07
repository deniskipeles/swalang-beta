# pylearn/examples/test_hv.py

import hv
import time

# Global flag to signal that our async callback has been called
http_callback_fired = False
response_status = 0

def on_http_response(response):
    """This is our user-defined callback for the HTTP request."""
    global http_callback_fired, response_status
    
    print("\n--- HTTP Response Received (from callback) ---")
    print(format_str("Status Code: {response.status_code}"))
    body = response.text
    print(format_str("Body (first 100 chars): '{body[:100]}...'"))
    
    response_status = response.status_code
    http_callback_fired = True

# --- Main application logic ---
if not hv.HV_AVAILABLE:
    print("libhv is not available, cannot run the example.")
else:
    print("libhv async HTTP client example started.")
    
    # 1. Make an asynchronous HTTP GET request. This call returns immediately.
    url = "https://httpbin.org/get?source=pylearn_libhv"
    print(format_str("Requesting URL: {url}"))
    hv.get(url, on_http_response)

    # 2. Wait for the callback to run.
    #    Since the request runs on a background thread, our main script
    #    needs to wait for it to complete. In a real application, an
    #    event loop would handle this. For a simple test, a sleep loop is fine.
    print("Main script is now waiting for the async callback...")
    timeout = 5 # 5 seconds
    start_time = time.time()
    while not http_callback_fired:
        time.sleep(0.1)
        if time.time() - start_time > timeout:
            print("ERROR: Test timed out!")
            break

    print("\nMain script finished waiting.")

    # 3. Verify that our async callback was executed successfully
    assert http_callback_fired, "The HTTP response callback did not fire!"
    assert response_status == 200, "Expected status code 200"
    print("All tests successful.")