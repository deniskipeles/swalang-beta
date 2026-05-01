import yyjson

print("===========================================")
print("🚀 Testing yyjson Fast FFI Wrapper")
print("===========================================")

json_string = """
{
    "project": "Swalang",
    "version": 1.0,
    "is_awesome": True,
    "features": ["FFI", "JSON", "Async"],
    "stats": {
        "speed": 99.9,
        "bugs": None
    }
}
"""

print("1. Testing loads()...")
data = yyjson.loads(json_string)

print(format_str("👉 Parsed Data: {data}"))
assert data["project"] == "Swalang"
assert data["version"] == 1.0
assert data["is_awesome"] is True
assert data["features"][1] == "JSON"
assert data["stats"]["bugs"] is None
print("✅ loads() passed!")

print("\n2. Testing dumps()...")
# Modify the data and reserialize it
data["new_feature"] = "Embedded C Structs"
data["features"].append("Magic")

output = yyjson.dumps(data, indent=True)
print(format_str("👉 Serialized Output:\n{output}"))

# Verify it survived the roundtrip
verify = yyjson.loads(output)
assert verify["new_feature"] == "Embedded C Structs"
assert verify["features"][-1] == "Magic"
print("✅ dumps() passed!")

print("\n🎉 All yyjson tests passed successfully!")
