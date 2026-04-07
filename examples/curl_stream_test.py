# examples/curl_stream_test.py
import curl as http

# A URL to a large file (e.g., a test file or a Linux ISO)
# Using a smaller file for this example
URL = "https://httpbin.org/stream-bytes/102400" # 100KB stream

print(format_str("Streaming from {URL}..."))

try:
    # Use the `stream=True` parameter
    response = http.get(URL, stream=True)
    
    print(format_str("Status Code: {response.status_code}"))
    print(format_str("Headers: {response.headers}"))

    total_bytes = 0
    chunk_count = 0
    
    # The 'content' attribute is not available until we iterate
    # This would raise an error if accessed before iteration is complete.
    # print(response.content) 

    # Iterate over the response content in chunks
    # This loop drives the actual download
    for chunk in response.iter_content(chunk_size=8192):
        if chunk:
            chunk_count = chunk_count + 1
            total_bytes = total_bytes + len(chunk)
            print(format_str("  Received chunk {chunk_count} of size {len(chunk)}. Total so far: {total_bytes} bytes."))

    print(format_str("\nDownload complete. Total size: {total_bytes} bytes."))

    # Now that the stream is exhausted, `response.content` would work
    # but it's empty because we consumed it. The handle is also already closed.

except http.CurlError as e:
    print(format_str("An error occurred: {e}"))
finally:
    # It's good practice to ensure the connection is closed,
    # though the iter_content generator does this automatically when exhausted.
    # if 'response' in locals() and response is not None:
    if response is not None:
        response.close()


# import ffi
# ptr = ffi.malloc(8)
# print(ptr.Address)  # prints memory address as integer>>>>>855268656