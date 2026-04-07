# pylearn/examples/test_stdlib.py
# from concurrent import futures
import http as http_module
import  http.get

import concurrent.futures
from concurrent.futures import ThreadPoolExecutor, TestThreadPoolExecutor
import string_utils # Should be found in the new `lib` directory
from regex import testRE

testRE()

s1 = "level"
s2 = "pylearnLKJHGFDSA"

test = TestThreadPoolExecutor()
print(test.getName())


print(format_str("Reversing '{s2}': {string_utils.reverse(s2)}"))
print(format_str("Is '{s2}' a palindrome? {string_utils.is_palindrome(s2)}"))
print(format_str("Is '{s1}' a palindrome? {string_utils.is_palindrome(s1)}"))

res = http_module.get("https://jsonplaceholder.typicode.com/todos/1")
print(res.json())