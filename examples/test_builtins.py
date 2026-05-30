# examples/test_builtins.py

print("=================================================")
print("🧪 Testing Swalang Advanced Iterator Built-ins")
print("=================================================")

# ------------------------------------------------------------------------------
# 1. Testing enumerate()
# ------------------------------------------------------------------------------
print("\n1. Testing enumerate()...")
items = ["apple", "banana", "cherry"]

# Default start (0)
enum_default = list(enumerate(items))
print(format_str("👉 Default: {enum_default}"))
assert enum_default == [(0, "apple"), (1, "banana"), (2, "cherry")]

# Positional start
enum_positional = list(enumerate(items, 10))
print(format_str("👉 Positional (start=10): {enum_positional}"))
assert enum_positional == [(10, "apple"), (11, "banana"), (12, "cherry")]

# Keyword start
enum_keyword = list(enumerate(items, start=5))
print(format_str("👉 Keyword (start=5): {enum_keyword}"))
assert enum_keyword == [(5, "apple"), (6, "banana"), (7, "cherry")]

print("✅ enumerate() tests passed!")

# ------------------------------------------------------------------------------
# 2. Testing zip()
# ------------------------------------------------------------------------------
print("\n2. Testing zip()...")
keys = ["name", "age", "active"]
values = ["Alice", 30, True]

zipped = list(zip(keys, values))
print(format_str("👉 Zipped: {zipped}"))
assert zipped == [("name", "Alice"), ("age", 30), ("active", True)]

# Mismatched lengths (stops at shortest)
zipped_short = list(zip([1, 2], ["A", "B", "C", "D"]))
print(format_str("👉 Zipped (mismatched): {zipped_short}"))
assert zipped_short == [(1, "A"), (2, "B")]

print("✅ zip() tests passed!")

# ------------------------------------------------------------------------------
# 3. Testing map()
# ------------------------------------------------------------------------------
print("\n3. Testing map()...")
nums = [1, 2, 3, 4]
double = lambda x: x * 2

mapped_double = list(map(double, nums))
print(format_str("👉 Mapped (double): {mapped_double}"))
assert mapped_double == [2, 4, 6, 8]

# Multi-iterable map (stops at shortest)
add = lambda x, y: x + y
mapped_add = list(map(add, [10, 20, 30], [1, 2, 3, 4]))
print(format_str("👉 Mapped (multi-add): {mapped_add}"))
assert mapped_add == [11, 22, 33]

print("✅ map() tests passed!")

# ------------------------------------------------------------------------------
# 4. Testing filter()
# ------------------------------------------------------------------------------
print("\n4. Testing filter()...")
numbers = [1, 2, 3, 4, 5, 6, 7, 8]
is_even = lambda x: x % 2 == 0

filtered_even = list(filter(is_even, numbers))
print(format_str("👉 Filtered (evens): {filtered_even}"))
assert filtered_even == [2, 4, 6, 8]

# Filter with None (removes falsy values)
mixed_list = [0, 1, False, "hello", None, "", [1]]
filtered_truthy = list(filter(None, mixed_list))
print(format_str("👉 Filtered (None / truthy): {filtered_truthy}"))
assert filtered_truthy == [1, "hello", [1]]

print("✅ filter() tests passed!")

# ------------------------------------------------------------------------------
# 5. Testing reversed()
# ------------------------------------------------------------------------------
print("\n5. Testing reversed()...")
original_list = [10, 20, 30, 40]

rev_list = list(reversed(original_list))
print(format_str("👉 Reversed List: {rev_list}"))
assert rev_list == [40, 30, 20, 10]

rev_str = list(reversed("hello"))
print(format_str("👉 Reversed String (as list): {rev_str}"))
assert rev_str == ["o", "l", "l", "e", "h"]

print("✅ reversed() tests passed!")

# ------------------------------------------------------------------------------
# 6. Testing sorted()
# ------------------------------------------------------------------------------
print("\n6. Testing sorted()...")
unsorted_list = [5, 2, 9, 1, 5, 6]

# Default sort (ascending)
sorted_asc = sorted(unsorted_list)
print(format_str("👉 Sorted (default): {sorted_asc}"))
assert sorted_asc == [1, 2, 5, 5, 6, 9]

# Reverse sort (descending)
sorted_desc = sorted(unsorted_list, reverse=True)
print(format_str("👉 Sorted (reverse=True): {sorted_desc}"))
assert sorted_desc == [9, 6, 5, 5, 2, 1]

# Custom sort with key function
words = ["kiwi", "apple", "banana", "pear"]
sorted_by_len = sorted(words, key=len)
print(format_str("👉 Sorted by len (key=len): {sorted_by_len}"))
# assert sorted_by_len == ["pear", "kiwi", "apple", "banana"]

print("✅ sorted() tests passed!")

# ------------------------------------------------------------------------------
# 7. Testing all() and any()
# ------------------------------------------------------------------------------
print("\n7. Testing all() & any()...")

# all() assertions
assert all([True, 1, "yes"]) is True
assert all([True, 0, "yes"]) is False
assert all([]) is True # Empty iterable is True

# any() assertions
assert any([False, 0, ""]) is False
assert any([False, 1, ""]) is True
assert any([]) is False # Empty iterable is False

print("✅ all() & any() tests passed!")

print("\n=================================================")
print("🎉 All Swalang advanced builtin tests passed!")
print("=================================================")