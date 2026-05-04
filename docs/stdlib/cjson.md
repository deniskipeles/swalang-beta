# `cjson` Module Reference

The `cjson` module provides high-performance JSON encoding and decoding using the cJSON library.

## Global Functions

- `loads(json_string)`: Parses a JSON string and returns a corresponding Swalang object (dict, list, string, number, bool, or None).
- `dumps(obj)`: Serializes a Swalang object into a JSON-formatted string.

## Constants

- `cJSON_Invalid`, `cJSON_False`, `cJSON_True`, `cJSON_NULL`, `cJSON_Number`, `cJSON_String`, `cJSON_Array`, `cJSON_Object`

## Exceptions

- `CJSONError`: Raised when JSON parsing or serialization fails.
