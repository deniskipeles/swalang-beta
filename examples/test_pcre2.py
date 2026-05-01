import pcre2

print("===========================================")
print("🧩 Testing PCRE2 Regex Wrapper")
print("===========================================")

text = "The quick brown fox jumps over 123 lazy dogs at 4:56 PM."

# 1. Search (find anywhere)
match = pcre2.search("brown (\\w+)", text)
print(format_str("👉 search('brown (\\w+)'): {match}"))
assert match.group(0) == "brown fox"
assert match.group(1) == "fox"
assert match.span() == (10, 19)

# 2. Match (must match at start)
m1 = pcre2.match("The quick", text)
print(format_str("👉 match('The quick'): {m1}"))
assert m1 is not None

m2 = pcre2.match("brown", text)
print(format_str("👉 match('brown'): {m2}"))
assert m2 is None

# 3. Findall
numbers = pcre2.findall("(\\d+)", text)
print(format_str("👉 findall('\\d+'): {numbers}"))
assert numbers == ["123", "4", "56"]

# 4. Substitution (sub)
redacted = pcre2.sub("\\d+", "[REDACTED]", text)
print(format_str("👉 sub('\\d+', '[REDACTED]'):\n   {redacted}"))
assert "123" not in redacted
assert "[REDACTED]" in redacted

# 5. Split
sentence = "apple,  banana; cherry | dates"
fruits = pcre2.split("[,;|]\\s*", sentence)
print(format_str("👉 split('[,;|]\\s*'): {fruits}"))
assert len(fruits) == 4
assert fruits[1] == "banana"

# 6. Ignore Case Flag
text_lower = "hello WORLD"
match_ignore = pcre2.search("world", text_lower, pcre2.I)
print(format_str("👉 search('world', flags=I): {match_ignore}"))
assert match_ignore is not None

print("\n🎉 PCRE2 Regex Wrapper tests passed successfully!")
