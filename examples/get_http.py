# examples/get_http.py

print("Test http.get()")
import http
# response = http.get("https://httpbin.org/get")
response = http.get("https://jsonplaceholder.typicode.com/posts/1")
response # Evaluate the response object
# for key in response:
#     print(key)
print(response.text)
print(type(response))
print(type(response.text))

# response = http.get("https://httpbin.org/get")
response = http.post("https://posttestserver.dev/p/3aw71rmipuqva227/post",'{"foo":"bar"}')
response # Evaluate the response object
# for key in response:
#     print(key)
print(response.text)
print(response.json)
print(type(response))
print(type(response.text))
print(type(response.json))
