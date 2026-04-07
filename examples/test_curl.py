# examples/test_curl.py

import curl as http

# Initialize curl
# http.init()

# # Simple GET request
response = http.get("https://httpbin.org/get")
print(response.status_code)
print(response.text)