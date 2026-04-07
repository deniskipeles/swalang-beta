# examples/test_pycurl.py

# Import our new high-level wrapper
import pycurl

print("--- Testing our custom pycurl module ---")

# Perform a simple GET request
response = pycurl.get("https://httpbin.org/html")

# The response.content is a Pylearn 'bytes' object.
# We'll print its representation, which will show the raw bytes.
# In a full-featured Pylearn, you might have `response.content.decode('utf-8')`.
print("Response from httpbin.org/html:")
print(response.content[:100].decode('utf-8') + "...")

# You can assert that the content contains expected HTML tags
# (This requires the `in` operator to work on bytes objects)
assert b"<h1>Herman Melville - Moby-Dick</h1>" in response.content

print("\n--- pycurl test successful! ---")


# response = http.get("https://httpbin.org/get")
data = {
    "foo":"bar"
}
# response = pycurl.post("https://posttestserver.dev/p/3aw71rmipuqva228/post", data=data)
# # responsedata='{"foo":"bar"}')
# # response # Evaluate the response object
# # for key in response:
# #     print(key)
# try:
#     print(response.text())
#     print(type(response.text()))
#     print(type(response))
#     # print(response.json())
#     # print(type(response.json()))
# except Exception as e:
#     print("Error>>>",e)

result = pycurl.get("https://jsonplaceholder.typicode.com/todos/1")
print(result.content)
print(result.content.decode())
print(type(result.content.decode()))
try:
    print(type(result.json()))
except Exception as e:
    print("Error>>>",e)

try:
    print(result.json())
except Exception as e:
    print("Error>>>",e)

try:
    print(result.text())
except Exception as e:
    print("Error>>>",e)


try:
    x = float(1)
except Exception as e:
    print("FLOAT ERR>>",e)

try:
    x = int(1)
except Exception as e:
    print("INT ERR>>",e)